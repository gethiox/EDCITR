package tasks

import (
	"context"
	"errors"
	"strings"
	"sync"
	"time"

	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/gethiox/EDCITR/internal/app/state"
	"github.com/gethiox/EDCITR/internal/config"
	"github.com/gethiox/EDCITR/internal/inara"
	"github.com/gethiox/EDCITR/internal/inara/inapi/events/commander_profile"
	"github.com/gethiox/EDCITR/internal/storage"
	"go.uber.org/zap"
)

type InaraService struct {
	Config         *config.EDCITRConfig
	State          *state.TheHolyGlobalState
	InaraClient    *inara.Client
	UIState        *state.UIState
	CacheStore     *storage.Cache
	PFPDownloader  *inara.PFPDownloader
	PFPResultsChan <-chan inara.DownloadImageResult
	Logger         *zap.Logger
}

func (j *InaraService) Run(ctx context.Context) {
	Logger := j.Logger.Named("InaraJob")

	Logger.Info("Starting Inara Background Jobs")
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		j.inaraBatchWorker(ctx)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		j.inaraCMDRStreamWorker(ctx)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-ctx.Done():
				Logger.Info("Context expired")
				return
			case result, ok := <-j.PFPResultsChan:
				if !ok {
					Logger.Info("PFP result channel closed")
					return
				}
				Logger.Debug("processing result", zap.String("CMDR", result.CmdrName))
				if result.Error != nil {
					if errors.Is(result.Error, inara.StatusCodeError{StatusCode: 404}) {
						err := j.CacheStore.SaveMissingProfilePicture(result.CmdrID)
						if err != nil {
							j.State.Inara.CMDRTexturesState.Set(result.CmdrName, state.StateOKButMissing)
							Logger.Error("Failed to save missing profile picture", zap.Error(err))
							continue
						}
						j.State.Inara.CMDRTexturesState.Set(result.CmdrName, state.StateOKButMissing)
						continue
					}

					j.State.Inara.CMDRTexturesState.Set(result.CmdrName, state.StateError)
					Logger.Error("download error", zap.Error(result.Error), zap.Any("result", result))
					continue
				}

				err := j.CacheStore.SaveProfilePicture(result.CmdrID, result.Image)
				if err != nil {
					Logger.Error("Failed to save profile picture", zap.Error(err))
					continue
				}
				Logger.Debug("saved profile picture", zap.String("CMDR", result.CmdrName))

				j.State.Inara.CMDRTexturesState.Set(result.CmdrName, state.StateDownloading)
				j.UIState.ProfilePicturesToLoadMtx.Lock()
				j.UIState.ProfilePicturesToLoad = append(j.UIState.ProfilePicturesToLoad, state.TexLoadInfo{
					CmdrName: result.CmdrName,
					Img:      result.Image,
				})
				j.UIState.ProfilePicturesToLoadMtx.Unlock()
				Logger.Debug("profile picture image sent for texture loading", zap.String("CMDR", result.CmdrName))
			}
		}
	}()

	defer func() {
		j.UIState.ProfilePicturesToLoadMtx.Lock()
		j.UIState.ProfilePicturesToLoad = make([]state.TexLoadInfo, 0, 64)
		j.UIState.ProfilePicturesToLoadMtx.Unlock()

		j.State.Inara.CMDRTextures.IterCb(func(cmdrName string, tex rl.Texture2D) {
			rl.UnloadTexture(tex) // TODO: this should (gradually) happen in the main thread, but I don't mind stutter for now
		})

		j.State.Inara.CMDRTextures.Clear()
		j.State.Inara.CMDRInfo.Clear()
		j.State.Inara.CMDRTexturesState.Clear()
		j.State.Inara.CMDRInfoState.Clear()
	}()

	wg.Wait()
	Logger.Info("Inara Background Jobs cleanly shut down")
}

