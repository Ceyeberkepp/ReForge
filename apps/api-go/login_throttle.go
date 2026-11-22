package main

import (
 "strings"
 "sync"
 "time"
)

// Login throttling is deliberately keyed by account name rather than the
// untrusted X-Forwarded-For header. A reverse proxy may hide client IPs.
type loginFailures struct {
 mu sync.Mutex
 entries map[string]loginFailure
}
type loginFailure struct {
 count int
 until time.Time
}
var failedLogins = loginFailures{entries: make(map[string]loginFailure)}

func (l *loginFailures) blocked(username string, now time.Time) bool {
 key := strings.ToLower(strings.TrimSpace(username))
 l.mu.Lock()
 defer l.mu.Unlock()
 e, ok := l.entries[key]
 if !ok { return false }
 if !now.Before(e.until) { delete(l.entries,key); return false }
 return e.count >= 5
}
func (l *loginFailures) fail(username string, now time.Time) {
 key := strings.ToLower(strings.TrimSpace(username))
 l.mu.Lock()
 defer l.mu.Unlock()
 // Bound memory for arbitrary usernames; evict expired records first.
 if len(l.entries) >= 10000 {
  for k, e := range l.entries { if !now.Before(e.until) { delete(l.entries,k) } }
  if len(l.entries) >= 10000 { return } // never evict active lockouts
 }
 e := l.entries[key]
 if !now.Before(e.until) { e = loginFailure{until: now.Add(15*time.Minute)} }
 e.count++
 l.entries[key] = e
}
func (l *loginFailures) clear(username string) {
 l.mu.Lock()
 defer l.mu.Unlock()
 delete(l.entries, strings.ToLower(strings.TrimSpace(username)))
}
