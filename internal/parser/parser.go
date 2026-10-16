package parser

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json/jsontext"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/gethiox/EDCITR/internal/jsonutil"
	"github.com/gethiox/EDCITR/internal/utils"
	"github.com/gethiox/EDCITR/internal/zaputils"
	"github.com/nxadm/tail"
	"github.com/orcaman/concurrent-map/v2"
	"go.uber.org/zap"
)

const (
	lineScannerBufferSize = 1024 * 1024 // 1MB should be enough to handle big journal events
	seekChunkSize         = 1024 * 4
)

var (
	maxConcurrentFiles      = 8
	processorWorkers        = 16
	processorWorkerChanSize = 64
	processorResultChanSize = 128
)

var Logger = zaputils.Logger.Named("Parser")

var journalDirectoriesWindows = []string{
	`Saved Games\Frontier Developments\Elite Dangerous`,
}

var journalDirectoriesLinux = []string{
	`.local/share/Steam/steamapps/compatdata/359320/pfx/drive_c/users/steamuser/Saved Games/Frontier Developments/Elite Dangerous/`,
}

var EliteJournalDirectory string

func init() {
	var candidates []string

	switch runtime.GOOS {
	case "windows":
		candidates = journalDirectoriesWindows
	case "linux":
		candidates = journalDirectoriesLinux
	default:
		Logger.Panic("Unsupported platform", zap.String("platform", runtime.GOOS))
	}

	home, err := os.UserHomeDir()
	if err != nil {
		Logger.Panic("Failed to get user home directory", zap.Error(err))
	}

	for _, candidate := range candidates {
		target := filepath.Join(home, candidate)
		if _, err := os.Stat(target); err == nil {
			EliteJournalDirectory = target
			return
		}
	}

	Logger.Panic("Journal directory for Elite not found", zap.String("candidates", strings.Join(candidates, ", ")))
}

type Settings struct {
	MaxConcurrentFiles      int
	ProcessorWorkers        int
	ProcessorWorkerChanSize int
	ProcessorResultChanSize int
}

func Configure(settings Settings) {
	maxConcurrentFiles = settings.MaxConcurrentFiles
	processorWorkers = settings.ProcessorWorkers
	processorWorkerChanSize = settings.ProcessorWorkerChanSize
	processorResultChanSize = settings.ProcessorResultChanSize
}

type EntryWithTime struct {
	FilePath string
	Name     string
	Time     time.Time
	Size     int64
}

func (e EntryWithTime) String() string {
	return fmt.Sprintf("[%s (size: %.2f MB)]", e.Name, float64(e.Size)/1024/1024)
}

func IsJournalFile(name string) bool {
	return strings.HasPrefix(name, "Journal.") && strings.HasSuffix(name, ".log") && len(name) == len("Journal.2025-12-05T221051.01.log") // todo: regex
}

func FileNameToEntry(path, name string, size int64) EntryWithTime {
	trimmed := name[8:25]
	t, err := time.ParseInLocation("2006-01-02T150405", trimmed, time.Local)
	if err != nil {
		panic(err)
	}
	return EntryWithTime{
		FilePath: path,
		Name:     name,
		Time:     t,
		Size:     size,
	}
}

func GetLastLineWithSeek(filepath string) string {
	fd, err := os.Open(filepath)
	if err != nil {
		return ""
	}
	defer fd.Close()

	stat, err := fd.Stat()
	if err != nil || stat.Size() == 0 {
		return ""
	}

	filesize := stat.Size()
	chunkSize := int64(seekChunkSize) // Read 4KB at once
	if chunkSize > filesize {
		chunkSize = filesize
	}

	buf := make([]byte, chunkSize)
	_, err = fd.ReadAt(buf, filesize-chunkSize)
	if err != nil && err != io.EOF {
		return ""
	}

	// Trim trailing newlines
	endIdx := len(buf) - 1
	for endIdx >= 0 && (buf[endIdx] == '\n' || buf[endIdx] == '\r') {
		endIdx--
	}

	startIdx := bytes.LastIndexByte(buf[:endIdx+1], '\n')
	if startIdx == -1 {
		return string(buf[:endIdx+1])
	}

	return string(buf[startIdx+1 : endIdx+1])
}

