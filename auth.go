package main

import (
	"crypto/pbkdf2"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"
)

var pbkdf2Iter = 600_000

const (
	pbkdf2Prefix = "pbkdf2-sha256"
	minPassword  = 10
	maxPassword  = 128
)

var b64 = base64.RawStdEncoding

func hashPassword(pw string) (string, error) {
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}
	key, err := pbkdf2.Key(sha256.New, pw, salt, pbkdf2Iter, 32)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s$%d$%s$%s", pbkdf2Prefix, pbkdf2Iter, b64.EncodeToString(salt), b64.EncodeToString(key)), nil
}

func verifyPassword(stored, pw string) bool {
	parts := strings.Split(stored, "$")
	if len(parts) != 4 || parts[0] != pbkdf2Prefix {
		return false
	}
	iter, err := strconv.Atoi(parts[1])
	if err != nil || iter < 1000 || iter > 10_000_000 {
		return false
	}
	salt, err1 := b64.DecodeString(parts[2])
	want, err2 := b64.DecodeString(parts[3])
	if err1 != nil || err2 != nil || len(want) == 0 {
		return false
	}
	got, err := pbkdf2.Key(sha256.New, pw, salt, iter, len(want))
	return err == nil && subtle.ConstantTimeCompare(got, want) == 1
}

func newToken() string {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	return base64.RawURLEncoding.EncodeToString(b)
}

func tokenHash(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func tokensEqual(a, b string) bool {
	return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}

type attempt struct {
	fails int
	first time.Time
	until time.Time
}

type limiter struct {
	mu         sync.Mutex
	max        int
	window     time.Duration
	block      time.Duration
	maxEntries int
	m          map[string]*attempt
}

func newLimiter(max int, window, block time.Duration) *limiter {
	return &limiter{max: max, window: window, block: block, maxEntries: 10000, m: map[string]*attempt{}}
}

func (l *limiter) allowed(key string, now time.Time) (bool, time.Duration) {
	l.mu.Lock()
	defer l.mu.Unlock()
	a := l.m[key]
	if a != nil && now.Before(a.until) {
		return false, a.until.Sub(now)
	}
	return true, 0
}

func (l *limiter) fail(key string, now time.Time) {
	l.mu.Lock()
	defer l.mu.Unlock()
	a := l.m[key]
	if a == nil || now.Sub(a.first) > l.window {
		if a == nil && len(l.m) >= l.maxEntries {
			l.purgeLocked(now)
			if len(l.m) >= l.maxEntries {
				l.evictOldestLocked()
			}
		}
		a = &attempt{first: now}
		l.m[key] = a
	}
	a.fails++
	if a.fails >= l.max {
		a.until = now.Add(l.block)
	}
}

func (l *limiter) reset(key string) {
	l.mu.Lock()
	delete(l.m, key)
	l.mu.Unlock()
}

func (l *limiter) purge(now time.Time) {
	l.mu.Lock()
	l.purgeLocked(now)
	l.mu.Unlock()
}

func (l *limiter) purgeLocked(now time.Time) {
	for k, a := range l.m {
		if now.Sub(a.first) > l.window && !now.Before(a.until) {
			delete(l.m, k)
		}
	}
}

func (l *limiter) evictOldestLocked() {
	var oldestKey string
	var oldest time.Time
	for k, a := range l.m {
		if oldestKey == "" || a.first.Before(oldest) {
			oldestKey, oldest = k, a.first
		}
	}
	delete(l.m, oldestKey)
}