func (j *InaraService) inaraBatchWorker(ctx context.Context) {
	Logger := j.Logger.Named("Inara")
	bulkSize := 1000
	waitTimeBeforeRequest := j.Config.InaraSettings.WaitBeforeRequest

	var collectedBulk = make([]string, 0, bulkSize)
	var collected = make([][]string, 0, 16)
	var collectedMtx sync.Mutex
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		timer := time.NewTimer(waitTimeBeforeRequest)
		defer timer.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case cmdr := <-j.State.Inara.CMDRInfoChan:
				timer.Reset(waitTimeBeforeRequest)

				collectedBulk = append(collectedBulk, cmdr)
				j.State.Inara.CMDRInfoQueued.Add(1)
				if len(collectedBulk) < bulkSize {
					continue
				}

				collectedMtx.Lock()
				collected = append(collected, collectedBulk)
				collectedMtx.Unlock()
				collectedBulk = make([]string, 0, bulkSize)
			case <-timer.C:
				timer.Reset(waitTimeBeforeRequest)

				if len(collectedBulk) == 0 {
					continue
				}

				collectedMtx.Lock()
				collected = append(collected, collectedBulk)
				collectedMtx.Unlock()
				collectedBulk = make([]string, 0, bulkSize)
			}
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()

		Logger.Info("waiting for inara api client to be ready")
		select {
		case <-ctx.Done():
			return
		case <-j.InaraClient.ReadyChan():
			Logger.Info("inara api client ready")
		}

		var lastRequestTime = time.Now().Add(-time.Second * 60)
		Logger.Info("Starting collecting commanders to fetch profile info", zap.Duration("wait_before_request", waitTimeBeforeRequest), zap.Int("bulk_size", bulkSize))

		for {
			select {
			case <-ctx.Done():
				return
			default:
			}

			collectedMtx.Lock()
			hasData := len(collected) > 0
			collectedMtx.Unlock()

			if !hasData {
				select {
				case <-ctx.Done():
					return
				case <-time.After(time.Second):
				}
				continue
			}

			collectedMtx.Lock()
			batch := collected[len(collected)-1]
			collected = collected[:len(collected)-1]
			collectedMtx.Unlock()

			Logger.Debug("fetching profile info", zap.Int("batch_size", len(batch)))

			for _, cmdrName := range batch {
				j.State.Inara.CMDRInfoState.Set(cmdrName, state.StateDownloading)
			}

			if wait := time.Until(lastRequestTime.Add(time.Second * 30)); wait > 0 {
				Logger.Info("waiting to not exceed 2 req per minute", zap.Duration("duration", wait))
				select {
				case <-ctx.Done():
					return
				case <-time.After(wait):
				}
			}

			respTime := time.Now()
			lastRequestTime = respTime
			resp, err := j.InaraClient.GetCommanderProfile(batch)
			if err != nil {
				Logger.Error("GetCommanderProfile failed", zap.Error(err))
				for _, cmdrName := range batch {
					j.State.Inara.CMDRInfoState.Set(cmdrName, state.StateError)
				}
				j.State.Inara.CMDRInfoQueued.Add(int64(-len(batch)))
				continue
			}

			Logger.Debug("Got response with profiles", zap.Int("total_responses", len(resp.Events)), zap.Int("batch_size", len(batch)))

			if len(resp.Events) != len(batch) {
				for _, cmdrName := range batch {
					j.State.Inara.CMDRInfoState.Set(cmdrName, state.StateError)
				}
				j.State.Inara.CMDRInfoQueued.Add(int64(-len(batch)))
				Logger.Error("GetCommanderProfile returned unexpected event list size", zap.Int("got", len(resp.Events)), zap.Int("expected", len(collected)))
				continue
			}

			var picturesToDownload = make([]inara.QueueEntry, 0, len(resp.Events))

			var present, missing int

			for i, ev := range resp.Events {
				cmdrName := batch[i]

				if ev.EventStatus != 200 {
					missing++
					j.State.Inara.CMDRInfoState.Set(cmdrName, state.StateOKButMissing)
					j.State.Inara.CMDRInfoQueued.Add(-1)
					err := j.CacheStore.SaveMissingProfileInfo(cmdrName)
					if err != nil {
						Logger.Error("SaveMissingProfileInfo failed", zap.Error(err), zap.String("CMDR", cmdrName))
						continue
					}
					continue
				}
				present++

				j.State.Inara.CMDRInfo.Set(cmdrName, storage.InfoWithTime{
					ProfileInfo:     ev.EventData,
					ProfileInfoTime: respTime,
				})
				j.State.Inara.CMDRInfoKeepAlive.Set(cmdrName, j.UIState.CurrentTime.Load())
				j.State.Inara.CMDRInfoState.Set(cmdrName, state.StateOK)
				j.State.Inara.CMDRInfoQueued.Add(-1)

				err := j.CacheStore.SaveProfileInfo(cmdrName, ev.EventData)
				if err != nil {
					Logger.Error("SaveProfileInfo failed", zap.Error(err))
					continue
				}
				Logger.Debug("Saved profile info", zap.String("CMDR", cmdrName))

				if strings.HasSuffix(ev.EventData.AvatarImageURL, commander_profile.NoProfilePictureSuffix) {
					j.State.Inara.CMDRTexturesState.Set(cmdrName, state.StateOKButMissing)
					err := j.CacheStore.SaveMissingProfilePicture(ev.EventData.UserID)
					if err != nil {
						Logger.Error("SaveMissingProfilePicture failed", zap.Error(err))
						continue
					}
					continue
				}

				j.State.Inara.CMDRTexturesState.Set(cmdrName, state.StateQueued)
				picturesToDownload = append(picturesToDownload, inara.QueueEntry{
					PictureURL:   ev.EventData.AvatarImageURL,
					CmdrName:     cmdrName,
					CmdrID:       ev.EventData.UserID,
					LastModified: time.Time{},
				})
				Logger.Debug("added picture to download queue buffer", zap.String("CMDR", cmdrName))
			}

			Logger.Info("Profile response processed, scheduling picture download", zap.Int("picture_count", len(picturesToDownload)), zap.Int("existing_profiles", present), zap.Int("missing_profiles", missing), zap.Int("total_profiles", len(resp.Events)))
			err = j.PFPDownloader.QueuePicturesDownload(picturesToDownload)
			if err != nil {
				Logger.Error("QueuePicturesDownload failed", zap.Error(err))
				continue
			}
			Logger.Debug("pictures to download queued successfully")
		}
	}()

	wg.Wait()
	Logger.Info("Inara batch worker cleanly shut down")
}

