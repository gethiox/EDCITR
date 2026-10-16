package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime/debug"

	"github.com/gethiox/EDCITR/internal/console"
	"github.com/gethiox/EDCITR/internal/utils"
	"github.com/gethiox/EDCITR/internal/zaputils"
	"go.uber.org/zap"
)

var Logger = zaputils.Logger.Named("crash")

// setupCrashHandler returns cleanup function that has to be executed at the end of program run.
// WARNING: DO NOT use defer statement on returned function.
// otherwise deferred call will close the pipe (used as synchronization) prematurely and spawned runCrashWatcher
// will wake up too early where Go would not produce any data on registered debug.SetCrashOutput yet
//
// Example:
//
//	cleanupCrashLog := setupCrashHandler()
//	myApp()
//	cleanupCrashLog()
func setupCrashHandler() func() {
	tempFile, err := os.CreateTemp("", "edcitr-crash-*.log")
	if err != nil {
		return func() {}
	}

	debug.SetCrashOutput(tempFile, debug.CrashOptions{})

	exe, _ := os.Executable()
	cmd := exec.Command(exe, "--crash-monitor", tempFile.Name())

	r, w, _ := os.Pipe()
	cmd.Stdin = r

	console.HideConsoleWindow(cmd)

	err = cmd.Start()
	if err != nil {
		Logger.Warn("Failed to start crash monitor", zap.Error(err))
	}

	return func() {
		w.Close()
		tempFile.Close()
		os.Remove(tempFile.Name())
	}
}

func runCrashWatcher(tempFilePath string) {
	io.ReadAll(os.Stdin) // block until parent dies

	info, err := os.Stat(tempFilePath)
	if err != nil || info.Size() == 0 {
		os.Remove(tempFilePath)
		return
	}

	timestamp := utils.GenerateTimestampForFilename()
	finalPath := fmt.Sprintf("EDCITR-crash-%s.log", timestamp)

	tempData, _ := os.ReadFile(tempFilePath)
	err = os.WriteFile(finalPath, tempData, 0666)

	var description string

	if err != nil {
		description = fmt.Sprintf(
			"A detailed crash log has failed to be written in current working directory:\n%v\n\n"+
				"Temporary file should still exist at:\n%s\n\n"+
				"Please include it when opening an issue.", err, tempFilePath,
		)
	} else {
		absPath, _ := filepath.Abs(finalPath)
		description = fmt.Sprintf(
			"A detailed crash log has been saved to:\n%s\n\n"+
				"Please include it when opening an issue.", absPath,
		)
		_ = os.Remove(tempFilePath)
	}

	msg := fmt.Sprintf("ED: Chat Is This Real encountered a fatal error.\n\n%s", description)

	console.ShowCrashDialog("EDCITR: Rainbow Crash", msg)
}
