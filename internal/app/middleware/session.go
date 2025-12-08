package middleware

import (
	"context"
	"net/http"

	"github.com/google/uuid"
)

type ctxKey int

const (
	sessionIDKey ctxKey = iota
)

// SessionIDFromContext returns session id stored by Session middleware
func SessionIDFromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	if v := ctx.Value(sessionIDKey); v != nil {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func Session(sessionName string, next http.Handler) http.Handler {
	if sessionName == "" || next == nil {
		return next
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var sid string

		if c, err := r.Cookie(sessionName); err == nil && c != nil && c.Value != "" {
			sid = c.Value
		} else {
			sid = uuid.NewString()
		}

		http.SetCookie(w, &http.Cookie{
			Name:     sessionName,
			Value:    sid,
			Path:     "/",
			HttpOnly: true,
			// Secure: true,
			SameSite: http.SameSiteLaxMode,
			MaxAge:   180 * 24 * 3600,
		})

		ctx := context.WithValue(r.Context(), sessionIDKey, sid)
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}