func IsCurrentLog(entry EntryWithTime) bool {
	line := GetLastLineWithSeek(entry.FilePath)
	ev, err := processLine(line)
	if err != nil {
		return true
	}

	switch ev.(type) {
	case *Shutdown:
		return false
	}

	return true
}

func listLogs(directory string) ([]EntryWithTime, error) {
	entries, err := os.ReadDir(directory)
	if err != nil {
		return nil, fmt.Errorf("failed to list journal logs: %w", err)
	}

	var interesting = make([]EntryWithTime, 0, len(entries))

	for _, entry := range entries {
		name := entry.Name()
		if !IsJournalFile(name) {
			continue
		}

		path := filepath.Join(directory, name)

		var size int64 = -1
		stat, err := os.Stat(path)
		if err == nil {
			size = stat.Size()
		}

		interesting = append(interesting, FileNameToEntry(path, entry.Name(), size))

	}

	sort.Slice(interesting, func(i, j int) bool {
		return interesting[i].Time.Before(interesting[j].Time)
	})

	return interesting, nil
}

func ListArchivedLogs(direcotry string) ([]EntryWithTime, error) {
	return listLogs(direcotry)
}

func ListCurrentLogs() ([]EntryWithTime, error) {
	return listLogs(EliteJournalDirectory)
}

func mergeLogs(current, archived []EntryWithTime) []EntryWithTime {
	var collected = make([]EntryWithTime, 0, max(len(current), len(archived)))
	var inArchive = make(map[string]EntryWithTime, max(len(current), len(archived)))
	var inCurrent = make(map[string]EntryWithTime, max(len(current), len(archived)))

	for _, entry := range archived {
		inArchive[entry.Name] = entry
	}

	for _, entry := range current {
		inCurrent[entry.Name] = entry
	}

	for name, archivedEntry := range inArchive {
		currentEntry, duplicate := inCurrent[name]

		if duplicate {
			if archivedEntry.Size != currentEntry.Size {
				Logger.Warn("duplicated journal size differs, picking bigger one", zap.Int64("archive_size", archivedEntry.Size), zap.Int64("current_size", currentEntry.Size), zap.String("archive", archivedEntry.String()), zap.String("current", currentEntry.String()))
				if archivedEntry.Size > currentEntry.Size {
					collected = append(collected, archivedEntry)
				} else {
					collected = append(collected, currentEntry)
				}
			} else {
				collected = append(collected, currentEntry) // doesn't matter which one
			}
		} else {
			collected = append(collected, archivedEntry)
		}
	}

	for name, currentEntry := range inCurrent {
		_, duplicate := inArchive[name]

		if duplicate { // duplicates already handled
			continue
		}

		collected = append(collected, currentEntry)
	}
	sort.Slice(collected, func(i, j int) bool {
		return collected[i].Time.Before(collected[j].Time)
	})

	return collected
}

func ListAllLogs(archiveDirectory string) ([]EntryWithTime, error) {
	archived, err1 := ListArchivedLogs(archiveDirectory)
	if err1 != nil {
		Logger.Warn("Failed to list archived logs", zap.Error(err1))
	}

	current, err2 := ListCurrentLogs()
	if err2 != nil {
		Logger.Warn("Failed to list current logs", zap.Error(err2))
	}

	switch {
	case err1 != nil && err2 != nil:
		return nil, fmt.Errorf("no logs available: %w, %w", err1, err2)
	case err1 == nil && err2 != nil:
		return archived, nil
	case err1 != nil && err2 == nil:
		return current, nil
	default:
		unique := mergeLogs(current, archived)
		Logger.Info("merging current and archived logs", zap.Int("current", len(current)), zap.Int("archived", len(archived)), zap.Int("unique", len(unique)), zap.Int("duplicates", len(archived)+len(current)-len(unique)))
		return unique, nil
	}
}

