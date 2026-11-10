//go:build linux

package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"syscall"
	"unsafe"
)

const (
	prSetKeepCaps = 8
	capNetRaw     = 13 // CAP_NET_RAW: raw ICMP sockets for the online check
	capVersion3   = 0x20080522
)

type capHeader struct {
	Version uint32
	Pid     int32
}

type capData struct {
	Effective, Permitted, Inheritable uint32
}

// hasPermittedCap reports whether the process currently holds the capability.
func hasPermittedCap(c uint) bool {
	hdr := capHeader{Version: capVersion3}
	var data [2]capData
	if _, _, e := syscall.RawSyscall(syscall.SYS_CAPGET, uintptr(unsafe.Pointer(&hdr)), uintptr(unsafe.Pointer(&data[0])), 0); e != 0 {
		return false
	}
	return data[c/32].Permitted&(1<<(c%32)) != 0
}

// keepNetRaw sets the capability as the only one kept after the user switch.
func keepNetRaw() error {
	hdr := capHeader{Version: capVersion3}
	var data [2]capData
	data[0].Effective = 1 << capNetRaw
	data[0].Permitted = 1 << capNetRaw
	if _, _, e := syscall.AllThreadsSyscall(syscall.SYS_CAPSET, uintptr(unsafe.Pointer(&hdr)), uintptr(unsafe.Pointer(&data[0])), 0); e != 0 {
		return e
	}
	return nil
}

// dropPrivileges lets the container start as root only for as long as it takes
// to hand the data directory to an unprivileged user, then switches to that
// user for good. It does nothing when the process is not root.
//
// Needed capabilities: CHOWN (hand over the data directory) and SETUID/SETGID
// (switch user). After the switch the process holds no privileges at all,
// except CAP_NET_RAW when the container was given it: that lets the online
// check send ICMP pings without any kernel setting. Drop it from cap_add to
// run without it (the online check then uses TCP ports).
func dropPrivileges(dataDir string, uid, gid int) error {
	if os.Geteuid() != 0 {
		return nil
	}
	if err := os.MkdirAll(dataDir, 0o700); err != nil {
		return fmt.Errorf("creating data directory: %w", err)
	}
	if err := chownDataDir(dataDir, uid, gid); err != nil {
		// Not fatal: the directory may already belong to the right user.
		log.Printf("warning: could not change owner of %s to %d:%d (%v). If saving fails, add the CHOWN capability or fix the owner by hand", dataDir, uid, gid, err)
	}
	keep := hasPermittedCap(capNetRaw)
	if keep {
		if _, _, e := syscall.AllThreadsSyscall(syscall.SYS_PRCTL, prSetKeepCaps, 1, 0); e != 0 {
			log.Printf("warning: cannot keep CAP_NET_RAW (%v), ping will be unavailable", e)
			keep = false
		}
	}
	if err := syscall.Setgroups([]int{}); err != nil {
		return fmt.Errorf("clearing groups: %w (add the SETGID capability)", err)
	}
	if err := syscall.Setgid(gid); err != nil {
		return fmt.Errorf("setgid %d: %w (add the SETGID capability)", gid, err)
	}
	if err := syscall.Setuid(uid); err != nil {
		return fmt.Errorf("setuid %d: %w (add the SETUID capability)", uid, err)
	}
	if os.Geteuid() != uid || os.Getuid() != uid || os.Getegid() != gid || os.Getgid() != gid {
		return fmt.Errorf("still running as %d:%d after switching user", os.Geteuid(), os.Getegid())
	}
	if keep {
		if err := keepNetRaw(); err != nil {
			log.Printf("warning: cannot keep CAP_NET_RAW (%v), ping will be unavailable", err)
		}
	}
	// Root must not be regainable.
	if syscall.Setuid(0) == nil {
		return fmt.Errorf("switched user but root can be regained")
	}
	log.Printf("running as user %d:%d (data directory: %s)", uid, gid, dataDir)
	return nil
}

// chownDataDir changes the owner of the directory and of the files directly in
// it. It does not descend further and does not follow symbolic links.
func chownDataDir(dir string, uid, gid int) error {
	if err := os.Lchown(dir, uid, gid); err != nil {
		return err
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	for _, e := range entries {
		if err := os.Lchown(filepath.Join(dir, e.Name()), uid, gid); err != nil {
			return err
		}
	}
	return nil
}
