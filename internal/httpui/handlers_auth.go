package httpui

import (
	"net/http"
	"strings"

	"ganache-admin-ui/internal/auth"
)

func (s *Server) showLogin(w http.ResponseWriter, r *http.Request) {
	data := TemplateData{Title: "Login"}
	s.render(w, "login.html", data, r)
}

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "invalid form", http.StatusBadRequest)
		return
	}
	username := strings.TrimSpace(r.FormValue("username"))
	password := r.FormValue("password")
	if !s.users.Validate(username, password) {
		data := TemplateData{Title: "Login", Error: "Invalid credentials"}
		s.render(w, "login.html", data, r)
		return
	}
	sess, err := s.sessions.Create(username)
	if err != nil {
		http.Error(w, "unable to create session", http.StatusInternalServerError)
		return
	}
	secure := secureCookie()
	if r.TLS != nil {
		secure = true
	}
	auth.SetSessionCookie(w, sess, secure, s.cookiePath())
	http.Redirect(w, r, s.path("/assets"), http.StatusFound)
}

func (s *Server) handleLogout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session")
	if err == nil {
		s.sessions.Delete(cookie.Value)
	}
	auth.ClearSessionCookie(w, s.cookiePath())
	http.Redirect(w, r, s.path("/login"), http.StatusFound)
}