func watchForNewJournalLogs(ctx context.Context) <-chan EntryWithTime {
	Logger := Logger.Named("fsnotify")
	var newEntries = make(chan EntryWithTime)

	w, err := fsnotify.NewWatcher()
	if err != nil {
		panic(err)
	}

	err = w.Add(EliteJournalDirectory)
	if err != nil {
		panic(fmt.Errorf("%w: %s", err, EliteJournalDirectory))
	}

	go func() {
		defer close(newEntries)
		defer w.Close()

		Logger.Info("Monitoring for new journal files started", zap.String("directory", EliteJournalDirectory))
	ring0:
		for {
			select {
			case event, ok := <-w.Events:
				if !ok {
					break ring0
				}

				if !event.Has(fsnotify.Create) {
					continue
				}

				fileName := filepath.Base(event.Name)

				if !IsJournalFile(fileName) {
					continue
				}

				Logger.Info("New journal file!", zap.String("file", fileName))

				var size int64 = -1
				stat, err := os.Stat(event.Name)
				if err == nil {
					size = stat.Size()
				}

				newEntries <- FileNameToEntry(event.Name, fileName, size)
			case err, ok := <-w.Errors:
				Logger.Error("error", zap.Error(err), zap.Bool("ok", ok))
				break ring0
			case <-ctx.Done():
				Logger.Info("context expired")
				break ring0
			}
		}
		Logger.Info("Monitoring for new journal files finished", zap.String("directory", EliteJournalDirectory))
	}()

	return newEntries
}

var WarningNan = WarningSanitize{`sanitization: -nan(ind)`}

type WarningSanitize struct {
	extra string
}

func (e WarningSanitize) Error() string {
	return e.extra
}

func sanitizeLine(line string) (string, error) {
	line = strings.TrimSuffix(line, "\r")

	if len(line) == 0 {
		return line, nil
	}

	if strings.Contains(line, `-nan(ind)`) {
		line = strings.ReplaceAll(line, `-nan(ind)`, `null`) // There is nothing better than json marshaling in C++
		return line, WarningNan
	}

	return line, nil
}

func GetEventType(data []byte) (string, error) {
	dec := jsontext.NewDecoder(bytes.NewReader(data), jsontext.AllowDuplicateNames(true))
	if _, err := dec.ReadToken(); err != nil { // skip '{'
		return "", err
	}

	for dec.PeekKind() != '}' && dec.PeekKind() != 0 {
		keyTok, err := dec.ReadToken()
		if err != nil {
			return "", err
		}

		if keyTok.String() == "event" {
			valTok, err := dec.ReadToken()
			if err != nil {
				return "", err
			}
			return utils.UniqueString(valTok.String()), nil
		}

		if err := dec.SkipValue(); err != nil {
			return "", err
		}
	}
	return "", ErrorEventKeyNotFound
}

var encounteredNewEvents = cmap.New[bool]()

type ProcessLineError struct {
	Err       error
	EventName string
}

func (e ProcessLineError) Error() string {
	if e.EventName != "" {
		return fmt.Sprintf("%v, event: '%s'", e.Err.Error(), e.EventName)
	}
	return e.Err.Error()
}

func (e ProcessLineError) Unwrap() error {
	return e.Err
}

var (
	ErrorEmptyEventData     = errors.New("empty event data received")
	ErrorEmptyEventName     = errors.New("empty event name value found")
	ErrorEventKeyNotFound   = errors.New("event key not found")
	ErrorUnsupportedEvent   = errors.New("unsupported event")
	ErrorSanitizationFailed = errors.New("sanitization failed")
	ErrorSupportedEvent     = errors.New("supported event")
)

func processLine(line string) (JournalEvent, error) {
	var sanErr error
	line, sanErr = sanitizeLine(line)
	if line == "" {
		return nil, ProcessLineError{Err: ErrorEmptyEventData}
	}

	raw := []byte(line)
	eventName, err := GetEventType(raw)
	if err != nil {
		//Logger.Warn("GetEventType failed", zap.Error(err), zap.String("line", line))
		return nil, ProcessLineError{Err: err}
	}

	if sanErr != nil {
		if errors.Is(sanErr, WarningNan) {
			Logger.Debug("sanitized", zap.Error(sanErr), zap.String("event", eventName))
		} else {
			//Logger.Error("sanitization failed", zap.Error(sanErr), zap.String("event", eventName))
			return nil, ProcessLineError{Err: ErrorSanitizationFailed, EventName: eventName}
		}
	}

	if eventName == "" {
		return nil, ProcessLineError{Err: ErrorEmptyEventName, EventName: eventName}
	}

	eventType, ok := EventRegistry[eventName]
	if !ok {
		if encounteredNewEvents.SetIfAbsent(eventName, true) {
			Logger.Debug("unsupported journal event", zap.String("event", eventName))
		}
		return nil, ProcessLineError{Err: ErrorUnsupportedEvent, EventName: eventName}
	}

	targetPtr := reflect.New(eventType).Interface()
	err = jsonutil.UnmarshalStrictV3(raw, targetPtr)
	if err != nil {
		//Logger.Warn("unmarshal failed", zap.Error(err), zap.String("Event", eventName))
		return nil, ProcessLineError{Err: err, EventName: eventName}
	}

	return targetPtr.(JournalEvent), nil
}

