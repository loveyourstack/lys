package lysformfile

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"log/slog"
	"mime"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// StreamFileResult represents the result of a single uploaded file streamed to disk.
type StreamFileResult struct {
	MimeType     string `json:"mime_type"`     // MIME type of the file
	OriginalName string `json:"original_name"` // original file name as uploaded by the user
	SizeBytes    int64  `json:"size_bytes"`    // size of the file in bytes
	StoredName   string `json:"stored_name"`   // name of the file as stored on the server
}

// StreamResponse represents the response returned after successfully streaming file(s) to disk.
type StreamResponse struct {
	FileResults []StreamFileResult `json:"file_results"`
}

func StreamToDisk(uploadFiles []UploadFile, destPath string, userId int64, logger *slog.Logger) (uploadResp StreamResponse, err error) {

	savedFilePaths := []string{}

	// function to delete all saved files in case of an error
	deletedSaved := func() {
		for _, savedFile := range savedFilePaths {
			if err = os.Remove(savedFile); err != nil {
				logger.Error("os.Remove failed", "file", savedFile, "err", err)
			}
		}
	}

	for _, uploadFile := range uploadFiles {

		file := uploadFile.File
		mimeType := uploadFile.MimeType
		fileHeader := uploadFile.FileHeader

		// generate random 4-byte hex string for unique file naming
		rnd := make([]byte, 4)
		if _, err := rand.Read(rnd); err != nil {
			deletedSaved()
			return StreamResponse{}, fmt.Errorf("rand.Read failed for file %s: %w", fileHeader.Filename, err)
		}

		// determine stored file extension based on MIME type and original file name
		ext := chooseStoredExtension(fileHeader, mimeType)

		// generate stored file name
		storedFileName := fmt.Sprintf("%s-u%d-%s%s", time.Now().Format("20060102"), userId, hex.EncodeToString(rnd), ext)

		// create destination file
		destFilePath := fmt.Sprintf("%s/%s", destPath, storedFileName)
		destFile, err := os.Create(destFilePath)
		if err != nil {
			deletedSaved()
			return StreamResponse{}, fmt.Errorf("os.Create failed for dest path %s: %w", destFilePath, err)
		}

		// stream uploaded file to destination file
		if _, err := io.Copy(destFile, file); err != nil {

			// partial file could remain: add to saved slice to trigger removal by deletedSaved()
			savedFilePaths = append(savedFilePaths, destFilePath)

			destFile.Close()

			deletedSaved()
			return StreamResponse{}, fmt.Errorf("io.Copy failed for file %s: %w", fileHeader.Filename, err)
		}

		// close destination file
		if err := destFile.Close(); err != nil {
			deletedSaved()
			return StreamResponse{}, fmt.Errorf("destFile.Close failed for file %s: %w", fileHeader.Filename, err)
		}

		savedFilePaths = append(savedFilePaths, destFilePath)

		// append file result to response
		fileResult := StreamFileResult{
			MimeType:     mimeType,
			OriginalName: fileHeader.Filename,
			SizeBytes:    fileHeader.Size,
			StoredName:   storedFileName,
		}
		uploadResp.FileResults = append(uploadResp.FileResults, fileResult)
	}

	return uploadResp, nil
}

// chooseStoredExtension determines the appropriate file extension for the stored file based on the detected MIME type and the original file name.
// It returns the chosen extension, including the leading dot (e.g., ".jpg").
// If no suitable extension can be determined, it returns an empty string.
func chooseStoredExtension(fileHeader *multipart.FileHeader, detectedMime string) string {

	// first, try to get extension from original file name
	if fileHeader != nil {
		ext := strings.TrimPrefix(filepath.Ext(fileHeader.Filename), ".")
		if ext != "" {
			if t := mime.TypeByExtension("." + ext); t != "" && strings.EqualFold(t, detectedMime) {
				return "." + strings.ToLower(ext)
			}
		}
	}

	// fallback: get first extension from detected MIME type
	exts, err := mime.ExtensionsByType(detectedMime)
	if err == nil && len(exts) > 0 {
		return strings.ToLower(exts[0])
	}
	return ""
}
