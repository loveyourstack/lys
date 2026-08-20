package lys

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/loveyourstack/lys/lyserr"
	"github.com/loveyourstack/lys/lysformfile"
)

// Upload is a handler that extracts file uploads from a multipart form request and streams them to the specified destination path.
func Upload(extractParams lysformfile.ExtractParams, destPath string, logger *slog.Logger) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// must be POST
		if r.Method != http.MethodPost {
			HandleUserError(lyserr.User{Message: "method must be POST"}, w)
			return
		}

		// get user id from ctx
		userId := GetUserIdFromCtx(ctx)
		if userId == 0 {
			HandleUserError(lyserr.User{Message: "user not authenticated", StatusCode: http.StatusForbidden}, w)
			return
		}

		// ensure request body is not nil
		if r.Body == nil {
			HandleUserError(lyserr.User{Message: "request body is empty"}, w)
			return
		}

		// extract uploaded files from multipart form request
		uploadFiles, err := lysformfile.ExtractFromRequest(r, extractParams)
		if err != nil {
			HandleError(ctx, fmt.Errorf("Upload: lysformfile.ExtractFromRequest failed: %w", err), logger, w)
			return
		}

		// ensure all opened files are closed at the end of processing
		defer func() {
			for _, uploadFile := range uploadFiles {
				uploadFile.File.Close()
			}
		}()

		// stream files to destination path
		streamResp, err := lysformfile.StreamToDisk(uploadFiles, destPath, userId, logger)
		if err != nil {
			HandleError(ctx, fmt.Errorf("Upload: lysformfile.StreamToDisk failed: %w", err), logger, w)
			return
		}

		// success
		resp := StdResponse{
			Status: ReqSucceeded,
			Data:   streamResp,
		}
		JsonResponse(resp, http.StatusCreated, w)
	}
}
