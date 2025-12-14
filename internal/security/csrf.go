package security

import (
	"net/http"

	"ganache-admin-ui/internal/auth"
)

func TokenFromRequest(r *http.Request) string {
	token := r.FormValue("csrf")
	if token != "" {
		return token
	}
	return r.Header.Get("X-CSRF-Token")
}

func Middleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPost && r.Method != http.MethodPut && r.Method != http.MethodPatch && r.Method != http.MethodDelete {
				next.ServeHTTP(w, r)
				return
			}
			sess, ok := auth.SessionFromContext(r.Context())
			if !ok {
				http.Error(w, "forbidden", http.StatusForbidden)
				return
			}
			token := TokenFromRequest(r)
			if token == "" || token != sess.CSRFToken {
				http.Error(w, "invalid csrf token", http.StatusForbidden)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func TokenFromSession(sess auth.Session) string {
	return sess.CSRFToken
}
