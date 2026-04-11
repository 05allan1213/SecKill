package observability

import (
	"compress/gzip"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
)

type RotationConfig struct {
	MaxSizeMB  int
	MaxBackups int
	KeepDays   int
	Compress   bool
}

type Writer struct {
	access *rotatingFile
	errs   *rotatingFile
	slow   *rotatingFile
	stat   *rotatingFile
	severe *rotatingFile
}

func NewWriter(logDir string, rotation RotationConfig) (*Writer, error) {
	if logDir == "" {
		return nil, errors.New("log directory is required")
	}
	if err := os.MkdirAll(logDir, 0o755); err != nil {
		return nil, fmt.Errorf("create log directory %s: %w", logDir, err)
	}

	build := func(name string) (*rotatingFile, error) {
		return newRotatingFile(filepath.Join(logDir, name), rotation)
	}

	access, err := build("access.log")
	if err != nil {
		return nil, err
	}
	errs, err := build("error.log")
	if err != nil {
		_ = access.Close()
		return nil, err
	}
	slow, err := build("slow.log")
	if err != nil {
		_ = access.Close()
		_ = errs.Close()
		return nil, err
	}
	stat, err := build("stat.log")
	if err != nil {
		_ = access.Close()
		_ = errs.Close()
		_ = slow.Close()
		return nil, err
	}
	severe, err := build("severe.log")
	if err != nil {
		_ = access.Close()
		_ = errs.Close()
		_ = slow.Close()
		_ = stat.Close()
		return nil, err
	}

	return &Writer{
		access: access,
		errs:   errs,
		slow:   slow,
		stat:   stat,
		severe: severe,
	}, nil
}

func (w *Writer) Close() error {
	if w == nil {
		return nil
	}

	var errs []error
	for _, file := range []*rotatingFile{w.access, w.errs, w.slow, w.stat, w.severe} {
		if file == nil {
			continue
		}
		if err := file.Close(); err != nil {
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}

func (w *Writer) Alert(v any) {
	w.write(w.severe, "alert", v, nil)
}

func (w *Writer) Debug(v any, fields ...logx.LogField) {
	w.write(w.access, "debug", v, fields)
}

func (w *Writer) Error(v any, fields ...logx.LogField) {
	w.write(w.errs, "error", v, fields)
}

func (w *Writer) Info(v any, fields ...logx.LogField) {
	w.write(w.access, "info", v, fields)
}

func (w *Writer) Severe(v any) {
	w.write(w.severe, "severe", v, nil)
}

func (w *Writer) Slow(v any, fields ...logx.LogField) {
	w.write(w.slow, "slow", v, fields)
}

func (w *Writer) Stack(v any) {
	w.write(w.severe, "stack", v, nil)
}

func (w *Writer) Stat(v any, fields ...logx.LogField) {
	w.write(w.stat, "stat", v, fields)
}

func (w *Writer) write(target *rotatingFile, level string, msg any, fields []logx.LogField) {
	if target == nil {
		return
	}

	entry := map[string]interface{}{
		"@timestamp": time.Now().Format(time.RFC3339Nano),
		"level":      level,
		"content":    normalizeValue(msg),
	}
	for _, field := range fields {
		entry[field.Key] = normalizeValue(field.Value)
	}

	line, err := json.Marshal(entry)
	if err != nil {
		line = []byte(fmt.Sprintf(`{"@timestamp":%q,"level":%q,"content":%q,"marshal_error":%q}`,
			time.Now().Format(time.RFC3339Nano), level, fmt.Sprint(msg), err.Error()))
	}
	line = append(line, '\n')
	_ = target.Write(line)
}

func normalizeValue(v interface{}) interface{} {
	switch value := v.(type) {
	case nil:
		return nil
	case error:
		return value.Error()
	case fmt.Stringer:
		return value.String()
	default:
		return value
	}
}

type rotatingFile struct {
	path      string
	maxSize   int64
	backups   int
	keepDays  int
	compress  bool
	mu        sync.Mutex
	file      *os.File
	sizeBytes int64
}

func newRotatingFile(path string, cfg RotationConfig) (*rotatingFile, error) {
	rf := &rotatingFile{
		path:     path,
		maxSize:  int64(cfg.MaxSizeMB) * 1024 * 1024,
		backups:  cfg.MaxBackups,
		keepDays: cfg.KeepDays,
		compress: cfg.Compress,
	}
	if err := rf.openOrCreate(); err != nil {
		return nil, err
	}
	if err := rf.cleanupLocked(); err != nil {
		_ = rf.file.Close()
		return nil, err
	}
	return rf, nil
}

func (f *rotatingFile) Close() error {
	if f == nil {
		return nil
	}

	f.mu.Lock()
	defer f.mu.Unlock()
	if f.file == nil {
		return nil
	}

	err := f.file.Close()
	f.file = nil
	f.sizeBytes = 0
	return err
}

func (f *rotatingFile) Write(p []byte) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.file == nil {
		if err := f.openOrCreate(); err != nil {
			return err
		}
	}
	if f.maxSize > 0 && f.sizeBytes > 0 && f.sizeBytes+int64(len(p)) > f.maxSize {
		if err := f.rotateLocked(); err != nil {
			return err
		}
	}

	n, err := f.file.Write(p)
	f.sizeBytes += int64(n)
	return err
}

func (f *rotatingFile) openOrCreate() error {
	if err := os.MkdirAll(filepath.Dir(f.path), 0o755); err != nil {
		return fmt.Errorf("create log directory: %w", err)
	}

	file, err := os.OpenFile(f.path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("open log file %s: %w", f.path, err)
	}

	info, err := file.Stat()
	if err != nil {
		_ = file.Close()
		return fmt.Errorf("stat log file %s: %w", f.path, err)
	}

	f.file = file
	f.sizeBytes = info.Size()
	return nil
}

func (f *rotatingFile) rotateLocked() error {
	if f.file != nil {
		if err := f.file.Close(); err != nil {
			return err
		}
		f.file = nil
	}

	rotated := fmt.Sprintf("%s.%s", f.path, time.Now().Format("20060102T150405.000000000"))
	if err := os.Rename(f.path, rotated); err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("rotate log file %s: %w", f.path, err)
	}
	if f.compress {
		if err := compressFile(rotated); err != nil && !errors.Is(err, os.ErrNotExist) {
			return err
		}
	}
	if err := f.openOrCreate(); err != nil {
		return err
	}

	return f.cleanupLocked()
}

