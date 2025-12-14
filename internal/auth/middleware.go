package auth

import (
	"context"
	"net/http"
	"time"
)

type contextKey string

const sessionContextKey contextKey = "session"

func RequireAuth(store *SessionStore) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie("session")
			if err != nil || cookie.Value == "" {
				http.Redirect(w, r, "/login", http.StatusFound)
				return
			}
			sess, ok := store.Get(cookie.Value)
			if !ok {
				http.Redirect(w, r, "/login", http.StatusFound)
				return
			}
			r = r.WithContext(ContextWithSession(r.Context(), sess))
			next.ServeHTTP(w, r)
		})
	}
}

func SessionFromContext(ctx context.Context) (Session, bool) {
	val := ctx.Value(sessionContextKey)
	if val == nil {
		return Session{}, false
	}
	sess, ok := val.(Session)
	return sess, ok
}

func ContextWithSession(ctx context.Context, sess Session) context.Context {
	return context.WithValue(ctx, sessionContextKey, sess)
}

func SetSessionCookie(w http.ResponseWriter, sess Session, secure bool) {
	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    sess.ID,
		Path:     "/",
		Expires:  sess.ExpiresAt,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   secure,
	})
}

func ClearSessionCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:    "session",
		Value:   "",
		Path:    "/",
		Expires: time.Unix(0, 0),
		MaxAge:  -1,
	})
}
