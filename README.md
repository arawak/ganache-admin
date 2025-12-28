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

## Official Docker image

A lightweight distroless image is published to GitHub Container Registry for every tagged release:

```
docker pull ghcr.io/arawak/ganache-admin:latest
```

- Tags include `latest`, the semantic version (e.g., `1.2.3`), and the exact git tag (e.g., `v1.2.3`).
- The image runs as a non-root user and needs access to `users.yaml` plus Ganache credentials.
- Mount your config directory and pass the required environment variables when running locally:

```bash
docker run --rm \
  -e GANACHE_BASE_URL="http://host.docker.internal:8081" \
  -e GANACHE_API_KEY="changeme" \
  -e UI_SESSION_SECRET="dev-session-secret" \
  -e UI_CSRF_SECRET="dev-csrf-secret" \
  -e UI_LISTEN_ADDR=":8080" \
  -v $(pwd)/users.yaml:/config/users.yaml:ro \
  -p 8080:8080 \
  ghcr.io/arawak/ganache-admin:latest
```

Expose the service behind TLS and set `UI_SECURE_COOKIE=true` to send Secure cookies in production.

## Running with Docker Compose

A sample `docker-compose.yml` is included in this repo and expects two local files:

- `.env` for environment variables (use `.env.example` as a starting point)
- `users.yaml` for local auth (mounted read-only)

```bash
cp .env.example .env
# edit .env

docker compose up -d

docker compose logs -f ganache-admin
```

Stop the stack:

```bash
docker compose down
```

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

## .env variables

Use `.env.example` as a starting point.

For production, generate strong secrets (example):

```bash
openssl rand -hex 32
```

Example `.env`:

```
UI_LISTEN_ADDR=:8080
# UI_USERS_FILE=./users.yaml        # default for `go run` (optional)
# UI_USERS_FILE=/config/users.yaml  # default in Docker image
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
