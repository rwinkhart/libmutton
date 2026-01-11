//go:build windows

package global

import (
	"syscall"

	"golang.org/x/sys/windows"
)

func GetSysProcAttr() *syscall.SysProcAttr {
	return &syscall.SysProcAttr{CreationFlags: syscall.CREATE_NEW_PROCESS_GROUP | windows.DETACHED_PROCESS}
}
