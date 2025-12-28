# Build stage
FROM golang:1.25-alpine AS builder
WORKDIR /src

ARG VERSION=dev

COPY go.mod go.sum ./
RUN go mod download
COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags "-s -w -X main.version=${VERSION}" \
    -trimpath \
    -o /out/ganache-admin-ui ./cmd/ganache-admin-ui

RUN cp -r web /out/web

# Final image
FROM gcr.io/distroless/base-debian12:nonroot
WORKDIR /app

COPY --from=builder /out/ganache-admin-ui /app/ganache-admin-ui
COPY --from=builder /out/web /app/web

ENV UI_LISTEN_ADDR=:8080 \
    UI_USERS_FILE=/config/users.yaml \
    GANACHE_TIMEOUT=10s

VOLUME ["/config"]
EXPOSE 8080
ENTRYPOINT ["/app/ganache-admin-ui"]
