package lysrpc

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/netip"
	"strings"
	"testing"

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

var middlewareTestLogger = slog.New(slog.NewTextHandler(io.Discard, nil))

type middlewareTestAddr string

func (a middlewareTestAddr) Network() string { return "test" }
func (a middlewareTestAddr) String() string  { return string(a) }

func TestGetRemoteAddr(t *testing.T) {
	tests := []struct {
		name    string
		ctx     context.Context
		want    string
		wantErr string
	}{
		{
			name: "remote address is present",
			ctx:  peer.NewContext(context.Background(), &peer.Peer{Addr: &net.TCPAddr{IP: net.ParseIP("203.0.113.1"), Port: 8080}}),
			want: "203.0.113.1:8080",
		},
		{
			name:    "peer context is missing",
			ctx:     context.Background(),
			wantErr: "failed to get remote address from peer context",
		},
		{
			name:    "remote address is empty",
			ctx:     peer.NewContext(context.Background(), &peer.Peer{Addr: middlewareTestAddr("")}),
			wantErr: "remote address is empty",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := getRemoteAddr(tc.ctx)
			if tc.wantErr != "" {
				if err == nil || err.Error() != tc.wantErr {
					t.Fatalf("getRemoteAddr error = %v, want %q", err, tc.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("getRemoteAddr returned unexpected error: %v", err)
			}
			if got != tc.want {
				t.Fatalf("getRemoteAddr = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestGetXRealIp(t *testing.T) {
	tests := []struct {
		name    string
		ctx     context.Context
		want    string
		wantErr string
	}{
		{
			name: "metadata value is read",
			ctx:  metadata.NewIncomingContext(context.Background(), metadata.Pairs("x-real-ip", "203.0.113.2")),
			want: "203.0.113.2",
		},
		{
			name:    "metadata is missing",
			ctx:     context.Background(),
			wantErr: "failed to get metadata from context",
		},
		{
			name:    "header is missing",
			ctx:     metadata.NewIncomingContext(context.Background(), metadata.MD{}),
			wantErr: "x-real-ip header not found",
		},
		{
			name: "multiple values are joined",
			ctx:  metadata.NewIncomingContext(context.Background(), metadata.Pairs("x-real-ip", "203.0.113.2", "x-real-ip", "203.0.113.3")),
			want: "203.0.113.2,203.0.113.3",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := getXRealIp(tc.ctx)
			if tc.wantErr != "" {
				if err == nil || err.Error() != tc.wantErr {
					t.Fatalf("getXRealIp error = %v, want %q", err, tc.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("getXRealIp returned unexpected error: %v", err)
			}
			if got != tc.want {
				t.Fatalf("getXRealIp = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestGetRemoteHostIp(t *testing.T) {
	tests := []struct {
		name            string
		ctx             context.Context
		useRealIpHeader bool
		want            netip.Addr
		wantErr         string
	}{
		{
			name: "remote address",
			ctx:  peer.NewContext(context.Background(), &peer.Peer{Addr: &net.TCPAddr{IP: net.ParseIP("203.0.113.4"), Port: 8080}}),
			want: netip.MustParseAddr("203.0.113.4"),
		},
		{
			name: "real IP header",
			ctx: metadata.NewIncomingContext(
				peer.NewContext(context.Background(), &peer.Peer{Addr: &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 8080}}),
				metadata.Pairs("x-real-ip", "203.0.113.5"),
			),
			useRealIpHeader: true,
			want:            netip.MustParseAddr("203.0.113.5"),
		},
		{
			name:    "peer context missing",
			ctx:     context.Background(),
			wantErr: "failed to get remote address from peer context",
		},
		{
			name:            "real IP metadata missing",
			ctx:             peer.NewContext(context.Background(), &peer.Peer{Addr: &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 8080}}),
			useRealIpHeader: true,
			wantErr:         "failed to get metadata from context",
		},
		{
			name: "real IP not trusted from non-loopback peer",
			ctx: metadata.NewIncomingContext(
				peer.NewContext(context.Background(), &peer.Peer{Addr: &net.TCPAddr{IP: net.ParseIP("203.0.113.6"), Port: 8080}}),
				metadata.Pairs("x-real-ip", "203.0.113.5"),
			),
			useRealIpHeader: true,
			wantErr:         "remote IP 203.0.113.6 is not loopback",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := getRemoteHostIp(tc.ctx, tc.useRealIpHeader)
			if tc.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
					t.Fatalf("getRemoteHostIp error = %v, want containing %q", err, tc.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("getRemoteHostIp returned unexpected error: %v", err)
			}
			if got != tc.want {
				t.Fatalf("getRemoteHostIp = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestMapError(t *testing.T) {
	tests := []struct {
		name        string
		err         error
		wantCode    codes.Code
		wantMessage string
	}{
		{
			name:        "success",
			wantMessage: "success",
		},
		{
			name:        "user error defaults to invalid argument",
			err:         lyserr.User{Message: "invalid input"},
			wantCode:    codes.InvalidArgument,
			wantMessage: "invalid input",
		},
		{
			name:        "forbidden user error",
			err:         lysauth.ErrBlocked,
			wantCode:    codes.PermissionDenied,
			wantMessage: lysauth.ErrBlocked.Message,
		},
		{
			name:        "rate limited user error",
			err:         lysauth.ErrTooManyRequests,
			wantCode:    codes.ResourceExhausted,
			wantMessage: lysauth.ErrTooManyRequests.Message,
		},
		{
			name: "unique constraint",
			err: &pgconn.PgError{
				Code:           pgerrcode.UniqueViolation,
				ConstraintName: "users_email_key",
			},
			wantCode:    codes.AlreadyExists,
			wantMessage: "unique constraint violation: users_email_key",
		},
		{
			name: "unique constraint detail preferred",
			err: &pgconn.PgError{
				Code:           pgerrcode.UniqueViolation,
				ConstraintName: "users_email_name_key",
				Detail:         "duplicate email and name",
			},
			wantCode:    codes.AlreadyExists,
			wantMessage: "unique constraint violation: duplicate email and name",
		},
		{
			name: "check constraint",
			err: &pgconn.PgError{
				Code:           pgerrcode.CheckViolation,
				ConstraintName: "age_positive",
			},
			wantCode:    codes.InvalidArgument,
			wantMessage: "check constraint violation: age_positive",
		},
		{
			name: "exclusion constraint",
			err: &pgconn.PgError{
				Code:   pgerrcode.ExclusionViolation,
				Detail: "overlapping reservation",
			},
			wantCode:    codes.InvalidArgument,
			wantMessage: "exclusion constraint violation: overlapping reservation",
		},
		{
			name: "wrapped foreign key violation",
			err: fmt.Errorf("db operation: %w", &pgconn.PgError{
				Code:   pgerrcode.ForeignKeyViolation,
				Detail: "missing parent",
			}),
			wantCode:    codes.InvalidArgument,
			wantMessage: "foreign key violation: missing parent",
		},
		{
			name: "invalid text representation",
			err: &pgconn.PgError{
				Code:    pgerrcode.InvalidTextRepresentation,
				Message: "invalid enum value",
			},
			wantCode:    codes.InvalidArgument,
			wantMessage: "invalid text: invalid enum value",
		},
		{
			name: "not-null violation",
			err: &pgconn.PgError{
				Code:       pgerrcode.NotNullViolation,
				ColumnName: "email",
			},
			wantCode:    codes.InvalidArgument,
			wantMessage: "missing required field: email",
		},
		{
			name: "string data truncation",
			err: &pgconn.PgError{
				Code:    pgerrcode.StringDataRightTruncationDataException,
				Message: "value too long",
			},
			wantCode:    codes.InvalidArgument,
			wantMessage: "value too long",
		},
		{
			name: "undefined object",
			err: &pgconn.PgError{
				Code:    pgerrcode.UndefinedObject,
				Message: "type missing",
			},
			wantCode:    codes.InvalidArgument,
			wantMessage: "undefined object: type missing",
		},
		{
			name:        "internal error is hidden",
			err:         errors.New("sensitive internal details"),
			wantCode:    codes.Internal,
			wantMessage: "internal server error",
		},
		{
			name:        "unavailable error is preserved",
			err:         status.Error(codes.Unavailable, "temporarily unavailable"),
			wantCode:    codes.Unavailable,
			wantMessage: "rpc error: code = Unavailable desc = temporarily unavailable",
		},
	}

	mapError := MapError(middlewareTestLogger)
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			handler := func(context.Context, any) (any, error) {
				if tc.err == nil {
					return "success", nil
				}
				return "response", tc.err
			}
			got, err := mapError(context.Background(), nil, &grpc.UnaryServerInfo{FullMethod: "/test.Service/Call"}, handler)
			if tc.err == nil {
				if err != nil {
					t.Fatalf("MapError returned unexpected error: %v", err)
				}
				if got != "success" {
					t.Fatalf("response = %v, want success", got)
				}
				return
			}
			if got != "response" {
				t.Fatalf("response = %v, want original handler response", got)
			}
			if status.Code(err) != tc.wantCode {
				t.Fatalf("status code = %s, want %s (error: %v)", status.Code(err), tc.wantCode, err)
			}
			if status.Convert(err).Message() != tc.wantMessage {
				t.Fatalf("status message = %q, want %q", status.Convert(err).Message(), tc.wantMessage)
			}
		})
	}
}

func TestRejectBlockedIp(t *testing.T) {
	blockedIP := netip.MustParseAddr("203.0.113.10")
	loginAttempts := lysauth.NewAppLoginAttempts(3)
	if err := loginAttempts.Block(blockedIP); err != nil {
		t.Fatalf("Block returned unexpected error: %v", err)
	}

	nextCalls := 0
	handler := RejectBlockedIp(loginAttempts, middlewareTestLogger)
	next := func(ctx context.Context, req any) (any, error) {
		nextCalls++
		if got, err := lysauth.ClientIpFromCtx(ctx); err != nil || got != netip.MustParseAddr("203.0.113.11") {
			t.Errorf("client IP in context = %q, error %v; want 203.0.113.11", got, err)
		}
		return req, nil
	}

	ctx := lysauth.WithClientIp(context.Background(), netip.MustParseAddr("203.0.113.11"))
	resp, err := handler(ctx, "request", nil, next)
	if err != nil {
		t.Fatalf("RejectBlockedIp returned unexpected error for unblocked IP: %v", err)
	}
	if resp != "request" || nextCalls != 1 {
		t.Fatalf("unblocked response/calls = %v/%d, want request/1", resp, nextCalls)
	}

	ctx = lysauth.WithClientIp(context.Background(), blockedIP)
	resp, err = handler(ctx, "request", nil, next)
	if status.Code(err) != codes.PermissionDenied || status.Convert(err).Message() != "blocked IP" {
		t.Fatalf("blocked error = %v, want PermissionDenied: blocked IP", err)
	}
	if resp != nil || nextCalls != 1 {
		t.Fatalf("blocked response/calls = %v/%d, want nil/1", resp, nextCalls)
	}

	resp, err = handler(context.Background(), "request", nil, next)
	if err == nil || !strings.Contains(err.Error(), "lysauth.ClientIpFromCtx failed") {
		t.Fatalf("missing client IP error = %v, want wrapped missing-IP error", err)
	}
	if resp != nil || nextCalls != 1 {
		t.Fatalf("missing-IP response/calls = %v/%d, want nil/1", resp, nextCalls)
	}
}

func TestSetClientIp(t *testing.T) {
	for _, tc := range []struct {
		name            string
		ctx             context.Context
		useRealIpHeader bool
		wantIP          netip.Addr
		wantCode        codes.Code
	}{
		{
			name:   "remote address",
			ctx:    peer.NewContext(context.Background(), &peer.Peer{Addr: &net.TCPAddr{IP: net.ParseIP("203.0.113.21"), Port: 8080}}),
			wantIP: netip.MustParseAddr("203.0.113.21"),
		},
		{
			name: "real IP header",
			ctx: metadata.NewIncomingContext(
				peer.NewContext(context.Background(), &peer.Peer{Addr: &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 8080}}),
				metadata.Pairs("x-real-ip", "203.0.113.22"),
			),
			useRealIpHeader: true,
			wantIP:          netip.MustParseAddr("203.0.113.22"),
		},
		{
			name:     "missing peer is internal error",
			ctx:      context.Background(),
			wantCode: codes.Internal,
		},
		{
			name:            "real IP header metadata missing",
			ctx:             peer.NewContext(context.Background(), &peer.Peer{Addr: &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 8080}}),
			useRealIpHeader: true,
			wantCode:        codes.Internal,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			nextCalls := 0
			handler := SetClientIp(tc.useRealIpHeader, middlewareTestLogger)
			resp, err := handler(tc.ctx, "request", &grpc.UnaryServerInfo{FullMethod: "/test.Service/Call"}, func(ctx context.Context, req any) (any, error) {
				nextCalls++
				got, err := lysauth.ClientIpFromCtx(ctx)
				if err != nil {
					t.Errorf("ClientIpFromCtx returned unexpected error: %v", err)
				} else if got != tc.wantIP {
					t.Errorf("client IP = %q, want %q", got, tc.wantIP)
				}
				return req, nil
			})
			if tc.wantCode != codes.OK {
				if status.Code(err) != tc.wantCode {
					t.Fatalf("status code = %s, want %s (error: %v)", status.Code(err), tc.wantCode, err)
				}
				if status.Convert(err).Message() != "internal server error" {
					t.Fatalf("status message = %q, want internal server error", status.Convert(err).Message())
				}
				if resp != nil || nextCalls != 0 {
					t.Fatalf("error response/calls = %v/%d, want nil/0", resp, nextCalls)
				}
				return
			}
			if err != nil {
				t.Fatalf("SetClientIp returned unexpected error: %v", err)
			}
			if resp != "request" || nextCalls != 1 {
				t.Fatalf("success response/calls = %v/%d, want request/1", resp, nextCalls)
			}
		})
	}
}

func TestUserErrCode(t *testing.T) {
	tests := []struct {
		httpStatus int
		want       codes.Code
	}{
		{httpStatus: http.StatusConflict, want: codes.AlreadyExists},
		{httpStatus: http.StatusForbidden, want: codes.PermissionDenied},
		{httpStatus: http.StatusNotFound, want: codes.NotFound},
		{httpStatus: http.StatusTooManyRequests, want: codes.ResourceExhausted},
		{httpStatus: http.StatusUnauthorized, want: codes.Unauthenticated},
		{httpStatus: http.StatusBadRequest, want: codes.InvalidArgument},
		{httpStatus: http.StatusInternalServerError, want: codes.InvalidArgument},
	}

	for _, tc := range tests {
		t.Run(http.StatusText(tc.httpStatus), func(t *testing.T) {
			if got := userErrCode(tc.httpStatus); got != tc.want {
				t.Fatalf("userErrCode(%d) = %s, want %s", tc.httpStatus, got, tc.want)
			}
		})
	}
}
