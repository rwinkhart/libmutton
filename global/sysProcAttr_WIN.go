//go:build windows

package global

import "syscall"

func GetSysProcAttr() *syscall.SysProcAttr {
	return &syscall.SysProcAttr{}
}
