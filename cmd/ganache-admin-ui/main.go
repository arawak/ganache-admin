package main

import (
	"log"
	"net/http"
	"time"

	"ganache-admin-ui/internal/auth"
	"ganache-admin-ui/internal/config"
	"ganache-admin-ui/internal/ganache"
	"ganache-admin-ui/internal/httpui"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	users, err := auth.LoadUsers(cfg.UsersFile)
	if err != nil {
		log.Fatal(err)
	}

	sessions := auth.NewSessionStore(12 * time.Hour)
	client := ganache.NewClient(cfg.Ganache.BaseURL, cfg.Ganache.APIKey, cfg.Ganache.Timeout)

	srv, err := httpui.NewServer(cfg, users, sessions, client)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("listening on %s", cfg.ListenAddr)
	if err := http.ListenAndServe(cfg.ListenAddr, srv.Router()); err != nil {
		log.Fatal(err)
	}
}