type UnsupportedEvent struct {
	Event string
	Raw   string
}

func processLineUnsupported(line string) (UnsupportedEvent, error) {
	line, _ = sanitizeLine(line)
	if line == "" {
		return UnsupportedEvent{}, ProcessLineError{Err: ErrorEmptyEventData}
	}

	raw := []byte(line)
	eventName, err := GetEventType(raw)
	if err != nil {
		return UnsupportedEvent{}, ProcessLineError{Err: err}
	}

	if eventName == "" {
		return UnsupportedEvent{}, ProcessLineError{Err: ErrorEmptyEventName}
	}

	_, ok := EventRegistry[eventName]
	if ok {
		return UnsupportedEvent{}, ProcessLineError{Err: ErrorSupportedEvent, EventName: eventName}
	}

	return UnsupportedEvent{Event: eventName, Raw: line}, nil
}

type JournalEventBatch[T any] struct {
	Entry        EntryWithTime
	Events       chan T
	CountResults chan CountResult
}

// ReadLogsGeneric is the universal engine for reading any file with any processor.
func ReadLogsGeneric[T any](entries []EntryWithTime, ctx context.Context, processFn ProcessLineFunc[T], reportErrors bool) []JournalEventBatch[T] {
	var batches []JournalEventBatch[T]
	var queueSlots = make(chan struct{}, maxConcurrentFiles)

	for _, entry := range entries {
		batches = append(batches, JournalEventBatch[T]{
			Entry:        entry,
			Events:       make(chan T, processorResultChanSize),
			CountResults: make(chan CountResult, 1),
		})
	}

	go func() {
		var wg sync.WaitGroup

		for _, batch := range batches {
			if ctx.Err() != nil {
				close(batch.Events)
				close(batch.CountResults)
				continue
			}

			wg.Add(1)
			queueSlots <- struct{}{} // acquiring queue slot

			go func(e EntryWithTime, out chan<- T, counters chan<- CountResult) {
				defer wg.Done()
				defer func() { <-queueSlots }() // releasing queue slot

				lp := NewLineProcessor[T](processorWorkers, processorWorkerChanSize, processorResultChanSize, processFn)

				var subWg sync.WaitGroup
				subWg.Add(1)
				go func() {
					defer subWg.Done()
					defer close(out)

					var line int
					for res := range lp.Results() {
						line++
						if res.Error != nil {
							if !reportErrors {
								continue
							}

							pe := ProcessLineError{}

							switch {
							case errors.Is(res.Error, ErrorUnsupportedEvent):
								break
							case errors.As(res.Error, &pe) && pe.EventName != "":
								Logger.Warn("process line failed", zap.Error(pe.Err), zap.String("entry", e.String()), zap.Int("line", line), zap.String("event", pe.EventName))
							default:
								Logger.Warn("process line failed", zap.Error(res.Error), zap.String("entry", e.String()), zap.Int("line", line))
							}

							continue
						}
						out <- res.Result
					}
				}()

				file, err := os.Open(e.FilePath)
				if err != nil {
					Logger.Error("Failed to open journal file", zap.Error(err), zap.String("path", e.FilePath))
					lp.Close()
					close(counters)
					return
				}
				defer file.Close()

				scanner := bufio.NewScanner(file)

				buf := make([]byte, lineScannerBufferSize)
				scanner.Buffer(buf, lineScannerBufferSize)

				subWg.Add(1)
				go func() {
					defer subWg.Done()

					var lineCounter int
					defer func() {
						count := lp.Close()
						count.Bytes += uint64(lineCounter * 2) // include "\r\n" occurrences in processed data amount
						counters <- count
						close(counters)
					}()

					for scanner.Scan() {
						if ctx.Err() != nil {
							return
						}
						lp.Push(scanner.Text()[:])
						lineCounter++
					}
				}()

				subWg.Wait()

				// TODO: maybe handle `bufio.ErrTooLong` and retry with a increased buffer size
				//       instead of "fixing" the problem with a bigger buffer from the beginning
				err = scanner.Err()
				if err != nil {
					Logger.Error("Scanner failed during file read", zap.Error(err), zap.String("path", e.FilePath))
				}

			}(batch.Entry, batch.Events, batch.CountResults)
		}

		wg.Wait()
	}()

	return batches
}

