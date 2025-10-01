package auth

import (
    "context"
    "net/http"
    "strings"
)

// ContextKey is used for storing auth data in context.
type ContextKey string

const (
    // CtxKeyPrincipal is the principal extracted from a validated request.
    CtxKeyPrincipal ContextKey = "principal"
)

// Authenticator validates bearer tokens.
type Authenticator struct {
    apiToken string
}

// NewAuthenticator creates a new Authenticator instance.
func NewAuthenticator(token string) *Authenticator {
    return &Authenticator{apiToken: token}
}

// Middleware checks bearer token and attaches a principal to the context.
func (a *Authenticator) Middleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        token := extractBearer(r.Header.Get("Authorization"))
        if token == "" || token != a.apiToken {
            http.Error(w, "unauthorized", http.StatusUnauthorized)
            return
        }
        ctx := context.WithValue(r.Context(), CtxKeyPrincipal, "bot")
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}

func extractBearer(header string) string {
    if header == "" {
        return ""
    }
    parts := strings.SplitN(header, " ", 2)
    if len(parts) != 2 {
        return ""
    }
    if !strings.EqualFold(parts[0], "bearer") {
        return ""
    }
    return strings.TrimSpace(parts[1])
}
