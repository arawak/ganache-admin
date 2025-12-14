package auth

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"golang.org/x/crypto/bcrypt"
)

func TestLoadUsersAndValidate(t *testing.T) {
	hash, err := bcrypt.GenerateFromPassword([]byte("secret"), bcrypt.MinCost)
	if err != nil {
		t.Fatalf("hash: %v", err)
	}
	yaml := "users:\n  - username: alice\n    passwordHash: \"" + string(hash) + "\"\n"
	file, err := os.CreateTemp(t.TempDir(), "users-*.yaml")
	if err != nil {
		t.Fatalf("tmp: %v", err)
	}
	if _, err := file.WriteString(yaml); err != nil {
		t.Fatalf("write: %v", err)
	}
	file.Close()

	store, err := LoadUsers(file.Name())
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if !store.Validate("alice", "secret") {
		t.Fatalf("expected valid credentials")
	}
	if store.Validate("alice", "wrong") {
		t.Fatalf("expected invalid password")
	}
}

func TestSessionStore(t *testing.T) {
	store := NewSessionStore(20 * time.Millisecond)
	sess, err := store.Create("bob")
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if _, ok := store.Get(sess.ID); !ok {
		t.Fatalf("expected session present")
	}
	time.Sleep(30 * time.Millisecond)
	if _, ok := store.Get(sess.ID); ok {
		t.Fatalf("expected session expired")
	}
}

func TestRequireAuthMiddleware(t *testing.T) {
	store := NewSessionStore(time.Minute)
	sess, _ := store.Create("tester")
	h := RequireAuth(store)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/assets", nil)
	req.AddCookie(&http.Cookie{Name: "session", Value: sess.ID})
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	req2 := httptest.NewRequest("GET", "/assets", nil)
	rec2 := httptest.NewRecorder()
	h.ServeHTTP(rec2, req2)
	if rec2.Code != http.StatusFound {
		t.Fatalf("expected redirect, got %d", rec2.Code)
	}
}