func ReadLogs(entries []EntryWithTime, ctx context.Context) []JournalEventBatch[JournalEvent] {
	return ReadLogsGeneric(entries, ctx, processLine, true)
}

func ReadLogsUnsupported(entries []EntryWithTime, ctx context.Context) []JournalEventBatch[UnsupportedEvent] {
	return ReadLogsGeneric(entries, ctx, processLineUnsupported, false)
}

func ReadLogAndFollowGeneric[T any](entry EntryWithTime, ctx context.Context, processFn ProcessLineFunc[T]) (<-chan T, <-chan struct{}) {
	Logger := Logger.Named("ReadLogAndFollow")

	eventChan := make(chan T, processorResultChanSize)
	caughtUpChan := make(chan struct{})

	go func() {
		defer close(eventChan)
		defer func() {
			select {
			case <-caughtUpChan:
			default:
				close(caughtUpChan)
			}
		}()

		Logger.Info("reading file", zap.String("entry", entry.String()))

		var initialEndOffset int64
		fileInfo, err := os.Stat(entry.FilePath)
		if err == nil {
			initialEndOffset = fileInfo.Size()
		} else {
			initialEndOffset = 0
		}

		t, err := tail.TailFile(
			entry.FilePath,
			tail.Config{
				Location:      nil,
				ReOpen:        false,
				MustExist:     true,
				Poll:          false,
				Pipe:          false,
				Follow:        true,
				MaxLineSize:   0,
				CompleteLines: false,
				RateLimiter:   nil,
				Logger:        nil,
			},
		)
		if err != nil {
			Logger.Error("tail error", zap.Error(err), zap.String("entry", entry.String()))
			return
		}

		isCaughtUp := false

		if initialEndOffset == 0 {
			close(caughtUpChan)
			isCaughtUp = true
		}

	main:
		for {
			select {
			case <-ctx.Done():
				Logger.Info("context finished", zap.String("entry", entry.String()))
				t.Stop()
				break main
			case line, ok := <-t.Lines:
				if !ok {
					Logger.Info("processing finished", zap.String("entry", entry.String()))
					break main
				}

				lineStr := line.Text // copy required
				ev, err := processFn(lineStr)

				// CRITICAL: Check position even if processLine returns !ok (e.g. empty line or parse error)
				// We still consumed bytes, so we might be caught up.
				if !isCaughtUp {
					currentOffset, err := t.Tell()
					// If we have passed the initial size, we are now "Live"
					if err == nil && currentOffset >= initialEndOffset {
						close(caughtUpChan)
						isCaughtUp = true
					}
				}

				if err != nil {
					continue // TODO: display error here
				}

				eventChan <- ev
				switch any(ev).(type) {
				case *Shutdown, *Continued:
					Logger.Info("Shutdown/Continued event received")
					t.Stop()
					break main
				}
			}
		}
		Logger.Info("finished")
	}()

	return eventChan, caughtUpChan
}

func ReadLogAndFollow(entry EntryWithTime, ctx context.Context) (<-chan JournalEvent, <-chan struct{}) {
	return ReadLogAndFollowGeneric(entry, ctx, processLine)
}

type JournalEvents JournalEventBatch[JournalEvent]

