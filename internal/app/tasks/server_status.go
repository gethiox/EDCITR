package tasks

import (
	"context"
	"sync"
	"time"

	"github.com/gethiox/EDCITR/internal/app/state"
	"github.com/gethiox/EDCITR/internal/config"
	"github.com/gethiox/EDCITR/internal/elite_status"
	"github.com/gethiox/rotational-velocidensity-buffer"
	"go.uber.org/zap"
)

type ServerStatus struct {
	Logger *zap.Logger
	Config *config.EDCITRConfig
	State  *state.TheHolyGlobalState
}

func (j *ServerStatus) Run(ctx context.Context) {
	j.Logger.Info("Starting Server Status Job")

	// Re-initialize buffers on fresh start
	j.State.ServerStatusRVB = rvb.NewBuffer[state.ServerStatus](j.Config.ServerStatus.HistorySize)
	j.State.ServerStatusTimestampsRVB = rvb.NewBuffer[time.Time](j.Config.ServerStatus.HistorySize)

	timeout := j.Config.ServerStatus.Timeout
	checkPeriod := j.Config.ServerStatus.Period
	statusClient := elite_status.GetProductionClient(timeout)

	var nextPing = time.Now()
	var fullPeriodCycles time.Duration

	// Track transient HTTP request goroutines
	var wg sync.WaitGroup

	for {
		j.State.ServerStatusTimestampsRVB.Push(nextPing)
		start := time.Now()

		select {
		case <-ctx.Done():
			wg.Wait() // Wait for any lingering HTTP requests to finish
			j.Logger.Info("Server Status Job cleanly shut down")
			return
		case <-time.After(time.Until(nextPing)):
			break
		}
		end := time.Now()

		wg.Add(1)
		go func(pingTime time.Time) {
			defer wg.Done()

			status, err := statusClient.GetEDServerStatus()
			if err != nil {
				j.Logger.Warn("request failed", zap.Error(err))
				j.State.ServerStatusRVB.Push(state.ServerStatus{
					Err:       err,
					Timestamp: pingTime,
					Status:    status,
				})
				return
			}
			j.State.ServerStatusRVB.Push(state.ServerStatus{
				Err:       nil,
				Timestamp: pingTime,
				Status:    status,
			})
		}(nextPing)

		fullPeriodCycles = end.Sub(start) / checkPeriod
		if fullPeriodCycles > 0 {
			nextPing = nextPing.Add(checkPeriod * fullPeriodCycles)
		} else {
			nextPing = nextPing.Add(checkPeriod)
		}
	}
}
