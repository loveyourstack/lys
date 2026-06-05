package lysauth

import (
	"net/netip"
	"testing"
	"time"

	"golang.org/x/time/rate"
)

func TestAppRateLimits_Allow(t *testing.T) {
	appRL := NewAppRateLimits(1, 2, time.Minute)

	key1 := netip.MustParseAddr("198.51.100.1").String()
	if !appRL.Allow(key1) {
		t.Fatal("Allow(key1) first call = false, want true")
	}
	if !appRL.Allow(key1) {
		t.Fatal("Allow(key1) second call = false, want true")
	}
	if appRL.Allow(key1) {
		t.Fatal("Allow(key1) third call = true, want false")
	}

	key2 := netip.MustParseAddr("198.51.100.2").String()
	if !appRL.Allow(key2) {
		t.Fatal("Allow(key2) first call = false, want true")
	}
	if !appRL.Allow(key2) {
		t.Fatal("Allow(key2) second call = false, want true")
	}
	if appRL.Allow(key2) {
		t.Fatal("Allow(key2) third call = true, want false")
	}

	if appRL.Allow(key1) {
		t.Fatal("Allow(key1) fourth call = true, want false")
	}
}

func TestAppRateLimits_Allow_CanonicalIPv6Key(t *testing.T) {
	appRL := NewAppRateLimits(1, 2, time.Minute)

	canonical := netip.MustParseAddr("2001:db8::1").String()
	expanded := netip.MustParseAddr("2001:0db8:0:0:0:0:0:1").String()

	if canonical != expanded {
		t.Fatalf("expected canonical and expanded to normalize to same key: %q vs %q", canonical, expanded)
	}

	if !appRL.Allow(canonical) {
		t.Fatal("Allow(canonical) first call = false, want true")
	}
	if !appRL.Allow(expanded) {
		t.Fatal("Allow(expanded) second call = false, want true")
	}
	if appRL.Allow(canonical) {
		t.Fatal("Allow(canonical) third call = true, want false")
	}
}

func TestAppRateLimits_CleanupExpired(t *testing.T) {
	appRL := NewAppRateLimits(1, 1, time.Minute)
	now := time.Now()

	appRL.mu.Lock()
	appRL.buckets["old"] = &clientBucket{
		limiter:  rate.NewLimiter(1, 1),
		lastSeen: now.Add(-2 * time.Minute),
	}
	appRL.buckets["recent"] = &clientBucket{
		limiter:  rate.NewLimiter(1, 1),
		lastSeen: now.Add(-30 * time.Second),
	}
	appRL.mu.Unlock()

	appRL.CleanupExpired()

	appRL.mu.Lock()
	defer appRL.mu.Unlock()

	if _, ok := appRL.buckets["old"]; ok {
		t.Fatal("CleanupExpired kept stale bucket, want it deleted")
	}
	if _, ok := appRL.buckets["recent"]; !ok {
		t.Fatal("CleanupExpired deleted recent bucket, want it kept")
	}
}
