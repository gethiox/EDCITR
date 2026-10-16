package inara

import (
	"context"
	"errors"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"net/http"
	"net/url"
	"path"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gethiox/EDCITR/internal/version"
	"github.com/gethiox/EDCITR/internal/zaputils"
	"go.uber.org/zap"
)

var Logger = zaputils.Logger.Named("PFPDownloader")

var (
	ErrorNotModified = errors.New("not modified")
	ErrorQueueClosed = errors.New("queue closed")
)

type StatusCodeError struct {
	StatusCode int
}

func (e StatusCodeError) Error() string {
	return fmt.Sprintf("unexpected status code: %d", e.StatusCode)
}

type QueueEntry struct {
	PictureURL   string
	CmdrID       uint64
	CmdrName     string
	LastModified time.Time
}

type PFPDownloader struct {
	httpClient            *http.Client
	maxDownloadsPerSecond uint
	results               chan<- DownloadImageResult
	closed                bool
	header                string
	headerSet             bool
	readyChan             chan struct{}
	ctx                   context.Context
	cancelProcessor       context.CancelFunc
	queue                 []QueueEntry
	mtx                   sync.Mutex
}

type DownloadImageResult struct {
	RequestedURL string
	CmdrID       uint64
	CmdrName     string

	Image        image.Image
	LastModified time.Time
	Error        error
}

func GetPFPDownloader(resultSize int) (downloader *PFPDownloader, results <-chan DownloadImageResult) {
	resultChan := make(chan DownloadImageResult, resultSize)

	ctx, cancel := context.WithCancel(context.Background())
	downloader = &PFPDownloader{
		httpClient: &http.Client{
			Timeout: time.Second * 10,
		},
		maxDownloadsPerSecond: 2,
		results:               resultChan,
		closed:                false,
		header:                "",
		headerSet:             false,
		readyChan:             make(chan struct{}),
		ctx:                   ctx,
		cancelProcessor:       cancel,
		queue:                 make([]QueueEntry, 0, 64),
		mtx:                   sync.Mutex{},
	}
	go downloader.queueProcessor()
	return downloader, resultChan
}

func (d *PFPDownloader) SetHeader(appName string, version version.Version) error {
	switch {
	case d.closed:
		return errors.New("client is closed")
	case appName == "":
		return errors.New("appName is required")
	case !version.IsValid():
		return fmt.Errorf("version is invalid: %v", version)
	default:
		d.header = fmt.Sprintf("%s/%s", appName, version.FormatFull())
		d.headerSet = true
		close(d.readyChan)
		return nil
	}
}

func (d *PFPDownloader) Ready() bool {
	select {
	case <-d.readyChan:
		return true
	default:
		return false
	}
}

func (d *PFPDownloader) ReadyChan() <-chan struct{} {
	return d.readyChan
}

func (d *PFPDownloader) Close() error {
	if !d.closed {
		d.mtx.Lock()
		d.closed = true
		d.cancelProcessor()
		d.mtx.Unlock()
		if !d.headerSet {
			close(d.readyChan)
		}
	}
	return nil
}

func (d *PFPDownloader) QueuePicturesDownload(entries []QueueEntry) error {
	d.mtx.Lock()
	if d.closed {
		return ErrorQueueClosed
	}
	for _, e := range entries {
		d.queue = append(d.queue, e)
	}
	d.mtx.Unlock()
	return nil
}

func (d *PFPDownloader) InQueue() int {
	d.mtx.Lock()
	defer d.mtx.Unlock()
	return len(d.queue)
}

