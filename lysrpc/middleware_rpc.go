package lysrpc

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/netip"
	"strings"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/loveyourstack/lys/lysauth"
	"github.com/loveyourstack/lys/lyserr"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
)

func getRemoteHostIp(ctx context.Context, useRealIpHeader bool) (netip.Addr, error) {

	// get remote addr from peer context
	remoteAddr, err := getRemoteAddr(ctx)
	if err != nil {
		return netip.Addr{}, err
	}

	// get x-real-ip header from metadata if needed
	xrealIpVal := ""
	if useRealIpHeader {
		xrealIpVal, err = getXRealIp(ctx)
		if err != nil {
			return netip.Addr{}, err
		}
	}

	// resolve remote ip
	remoteHostIp, err := lysauth.GetIpAddr(remoteAddr, xrealIpVal, useRealIpHeader)
	if err != nil {
		return netip.Addr{}, fmt.Errorf("lysauth.GetIpAddr failed: %w", err)
	}

	return remoteHostIp, nil
}

func getRemoteAddr(ctx context.Context) (remoteAddr string, err error) {

	p, ok := peer.FromContext(ctx)
	if ok {
		remoteAddr = p.Addr.String()
	} else {
		return "", fmt.Errorf("failed to get remote address from peer context")
	}

	if remoteAddr == "" {
		return "", fmt.Errorf("remote address is empty")
	}

	return remoteAddr, nil
}

func getXRealIp(ctx context.Context) (xrealIpVal string, err error) {

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", fmt.Errorf("failed to get metadata from context")
	}

	xrealIp := md.Get("x-real-ip")
	if len(xrealIp) == 0 {
		return "", fmt.Errorf("x-real-ip header not found")
	}

	return strings.Join(xrealIp, ","), nil
}

// MapError is gRPC middleware that maps database and other errors to appropriate gRPC status codes.
func MapError(logger *slog.Logger) func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {

		// call the handler and return if no error occurred
		resp, err := handler(ctx, req)
		if err == nil {
			return resp, nil
		}

		// an error occurred

		code := status.Code(err)
		errMsg := err.Error()

		// see if err can be unwrapped to a pgx PgError
		if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok {

			switch pgErr.Code {

			// define text and status for errors that can be attributed to bad requests or conflicts

			case pgerrcode.CheckViolation:
				code = codes.InvalidArgument
				errMsg = fmt.Sprintf("check constraint violation: %s", pgErr.ConstraintName)

			case pgerrcode.ExclusionViolation:
				code = codes.InvalidArgument
				errMsg = fmt.Sprintf("exclusion constraint violation: %s", pgErr.Detail)

			case pgerrcode.ForeignKeyViolation:
				code = codes.InvalidArgument
				errMsg = fmt.Sprintf("foreign key violation: %s", pgErr.Detail)

			case pgerrcode.InvalidTextRepresentation: // e.g. enum value does not exist
				code = codes.InvalidArgument
				errMsg = fmt.Sprintf("invalid text: %s", pgErr.Message)

			case pgerrcode.NotNullViolation:
				code = codes.InvalidArgument
				errMsg = fmt.Sprintf("missing required field: %s", pgErr.ColumnName)

			case pgerrcode.StringDataRightTruncationDataException: // e.g. text too long for varchar(i) column
				code = codes.InvalidArgument
				errMsg = pgErr.Message

			case pgerrcode.UndefinedObject: // e.g. enum type does not exist
				code = codes.InvalidArgument
				errMsg = fmt.Sprintf("undefined object: %s", pgErr.Message)

			case pgerrcode.UniqueViolation:
				code = codes.AlreadyExists
				errStr := pgErr.ConstraintName // single column unique key
				if pgErr.Detail != "" {
					errStr = pgErr.Detail // is better, but only filled on multiple column unique keys
				}
				errMsg = fmt.Sprintf("unique constraint violation: %s", errStr)

			default:
				// other pgx error, no special treatment
			}
		}

		// user-facing errors (lysauth.Err*, bad credentials, etc): map HTTP status to gRPC code
		if userErr, ok := errors.AsType[lyserr.User](err); ok {
			code = userErrCode(userErr.StatusCode)
			errMsg = userErr.Message
		}

		// distinguish between error types
		switch code {

		// Expected client errors: return, but don't log
		case codes.Aborted,
			codes.AlreadyExists,
			codes.Canceled,
			codes.DeadlineExceeded,
			codes.FailedPrecondition,
			codes.InvalidArgument,
			codes.NotFound,
			codes.PermissionDenied,
			codes.ResourceExhausted,
			codes.Unauthenticated,
			codes.Unavailable:
			return resp, status.Error(code, errMsg)

		default:
			// Unexpected server errors: log and return generic error to the client.
			logger.Error("gRPC method failed", "method", info.FullMethod, "code", code.String(), "error", err)
			return resp, status.Error(codes.Internal, "internal server error")
		}
	}
}

// RejectBlockedIp is middleware that rejects requests from blocked IPs based on login attempts.
func RejectBlockedIp(loginAttempts *lysauth.AppLoginAttempts, logger *slog.Logger) func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {

		remoteHostIp, err := lysauth.ClientIpFromCtx(ctx)
		if err != nil {
			return nil, fmt.Errorf("lysauth.ClientIpFromCtx failed: %w", err)
		}

		isBlocked, err := loginAttempts.IsBlocked(remoteHostIp)
		if err != nil {
			return nil, fmt.Errorf("loginAttempts.IsBlocked failed for IP %s: %w", remoteHostIp, err)
		}

		// reject if blocked
		if isBlocked {
			logger.Debug(fmt.Sprintf("RejectBlockedIp: blocked request from IP %s", remoteHostIp))
			return nil, status.Error(codes.PermissionDenied, "blocked IP")
		}

		return handler(ctx, req)
	}
}

// SetClientIp is gRPC middleware that resolves the client IP and stores it in context.
func SetClientIp(useRealIpHeader bool, logger *slog.Logger) func(ctx context.Context, req any,
	info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {

		remoteHostIP, err := getRemoteHostIp(ctx, useRealIpHeader)
		if err != nil {
			logger.Error("SetClientIp: GetRemoteHostIp failed", "method", info.FullMethod, "error", err)
			return nil, status.Error(codes.Internal, "internal server error")
		}
		ctx = lysauth.WithClientIp(ctx, remoteHostIP)

		return handler(ctx, req)
	}
}

// userErrCode maps HTTP user error status codes to gRPC codes.
func userErrCode(httpStatus int) codes.Code {
	switch httpStatus {
	case http.StatusConflict:
		return codes.AlreadyExists
	case http.StatusForbidden:
		return codes.PermissionDenied
	case http.StatusNotFound:
		return codes.NotFound
	case http.StatusTooManyRequests:
		return codes.ResourceExhausted
	case http.StatusUnauthorized:
		return codes.Unauthenticated
	default: // 400 and any other user error
		return codes.InvalidArgument
	}
}
