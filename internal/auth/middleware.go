package auth

import (
	"context"
	"net/http"
	"strings"
)

type Identity struct {
	UserID string
	Role   string
}

type ctxKey struct{}

// Middleware attaches the caller's identity to the context when a valid Bearer token is sent.
// It never rejects: enforcement is done per-field by the @hasRole directive.
func Middleware(m *Manager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if h := r.Header.Get("Authorization"); strings.HasPrefix(h, "Bearer ") {
				if c, err := m.Parse(strings.TrimPrefix(h, "Bearer ")); err == nil {
					ctx := context.WithValue(r.Context(), ctxKey{}, &Identity{UserID: c.Subject, Role: c.Role})
					r = r.WithContext(ctx)
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}

func FromContext(ctx context.Context) (*Identity, bool) {
	id, ok := ctx.Value(ctxKey{}).(*Identity)
	return id, ok && id != nil
}
