package lysformfile

import (
	"errors"
	"io"
	"log/slog"
	"mime/multipart"
	"os"
	"testing"
)

type errReadFile struct {
	readOnce bool
}

func (f *errReadFile) Read(p []byte) (int, error) {
	if f.readOnce {
		return 0, errors.New("simulated read failure")
	}
	f.readOnce = true
	copyLen := copy(p, pngBytes())
	return copyLen, nil
}

func (f *errReadFile) ReadAt(p []byte, offset int64) (int, error) {
	return 0, errors.New("simulated ReadAt failure")
}

func (f *errReadFile) Seek(offset int64, whence int) (int64, error) { return 0, nil }
func (f *errReadFile) Close() error                                 { return nil }

func TestStreamFilesToDestPath_RollsBackOnCopyFailure(t *testing.T) {
	dir := t.TempDir()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	uploads := []UploadFile{{
		File:       &errReadFile{},
		FileHeader: &multipart.FileHeader{Filename: "broken.png", Size: int64(len(pngBytes()))},
		MimeType:   "image/png",
	}}

	_, err := StreamToDisk(uploads, dir, 99, logger)
	if err == nil {
		t.Fatal("expected StreamToDisk to fail on copy error")
	}

	entries, readErr := os.ReadDir(dir)
	if readErr != nil {
		t.Fatalf("os.ReadDir failed: %v", readErr)
	}
	if len(entries) != 0 {
		t.Fatalf("expected no files left after rollback, found %d entries", len(entries))
	}
}
