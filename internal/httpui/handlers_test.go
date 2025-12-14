package httpui

import (
	"bytes"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"ganache-admin-ui/internal/auth"
	"ganache-admin-ui/internal/config"
	"ganache-admin-ui/internal/ganache"
)

func newTestServer(t *testing.T, ganacheHandler http.HandlerFunc) (*Server, *auth.SessionStore) {
	t.Helper()
	backend := httptest.NewServer(ganacheHandler)
	t.Cleanup(backend.Close)

	cfg := &config.Config{
		ListenAddr:    ":0",
		UsersFile:     "./users.yaml",
		SessionSecret: []byte("secret"),
		CSRFSecret:    []byte("csrf"),
		Ganache: config.GanacheConfig{
			BaseURL: backend.URL,
			APIKey:  "key",
			Timeout: time.Second,
		},
	}
	users, err := auth.NewUserStore([]auth.User{{Username: "tester", PasswordHash: "hash"}})
	if err != nil {
		t.Fatalf("users: %v", err)
	}
	sessions := auth.NewSessionStore(time.Hour)
	client := ganache.NewClient(cfg.Ganache.BaseURL, cfg.Ganache.APIKey, cfg.Ganache.Timeout)
	srv, err := NewServer(cfg, users, sessions, client)
	if err != nil {
		t.Fatalf("server: %v", err)
	}
	return srv, sessions
}

func TestAssetsIndexCallsSearch(t *testing.T) {
	var captured *http.Request
	srv, sessions := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/assets" {
			http.NotFound(w, r)
			return
		}
		captured = r
		io.WriteString(w, `{"assets":[{"id":"1","title":"Cat","variants":{}}],"page":1,"pageSize":20,"total":1}`)
	})
	router := srv.Router()

	sess, _ := sessions.Create("tester")
	req := httptest.NewRequest(http.MethodGet, "/assets?q=cat&tag=news&sort=oldest&page=2&pageSize=5", nil)
	req.AddCookie(&http.Cookie{Name: "session", Value: sess.ID})
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if captured == nil {
		t.Fatalf("ganache search not called")
	}
	if captured.Header.Get("X-Api-Key") != "key" {
		t.Fatalf("expected api key header")
	}
	q := captured.URL.Query()
	if q.Get("q") != "cat" || q.Get("sort") != "oldest" || q.Get("page") != "2" || q.Get("pageSize") != "5" {
		t.Fatalf("query params not forwarded: %v", q)
	}
	tags := q["tag"]
	if len(tags) != 1 || tags[0] != "news" {
		t.Fatalf("tags not forwarded: %v", tags)
	}
	if !strings.Contains(rec.Body.String(), "Cat") {
		t.Fatalf("expected rendered assets")
	}
}

func TestAssetsUploadForwardsMultipart(t *testing.T) {
	var filename, title string
	var tags []string
	srv, sessions := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/assets" || r.Method != http.MethodPost {
			http.NotFound(w, r)
			return
		}
		if err := r.ParseMultipartForm(1024); err != nil {
			t.Fatalf("parse multipart: %v", err)
		}
		if fhs := r.MultipartForm.File["file"]; len(fhs) > 0 {
			filename = fhs[0].Filename
		}
		title = r.FormValue("title")
		tags = append(tags, r.MultipartForm.Value["tags[]"]...)
		io.WriteString(w, `{"id":"xyz"}`)
	})
	router := srv.Router()

	sess, _ := sessions.Create("tester")
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	fileWriter, _ := writer.CreateFormFile("file", "pic.png")
	io.Copy(fileWriter, strings.NewReader("hello"))
	writer.WriteField("title", "Cover")
	writer.WriteField("tags[]", "one")
	writer.WriteField("csrf", sess.CSRFToken)
	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/assets/upload", &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.AddCookie(&http.Cookie{Name: "session", Value: sess.ID})
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusFound {
		t.Fatalf("expected redirect, got %d", rec.Code)
	}
	if loc := rec.Header().Get("Location"); loc != "/assets/xyz" {
		t.Fatalf("unexpected redirect: %s", loc)
	}
	if filename != "pic.png" || title != "Cover" {
		t.Fatalf("multipart fields missing: %s %s", filename, title)
	}
	if len(tags) != 1 || tags[0] != "one" {
		t.Fatalf("tags missing: %v", tags)
	}
}

func TestAssetEditSendsPatchAndRendersPartial(t *testing.T) {
	var update ganache.AssetUpdate
	srv, sessions := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/assets/123" || r.Method != http.MethodPatch {
			http.NotFound(w, r)
			return
		}
		if err := json.NewDecoder(r.Body).Decode(&update); err != nil {
			t.Fatalf("decode: %v", err)
		}
		resp := ganache.Asset{ID: "123", Title: update.Title, Tags: update.Tags, Variants: ganache.Variants{}}
		json.NewEncoder(w).Encode(resp)
	})
	router := srv.Router()

	sess, _ := sessions.Create("tester")
	form := strings.NewReader("title=Updated&tags=one,two&csrf=" + sess.CSRFToken)
	req := httptest.NewRequest(http.MethodPost, "/assets/123/edit", form)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("HX-Request", "true")
	req.AddCookie(&http.Cookie{Name: "session", Value: sess.ID})

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if update.Title != "Updated" || len(update.Tags) != 2 || update.Tags[0] != "one" {
		t.Fatalf("update payload wrong: %+v", update)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "Updated") {
		t.Fatalf("expected updated partial")
	}
}
