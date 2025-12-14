package config

import (
	"crypto/rand"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/joho/godotenv"
)

const defaultListenAddr = ":8080"
const defaultUsersFile = "./users.yaml"
const defaultTimeout = 10 * time.Second
const secretLength = 32

type GanacheConfig struct {
	BaseURL string
	APIKey  string
	Timeout time.Duration
}

type Config struct {
	ListenAddr    string
	UsersFile     string
	SessionSecret []byte
	CSRFSecret    []byte
	Ganache       GanacheConfig
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	listenAddr := valueOrDefault("UI_LISTEN_ADDR", defaultListenAddr)
	usersFile := valueOrDefault("UI_USERS_FILE", defaultUsersFile)

	ganacheBase := os.Getenv("GANACHE_BASE_URL")
	if ganacheBase == "" {
		return nil, errors.New("GANACHE_BASE_URL is required")
	}
	ganacheKey := os.Getenv("GANACHE_API_KEY")
	if ganacheKey == "" {
		return nil, errors.New("GANACHE_API_KEY is required")
	}

	timeoutStr := valueOrDefault("GANACHE_TIMEOUT", defaultTimeout.String())
	timeout, err := time.ParseDuration(timeoutStr)
	if err != nil {
		return nil, fmt.Errorf("invalid GANACHE_TIMEOUT: %w", err)
	}

	sessionSecret, err := readSecret("UI_SESSION_SECRET")
	if err != nil {
		return nil, err
	}
	csrfSecret, err := readSecret("UI_CSRF_SECRET")
	if err != nil {
		return nil, err
	}

	return &Config{
		ListenAddr:    listenAddr,
		UsersFile:     usersFile,
		SessionSecret: sessionSecret,
		CSRFSecret:    csrfSecret,
		Ganache: GanacheConfig{
			BaseURL: ganacheBase,
			APIKey:  ganacheKey,
			Timeout: timeout,
		},
	}, nil
}

func valueOrDefault(key, def string) string {
	val := os.Getenv(key)
	if val == "" {
		return def
	}
	return val
}

func readSecret(key string) ([]byte, error) {
	val := os.Getenv(key)
	if val != "" {
		return []byte(val), nil
	}
	buf := make([]byte, secretLength)
	_, err := rand.Read(buf)
	if err != nil {
		return nil, fmt.Errorf("generate secret: %w", err)
	}
	return buf, nil
}
