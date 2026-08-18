package lysformfile

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

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

func TestExtractFromRequest_AllowsValidPNG(t *testing.T) {
	params := ExtractParams{
		AllowedMimeTypes: []string{"image/png"},
		MaxSizePerFile:   1024 * 1024,
		MaxFiles:         new(2),
	}

	req := newMultipartRequest(t, map[string]struct {
		filename string
		content  []byte
		mimeType string
	}{
		"file": {filename: "test.png", content: pngBytes(), mimeType: "image/png"},
	})

	files, err := ExtractFromRequest(req, params)
	if err != nil {
		t.Fatalf("ExtractFromRequest returned error for valid PNG: %v", err)
	}
	if len(files) != 1 {
		t.Fatalf("expected 1 file, got %d", len(files))
	}
	if files[0].MimeType != "image/png" {
		t.Fatalf("expected MIME image/png, got %q", files[0].MimeType)
	}
	if files[0].FileHeader.Filename != "test.png" {
		t.Fatalf("expected original filename test.png, got %q", files[0].FileHeader.Filename)
	}
}

func TestExtractFromRequest_RejectsDisallowedMimeType(t *testing.T) {
	params := ExtractParams{
		AllowedMimeTypes: []string{"image/png"},
		MaxSizePerFile:   1024 * 1024,
		MaxFiles:         new(2),
	}

	req := newMultipartRequest(t, map[string]struct {
		filename string
		content  []byte
		mimeType string
	}{
		"file": {filename: "notes.txt", content: []byte("plain text payload"), mimeType: "text/plain"},
	})

	_, err := ExtractFromRequest(req, params)
	if err == nil {
		t.Fatal("expected error for disallowed MIME type")
	}
	if !strings.Contains(err.Error(), "disallowed MIME type") {
		t.Fatalf("expected disallowed MIME error, got: %v", err)
	}
}

func TestExtractFromRequest_EnforcesMaxFiles(t *testing.T) {
	params := ExtractParams{
		AllowedMimeTypes: []string{"image/png"},
		MaxSizePerFile:   1024 * 1024,
		MaxFiles:         new(1),
	}

	req := newMultipartRequest(t, map[string]struct {
		filename string
		content  []byte
		mimeType string
	}{
		"file1": {filename: "first.png", content: pngBytes(), mimeType: "image/png"},
		"file2": {filename: "second.png", content: pngBytes(), mimeType: "image/png"},
	})

	_, err := ExtractFromRequest(req, params)
	if err == nil {
		t.Fatal("expected error for exceeding max file count")
	}
	if !strings.Contains(err.Error(), "too many files uploaded") {
		t.Fatalf("expected too-many-files error, got: %v", err)
	}
}

