package lysauth

import (
	"net/http"
	"sync"
	"time"

	"github.com/loveyourstack/lys/lyserr"
	"golang.org/x/time/rate"
)

var (
	ErrTooManyRequests = lyserr.User{Message: "Too many requests", StatusCode: http.StatusTooManyRequests}
)

type AppRateLimits struct {
	mu      sync.Mutex
	buckets map[string]*clientBucket
	rps     rate.Limit
	burst   int
	ttl     time.Duration // time to live (how long to keep unused buckets before cleanup)
}

type clientBucket struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

func NewAppRateLimits(rps float64, burst int, ttl time.Duration) *AppRateLimits {
	return &AppRateLimits{
		buckets: make(map[string]*clientBucket),
		rps:     rate.Limit(rps),
		burst:   burst,
		ttl:     ttl,
	}
}

func (m *AppRateLimits) Allow(key string) bool {
	now := time.Now()

	m.mu.Lock()
	defer m.mu.Unlock()

	b, ok := m.buckets[key]
	if !ok {
		b = &clientBucket{
			limiter:  rate.NewLimiter(m.rps, m.burst),
			lastSeen: now,
		}
		m.buckets[key] = b
	}
	b.lastSeen = now
	return b.limiter.Allow()
}

func (m *AppRateLimits) CleanupExpired() {
	cutoff := time.Now().Add(-m.ttl)

	m.mu.Lock()
	defer m.mu.Unlock()

	for k, b := range m.buckets {
		if b.lastSeen.Before(cutoff) {
			delete(m.buckets, k)
		}
	}
}
