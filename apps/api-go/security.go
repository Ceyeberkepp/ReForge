package main

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/argon2"
)

type contextKey string

const userKey contextKey = "user"

func randomToken(n int) string {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	return base64.RawURLEncoding.EncodeToString(b)
}

func hashPassword(password string) string {
	salt := make([]byte, 16)
	_, _ = rand.Read(salt)
	hash := argon2.IDKey([]byte(password), salt, 3, 64*1024, 2, 32)
	return base64.RawStdEncoding.EncodeToString(salt) + "." + base64.RawStdEncoding.EncodeToString(hash)
}

func verifyPassword(encoded, password string) bool {
	parts := strings.Split(encoded, ".")
	if len(parts) != 2 {
		return false
	}
	salt, err1 := base64.RawStdEncoding.DecodeString(parts[0])
	want, err2 := base64.RawStdEncoding.DecodeString(parts[1])
	if err1 != nil || err2 != nil {
		return false
	}
	got := argon2.IDKey([]byte(password), salt, 3, 64*1024, 2, uint32(len(want)))
	return subtle.ConstantTimeCompare(want, got) == 1
}

func (a *App) bootstrapAdmin() error {
	var count int64
	if err := a.db.Model(&User{}).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	if len(a.cfg.AdminPassword) < 12 {
		return errors.New("REFORGE_ADMIN_PASSWORD must be at least 12 characters on first startup")
	}
	u := User{
		ID:           uuid.NewString(),
		Username:     a.cfg.AdminUser,
		PasswordHash: hashPassword(a.cfg.AdminPassword),
		Role:         "admin",
		CreatedAt:    time.Now().UTC(),
	}
	return a.db.Create(&u).Error
}

func (a *App) currentUser(r *http.Request) *User {
	value := r.Context().Value(userKey)
	if value == nil {
		return nil
	}
	u, _ := value.(*User)
	return u
}

func clientIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil {
		return host
	}
	return r.RemoteAddr
}

func (a *App) audit(r *http.Request, action, resource, id string, success bool, detail string) {
	actor := "anonymous"
	if u := a.currentUser(r); u != nil {
		actor = u.Username
	}
	_ = a.db.Create(&AuditEvent{
		ID:         uuid.NewString(),
		At:         time.Now().UTC(),
		Actor:      actor,
		Action:     action,
		Resource:   resource,
		ResourceID: id,
		RemoteIP:   clientIP(r),
		Success:    success,
		Detail:     detail,
	}).Error
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func (a *App) securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Referrer-Policy", "no-referrer")
		w.Header().Set("Permissions-Policy", "camera=(), microphone=(), geolocation=()")
		next.ServeHTTP(w, r)
	})
}

func (a *App) cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin != "" && origin == a.cfg.AllowedOrigin {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
			w.Header().Set("Access-Control-Allow-Methods", "GET,POST,PUT,PATCH,DELETE,OPTIONS")
			w.Header().Add("Vary", "Origin")
		}
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (a *App) auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("reforge_session")
		if err != nil {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"detail": "authentication required"})
			return
		}
		var s Session
		if a.db.First(&s, "token = ? AND expires_at > ?", cookie.Value, time.Now().UTC()).Error != nil {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"detail": "session expired"})
			return
		}
		var u User
		if a.db.First(&u, "id = ? AND disabled = ?", s.UserID, false).Error != nil {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"detail": "authentication required"})
			return
		}
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), userKey, &u)))
	})
}


func (a *App) workerAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if len(a.cfg.WorkerToken) < 32 {
			writeJSON(w, http.StatusServiceUnavailable, map[string]string{"detail": "worker token is not configured"})
			return
		}
		const prefix = "Bearer "
		auth := r.Header.Get("Authorization")
		if !strings.HasPrefix(auth, prefix) {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"detail": "worker authentication required"})
			return
		}
		got := []byte(strings.TrimPrefix(auth, prefix))
		want := []byte(a.cfg.WorkerToken)
		if len(got) != len(want) || subtle.ConstantTimeCompare(got, want) != 1 {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"detail": "invalid worker token"})
			return
		}
		next.ServeHTTP(w, r)
	})
}
