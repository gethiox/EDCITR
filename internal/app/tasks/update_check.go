package tasks

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gethiox/EDCITR/internal/app/state"
	"github.com/gethiox/EDCITR/internal/version"
	"github.com/google/go-github/v83/github"
	"go.uber.org/zap"
)

type CheckForUpdates struct {
	Logger                          *zap.Logger
	State                           *state.TheHolyGlobalState
	RepoOwner, RepoName, APIAppName string
}

func (j *CheckForUpdates) Run(ctx context.Context) {
	cc := NewCheckerClient(time.Second*10, j.RepoOwner, j.RepoName, j.APIAppName, version.Release.FormatFull())
	var Period = time.Hour

	defer func() {
		j.State.UpdateCheckState.Store(0)
		j.State.UpdateCheckLast.Store(time.Time{})
		j.State.UpdateCheckVersion.Store(version.Version{})
		j.State.UpdateCheckReleaseTime.Store(time.Time{})
	}()

	var toSleep time.Duration
	var retryPool = 3
	j.State.UpdateCheckLast.Store(time.Now())

root:
	for {
		select {
		case <-ctx.Done():
			break root
		case <-time.After(toSleep):
			break
		}

		if time.Since(j.State.UpdateCheckLast.Load()) > Period/2 {
			retryPool = 3
		}

		start := time.Now()
		j.State.UpdateCheckState.Store(state.StateDownloading)
		release, err := cc.GetLastRelease()
		stop := time.Now()

		j.State.UpdateCheckLast.Store(start)

		duration := stop.Sub(start)

		toSleep = Period - duration

		if err != nil {
			if errors.Is(err, ErrNoReleasesAvailable) {
				j.State.UpdateCheckState.Store(state.StateOKButMissing)
			} else {
				j.State.UpdateCheckState.Store(state.StateError)
				if retryPool > 0 {
					toSleep = time.Minute
					retryPool--
					j.Logger.Error("get last release failed, retrying soon", zap.Error(err), zap.Duration("sleeping", toSleep))
				} else {
					j.Logger.Error("get last release failed", zap.Error(err), zap.Duration("sleeping", toSleep))
				}
			}
			continue
		}

		ver, err := version.NewVersionFromString(release.GetTagName())
		if err != nil {
			j.State.UpdateCheckState.Store(state.StateError)
			j.Logger.Error("creating version from tag failed", zap.Error(err), zap.String("tag", release.GetTagName()), zap.Duration("sleeping", toSleep))
			continue
		}

		j.State.UpdateCheckReleaseTime.Store(release.GetCreatedAt().UTC())
		j.State.UpdateCheckVersion.Store(ver)
		j.State.UpdateCheckState.Store(state.StateOK)
	}
}

type CheckerClient struct {
	ghClient  *github.Client
	repoOwner string
	repoName  string
}

func NewCheckerClient(timeout time.Duration, repoOwner, repoName, appNameUA, versionUA string) *CheckerClient {
	clientHttp := &http.Client{Timeout: timeout}
	ghClient := github.NewClient(clientHttp)
	ghClient.UserAgent = fmt.Sprintf("%s/%s (%s)", appNameUA, versionUA, ghClient.UserAgent)

	return &CheckerClient{ghClient, repoOwner, repoName}
}

var ErrNoReleasesAvailable = errors.New("no releases available")

func (c *CheckerClient) GetLastRelease() (*github.RepositoryRelease, error) {
	release, resp, err := c.ghClient.Repositories.GetLatestRelease(context.Background(), c.repoOwner, c.repoName)
	if err != nil {
		if resp != nil && resp.StatusCode == http.StatusNotFound {
			return nil, ErrNoReleasesAvailable
		}
		return nil, err
	}
	return release, nil
}