func TestExtractFromRequest_RejectsNoFiles(t *testing.T) {
	params := ExtractParams{
		AllowedMimeTypes: []string{"image/png"},
		MaxSizePerFile:   1024 * 1024,
		MaxFiles:         new(2),
	}

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	if err := writer.Close(); err != nil {
		t.Fatalf("writer.Close failed: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/upload", &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	_, err := ExtractFromRequest(req, params)
	if err == nil {
		t.Fatal("expected error when no files uploaded")
	}
	if !strings.Contains(err.Error(), "no files uploaded") {
		t.Fatalf("expected no-files error, got: %v", err)
	}
}

func TestExtractFromRequest_ValidatesContentTypeHeader(t *testing.T) {
	t.Run("rejects missing content type", func(t *testing.T) {
		params := ExtractParams{
			AllowedMimeTypes: []string{"image/png"},
			MaxSizePerFile:   1024 * 1024,
			MaxFiles:         new(1),
		}

		req := httptest.NewRequest(http.MethodPost, "/upload", strings.NewReader("ignored"))
		_, err := ExtractFromRequest(req, params)
		if err == nil {
			t.Fatal("expected error for missing Content-Type header")
		}
		if !strings.Contains(err.Error(), "Content-Type header is missing") {
			t.Fatalf("expected missing header error, got: %v", err)
		}
	})

	t.Run("rejects non multipart content type", func(t *testing.T) {
		params := ExtractParams{
			AllowedMimeTypes: []string{"image/png"},
			MaxSizePerFile:   1024 * 1024,
			MaxFiles:         new(1),
		}

		req := httptest.NewRequest(http.MethodPost, "/upload", strings.NewReader("ignored"))
		req.Header.Set("Content-Type", "application/json")
		_, err := ExtractFromRequest(req, params)
		if err == nil {
			t.Fatal("expected error for non multipart Content-Type")
		}
		if !strings.Contains(err.Error(), "Content-Type must be multipart/form-data") {
			t.Fatalf("expected multipart header restriction, got: %v", err)
		}
	})
}

func TestExtractParamsValidateAndProcess(t *testing.T) {
	t.Run("rejects empty allowed mime types", func(t *testing.T) {
		maxFiles := 1
		params := ExtractParams{MaxSizePerFile: 1024, MaxFiles: &maxFiles}
		if err := params.validateAndProcess(); err == nil {
			t.Fatal("expected error for empty AllowedMimeTypes")
		}
	})

	t.Run("rejects invalid max files", func(t *testing.T) {
		maxFiles := 0
		params := ExtractParams{
			AllowedMimeTypes: []string{"image/png"},
			MaxSizePerFile:   1024,
			MaxFiles:         &maxFiles,
		}
		if err := params.validateAndProcess(); err == nil {
			t.Fatal("expected error for MaxFiles <= 0")
		}
	})

	t.Run("accepts valid params", func(t *testing.T) {
		maxFiles := 3
		params := ExtractParams{
			AllowedMimeTypes: []string{"image/png", "image/jpeg"},
			MaxSizePerFile:   1024,
			MaxFiles:         &maxFiles,
		}
		if err := params.validateAndProcess(); err != nil {
			t.Fatalf("validateAndProcess returned unexpected error: %v", err)
		}
		if params.calcMaxFiles != 3 {
			t.Fatalf("expected calcMaxFiles=3, got %d", params.calcMaxFiles)
		}
		if params.maxBodyBytes <= 0 {
			t.Fatalf("expected positive maxBodyBytes, got %d", params.maxBodyBytes)
		}
	})
}

func TestExtractParamsValidateAndProcess_ImageDimensionConstraints(t *testing.T) {
	t.Run("rejects negative dimensions", func(t *testing.T) {
		minWidth := -1
		params := ExtractParams{
			AllowedMimeTypes: []string{"image/png"},
			MaxSizePerFile:   1024,
			ImgMinWidthPx:    &minWidth,
		}
		if err := params.validateAndProcess(); err == nil {
			t.Fatal("expected error for negative ImgMinWidthPx")
		}
	})

	t.Run("rejects min greater than max", func(t *testing.T) {
		minWidth := 150
		maxWidth := 100
		params := ExtractParams{
			AllowedMimeTypes: []string{"image/png"},
			MaxSizePerFile:   1024,
			ImgMinWidthPx:    &minWidth,
			ImgMaxWidthPx:    &maxWidth,
		}
		if err := params.validateAndProcess(); err == nil {
			t.Fatal("expected error when ImgMinWidthPx > ImgMaxWidthPx")
		}
	})

	t.Run("accepts valid dimension bounds", func(t *testing.T) {
		minWidth := 8
		maxWidth := 16
		minHeight := 8
		maxHeight := 16
		params := ExtractParams{
			AllowedMimeTypes: []string{"image/png"},
			MaxSizePerFile:   1024,
			ImgMinWidthPx:    &minWidth,
			ImgMaxWidthPx:    &maxWidth,
			ImgMinHeightPx:   &minHeight,
			ImgMaxHeightPx:   &maxHeight,
		}
		if err := params.validateAndProcess(); err != nil {
			t.Fatalf("validateAndProcess rejected valid image bounds: %v", err)
		}
	})
}

func TestExtractFromRequest_ValidatesImageDimensions(t *testing.T) {
	t.Run("rejects width below minimum", func(t *testing.T) {
		minWidth := 9
		params := ExtractParams{
			AllowedMimeTypes: []string{"image/png"},
			MaxSizePerFile:   1024 * 1024,
			MaxFiles:         new(1),
			ImgMinWidthPx:    &minWidth,
		}

		req := newMultipartRequest(t, map[string]struct {
			filename string
			content  []byte
			mimeType string
		}{
			"file": {filename: "small.png", content: pngBytes(), mimeType: "image/png"},
		})

		_, err := ExtractFromRequest(req, params)
		if err == nil {
			t.Fatal("expected error for image width below minimum")
		}
		if !strings.Contains(err.Error(), "image width 8 px is less than the minimum allowed 9 px") {
			t.Fatalf("expected width-too-small error, got: %v", err)
		}
	})

	t.Run("rejects height above maximum", func(t *testing.T) {
		maxHeight := 7
		params := ExtractParams{
			AllowedMimeTypes: []string{"image/png"},
			MaxSizePerFile:   1024 * 1024,
			MaxFiles:         new(1),
			ImgMaxHeightPx:   &maxHeight,
		}

		req := newMultipartRequest(t, map[string]struct {
			filename string
			content  []byte
			mimeType string
		}{
			"file": {filename: "tall.png", content: pngBytes(), mimeType: "image/png"},
		})

		_, err := ExtractFromRequest(req, params)
		if err == nil {
			t.Fatal("expected error for image height above maximum")
		}
		if !strings.Contains(err.Error(), "image height 8 px exceeds the maximum allowed 7 px") {
			t.Fatalf("expected height-too-large error, got: %v", err)
		}
	})

	t.Run("accepts images within bounds", func(t *testing.T) {
		minWidth := 8
		maxWidth := 8
		minHeight := 8
		maxHeight := 8
		params := ExtractParams{
			AllowedMimeTypes: []string{"image/png"},
			MaxSizePerFile:   1024 * 1024,
			MaxFiles:         new(1),
			ImgMinWidthPx:    &minWidth,
			ImgMaxWidthPx:    &maxWidth,
			ImgMinHeightPx:   &minHeight,
			ImgMaxHeightPx:   &maxHeight,
		}

		req := newMultipartRequest(t, map[string]struct {
			filename string
			content  []byte
			mimeType string
		}{
			"file": {filename: "valid.png", content: pngBytes(), mimeType: "image/png"},
		})

		files, err := ExtractFromRequest(req, params)
		if err != nil {
			t.Fatalf("ExtractFromRequest rejected valid image dimensions: %v", err)
		}
		if len(files) != 1 {
			t.Fatalf("expected 1 valid uploaded file, got %d", len(files))
		}
	})
}

func TestExtractFromRequest_UsesOriginalFilenameAndKeepsReaderAtStart(t *testing.T) {
	params := ExtractParams{
		AllowedMimeTypes: []string{"image/png"},
		MaxSizePerFile:   1024 * 1024,
		MaxFiles:         new(1),
	}

	req := newMultipartRequest(t, map[string]struct {
		filename string
		content  []byte
		mimeType string
	}{
		"file": {filename: "example.png", content: pngBytes(), mimeType: "image/png"},
	})

	files, err := ExtractFromRequest(req, params)
	if err != nil {
		t.Fatalf("ExtractFromRequest returned error: %v", err)
	}
	if len(files) != 1 {
		t.Fatalf("expected 1 file, got %d", len(files))
	}

	buf := make([]byte, 8)
	n, err := files[0].File.Read(buf)
	if err != nil && err.Error() != "EOF" {
		t.Fatalf("unexpected read error: %v", err)
	}
	if n <= 0 {
		t.Fatal("expected readable content from uploaded file")
	}

	if _, err := files[0].File.Seek(0, 0); err != nil {
		t.Fatalf("expected file to be seekable to start: %v", err)
	}
	if got := fmt.Sprintf("%x", buf[:n]); !strings.Contains(got, "89504e470d0a1a0a") {
		t.Fatalf("expected PNG signature bytes at start, got %s", got)
	}
}
