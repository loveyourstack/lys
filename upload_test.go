package lys

import (
	"bytes"
	"context"
	"encoding/json"
	"image"
	"image/color"
	"image/png"
	"io"
	"log/slog"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/loveyourstack/lys/lysformfile"
)

type testUserInfo struct{ userID int64 }

func (u testUserInfo) GetUserId() int64 { return u.userID }

func pngBytes() []byte {
	img := image.NewNRGBA(image.Rect(0, 0, 8, 8))
	for x := 0; x < 8; x++ {
		for y := 0; y < 8; y++ {
			img.Set(x, y, color.NRGBA{R: 255, G: 0, B: 0, A: 255})
		}
	}

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		panic(err)
	}
	return buf.Bytes()
}

func newMultipartRequest(t *testing.T, parts map[string]struct {
	filename string
	content  []byte
	mimeType string
}) *http.Request {
	t.Helper()

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	for fieldName, part := range parts {
		fh, err := writer.CreateFormFile(fieldName, part.filename)
		if err != nil {
			t.Fatalf("CreateFormFile failed: %v", err)
		}
		if _, err := fh.Write(part.content); err != nil {
			t.Fatalf("Write part content failed: %v", err)
		}
	}

	if err := writer.Close(); err != nil {
		t.Fatalf("writer.Close failed: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/upload", &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	return req
}

func TestValidateDestPath(t *testing.T) {
	t.Run("accepts directory", func(t *testing.T) {
		dir := t.TempDir()
		if err := validateDestPath(dir); err != nil {
			t.Fatalf("validateDestPath returned error for valid directory: %v", err)
		}
	})

	t.Run("rejects empty path", func(t *testing.T) {
		if err := validateDestPath(""); err == nil {
			t.Fatal("expected error for empty destPath")
		}
	})

	t.Run("rejects file path", func(t *testing.T) {
		file, err := os.CreateTemp(t.TempDir(), "not-dir-*.txt")
		if err != nil {
			t.Fatalf("CreateTemp failed: %v", err)
		}
		defer file.Close()

		if err := validateDestPath(file.Name()); err == nil {
			t.Fatal("expected error when destPath is a file")
		}
	})
}

func TestUpload_HandlerSuccess(t *testing.T) {
	dir := t.TempDir()
	params := lysformfile.ExtractParams{
		AllowedMimeTypes: []string{"image/png"},
		MaxSizePerFile:   1024 * 1024,
		MaxFiles:         new(2),
	}

	req := newMultipartRequest(t, map[string]struct {
		filename string
		content  []byte
		mimeType string
	}{
		"file": {filename: "avatar.png", content: pngBytes(), mimeType: "image/png"},
	})
	req = req.WithContext(context.WithValue(req.Context(), UserInfoCtxKey, testUserInfo{userID: 42}))

	rec := httptest.NewRecorder()
	handler := Upload(params, dir, slog.New(slog.NewTextHandler(io.Discard, nil)))
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d, body=%s", http.StatusCreated, rec.Code, rec.Body.String())
	}

	var resp struct {
		Status string `json:"status"`
		Data   struct {
			FileResults []struct {
				OriginalName string `json:"original_name"`
				StoredName   string `json:"stored_name"`
			} `json:"file_results"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("json.Unmarshal failed: %v; body=%s", err, rec.Body.String())
	}
	if resp.Status != ReqSucceeded {
		t.Fatalf("expected status %q, got %q", ReqSucceeded, resp.Status)
	}
	if len(resp.Data.FileResults) != 1 {
		t.Fatalf("expected 1 uploaded file result, got %d", len(resp.Data.FileResults))
	}
	if resp.Data.FileResults[0].OriginalName != "avatar.png" {
		t.Fatalf("expected original name avatar.png, got %q", resp.Data.FileResults[0].OriginalName)
	}
	if resp.Data.FileResults[0].StoredName == "" {
		t.Fatal("expected stored_name to be populated")
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("os.ReadDir failed: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 file saved to dest dir, got %d", len(entries))
	}
}

func TestUpload_HandlerRejectsUnauthenticatedUser(t *testing.T) {
	dir := t.TempDir()
	params := lysformfile.ExtractParams{
		AllowedMimeTypes: []string{"image/png"},
		MaxSizePerFile:   1024 * 1024,
		MaxFiles:         new(2),
	}

	req := newMultipartRequest(t, map[string]struct {
		filename string
		content  []byte
		mimeType string
	}{
		"file": {filename: "avatar.png", content: pngBytes(), mimeType: "image/png"},
	})

	rec := httptest.NewRecorder()
	handler := Upload(params, dir, slog.New(slog.NewTextHandler(io.Discard, nil)))
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected status %d, got %d, body=%s", http.StatusForbidden, rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "user not authenticated") {
		t.Fatalf("expected unauthenticated error in body, got %s", rec.Body.String())
	}
}
