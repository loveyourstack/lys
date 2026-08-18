package lysformfile

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"math"
	"mime"
	"mime/multipart"
	"net/http"
	"slices"
	"strings"

	"github.com/loveyourstack/lys/lyserr"
	"github.com/loveyourstack/lys/lysset"
	"github.com/loveyourstack/lys/lysstring"
)

// ExtractParams defines the parameters for extracting uploaded files from a multipart form request.
type ExtractParams struct {
	AllowedMimeTypes []string // list of allowed MIME types for uploaded files, e.g. ["image/jpeg", "image/png"].
	MaxSizePerFile   int64    // maximum size in bytes for each individual file, e.g. 10 << 20 for 10MB.

	ImgMaxHeightPx *int // optional: if an image is uploaded, its maximum height in pixels. If omitted, no maximum height check is performed.
	ImgMinHeightPx *int // optional: if an image is uploaded, its minimum height in pixels. If omitted, no minimum height check is performed.
	ImgMaxWidthPx  *int // optional: if an image is uploaded, its maximum width in pixels. If omitted, no maximum width check is performed.
	ImgMinWidthPx  *int // optional: if an image is uploaded, its minimum width in pixels. If omitted, no minimum width check is performed.
	MaxFiles       *int // optional: maximum number of files allowed to be uploaded in a single request. If omitted, defaults to 1.

	calcMaxFiles               int                // from MaxFiles, defaulting to 1 if nil
	maxBodyBytes               int64              // calculated max body size for the request, based on calcMaxFiles and maxSizePerFileWithOverhead
	maxSizePerFileWithOverhead int64              // MaxSizePerFile + overhead allowance for multipart form data
	normMimeTypes              lysset.Set[string] // validated and normalized AllowedMimeTypes
}

func (params *ExtractParams) validateAndProcess() (err error) {

	// AllowedMimeTypes & normMimeTypes
	if len(params.AllowedMimeTypes) == 0 {
		return lyserr.User{Message: "AllowedMimeTypes must not be empty"}
	}
	if slices.Contains(params.AllowedMimeTypes, "") {
		return lyserr.User{Message: "AllowedMimeTypes contains an empty MIME type"}
	}
	for _, mimeType := range params.AllowedMimeTypes {
		lcName, _, err := mime.ParseMediaType(mimeType)
		if err != nil {
			return lyserr.User{Message: fmt.Sprintf("invalid MIME type in AllowedMimeTypes: %s", mimeType)}
		}
		params.normMimeTypes.Add(lcName)
	}

	// ImgMinHeightPx, ImgMaxHeightPx, ImgMinWidthPx, ImgMaxWidthPx
	if params.ImgMinHeightPx != nil && *params.ImgMinHeightPx < 0 {
		return lyserr.User{Message: "ImgMinHeightPx must be >= 0"}
	}
	if params.ImgMaxHeightPx != nil && *params.ImgMaxHeightPx < 0 {
		return lyserr.User{Message: "ImgMaxHeightPx must be >= 0"}
	}
	if params.ImgMinWidthPx != nil && *params.ImgMinWidthPx < 0 {
		return lyserr.User{Message: "ImgMinWidthPx must be >= 0"}
	}
	if params.ImgMaxWidthPx != nil && *params.ImgMaxWidthPx < 0 {
		return lyserr.User{Message: "ImgMaxWidthPx must be >= 0"}
	}
	if params.ImgMinHeightPx != nil && params.ImgMaxHeightPx != nil && *params.ImgMinHeightPx > *params.ImgMaxHeightPx {
		return lyserr.User{Message: "ImgMinHeightPx must be <= ImgMaxHeightPx"}
	}
	if params.ImgMinWidthPx != nil && params.ImgMaxWidthPx != nil && *params.ImgMinWidthPx > *params.ImgMaxWidthPx {
		return lyserr.User{Message: "ImgMinWidthPx must be <= ImgMaxWidthPx"}
	}

	// MaxFiles & calcMaxFiles
	params.calcMaxFiles = 1
	if params.MaxFiles != nil {
		if *params.MaxFiles <= 0 {
			return lyserr.User{Message: "MaxFiles must be greater than 0"}
		} else {
			params.calcMaxFiles = *params.MaxFiles
		}
	}

	// MaxSizePerFile + maxSizePerFileWithOverhead
	if params.MaxSizePerFile <= 0 {
		return lyserr.User{Message: "MaxSizePerFile must be greater than 0"}
	}
	params.maxSizePerFileWithOverhead = params.MaxSizePerFile + 1024 // add 1KB overhead for multipart form data

	// maxBodyBytes: ensure it does not overflow int64.
	if int64(params.calcMaxFiles) > math.MaxInt64/params.maxSizePerFileWithOverhead {
		return fmt.Errorf("calculated maxBodyBytes would overflow int64")
	}
	params.maxBodyBytes = int64(params.calcMaxFiles) * params.maxSizePerFileWithOverhead

	return nil
}

// UploadFile represents an uploaded file with its associated metadata.
type UploadFile struct {
	File       multipart.File        // file is opened and ready for reading. It must be closed by caller.
	FileHeader *multipart.FileHeader // includes the original file name and size.
	MimeType   string
}

