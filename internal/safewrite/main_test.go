package safewrite

import (
	"os"
	"path/filepath"
	"testing"
)

func TestOpenWrite_DevNull(t *testing.T) {
	w, err := Open(os.DevNull, func() bool {
		t.Fatal("confirmOverwrite should not be called for device")
		return false
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, err := w.Write([]byte("test")); err != nil {
		t.Fatalf("write failed: %v", err)
	}

	if err := w.Close(); err != nil {
		t.Fatalf("close failed: %v", err)
	}
}

func TestOpenWrite_InvalidPath(t *testing.T) {
	dir := t.TempDir()
	badPath := filepath.Join(dir, "no-such-dir", "file.bin")

	_, err := Open(badPath, func() bool { return true })
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestOpenWrite_CreateNewFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "new.bin")

	w, err := Open(path, func() bool {
		t.Fatal("confirmOverwrite should not be called")
		return false
	})
	if err != nil {
		t.Fatalf("openWrite failed: %v", err)
	}

	data := []byte("hello")
	if _, err := w.Write(data); err != nil {
		t.Fatalf("write failed: %v", err)
	}
	if err := w.Close(); err != nil {
		t.Fatalf("close failed: %v", err)
	}

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read failed: %v", err)
	}
	if string(got) != string(data) {
		t.Fatalf("unexpected content: %q", got)
	}

	if _, err := os.Stat(path + "~"); !os.IsNotExist(err) {
		t.Fatalf("backup file should not exist")
	}
}

func TestOpenWrite_OverwriteWithBackup(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "file.bin")

	if err := os.WriteFile(path, []byte("old"), 0666); err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	w, err := Open(path, func() bool { return true })
	if err != nil {
		t.Fatalf("openWrite failed: %v", err)
	}

	if _, err := w.Write([]byte("new")); err != nil {
		t.Fatalf("write failed: %v", err)
	}
	if err := w.Close(); err != nil {
		t.Fatalf("close failed: %v", err)
	}

	cur, _ := os.ReadFile(path)
	if string(cur) != "new" {
		t.Fatalf("unexpected content: %q", cur)
	}

	bak, _ := os.ReadFile(path + "~")
	if string(bak) != "old" {
		t.Fatalf("unexpected backup content: %q", bak)
	}
}
