//go:build linux

package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestChownDataDirToSelf(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "state.json"), []byte("{}"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := chownDataDir(dir, os.Getuid(), os.Getgid()); err != nil {
		t.Fatalf("chownDataDir: %v", err)
	}
}

func TestDropPrivilegesNoopWhenNotRoot(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("running as root")
	}
	if err := dropPrivileges(t.TempDir(), 65532, 65532); err != nil {
		t.Fatalf("expected no-op, got %v", err)
	}
}

func TestEnvIDRejectsRoot(t *testing.T) {
	t.Setenv("WOL_UID", "0")
	if _, err := envID("WOL_UID", 65532); err == nil {
		t.Fatal("uid 0 must be refused")
	}
	t.Setenv("WOL_UID", "1000")
	if n, err := envID("WOL_UID", 65532); err != nil || n != 1000 {
		t.Fatalf("got %d, %v", n, err)
	}
}
