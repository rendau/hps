package middleware

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/google/uuid"
)

type ctxKey string

const ctxKeySessionID ctxKey = "session_id"

// ContextSessionID returns session id from context if set, otherwise empty string
func ContextSessionID(ctx context.Context) string {
	v := ctx.Value(ctxKeySessionID)
	if v != nil {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

type Session struct {
	name string
}

func NewSession(name string) *Session {
	return &Session{name: name}
}

func (m *Session) Middleware(next http.Handler) http.Handler {
	if m.name == "" {
		return next
	}

	slog.Info("Session middleware created", "name", m.name)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// read existing cookie
		var sid string
		if c, err := r.Cookie(m.name); err == nil && c != nil && c.Value != "" {
			sid = c.Value
		}

		// generate if missing
		if sid == "" {
			sid = genSessionID()
			// set cookie immediately
			http.SetCookie(w, &http.Cookie{
				Name:     m.name,
				Value:    sid,
				Path:     "/",
				HttpOnly: true,
				SameSite: http.SameSiteLaxMode,
			})
		}

		// put into context
		r = r.WithContext(context.WithValue(r.Context(), ctxKeySessionID, sid))

		next.ServeHTTP(w, r)
	})
}

func genSessionID() string {
	return uuid.NewString()
}
