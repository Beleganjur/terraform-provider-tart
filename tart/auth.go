package tart

import (
	"net/http"
	"strings"
)

func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		// For development/testing: allow requests with no Authorization header.
		// If a header is provided, ensure it looks like a Bearer token.
		if auth != "" {
			if !strings.HasPrefix(auth, "Bearer ") || len(auth) < len("Bearer ")+1 {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
		}
		// Add JWT validation here
		next(w, r)
	}
}
