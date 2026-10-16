package generate

import (
	"context"
	"io"
	"slices"
	"time"

	"github.com/gethiox/EDCITR/internal/app/state"
	"github.com/gethiox/EDCITR/internal/parser"
	"go.uber.org/zap"
)

type GenericGenerator[T any] struct {
	Logger     *zap.Logger
	StatePtr   *state.DataState
	ArchiveDir string

	LogProducer  func(logs []parser.EntryWithTime, ctx context.Context) []parser.JournalEventBatch[T]
	FileProducer func(identifier string) (io.WriteCloser, error)
	Evaluator    func(event T) (keep bool, identifier string, payload []byte)
}

func (p *GenericGenerator[T]) Run(ctx context.Context) {
	*p.StatePtr = state.StateQueued

	var stateOnReturn = state.StateError
	defer func() {
		*p.StatePtr = stateOnReturn
	}()

	logs, err := parser.ListAllLogs(p.ArchiveDir)
	if err != nil {
		p.Logger.Error("Listing logs failed", zap.Error(err))
		return
	}

	p.Logger.Info("File generation started", zap.Int("journals", len(logs)))

	encountered := make(map[string]io.WriteCloser)
	start := time.Now()

	defer func() {
		for _, fd := range encountered {
			fd.Write([]byte("\n]\n"))
			fd.Close()
		}
		p.Logger.Info("File generation completed", zap.Duration("duration", time.Since(start)), zap.Int("produced_files", len(encountered)))
	}()

	counters := make(map[string]int)

	defer func() {
		type counterEntry struct {
			identifier string
			count      int
		}

		var toSort []counterEntry
		for k, v := range counters {
			toSort = append(toSort, counterEntry{identifier: k, count: v})
		}

		slices.SortFunc(toSort, func(a, b counterEntry) int {
			if a.count > b.count {
				return 1
			}
			if a.count < b.count {
				return -1
			}
			return 0
		})
		slices.Reverse(toSort)

		p.Logger.Info("Counters report:")
		for _, v := range toSort {
			p.Logger.Info("> "+v.identifier, zap.Int("count", v.count))
		}
	}()

	entries := p.LogProducer(logs, ctx)
	for _, entry := range entries {
		p.Logger.Debug("Processing Entry", zap.String("entry", entry.Entry.String()))

		for ev := range entry.Events {
			keep, identifier, payload := p.Evaluator(ev)
			if !keep {
				continue
			}

			counters[identifier]++

			fd, ok := encountered[identifier]
			if !ok {
				fd, err = p.FileProducer(identifier)
				if err != nil {
					p.Logger.Error("failed to create file", zap.Error(err), zap.String("identifier", identifier), zap.String("entry", entry.Entry.String()))
					return
				}
				fd.Write([]byte("[\n"))
				encountered[identifier] = fd

				fd.Write(payload)
				continue
			}

			fd.Write([]byte(",\n"))
			fd.Write(payload)
		}
	}

	stateOnReturn = state.StateOK
}