func (j *InaraService) inaraCMDRStreamWorker(ctx context.Context) {
	j.State.Inara.CMDRNamesStreamActive.Store(true)
	j.State.Inara.CMDRNamesRefreshAvailable.Store(true)

	for {
		select {
		case CMDR := <-j.State.Inara.CMDRNamesStream:
			j.QueueInaraProfileFetch(CMDR)
		case <-ctx.Done():
			j.State.Inara.CMDRNamesStreamActive.Store(false)

			for {
				select {
				case <-time.After(time.Millisecond * 500):
					return
				case <-j.State.Inara.CMDRNamesStream:
					// drain until timeout
				}
			}
		}
	}
}

func (j *InaraService) QueueInaraProfileFetch(cmdrName string) {
	if len(cmdrName) == 0 {
		j.Logger.Error("cmdrName is empty")
		return
	}

	if !j.State.ProcessingLiveEvents.Load() {
		return
	}

	if j.State.Inara.CMDRInfoState.Has(cmdrName) {
		return
	}

	info, infoTime, err := j.CacheStore.LoadProfileInfoFromName(cmdrName)

	//if err == nil && time.Since(infoTime) > 24*time.Hour { // todo: make it handle well when "awaiting for full login" (inara client not ready yet)
	//	err = storage.ProfileNotFound // Force a refetch
	//}

	if err != nil {
		switch {
		case errors.Is(err, storage.HasNoProfileError):
			j.State.Inara.CMDRInfoState.Set(cmdrName, state.StateOKButMissing)
		case errors.Is(err, storage.ProfileNotFound):
			select {
			case j.State.Inara.CMDRInfoChan <- cmdrName:
				j.State.Inara.CMDRInfoState.Set(cmdrName, state.StateQueued)
			default:
				// Channel is full (worker is busy/dead). Drop the request to prevent deadlock.
				j.Logger.Warn("CMDRInfoChan is full, dropping profile fetch", zap.String("CMDR", cmdrName))
			}
		default:
			j.State.Inara.CMDRInfoState.Set(cmdrName, state.StateError)
			j.Logger.Warn("Failed to load profile info", zap.Error(err))
		}
		return
	}

	j.State.Inara.CMDRInfo.Set(cmdrName, storage.InfoWithTime{
		info,
		infoTime,
	})
	j.State.Inara.CMDRInfoKeepAlive.Set(cmdrName, j.UIState.CurrentTime.Load())
	j.State.Inara.CMDRInfoState.Set(cmdrName, state.StateOK)

	if j.State.Inara.CMDRTexturesState.Has(cmdrName) {
		return
	}

	pic, _, err := j.CacheStore.LoadProfilePicture(info.UserID)
	if err != nil {
		switch {
		case errors.Is(err, storage.HasNoPictureError):
			j.State.Inara.CMDRTexturesState.Set(cmdrName, state.StateOKButMissing)
		case errors.Is(err, storage.PictureNotFound):
			_ = j.PFPDownloader.QueuePicturesDownload([]inara.QueueEntry{
				{
					PictureURL:   info.AvatarImageURL,
					CmdrID:       info.UserID,
					CmdrName:     cmdrName,
					LastModified: time.Time{},
				},
			})
			j.State.Inara.CMDRTexturesState.Set(cmdrName, state.StateQueued)
		default:
			j.State.Inara.CMDRTexturesState.Set(cmdrName, state.StateError)
			j.Logger.Warn("Failed to load profile info", zap.Error(err))
		}
		return
	}

	j.UIState.ProfilePicturesToLoadMtx.Lock()
	j.UIState.ProfilePicturesToLoad = append(j.UIState.ProfilePicturesToLoad, state.TexLoadInfo{
		CmdrName: cmdrName,
		Img:      pic,
	})
	j.UIState.ProfilePicturesToLoadMtx.Unlock()
	j.State.Inara.CMDRTexturesState.Set(cmdrName, state.StateQueued)
}
