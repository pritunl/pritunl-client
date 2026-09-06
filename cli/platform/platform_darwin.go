package platform

import (
	"syscall"
)

func DetachAttr() *syscall.SysProcAttr {
	return &syscall.SysProcAttr{
		Setsid: true,
	}
}
