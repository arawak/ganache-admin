# Ganache Media Admin UI

Server-side rendered admin UI for the Ganache media service. It uses Go + `html/template` with HTMX partials to search, upload, and edit assets. Authentication is handled locally via `users.yaml` (bcrypt hashes) with in-memory sessions and CSRF protection; Ganache API calls stay server-side so the API key never reaches the browser.

## Features
- Login with username/password from `users.yaml`; session cookie with 12h TTL
- CSRF token on all mutating requests; SameSite Lax cookies
- Search/browse assets with HTMX results, sorting, and paging
- Upload images via file input or clipboard paste; 25MB max
- Edit metadata inline; tag autocomplete powered by Ganache `/api/tags`
- Copy variant URLs (thumb/content/original) from the detail page

## Prerequisites
- Go toolchain (Go 1.20+)
- Ganache API base URL and API key
- `users.yaml` file with bcrypt password hashes (see below)

## Running locally
1. Copy `.env.example` to `.env` and fill values. Environment variables override `.env`.
2. Create `users.yaml` (format below) and generate bcrypt hashes with the CLI helper.
3. Run the server: `go run ./cmd/ganache-admin-ui` (defaults to `:8080`).
4. Visit `http://localhost:8080/login` and sign in with a user from `users.yaml`.

Static assets and templates are under `web/`; Ganache requests are made server-to-server with the `X-Api-Key` header and request timeouts.

## users.yaml format
```yaml
users:
  - username: admin
    passwordHash: "$2a$12$..."
```

## CLI helper (bcrypt hashes)
Generate a hash (reads password from stdin):

```bash
go run ./cmd/ganache-admin-cli hashpw
```

Verify a password against an existing hash:

```bash
go run ./cmd/ganache-admin-cli verify "$2a$12$example..."
```

## Example .env
```
UI_LISTEN_ADDR=:8080
UI_USERS_FILE=./users.yaml
UI_SESSION_SECRET=dev-session-secret
UI_CSRF_SECRET=dev-csrf-secret
GANACHE_BASE_URL=http://localhost:8081
GANACHE_API_KEY=changeme
GANACHE_TIMEOUT=10s
```

## Security notes
- Ganache API key is only used in server-to-server requests and is not exposed to templates or JavaScript.
- Session cookies are HttpOnly and SameSite=Lax; set `UI_SECURE_COOKIE=true` or run behind TLS to send the Secure flag.
- CSRF tokens are required for POST/PATCH/DELETE routes (HTMX uses the hidden input in forms).
- Uploads are capped at 25MB before forwarding to Ganache.