func (f *rotatingFile) cleanupLocked() error {
	rotated, err := f.rotatedFiles()
	if err != nil {
		return err
	}

	if f.keepDays > 0 {
		deadline := time.Now().Add(-time.Duration(f.keepDays) * 24 * time.Hour)
		filtered := rotated[:0]
		for _, candidate := range rotated {
			if candidate.ModTime.Before(deadline) {
				if err := os.Remove(candidate.Path); err != nil && !errors.Is(err, os.ErrNotExist) {
					return err
				}
				continue
			}
			filtered = append(filtered, candidate)
		}
		rotated = filtered
	}

	if f.backups > 0 && len(rotated) > f.backups {
		for _, candidate := range rotated[f.backups:] {
			if err := os.Remove(candidate.Path); err != nil && !errors.Is(err, os.ErrNotExist) {
				return err
			}
		}
	}

	return nil
}

type rotatedFileInfo struct {
	Path    string
	ModTime time.Time
}

func (f *rotatingFile) rotatedFiles() ([]rotatedFileInfo, error) {
	entries, err := os.ReadDir(filepath.Dir(f.path))
	if err != nil {
		return nil, fmt.Errorf("read log directory: %w", err)
	}

	base := filepath.Base(f.path)
	files := make([]rotatedFileInfo, 0, len(entries))
	for _, entry := range entries {
		name := entry.Name()
		if name == base {
			continue
		}
		if !strings.HasPrefix(name, base+".") && !strings.HasPrefix(name, base+"-") {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			return nil, fmt.Errorf("stat rotated log file %s: %w", name, err)
		}
		files = append(files, rotatedFileInfo{
			Path:    filepath.Join(filepath.Dir(f.path), name),
			ModTime: info.ModTime(),
		})
	}

	sort.Slice(files, func(i, j int) bool {
		return files[i].ModTime.After(files[j].ModTime)
	})
	return files, nil
}

func compressFile(path string) error {
	source, err := os.Open(path)
	if err != nil {
		return err
	}
	defer source.Close()

	target, err := os.Create(path + ".gz")
	if err != nil {
		return fmt.Errorf("create compressed log file: %w", err)
	}

	gzipWriter := gzip.NewWriter(target)
	if _, err := source.WriteTo(gzipWriter); err != nil {
		_ = gzipWriter.Close()
		_ = target.Close()
		_ = os.Remove(target.Name())
		return fmt.Errorf("compress log file %s: %w", path, err)
	}
	if err := gzipWriter.Close(); err != nil {
		_ = target.Close()
		_ = os.Remove(target.Name())
		return fmt.Errorf("finalize compressed log file %s: %w", path, err)
	}
	if err := target.Close(); err != nil {
		_ = os.Remove(target.Name())
		return fmt.Errorf("close compressed log file %s: %w", path, err)
	}

	if err := os.Remove(path); err != nil {
		return fmt.Errorf("remove rotated log file %s: %w", path, err)
	}
	return nil
}
