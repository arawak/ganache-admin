package ganache

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestSearchAssetsBuildsRequest(t *testing.T) {
	var captured *http.Request
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		captured = r
		io.WriteString(w, `{"assets":[],"page":1,"pageSize":10,"total":0}`)
	}))
	t.Cleanup(ts.Close)

	client := NewClient(ts.URL, "key", time.Second)
	_, err := client.SearchAssets(context.Background(), "cat", []string{"news", "sports"}, 2, 10, "newest")
	if err != nil {
		t.Fatalf("search: %v", err)
	}
	if captured.Header.Get("X-Api-Key") != "key" {
		t.Fatalf("missing api key")
	}
	q := captured.URL.Query()
	if q.Get("q") != "cat" || q.Get("sort") != "newest" || q.Get("page") != "2" {
		t.Fatalf("query not set")
	}
	tags := q["tag"]
	if len(tags) != 2 || tags[0] != "news" {
		t.Fatalf("tags not forwarded")
	}
}

func TestSearchAssetsMapsItemsToAssets(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, `{"items":[{"id":"13","title":"cricket","variants":{"thumb":"/media/13/thumb"}}],"page":1,"pageSize":30,"total":1}`)
	}))
	t.Cleanup(ts.Close)

	client := NewClient(ts.URL, "key", time.Second)
	resp, err := client.SearchAssets(context.Background(), "cricket", nil, 1, 30, "newest")
	if err != nil {
		t.Fatalf("search: %v", err)
	}
	if len(resp.Assets) != 1 || resp.Assets[0].ID != "13" {
		t.Fatalf("items not mapped to assets: %+v", resp.Assets)
	}
	expectedThumb := ts.URL + "/media/13/thumb"
	if resp.Assets[0].Variants.Thumb != expectedThumb {
		t.Fatalf("thumb not absolutized: %s", resp.Assets[0].Variants.Thumb)
	}
}

func TestUpdateAssetSendsJSON(t *testing.T) {
	var body AssetUpdate
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		json.NewDecoder(r.Body).Decode(&body)
		io.WriteString(w, `{"id":"1"}`)
	}))
	t.Cleanup(ts.Close)

	client := NewClient(ts.URL, "key", time.Second)
	_, err := client.UpdateAsset(context.Background(), "1", AssetUpdate{Title: "New"})
	if err != nil {
		t.Fatalf("update: %v", err)
	}
	if body.Title != "New" {
		t.Fatalf("expected json body")
	}
}

func TestCreateAssetMultipart(t *testing.T) {
	var filename string
	var tags []string
	var fields map[string]string
	fields = make(map[string]string)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.ParseMultipartForm(1024)
		if fhs := r.MultipartForm.File["file"]; len(fhs) > 0 {
			filename = fhs[0].Filename
		}
		for k, v := range r.MultipartForm.Value {
			if k == "tags[]" {
				tags = append(tags, v...)
			} else {
				fields[k] = strings.Join(v, ",")
			}
		}
		io.WriteString(w, `{"id":"55"}`)
	}))
	t.Cleanup(server.Close)

	client := NewClient(server.URL, "key", time.Second)
	_, err := client.CreateAssetMultipart(context.Background(), strings.NewReader("hello"), "hello.png", map[string]string{"title": "Hi"}, []string{"a", "b"})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if filename != "hello.png" {
		t.Fatalf("filename not forwarded")
	}
	if len(tags) != 2 || tags[0] != "a" {
		t.Fatalf("tags missing")
	}
	if fields["title"] != "Hi" {
		t.Fatalf("field missing")
	}
}

func TestParseErrorUsesMessage(t *testing.T) {
	rec := httptest.NewRecorder()
	rec.WriteHeader(http.StatusBadRequest)
	rec.Body.WriteString(`{"error":{"message":"nope"}}`)
	err := parseError(rec.Result())
	if err == nil || err.Error() != "nope" {
		t.Fatalf("expected error message")
	}
}
