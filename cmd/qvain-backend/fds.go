// +build linux

package main

import (
	"errors"
	"os"
	"syscall"
	"unsafe"
)

// Call syscall.SYS_CAPGET for CAP_NET_BIND_SERVICE (= 10).
// See also:
//   https://github.com/syndtr/gocapability/
//   linux/kernel/capability.c
// setcap cap_net_bind_service=+ep ./this

const (
	linuxCapVer1 = 0x19980330
	linuxCapVer2 = 0x20071026
	linuxCapVer3 = 0x20080522
)

var errUnknownVersion = errors.New("unknown capability version")

type capHeader struct {
	version uint32
	pid     int
}

type capData struct {
	effective   uint32
	permitted   uint32
	inheritable uint32
}

type capsV1 struct {
	hdr  capHeader
	data capData
}

type capsV3 struct {
	hdr     capHeader
	data    [2]capData
	bounds  [2]uint32
	ambient [2]uint32
}

// linux/capability.h
const CAP_NET_BIND_SERVICE = 10

func callCapGet() (caps *capData, err error) {
	hdr := new(capHeader)
	data := (*capData)(nil)

	// get version
	_, _, errptr := syscall.Syscall(syscall.SYS_CAPGET, uintptr(unsafe.Pointer(hdr)), uintptr(unsafe.Pointer(data)), 0)
	if errptr != 0 {
		return nil, errptr // syscall.Errno, probably "invalid argument"
	}

	switch hdr.version {
	case linuxCapVer1:
		c := new(capsV1)
		c.hdr.version = hdr.version
		_, _, errptr = syscall.Syscall(syscall.SYS_CAPGET, uintptr(unsafe.Pointer(&c.hdr)), uintptr(unsafe.Pointer(&c.data)), 0)
		caps = &c.data
	case linuxCapVer2, linuxCapVer3:
		c := new(capsV3)
		c.hdr.version = hdr.version
		_, _, errptr = syscall.Syscall(syscall.SYS_CAPGET, uintptr(unsafe.Pointer(&c.hdr)), uintptr(unsafe.Pointer(&c.data[0])), 0)
		caps = &c.data[0]
	default:
		return nil, errUnknownVersion
	}

	if errptr != 0 {
		return nil, errptr
	}

	return
}

func canNetBindService() (bool, error) {
	caps, err := callCapGet()
	if err != nil {
		return false, err
	}

	return caps.effective&(1<<CAP_NET_BIND_SERVICE) != 0, nil
}

func getOpenFDs() (int, error) {
	f, err := os.Open("/proc/self/fd")
	if err != nil {
		return -1, err
	}
	list, err := f.Readdirnames(-1)
	f.Close()
	if err != nil {
		return -1, err
	}
	return len(list), nil
}

func getRlimit() (uint64, uint64, error) {
	var limit syscall.Rlimit
	if err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, &limit); err != nil {
		return 0, 0, err
	}
	return limit.Cur, limit.Max, nil
}
