package security

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"ganache-admin-ui/internal/auth"
)

func TestCSRFMiddleware(t *testing.T) {
	store := auth.NewSessionStore(time.Minute)
	sess, _ := store.Create("alice")
	h := Middleware()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodPost, "/save", strings.NewReader(""))
	req = req.WithContext(auth.ContextWithSession(req.Context(), sess))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rec.Code)
	}

	body := "csrf=" + sess.CSRFToken
	req2 := httptest.NewRequest(http.MethodPost, "/save", strings.NewReader(body))
	req2.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req2 = req2.WithContext(auth.ContextWithSession(req2.Context(), sess))
	rec2 := httptest.NewRecorder()
	h.ServeHTTP(rec2, req2)
	if rec2.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec2.Code)
	}
}