func (d *PFPDownloader) queueProcessor() {
	var minReqDuration = time.Second / time.Duration(d.maxDownloadsPerSecond)
	var reqStart, pushStart time.Time
	var reqDur, pushDur time.Duration

	Logger := Logger.Named("Downloader")

	Logger.Info("Waiting for client to be ready")
	select {
	case <-d.readyChan:
		break
	case <-d.ctx.Done():
		Logger.Info("Context expired")
		close(d.results)
		return
	}

	var (
		successes atomic.Int64
		fails     atomic.Int64
	)

	go func() {
		reportPeriod := time.Minute
		for {
			select {
			case <-d.ctx.Done():
				return
			case <-time.After(reportPeriod):
				s, f := successes.Swap(0), fails.Swap(0)
				if s == 0 && f == 0 {
					continue
				}
				Logger.Info("Periodic activity download report", zap.Duration("duration", reportPeriod), zap.Int("downloaded", int(s)), zap.Int("errors", int(fails.Load())))
			}
		}
	}()

	Logger.Info("Client ready, starting processing queue")
	for {
		select {
		case <-d.ctx.Done():
			Logger.Info("Context expired")
			close(d.results)
			return
		default:
			d.mtx.Lock()
			if len(d.queue) == 0 {
				if cap(d.queue) > 64 {
					d.queue = make([]QueueEntry, 0, 64)
				}
				d.mtx.Unlock()
				time.Sleep(time.Second)
				continue
			}

			e := d.queue[len(d.queue)-1]
			d.queue = d.queue[:len(d.queue)-1]
			d.mtx.Unlock()

			Logger.Debug("Downloading image", zap.String("CMDR", e.CmdrName), zap.String("url", e.PictureURL), zap.Time("last_modified", e.LastModified))
			reqStart = time.Now()
			img, lmtime, err := d.downloadPicture(e.PictureURL, e.LastModified)
			reqDur = time.Since(reqStart)
			if err != nil {
				fails.Add(1)
				Logger.Debug("Sending unsuccessful result", zap.Error(err), zap.String("CMDR", e.CmdrName), zap.String("url", e.PictureURL), zap.Time("last_modified", e.LastModified), zap.Time("last_modified_server", lmtime))
			} else {
				successes.Add(1)
				Logger.Debug("Sending successful result", zap.String("CMDR", e.CmdrName), zap.String("url", e.PictureURL), zap.Time("last_modified", e.LastModified), zap.Time("last_modified_server", lmtime))
			}

			result := DownloadImageResult{
				RequestedURL: e.PictureURL,
				CmdrID:       e.CmdrID,
				CmdrName:     e.CmdrName,
				Image:        img,
				LastModified: lmtime,
				Error:        err,
			}
			pushStart = time.Now()

			select {
			case d.results <- result:
				// ok
			case <-time.After(time.Second * 2):
				Logger.Warn("Results channel full or consumer dead, dropping image", zap.String("CMDR", e.CmdrName))
			case <-d.ctx.Done():
				Logger.Info("Context expired")
				return
			}

			pushDur = time.Since(pushStart)
			if pushDur+reqDur < minReqDuration {
				time.Sleep(minReqDuration - (pushDur + reqDur))
			}
		}
	}
}

func (d *PFPDownloader) downloadPicture(pictureURL string, lastModifiedSrc time.Time) (image.Image, time.Time, error) {
	u, err := url.Parse(pictureURL)
	if err != nil {
		return nil, time.Time{}, fmt.Errorf("url parse failed: %w", err)
	}

	req, err := http.NewRequest(http.MethodGet, pictureURL, nil)
	if err != nil {
		return nil, time.Time{}, fmt.Errorf("create request failed: %w", err)
	}
	req.Header.Set("User-Agent", "EDCITR/0.0.0")
	if !lastModifiedSrc.IsZero() {
		req.Header.Set("If-Modified-Since", lastModifiedSrc.UTC().Format(http.TimeFormat))
	}

	resp, err := d.httpClient.Do(req)
	if err != nil {
		return nil, time.Time{}, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	var lastModified time.Time
	lastModifiedString := resp.Header.Get("Last-Modified")
	if lastModifiedString != "" {
		lastModified, err = time.Parse(http.TimeFormat, lastModifiedString)
		if err != nil {
			return nil, time.Time{}, fmt.Errorf("parse last-modified failed: %w", err)
		}
	}

	if resp.StatusCode == http.StatusNotModified {
		return nil, lastModified, fmt.Errorf("image has not changed: %w", ErrorNotModified)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, lastModified, StatusCodeError{resp.StatusCode}
	}

	var img image.Image
	elements := strings.Split(path.Base(u.String()), ".")
	switch ext := elements[len(elements)-1]; strings.ToLower(ext) {
	case "jpg", "jpeg":
		img, err = jpeg.Decode(resp.Body)
	case "png":
		img, err = png.Decode(resp.Body)
	default:
		return nil, lastModified, fmt.Errorf("unsupported image format: .%s", ext)
	}
	if err != nil {
		return nil, lastModified, fmt.Errorf("image decode failed: %w", err)
	}
	return img, lastModified, nil
}
