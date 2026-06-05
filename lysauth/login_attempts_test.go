package lysauth

import (
	"errors"
	"net/netip"
	"sync"
	"testing"
	"time"

	"github.com/loveyourstack/lys/lystype"
	"github.com/stretchr/testify/assert"
)

func TestAppLoginAttempts_Add(t *testing.T) {
	appLa := NewAppLoginAttempts(5)
	ip := netip.MustParseAddr("203.0.113.1")

	if got := appLa.Count(); got != 0 {
		t.Fatalf("initial Count mismatch: got %d, want 0", got)
	}

	for i := 1; i <= 5; i++ {
		err := appLa.Add(ip)
		if err != nil {
			t.Fatalf("Add(%d) returned unexpected error: %v", i, err)
		}

		got := appLa.ListByCreatedAt(true)
		if len(got) != 1 {
			t.Fatalf("ListByCreatedAt len mismatch after Add(%d): got %d, want 1", i, len(got))
		}
		if got[0].Ip != ip {
			t.Fatalf("IP mismatch after Add(%d): got %q, want %q", i, got[0].Ip, ip)
		}
		if got[0].NumAttempts != i {
			t.Fatalf("NumAttempts mismatch after Add(%d): got %d, want %d", i, got[0].NumAttempts, i)
		}
		if got[0].IsBlocked {
			t.Fatalf("IsBlocked mismatch after Add(%d): got true, want false", i)
		}
	}

	err := appLa.Add(ip)
	if !errors.Is(err, ErrMaxAttemptsExceeded) {
		t.Fatalf("Add beyond max mismatch: got %v, want %v", err, ErrMaxAttemptsExceeded)
	}

	got := appLa.ListByCreatedAt(true)
	if len(got) != 1 {
		t.Fatalf("ListByCreatedAt len mismatch after blocking: got %d, want 1", len(got))
	}
	if !got[0].IsBlocked {
		t.Fatalf("IsBlocked mismatch after blocking add: got false, want true")
	}
	if got[0].NumAttempts != 6 {
		t.Fatalf("NumAttempts mismatch after blocking add: got %d, want %d", got[0].NumAttempts, 6)
	}

	err = appLa.Add(ip)
	if !errors.Is(err, ErrBlocked) {
		t.Fatalf("Add on blocked IP mismatch: got %v, want %v", err, ErrBlocked)
	}
}

func TestAppLoginAttempts_Add_ConcurrentSingleIP(t *testing.T) {
	appLa := NewAppLoginAttempts(5)
	ip := netip.MustParseAddr("203.0.113.111")

	const workers = 256

	start := make(chan struct{})
	done := make(chan struct{})
	errCh := make(chan error, 1)

	// Observe the current attempts while goroutines are running and ensure
	// they never decrease.
	go func() {
		ticker := time.NewTicker(1 * time.Millisecond)
		defer ticker.Stop()

		prev := 0
		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				got := appLa.ListByCreatedAt(true)
				if len(got) == 0 {
					continue
				}

				curr := got[0].NumAttempts
				if curr < prev {
					select {
					case errCh <- errors.New("NumAttempts decreased during concurrent Add calls"):
					default:
					}
					return
				}
				prev = curr
			}
		}
	}()

	var wg sync.WaitGroup
	wg.Add(workers)
	for i := 0; i < workers; i++ {
		go func() {
			defer wg.Done()
			<-start
			_ = appLa.Add(ip)
		}()
	}

	close(start)
	wg.Wait()
	close(done)

	select {
	case err := <-errCh:
		t.Fatal(err)
	default:
	}

	got := appLa.ListByCreatedAt(true)
	if len(got) != 1 {
		t.Fatalf("ListByCreatedAt len mismatch after concurrent Add: got %d, want 1", len(got))
	}
	if got[0].Ip != ip {
		t.Fatalf("IP mismatch after concurrent Add: got %q, want %q", got[0].Ip, ip)
	}
	if !got[0].IsBlocked {
		t.Fatalf("IsBlocked mismatch after concurrent Add: got false, want true")
	}
	if got[0].NumAttempts != 6 {
		t.Fatalf("NumAttempts mismatch after concurrent Add: got %d, want %d", got[0].NumAttempts, 6)
	}

	err := appLa.Add(ip)
	if !errors.Is(err, ErrBlocked) {
		t.Fatalf("Add on blocked IP mismatch after concurrent Add: got %v, want %v", err, ErrBlocked)
	}
}

