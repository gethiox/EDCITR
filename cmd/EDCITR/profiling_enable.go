//go:build profile

package main

import (
	"net/http"
	_ "net/http/pprof"

	"go.uber.org/zap"
)

func profile() {
	Logger.Info("profiling enabled, usage:")
	Logger.Info("> go tool pprof -http localhost:8080 http://localhost:6060/debug/pprof/profile?seconds=30")
	Logger.Info("> go tool pprof -http localhost:8080 http://localhost:6060/debug/pprof/heap")
	err := http.ListenAndServe("localhost:6060", nil)
	if err != nil {
		Logger.Error("profile server error", zap.Error(err))
	}
	Logger.Info("profile server finished")
}
