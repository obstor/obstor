/*
 * PGG Obstor, (C) 2021-2026 PGG, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package cmd

import (
	"errors"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

var errLDAPAuthFailed = errors.New("LDAP authentication failed")

type ldapSTSThrottle struct {
	mu      sync.Mutex
	perIP   map[string]*throttleEntry
	perUser map[string]*throttleEntry
}

type throttleEntry struct {
	lim  *rate.Limiter
	seen time.Time
}

const (
	// Allow a short burst, then ~1 attempt/sec sustained per key.
	ldapThrottleBurst   = 5
	ldapThrottleRefill  = rate.Limit(1)
	ldapThrottleIdleTTL = 10 * time.Minute
	// Soft cap on tracked keys; an idle sweep runs once a map grows past it.
	ldapThrottleSoftMax = 4096
)

var globalLDAPSTSThrottle = newLDAPSTSThrottle()

func newLDAPSTSThrottle() *ldapSTSThrottle {
	return &ldapSTSThrottle{
		perIP:   make(map[string]*throttleEntry),
		perUser: make(map[string]*throttleEntry),
	}
}

// LDAP STS attempt from the given client IP
func (t *ldapSTSThrottle) allow(ip, user string) bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	now := UTCNow()
	ipOK := t.allowKeyLocked(t.perIP, ip, now)
	userOK := t.allowKeyLocked(t.perUser, user, now)
	return ipOK && userOK
}

func (t *ldapSTSThrottle) allowKeyLocked(m map[string]*throttleEntry, key string, now time.Time) bool {
	if key == "" {
		return true
	}
	if len(m) > ldapThrottleSoftMax {
		t.evictIdleLocked(m, now)
	}
	e, ok := m[key]
	if !ok {
		e = &throttleEntry{lim: rate.NewLimiter(ldapThrottleRefill, ldapThrottleBurst)}
		m[key] = e
	}
	e.seen = now
	return e.lim.Allow()
}

func (t *ldapSTSThrottle) evictIdleLocked(m map[string]*throttleEntry, now time.Time) {
	for k, e := range m {
		if now.Sub(e.seen) > ldapThrottleIdleTTL {
			delete(m, k)
		}
	}
}
