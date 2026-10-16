package tasks

import (
	"context"
	"time"

	"github.com/gethiox/EDCITR/internal/app/state"
	"github.com/gethiox/sleep-detect-go"
	"go.uber.org/zap"
)

type WakeupDetector struct {
	Logger *zap.Logger
	State  *state.TheHolyGlobalState
}

func (j *WakeupDetector) Run(ctx context.Context) {
	period := time.Second * 2
	margin := time.Millisecond * 250
	trigger := time.Second * 10

	bob := sleep_detect.NewBobTheStrider(period, margin)
	defer func() { _ = bob.Close() }()

	wakeups := bob.DetectWakeupEvents()

	for {
		select {
		case <-ctx.Done():
			return
		case ev := <-wakeups:
			if ev < trigger {
				continue
			}

			go func() {
				j.Logger.Warn("sleep detected", zap.Duration("duration", ev))
				time.Sleep(time.Second * 10)
				//time.Sleep(time.Millisecond * 500)
				j.Logger.Info("audio reload requested")
				j.State.AudioReloadRequired = true
			}()
		}
	}
}
