package main

import (
 "encoding/base64"
 "strings"
 "testing"
)

func TestVerifyPasswordRejectsMalformedHashDimensions(t *testing.T) {
 valid := hashPassword("correct horse battery staple")
 if !verifyPassword(valid, "correct horse battery staple") {
  t.Fatal("valid password rejected")
 }
 if verifyPassword(valid, "wrong password") {
  t.Fatal("wrong password accepted")
 }
 parts := strings.Split(valid, ".")
 cases := map[string]string{
  "empty": "",
  "invalid-base64": "%%%%.%%%%",
  "missing-hash": parts[0] + ".",
  "short-salt": base64.RawStdEncoding.EncodeToString(make([]byte, 8)) + "." + parts[1],
  "short-key": parts[0] + "." + base64.RawStdEncoding.EncodeToString(make([]byte, 16)),
  "oversized-key": parts[0] + "." + base64.RawStdEncoding.EncodeToString(make([]byte, 4096)),
  "extra-separator": valid + ".extra",
 }
 for name, encoded := range cases {
  t.Run(name, func(t *testing.T) {
   if verifyPassword(encoded, "correct horse battery staple") {
    t.Fatal("malformed stored password hash accepted")
   }
  })
 }
}
