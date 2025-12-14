package httpui

import (
	"bytes"
	"fmt"
	"html/template"
	"net/http"
	"path/filepath"
	"runtime"
	"strings"

	"ganache-admin-ui/internal/auth"
)

type Templates struct {
	t *template.Template
}

type TemplateData struct {
	Title   string
	User    string
	CSRF    string
	Flash   string
	Error   string
	Query   string
	Tags    []string
	Search  any
	Asset   any
	Assets  any
	Extra   map[string]any
	Content template.HTML
}

func ParseTemplates() (*Templates, error) {
	files, err := templateFiles()
	if err != nil {
		return nil, err
	}
	base := template.New("layout.html")
	t, err := base.Funcs(template.FuncMap{
		"join": func(list []string, sep string) string {
			return template.HTMLEscapeString(strings.Join(list, sep))
		},
	}).ParseFiles(files...)
	if err != nil {
		return nil, err
	}
	return &Templates{t: t}, nil
}

func templateFiles() ([]string, error) {
	_, file, _, _ := runtime.Caller(0)
	base := filepath.Dir(file)
	patterns := []string{
		filepath.Join(base, "..", "..", "web", "templates", "*.html"),
		filepath.Join("web", "templates", "*.html"),
	}
	for _, pattern := range patterns {
		matches, err := filepath.Glob(pattern)
		if err != nil {
			return nil, err
		}
		if len(matches) > 0 {
			return matches, nil
		}
	}
	return nil, fmt.Errorf("no template files found")
}

func (t *Templates) Render(w http.ResponseWriter, name string, data TemplateData, r *http.Request) {
	sess, ok := auth.SessionFromContext(r.Context())
	if ok {
		data.User = sess.Username
		data.CSRF = sess.CSRFToken
	}

	contentName := string(data.Content)
	if contentName == "" {
		fallback := strings.TrimSuffix(name, filepath.Ext(name)) + "_content"
		if t.t.Lookup(fallback) != nil {
			contentName = fallback
		} else {
			contentName = name
		}
	}

	var buf bytes.Buffer
	if err := t.t.ExecuteTemplate(&buf, contentName, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	data.Content = template.HTML(buf.String())

	if err := t.t.ExecuteTemplate(w, name, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
