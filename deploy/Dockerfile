# syntax=docker/dockerfile:1.7

# --- Stage 1: build the static dist/ ---
FROM golang:1.23-alpine AS builder
WORKDIR /src

# Cache deps separately
COPY go.mod go.sum ./
RUN go mod download

# Copy everything else and build
COPY . .
RUN go build -o /out/build-index ./tools/build-index \
    && /out/build-index -lessons /src/lessons -shared /src/shared/reveal -out /dist

# --- Stage 2: serve dist/ with non-root nginx ---
FROM nginxinc/nginx-unprivileged:alpine
USER root
COPY deploy/nginx.conf.template /etc/nginx/templates/default.conf.template
COPY --from=builder /dist /usr/share/nginx/html
# Drop back to the unprivileged user the image ships with (UID 101).
USER 101
EXPOSE 8080
