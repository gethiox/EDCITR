//go:build windows

package console

import (
	"os"
	"os/exec"
	"syscall"
	"unsafe"
)

func init() {
	if os.Getenv("EDCITR_FORCE_STDOUT") != "" {
		return
	}

	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	procGetConsoleProcessList := kernel32.NewProc("GetConsoleProcessList")
	procFreeConsole := kernel32.NewProc("FreeConsole")

	var pids [2]uint32
	count, _, _ := syscall.SyscallN(procGetConsoleProcessList.Addr(), uintptr(unsafe.Pointer(&pids[0])), 2)

	if count == 1 { // no processes are attached to this console
		syscall.SyscallN(procFreeConsole.Addr())

		nullDev, err := os.OpenFile("NUL", os.O_WRONLY, 0666)
		if err != nil {
			return
		}

		os.Stdout = nullDev
		os.Stderr = nullDev
	}
}

func HideConsoleWindow(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{CreationFlags: 0x08000000} // CREATE_NO_WINDOW
}
