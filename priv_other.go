//go:build !linux

package main

// dropPrivileges is only needed in the Linux container.
func dropPrivileges(dataDir string, uid, gid int) error { return nil }
