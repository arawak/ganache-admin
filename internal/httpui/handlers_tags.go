package httpui

import (
	"fmt"
	"html"
	"net/http"
)

func (s *Server) tagsList(w http.ResponseWriter, r *http.Request) {
	prefix := r.URL.Query().Get("prefix")
	if prefix == "" {
		prefix = r.URL.Query().Get("tags")
	}
	page := parseInt(r.URL.Query().Get("page"), 1)
	pageSize := parseInt(r.URL.Query().Get("pageSize"), 10)
	resp, err := s.client.ListTags(r.Context(), prefix, page, pageSize)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	for _, tag := range resp.Tags {
		name := html.EscapeString(tag.Name)
		fmt.Fprintf(w, "<button type=\"button\" class=\"tag-suggestion\" data-tag=\"%s\">%s</button>", name, name)
	}
}
