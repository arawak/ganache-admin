package httpui

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"ganache-admin-ui/internal/ganache"

	"github.com/go-chi/chi/v5"
)

func (s *Server) assetsIndex(w http.ResponseWriter, r *http.Request) {
	q := strings.TrimSpace(r.URL.Query().Get("q"))
	tags := r.URL.Query()["tag"]
	sort := r.URL.Query().Get("sort")
	page := parseInt(r.URL.Query().Get("page"), 1)
	pageSize := parseInt(r.URL.Query().Get("pageSize"), 20)

	resp, err := s.client.SearchAssets(r.Context(), q, tags, page, pageSize, sort)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	extra := map[string]any{
		"sort":     sort,
		"hasPrev":  resp.Page > 1,
		"hasNext":  resp.Page*resp.PageSize < resp.Total,
		"prevPage": resp.Page - 1,
		"nextPage": resp.Page + 1,
		"pageSize": resp.PageSize,
	}
	s.render(w, "assets_index.html", TemplateData{
		Title:  "Assets",
		Query:  q,
		Tags:   tags,
		Search: resp,
		Assets: resp.Assets,
		Extra:  extra,
	}, r)
}

func (s *Server) assetsResults(w http.ResponseWriter, r *http.Request) {
	q := strings.TrimSpace(r.URL.Query().Get("q"))
	tags := r.URL.Query()["tag"]
	sort := r.URL.Query().Get("sort")
	page := parseInt(r.URL.Query().Get("page"), 1)
	pageSize := parseInt(r.URL.Query().Get("pageSize"), 20)
	resp, err := s.client.SearchAssets(r.Context(), q, tags, page, pageSize, sort)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	extra := map[string]any{
		"sort":     sort,
		"hasPrev":  resp.Page > 1,
		"hasNext":  resp.Page*resp.PageSize < resp.Total,
		"prevPage": resp.Page - 1,
		"nextPage": resp.Page + 1,
		"pageSize": resp.PageSize,
	}

	if r.Header.Get("HX-Request") != "true" {
		s.render(w, "assets_index.html", TemplateData{
			Title:  "Assets",
			Query:  q,
			Tags:   tags,
			Search: resp,
			Assets: resp.Assets,
			Extra:  extra,
		}, r)
		return
	}

	s.render(w, "assets_results_partial.html", TemplateData{
		Query:  q,
		Tags:   tags,
		Search: resp,
		Assets: resp.Assets,
		Extra:  extra,
	}, r)
}

func (s *Server) assetsNew(w http.ResponseWriter, r *http.Request) {
	s.render(w, "assets_index.html", TemplateData{
		Title: "Upload Asset",
		Extra: map[string]any{"new": true},
	}, r)
}

func (s *Server) assetsUpload(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)
	if err := r.ParseMultipartForm(maxUploadSize); err != nil {
		http.Error(w, "upload too large", http.StatusRequestEntityTooLarge)
		return
	}
	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "file is required", http.StatusBadRequest)
		return
	}
	defer file.Close()

	fields := map[string]string{
		"title":      r.FormValue("title"),
		"caption":    r.FormValue("caption"),
		"credit":     r.FormValue("credit"),
		"source":     r.FormValue("source"),
		"usageNotes": r.FormValue("usageNotes"),
	}
	tags := parseTags(r)

	asset, err := s.client.CreateAssetMultipart(r.Context(), file, header.Filename, fields, tags)
	if err != nil {
		data := TemplateData{Title: "Upload Asset", Error: err.Error(), Extra: map[string]any{"new": true}}
		s.render(w, "assets_index.html", data, r)
		return
	}
	http.Redirect(w, r, s.path(fmt.Sprintf("/assets/%s", asset.ID)), http.StatusFound)
}

func (s *Server) assetDetail(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	asset, err := s.client.GetAsset(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	s.render(w, "asset_detail.html", TemplateData{Title: asset.Title, Asset: asset}, r)
}

func (s *Server) assetEdit(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := r.ParseForm(); err != nil {
		http.Error(w, "invalid form", http.StatusBadRequest)
		return
	}
	update := ganache.AssetUpdate{
		Title:      r.FormValue("title"),
		Caption:    r.FormValue("caption"),
		Credit:     r.FormValue("credit"),
		Source:     r.FormValue("source"),
		UsageNotes: r.FormValue("usageNotes"),
		Tags:       parseTags(r),
	}
	asset, err := s.client.UpdateAsset(r.Context(), id, update)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	if r.Header.Get("HX-Request") == "true" {
		s.render(w, "asset_meta_partial.html", TemplateData{Asset: asset}, r)
		return
	}
	http.Redirect(w, r, s.path(fmt.Sprintf("/assets/%s", asset.ID)), http.StatusFound)
}

func (s *Server) assetDelete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := s.client.DeleteAsset(r.Context(), id); err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	http.Redirect(w, r, s.path("/assets"), http.StatusFound)
}

func parseTags(r *http.Request) []string {
	var inputs []string
	inputs = append(inputs, r.Form["tags"]...)
	inputs = append(inputs, r.Form["tags[]"]...)
	if v := r.FormValue("tags"); v != "" {
		inputs = append(inputs, v)
	}

	seen := map[string]struct{}{}
	var result []string
	for _, raw := range inputs {
		for _, part := range strings.Split(raw, ",") {
			t := strings.TrimSpace(part)
			if t == "" {
				continue
			}
			if _, ok := seen[t]; ok {
				continue
			}
			seen[t] = struct{}{}
			result = append(result, t)
		}
	}
	return result
}

func parseInt(val string, def int) int {
	if val == "" {
		return def
	}
	i, err := strconv.Atoi(val)
	if err != nil || i <= 0 {
		return def
	}
	return i
}