func TestAppLoginAttempts_DeleteAndCount(t *testing.T) {
	appLa := NewAppLoginAttempts(5)

	found, err := appLa.DeleteByIp(netip.MustParseAddr("198.51.100.9"))
	assert.NoError(t, err)
	if found {
		t.Fatalf("Delete on missing IP mismatch: got true, want false")
	}

	if err := appLa.Add(netip.MustParseAddr("198.51.100.9")); err != nil {
		t.Fatalf("Add returned unexpected error: %v", err)
	}
	if err := appLa.Add(netip.MustParseAddr("203.0.113.9")); err != nil {
		t.Fatalf("Add returned unexpected error: %v", err)
	}

	if got := appLa.Count(); got != 2 {
		t.Fatalf("Count mismatch after adds: got %d, want 2", got)
	}

	found, err = appLa.DeleteByIp(netip.MustParseAddr("198.51.100.9"))
	assert.NoError(t, err)
	if !found {
		t.Fatalf("Delete existing IP mismatch: got false, want true")
	}

	if got := appLa.Count(); got != 1 {
		t.Fatalf("Count mismatch after delete: got %d, want 1", got)
	}
}

func TestAppLoginAttempts_ListByCreatedAt(t *testing.T) {
	appLa := NewAppLoginAttempts(5)

	if got := appLa.ListByCreatedAt(true); len(got) != 0 {
		t.Fatalf("ListByCreatedAt after initialization: got non-empty slice, want empty")
	}

	t1 := time.Now().Add(-3 * time.Hour)
	t2 := time.Now().Add(-2 * time.Hour)
	t3 := time.Now().Add(-1 * time.Hour)

	appLa.mu.Lock()
	appLa.all[netip.MustParseAddr("203.0.113.10")] = LoginAttempt{CreatedAt: lystype.Datetime(t2), Ip: netip.MustParseAddr("203.0.113.10"), NumAttempts: 1}
	appLa.all[netip.MustParseAddr("203.0.113.20")] = LoginAttempt{CreatedAt: lystype.Datetime(t1), Ip: netip.MustParseAddr("203.0.113.20"), NumAttempts: 2}
	appLa.all[netip.MustParseAddr("203.0.113.30")] = LoginAttempt{CreatedAt: lystype.Datetime(t3), Ip: netip.MustParseAddr("203.0.113.30"), NumAttempts: 3}
	appLa.mu.Unlock()

	asc := appLa.ListByCreatedAt(true)
	if len(asc) != 3 {
		t.Fatalf("ListByCreatedAt asc len mismatch: got %d, want 3", len(asc))
	}
	if time.Time(asc[0].CreatedAt) != t1 || time.Time(asc[1].CreatedAt) != t2 || time.Time(asc[2].CreatedAt) != t3 {
		t.Fatalf("ListByCreatedAt asc order mismatch")
	}

	desc := appLa.ListByCreatedAt(false)
	if len(desc) != 3 {
		t.Fatalf("ListByCreatedAt desc len mismatch: got %d, want 3", len(desc))
	}
	if time.Time(desc[0].CreatedAt) != t3 || time.Time(desc[1].CreatedAt) != t2 || time.Time(desc[2].CreatedAt) != t1 {
		t.Fatalf("ListByCreatedAt desc order mismatch")
	}
}

