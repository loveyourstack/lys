package lysauth

import (
	"fmt"
	"net/http"
	"net/netip"
	"slices"
	"sync"
	"time"

	"github.com/loveyourstack/lys/lyserr"
	"github.com/loveyourstack/lys/lystype"
)

var (
	ErrBlocked             = lyserr.User{Message: "IP is blocked", StatusCode: http.StatusForbidden}
	ErrMaxAttemptsExceeded = lyserr.User{Message: "Max login attempts exceeded", StatusCode: http.StatusForbidden}
)

// LoginAttempt is an unsuccessful login attempt from a given IP.
type LoginAttempt struct {
	CreatedAt   lystype.Datetime `json:"created_at"`
	Ip          netip.Addr       `json:"ip"`
	IsBlocked   bool             `json:"is_blocked"`
	NumAttempts int              `json:"num_attempts"`
}

// AppLoginAttempts contains login attempts and methods to manage them.
type AppLoginAttempts struct {
	all         map[netip.Addr]LoginAttempt // map of IP to login attempt
	maxAttempts int
	mu          sync.RWMutex
}

// NewAppLoginAttempts creates a new AppLoginAttempts instance.
func NewAppLoginAttempts(maxAttempts int) *AppLoginAttempts {
	return &AppLoginAttempts{
		all:         make(map[netip.Addr]LoginAttempt),
		maxAttempts: maxAttempts,
	}
}

// Add tries to increment the number of attempts of the supplied IP if it already exists, or adds it if it doesn't exist.
func (appLa *AppLoginAttempts) Add(ip netip.Addr) (err error) {

	if !ip.IsValid() {
		return fmt.Errorf("empty IP")
	}

	appLa.mu.Lock()
	defer appLa.mu.Unlock()

	// try to get login attempt for this ip
	la, exists := appLa.all[ip]

	// doesn't exist: add new login attempt for this IP
	if !exists {
		appLa.all[ip] = LoginAttempt{
			CreatedAt:   lystype.Datetime(time.Now()),
			Ip:          ip,
			NumAttempts: 1,
		}
		return nil
	}

	// ip exists

	// check if already blocked
	if la.IsBlocked {
		return ErrBlocked
	}

	// increment num attempts and check against max
	la.NumAttempts++
	if la.NumAttempts > appLa.maxAttempts {

		// block IP and update login attempt
		la.IsBlocked = true
		appLa.all[ip] = la

		return ErrMaxAttemptsExceeded
	}

	// update login attempt
	appLa.all[ip] = la

	return nil
}

// All returns all login attempts.
func (appLa *AppLoginAttempts) All() (loginAttempts []LoginAttempt) {
	appLa.mu.RLock()
	defer appLa.mu.RUnlock()

	for _, la := range appLa.all {
		loginAttempts = append(loginAttempts, la)
	}
	return loginAttempts
}

func (appLa *AppLoginAttempts) Block(ip netip.Addr) error {

	if !ip.IsValid() {
		return fmt.Errorf("empty IP")
	}

	appLa.mu.Lock()
	defer appLa.mu.Unlock()

	la, exists := appLa.all[ip]
	if !exists {
		la = LoginAttempt{
			CreatedAt: lystype.Datetime(time.Now()),
			Ip:        ip,
		}
	}

	la.IsBlocked = true
	appLa.all[ip] = la

	return nil
}

// Count returns the number of login attempts.
func (appLa *AppLoginAttempts) Count() int {
	appLa.mu.RLock()
	defer appLa.mu.RUnlock()
	return len(appLa.all)
}

// DeleteByIp removes the login attempt for a given IP.
func (appLa *AppLoginAttempts) DeleteByIp(ip netip.Addr) (found bool, err error) {

	if !ip.IsValid() {
		return false, fmt.Errorf("empty IP")
	}

	appLa.mu.Lock()
	defer appLa.mu.Unlock()

	_, exists := appLa.all[ip]
	if !exists {
		return false, nil
	}

	delete(appLa.all, ip)

	return true, nil
}

func (appLa *AppLoginAttempts) IsBlocked(ip netip.Addr) (bool, error) {

	if !ip.IsValid() {
		return false, fmt.Errorf("empty IP")
	}

	appLa.mu.RLock()
	defer appLa.mu.RUnlock()

	la, exists := appLa.all[ip]
	if !exists {
		return false, nil
	}

	return la.IsBlocked, nil
}

// ListByCreatedAt returns all login attempts sorted by CreatedAt.
func (appLa *AppLoginAttempts) ListByCreatedAt(asc bool) (sortedLoginAttempts []LoginAttempt) {

	appLa.mu.RLock()
	defer appLa.mu.RUnlock()

	if len(appLa.all) == 0 {
		return []LoginAttempt{} // for JSON encoding to return [] instead of null
	}

	for _, la := range appLa.all {
		sortedLoginAttempts = append(sortedLoginAttempts, la)
	}

	slices.SortFunc(sortedLoginAttempts, func(a, b LoginAttempt) int {
		if time.Time(a.CreatedAt).Equal(time.Time(b.CreatedAt)) {
			return 0
		}
		if time.Time(a.CreatedAt).Before(time.Time(b.CreatedAt)) {
			return -1
		}
		return 1
	})
	if asc {
		return sortedLoginAttempts
	}

	slices.Reverse(sortedLoginAttempts)
	return sortedLoginAttempts
}

// Load loads the supplied login attempts into a freshly created AppLoginAttempts instance.
func (appLa *AppLoginAttempts) Load(loginAttempts []LoginAttempt) (err error) {
	if len(loginAttempts) == 0 {
		return fmt.Errorf("loginAttempts has len 0")
	}

	appLa.mu.Lock()
	defer appLa.mu.Unlock()

	if appLa.all == nil {
		return fmt.Errorf("AppLoginAttempts is not initialized")
	}
	if len(appLa.all) > 0 {
		return fmt.Errorf("expected empty AppLoginAttempts instance, but instance has %d login attempts", len(appLa.all))
	}

	for _, la := range loginAttempts {
		appLa.all[la.Ip] = la
	}

	return nil
}
