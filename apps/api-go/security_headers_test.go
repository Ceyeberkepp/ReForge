package main

import (
    "net/http"
    "net/http/httptest"
    "testing"
)

func TestSensitiveRoutesDisableCaching(t *testing.T) {
    app := &App{}
    handler := app.securityHeaders(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
    }))
    cases := []struct {
        path string
        noStore bool
    }{
        {"/api/auth/login", true},
        {"/api/directory", true},
        {"/api/deployments/123/plan", true},
        {"/boot/catalog.ipxe", true},
        {"/boot/sources.ipxe?token=temporary", true},
        {"/health", false},
        {"/", false},
    }
    for _, tc := range cases {
        t.Run(tc.path, func(t *testing.T) {
            rr := httptest.NewRecorder()
            handler.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, tc.path, nil))
            got := rr.Header().Get("Cache-Control")
            if tc.noStore && got != "no-store" {
                t.Fatalf("Cache-Control = %q, want no-store", got)
            }
            if !tc.noStore && got != "" {
                t.Fatalf("unexpected Cache-Control %q", got)
            }
        })
    }
}