func TestAppLoginAttempts_InvalidIPRejected(t *testing.T) {
	appLa := NewAppLoginAttempts(5)

	err := appLa.Add(netip.Addr{})
	if err == nil {
		t.Fatalf("Add(zero-value IP) expected error, got nil")
	}
	if !assert.ErrorContains(t, err, "empty IP") {
		t.Fatalf("Add(zero-value IP) error mismatch: got %q, want contains %q", err.Error(), "empty IP")
	}

	_, err = appLa.DeleteByIp(netip.Addr{})
	if err == nil {
		t.Fatalf("DeleteByIp(zero-value IP) expected error, got nil")
	}
	if !assert.ErrorContains(t, err, "empty IP") {
		t.Fatalf("DeleteByIp(zero-value IP) error mismatch: got %q, want contains %q", err.Error(), "empty IP")
	}

	_, err = appLa.IsBlocked(netip.Addr{})
	if err == nil {
		t.Fatalf("IsBlocked(zero-value IP) expected error, got nil")
	}
	if !assert.ErrorContains(t, err, "empty IP") {
		t.Fatalf("IsBlocked(zero-value IP) error mismatch: got %q, want contains %q", err.Error(), "empty IP")
	}

	err = appLa.Block(netip.Addr{})
	if err == nil {
		t.Fatalf("Block(zero-value IP) expected error, got nil")
	}
	if !assert.ErrorContains(t, err, "empty IP") {
		t.Fatalf("Block(zero-value IP) error mismatch: got %q, want contains %q", err.Error(), "empty IP")
	}
}

func TestAppLoginAttempts_BlockAndIsBlocked(t *testing.T) {
	appLa := NewAppLoginAttempts(5)
	ip := netip.MustParseAddr("203.0.113.200")

	blocked, err := appLa.IsBlocked(ip)
	if err != nil {
		t.Fatalf("IsBlocked(non-existing) returned unexpected error: %v", err)
	}
	if blocked {
		t.Fatalf("IsBlocked(non-existing) mismatch: got true, want false")
	}

	err = appLa.Block(ip)
	if err != nil {
		t.Fatalf("Block returned unexpected error: %v", err)
	}

	blocked, err = appLa.IsBlocked(ip)
	if err != nil {
		t.Fatalf("IsBlocked(after block) returned unexpected error: %v", err)
	}
	if !blocked {
		t.Fatalf("IsBlocked(after block) mismatch: got false, want true")
	}

	if got := appLa.Count(); got != 1 {
		t.Fatalf("Count mismatch after Block(non-existing): got %d, want 1", got)
	}
}

func TestAppLoginAttempts_LoadAndAll(t *testing.T) {
	t1 := time.Now().Add(-2 * time.Hour)
	t2 := time.Now().Add(-1 * time.Hour)

	seed := []LoginAttempt{
		{CreatedAt: lystype.Datetime(t1), Ip: netip.MustParseAddr("198.51.100.101"), NumAttempts: 2, IsBlocked: false},
		{CreatedAt: lystype.Datetime(t2), Ip: netip.MustParseAddr("198.51.100.102"), NumAttempts: 6, IsBlocked: true},
	}

	appLa := NewAppLoginAttempts(5)
	err := appLa.Load(seed)
	if err != nil {
		t.Fatalf("Load(valid seed) returned unexpected error: %v", err)
	}

	all := appLa.All()
	if len(all) != len(seed) {
		t.Fatalf("All len mismatch after Load: got %d, want %d", len(all), len(seed))
	}

	gotByIP := map[netip.Addr]LoginAttempt{}
	for _, la := range all {
		gotByIP[la.Ip] = la
	}
	for _, want := range seed {
		got, ok := gotByIP[want.Ip]
		if !ok {
			t.Fatalf("All missing IP: %s", want.Ip)
		}
		if got.NumAttempts != want.NumAttempts {
			t.Fatalf("NumAttempts mismatch for %s: got %d, want %d", want.Ip, got.NumAttempts, want.NumAttempts)
		}
		if got.IsBlocked != want.IsBlocked {
			t.Fatalf("IsBlocked mismatch for %s: got %v, want %v", want.Ip, got.IsBlocked, want.IsBlocked)
		}
	}

	err = appLa.Load(seed)
	if err == nil {
		t.Fatalf("Load(non-empty app) expected error, got nil")
	}
	if !assert.ErrorContains(t, err, "expected empty AppLoginAttempts instance") {
		t.Fatalf("Load(non-empty app) error mismatch: got %q, want contains %q", err.Error(), "expected empty AppLoginAttempts instance")
	}

	fresh := NewAppLoginAttempts(5)
	err = fresh.Load([]LoginAttempt{})
	if err == nil {
		t.Fatalf("Load(empty seed) expected error, got nil")
	}
	if !assert.ErrorContains(t, err, "loginAttempts has len 0") {
		t.Fatalf("Load(empty seed) error mismatch: got %q, want contains %q", err.Error(), "loginAttempts has len 0")
	}
}
