package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDedupe(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "a.txt"), []byte("hello"), 0o644)
	os.WriteFile(filepath.Join(dir, "b.txt"), []byte("hello"), 0o644)
	os.WriteFile(filepath.Join(dir, "c.txt"), []byte("world"), 0o644)

	groups, err := Dedupe(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(groups) != 1 {
		t.Fatalf("expected 1 group, got %d", len(groups))
	}
}
