package main

import (
 "testing"
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
