package observability

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestWriterRoutesByLevel(t *testing.T) {
	dir := t.TempDir()
	writer, err := NewWriter(dir, RotationConfig{MaxSizeMB: 10, MaxBackups: 7, KeepDays: 7, Compress: false})
	if err != nil {
		t.Fatalf("create writer failed: %v", err)
	}
	defer writer.Close()

	writer.Info("info-message")
	writer.Error("error-message")
	writer.Slow("slow-message")
	writer.Stat("stat-message")
	writer.Severe("severe-message")

	assertContains(t, filepath.Join(dir, "access.log"), "info-message")
	assertContains(t, filepath.Join(dir, "error.log"), "error-message")
	assertContains(t, filepath.Join(dir, "slow.log"), "slow-message")
	assertContains(t, filepath.Join(dir, "stat.log"), "stat-message")
	assertContains(t, filepath.Join(dir, "severe.log"), "severe-message")
}

func TestWriterRotatesAndCompresses(t *testing.T) {
	dir := t.TempDir()
	writer, err := NewWriter(dir, RotationConfig{
		MaxSizeMB:  1,
		MaxBackups: 2,
		KeepDays:   7,
		Compress:   true,
	})
	if err != nil {
		t.Fatalf("create writer failed: %v", err)
	}
	defer writer.Close()

	line := strings.Repeat("a", 700*1024)
	writer.Info(line)
	writer.Info(line)
	writer.Info(line)

	matches, err := filepath.Glob(filepath.Join(dir, "access.log.*.gz"))
	if err != nil {
		t.Fatalf("glob rotated files failed: %v", err)
	}
	if len(matches) == 0 {
		t.Fatalf("expected compressed rotated logs, got none")
	}
	if len(matches) > 2 {
		t.Fatalf("expected backup cap to apply, got %d rotated files", len(matches))
	}
}

func TestWriterRemovesExpiredRotatedFiles(t *testing.T) {
	dir := t.TempDir()
	writer, err := NewWriter(dir, RotationConfig{MaxSizeMB: 10, MaxBackups: 7, KeepDays: 1, Compress: false})
	if err != nil {
		t.Fatalf("create writer failed: %v", err)
	}
	defer writer.Close()

	oldFile := filepath.Join(dir, "access.log-2026-04-01")
	if err := os.WriteFile(oldFile, []byte("old"), 0o644); err != nil {
		t.Fatalf("write old file failed: %v", err)
	}
	oldTime := time.Now().Add(-48 * time.Hour)
	if err := os.Chtimes(oldFile, oldTime, oldTime); err != nil {
		t.Fatalf("set old file times failed: %v", err)
	}

	if err := writer.access.cleanupLocked(); err != nil {
		t.Fatalf("cleanup old files failed: %v", err)
	}
	if _, err := os.Stat(oldFile); !os.IsNotExist(err) {
		t.Fatalf("expected expired file to be removed, stat err=%v", err)
	}
}

func TestTruncateString(t *testing.T) {
	got := TruncateString("abcdefghijklmnopqrstuvwxyz", 20)
	if got != "abcdef...(truncated)" {
		t.Fatalf("unexpected truncated string: %q", got)
	}
}

func assertContains(t *testing.T, path, want string) {
	t.Helper()

	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s failed: %v", path, err)
	}
	if !strings.Contains(string(body), want) {
		t.Fatalf("expected %s to contain %q, got %s", path, want, string(body))
	}
}
