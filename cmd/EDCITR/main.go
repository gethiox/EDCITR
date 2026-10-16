package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/gethiox/EDCITR/internal/app"
	"github.com/gethiox/EDCITR/internal/console"
	"github.com/gethiox/EDCITR/internal/utils"
	"github.com/spf13/pflag"
)

var runChild = pflag.Bool("run", false, "internal use")

func main() {
	_ = pflag.CommandLine.MarkHidden("run")
	pflag.Parse()

	if *runChild {
		go profile()
		app.NewApp().Run()
		return
	}

	exe, _ := os.Executable()
	cmd := exec.Command(exe, "--run")

	var stderrBuf bytes.Buffer

	cmd.Stderr = io.MultiWriter(os.Stderr, &stderrBuf)
	cmd.Stdout = os.Stdout

	err := cmd.Run()

	if err != nil {
		timestamp := utils.GenerateTimestampForFilename()
		finalPath, _ := filepath.Abs(fmt.Sprintf("EDCITR-crash-%s.log", timestamp))
		os.WriteFile(finalPath, stderrBuf.Bytes(), 0666)

		description := fmt.Sprintf(
			"A detailed crash log has been saved to:\n%s\n\n"+
				"Please include it when opening an issue.", finalPath,
		)
		msg := fmt.Sprintf("ED: Chat Is This Real encountered a fatal error.\n\n%s", description)

		console.ShowCrashDialog("EDCITR: Rainbow Crash", msg)
	}
}
