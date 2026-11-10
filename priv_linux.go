//go:build linux

package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"syscall"
)

// dropPrivileges lets the container start as root only for as long as it takes
// to hand the data directory to an unprivileged user, then switches to that
// user for good. It does nothing when the process is not root.
//
// Needed capabilities: CHOWN (hand over the data directory) and SETUID/SETGID
// (switch user). After the switch the process holds no privileges at all.
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
