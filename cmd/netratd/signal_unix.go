//go:build linux || darwin

package main

import (
	"os"
	"syscall"
)

var (
	signalsToHandle = []os.Signal{
		syscall.SIGINT,
		syscall.SIGTERM,
	}
)
