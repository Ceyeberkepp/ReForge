package main

import (
    "net/http"
    "net/http/httptest"
    "strings"
    "testing"
)

func TestLoginRejectsMalformedOrOversizedBodyBeforeDatabase(t *testing.T) {
    cases := []struct {
        name string
        body string
    }{
        {"malformed", "{not-json"},
        {"trailing-object", `{"username":"test","password":"wrong"}{"extra":true}`},
        {"trailing-garbage", `{"username":"test","password":"wrong"} trailing`},
        {"oversized", `{"username":"` + strings.Repeat("x", 8192) + `","password":"x"}`},
    }
    app := &App{} // Invalid requests must fail without touching the database.
    for _, tc := range cases {
        t.Run(tc.name, func(t *testing.T) {
            req := httptest.NewRequest(http.MethodPost, "/api/auth/login", strings.NewReader(tc.body))
            recorder := httptest.NewRecorder()
            app.login(recorder, req)
            if recorder.Code != http.StatusBadRequest {
                t.Fatalf("status = %d; want 400; body = %s", recorder.Code, recorder.Body.String())
            }
        })
    }
}
