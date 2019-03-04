// +build !linux

package main

import (
	"errors"
)

var errSysNotSupported = errors.New("system funtion is not supported on this platform")

func canNetBindService() (bool, error) {
	return false, nil
}

func getOpenFDs() (int, error) {
	return 0, errSysNotSupported
}

func getRlimit() (uint64, uint64, error) {
	return 0, 0, errSysNotSupported
}
