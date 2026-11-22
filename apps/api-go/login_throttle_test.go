package main

import (
 "testing"
 "fmt"
 "time"
)
func TestLoginFailuresWindow(t *testing.T) {
 l := loginFailures{entries: map[string]loginFailure{}}
 now := time.Date(2026, 11, 22, 0, 0, 0, 0, time.UTC)
 for i:=0; i<4; i++ { l.fail("Alice",now) }
 if l.blocked("alice",now) { t.Fatal("blocked too early") }
 l.fail("ALICE",now)
 if !l.blocked(" alice ",now) { t.Fatal("fifth failure must block") }
 if l.blocked("bob",now) { t.Fatal("other account blocked") }
 if l.blocked("alice",now.Add(16*time.Minute)) { t.Fatal("expired window still blocked") }
}
func TestLoginFailuresClear(t *testing.T) {
 l := loginFailures{entries: map[string]loginFailure{}}
 now:=time.Now()
 for i:=0;i<5;i++ { l.fail("admin",now) }
 l.clear("admin")
 if l.blocked("admin",now) { t.Fatal("successful login must clear failures") }
}

func TestLoginFailuresSaturatedMapStillCountsNewUser(t *testing.T) {
 l := loginFailures{entries: map[string]loginFailure{}}
 now := time.Date(2026, 11, 22, 0, 0, 0, 0, time.UTC)
 for i := 0; i < 10000; i++ {
  l.entries[fmt.Sprintf("account-%d", i)] = loginFailure{count: 1, until: now.Add(15*time.Minute)}
 }
 for i := 0; i < 5; i++ { l.fail("target", now) }
 if !l.blocked("target", now) { t.Fatal("saturated map bypassed throttle") }
 if len(l.entries) > 10000 { t.Fatal("throttle map grew past limit") }
}