func AttachToLastLog(ctx context.Context, loadNPrev int, archive string) (prev []JournalEventBatch[JournalEvent], current <-chan JournalEvents, caughtUp <-chan struct{}) {
	newJournalFileAppeared := watchForNewJournalLogs(ctx)
	globalCaughtUpChan := make(chan struct{})

	var currentEvents = make(chan JournalEvents)

	entries, err := ListAllLogs(archive)
	if err != nil {
		close(globalCaughtUpChan)
		close(currentEvents)
		Logger.Error("Failed to list logs", zap.Error(err))
		return []JournalEventBatch[JournalEvent]{}, currentEvents, globalCaughtUpChan

	}

	if len(entries) < 1 {
		close(globalCaughtUpChan)
		close(currentEvents)
		return []JournalEventBatch[JournalEvent]{}, currentEvents, globalCaughtUpChan
	}

	var prevEvents []JournalEventBatch[JournalEvent]

	lastEntry := len(entries) - 1
	lastEntryIsCurrent := IsCurrentLog(entries[lastEntry])

	if loadNPrev > 0 {
		var entriesToRead []EntryWithTime

		if lastEntryIsCurrent {
			entriesToRead = entries[len(entries)-min(len(entries), loadNPrev) : len(entries)-1]
		} else {
			entriesToRead = entries[len(entries)-min(len(entries), loadNPrev):]
		}
		prevEvents = ReadLogs(entriesToRead, ctx)
	} else {
		prevEvents = []JournalEventBatch[JournalEvent]{}
	}

	go func() {
		defer Logger.Info("AttachToLastLog finished")
		defer close(currentEvents)

		var events <-chan JournalEvent
		var caughtUpChan <-chan struct{}
		var caughtUpHandled bool
		var nextEntryToFollow EntryWithTime
		var waitForNewEntry bool

		if lastEntryIsCurrent {
			waitForNewEntry = false
			nextEntryToFollow = entries[lastEntry]
		} else {
			waitForNewEntry = true
			caughtUpHandled = true
			close(globalCaughtUpChan)
		}

		for {
			if waitForNewEntry {
				select {
				case <-ctx.Done():
					return
				case newEntry := <-newJournalFileAppeared:
					nextEntryToFollow = newEntry
					waitForNewEntry = false
					break
				}
			}

			je := JournalEvents{
				Entry:        nextEntryToFollow,
				Events:       make(chan JournalEvent, processorResultChanSize),
				CountResults: make(chan CountResult, 1),
			}
			currentEvents <- je

			followCtx, followCancel := context.WithCancel(context.Background())
			events, caughtUpChan = ReadLogAndFollow(nextEntryToFollow, followCtx)

			if !caughtUpHandled {
				caughtUpHandled = true
				go func(caughtUpChan <-chan struct{}) {
					<-caughtUpChan
					close(globalCaughtUpChan)
				}(caughtUpChan)
			}

			var successEvents uint64

		root:
			for {
				select {
				case event, ok := <-events:
					if !ok { // follow closed itself
						followCancel()
						close(je.Events)
						je.CountResults <- CountResult{
							Success: successEvents,
							Fail:    0,
							Bytes:   0,
						}
						close(je.CountResults)
						waitForNewEntry = true
						break root
					}
					je.Events <- event
				case newEntry, ok := <-newJournalFileAppeared:
					if !ok { // new file watcher closed
						followCancel()
						close(je.Events)
						je.CountResults <- CountResult{
							Success: successEvents,
							Fail:    0,
							Bytes:   0,
						}
						close(je.CountResults)
						waitForNewEntry = true
						break root
					}
					followCancel()
					nextEntryToFollow = newEntry
					// consume leftovers
					for event := range events {
						je.Events <- event
					}
					close(je.Events)
					je.CountResults <- CountResult{
						Success: successEvents,
						Fail:    0,
						Bytes:   0,
					}
					close(je.CountResults)
					break root
					//case <-ctx.Done():
					//	followCancel()
					//	// consume leftovers safely
					//	for event := range events {
					//		select {
					//		case je.Events <- event:
					//		default:
					//			// If the buffer is full during a shutdown, drop the leftover
					//			// to guarantee we don't block the shutdown process.
					//		}
					//	}
					//	close(je.Events)
					//	return
				}
			}

		}

	}()

	return prevEvents, currentEvents, globalCaughtUpChan
}
