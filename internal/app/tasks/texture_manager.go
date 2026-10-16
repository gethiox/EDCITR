package tasks

import (
	"context"

	"github.com/gethiox/EDCITR/internal/app/state"
	"github.com/gethiox/EDCITR/internal/config"
	"github.com/gethiox/EDCITR/internal/storage"
	"go.uber.org/zap"
)

type TextureManagerWorker struct {
	Logger     *zap.Logger
	Config     *config.EDCITRConfig
	State      *state.TheHolyGlobalState
	UIState    *state.UIState
	CacheStore *storage.Cache
}

func (t *TextureManagerWorker) Run(ctx context.Context) {
	//workers := make(chan struct{}, 32)
	//
	//for CMDR := range t.State.CMDRNamesStream {
	//	workers <- struct{}{}
	//	go func(CMDR string) {
	//		defer func() { <-workers }()
	//		t.QueueInaraProfileFetch(CMDR)
	//	}(CMDR)
	//}
}