/*
ExtractFromRequest extracts uploaded files from a multipart form request and validates them against the provided ExtractParams.

The returned files must be closed after processing, e.g. using defer:

	defer func() {
		for _, uploadFile := range uploadFiles {
			uploadFile.File.Close()
		}
	}()
*/
func ExtractFromRequest(r *http.Request, params ExtractParams) (uploadFiles []UploadFile, err error) {

	// content-type header must be multipart/form-data
	contentType := r.Header.Get("Content-Type")
	if contentType == "" {
		return nil, lyserr.User{Message: "Content-Type header is missing"}
	}
	mediaType, _, err := mime.ParseMediaType(contentType)
	if err != nil {
		return nil, fmt.Errorf("invalid Content-Type header: %w", err)
	}
	if mediaType != "multipart/form-data" {
		return nil, lyserr.User{Message: fmt.Sprintf("Content-Type must be multipart/form-data, got %s", mediaType)}
	}

	// validate and process params
	if err := params.validateAndProcess(); err != nil {
		return nil, fmt.Errorf("validateAndProcess failed: %w", err)
	}

	// cap body size before multipart parsing
	r.Body = http.MaxBytesReader(nil, r.Body, params.maxBodyBytes)

	// parse multipart form
	if err := r.ParseMultipartForm(params.maxBodyBytes); err != nil {
		if _, ok := errors.AsType[*http.MaxBytesError](err); ok {
			return nil, lyserr.User{Message: fmt.Sprintf("request body exceeds the maximum allowed size of %s", lysstring.FormatBytes(params.maxBodyBytes))}
		}
		return nil, fmt.Errorf("r.ParseMultipartForm failed: %w", err)
	}
	defer r.MultipartForm.RemoveAll()

	fileMap := r.MultipartForm.File
	if len(fileMap) == 0 {
		return nil, lyserr.User{Message: "no files uploaded"}
	}

	// function to close all opened files in case of an error
	closeOpened := func() {
		for _, uploadFile := range uploadFiles {
			uploadFile.File.Close()
		}
	}

	for _, fileHeaders := range fileMap {
		for _, fileHeader := range fileHeaders {

			// enforce max files limit
			if len(uploadFiles) >= params.calcMaxFiles {
				closeOpened()
				return nil, lyserr.User{Message: fmt.Sprintf("too many files uploaded: %d, maximum allowed is %d", len(uploadFiles), params.calcMaxFiles)}
			}

			// check file size against max
			if fileHeader.Size > params.MaxSizePerFile {
				closeOpened()
				return nil, lyserr.User{Message: fmt.Sprintf("file %s exceeds the maximum allowed size of %d bytes", fileHeader.Filename, params.MaxSizePerFile)}
			}

			// open file
			file, err := fileHeader.Open()
			if err != nil {
				closeOpened()
				return nil, fmt.Errorf("ExtractFromRequest: fileHeader.Open failed for file %s: %w", fileHeader.Filename, err)
			}

			// read first 512 bytes to detect MIME type
			buffer := make([]byte, 512)
			n, err := file.Read(buffer)
			if err != nil && err != io.EOF {
				file.Close()
				closeOpened()
				return nil, fmt.Errorf("ExtractFromRequest: file.Read failed for file %s: %w", fileHeader.Filename, err)
			}

			// detect MIME type and validate against allowed types
			mimeType := http.DetectContentType(buffer[:n])
			if !params.normMimeTypes.Contains(strings.ToLower(mimeType)) {
				file.Close()
				closeOpened()
				return nil, lyserr.User{Message: fmt.Sprintf("file %s has disallowed MIME type: %s", fileHeader.Filename, mimeType)}
			}

			// if the file is an image, validate its dimensions if any dimension constraints are set
			if strings.HasPrefix(mimeType, "image/") && (params.ImgMinHeightPx != nil || params.ImgMaxHeightPx != nil || params.ImgMinWidthPx != nil || params.ImgMaxWidthPx != nil) {
				imgCfg, _, err := image.DecodeConfig(io.MultiReader(bytes.NewReader(buffer[:n]), file))
				if err != nil {
					file.Close()
					closeOpened()
					return nil, fmt.Errorf("ExtractFromRequest: image.DecodeConfig failed for file %s: %w", fileHeader.Filename, err)
				}
				if err := validateImageDimensions(imgCfg, params); err != nil {
					file.Close()
					closeOpened()
					return nil, err
				}
			}

			// rewind file read pointer to the beginning for later processing
			if _, err := file.Seek(0, io.SeekStart); err != nil {
				file.Close()
				closeOpened()
				return nil, fmt.Errorf("ExtractFromRequest: file.Seek failed for file %s: %w", fileHeader.Filename, err)
			}

			// append to the list of valid uploaded files
			uploadFiles = append(uploadFiles, UploadFile{
				File:       file,
				FileHeader: fileHeader,
				MimeType:   mimeType,
			})

		} // next fileHeader

	} // next fileMap / fileHeaders

	return uploadFiles, nil
}

func validateImageDimensions(imgCfg image.Config, params ExtractParams) error {
	if params.ImgMinWidthPx != nil && imgCfg.Width < *params.ImgMinWidthPx {
		return lyserr.User{Message: fmt.Sprintf("image width %d px is less than the minimum allowed %d px", imgCfg.Width, *params.ImgMinWidthPx)}
	}
	if params.ImgMaxWidthPx != nil && imgCfg.Width > *params.ImgMaxWidthPx {
		return lyserr.User{Message: fmt.Sprintf("image width %d px exceeds the maximum allowed %d px", imgCfg.Width, *params.ImgMaxWidthPx)}
	}
	if params.ImgMinHeightPx != nil && imgCfg.Height < *params.ImgMinHeightPx {
		return lyserr.User{Message: fmt.Sprintf("image height %d px is less than the minimum allowed %d px", imgCfg.Height, *params.ImgMinHeightPx)}
	}
	if params.ImgMaxHeightPx != nil && imgCfg.Height > *params.ImgMaxHeightPx {
		return lyserr.User{Message: fmt.Sprintf("image height %d px exceeds the maximum allowed %d px", imgCfg.Height, *params.ImgMaxHeightPx)}
	}
	return nil
}
