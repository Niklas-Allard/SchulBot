package store

import (
	"os"
	"path/filepath"
	"testing"
)

func TestStore(t *testing.T) {
	dir := t.TempDir()
	s, err := New(filepath.Join(dir, "test.db"))
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	defer s.Close()

	const id = "msg-001@example.com"

	ok, err := s.IsProcessed(id)
	if err != nil || ok {
		t.Fatalf("expected not processed, got ok=%v err=%v", ok, err)
	}

	if err := s.MarkProcessed(id); err != nil {
		t.Fatalf("MarkProcessed: %v", err)
	}

	ok, err = s.IsProcessed(id)
	if err != nil || !ok {
		t.Fatalf("expected processed, got ok=%v err=%v", ok, err)
	}
}

func TestStoreReopens(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.db")

	s, _ := New(path)
	_ = s.MarkProcessed("abc@test")
	s.Close()

	s2, err := New(path)
	if err != nil {
		t.Fatalf("reopen: %v", err)
	}
	defer s2.Close()

	ok, err := s2.IsProcessed("abc@test")
	if err != nil || !ok {
		t.Fatalf("expected persisted, got ok=%v err=%v", ok, err)
	}
}

func TestStoreMkDir(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "sub", "dir", "test.db")
	s, err := New(path)
	if err != nil {
		t.Fatalf("New with nested path: %v", err)
	}
	s.Close()
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("db file missing: %v", err)
	}
}