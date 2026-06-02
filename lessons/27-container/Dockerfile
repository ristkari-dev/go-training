# syntax=docker/dockerfile:1

# ---- build stage: full Go toolchain ----
FROM golang:1.23 AS build
WORKDIR /src

# Cache module downloads first (this layer only changes when deps change).
COPY go.mod go.sum ./
RUN go mod download

# Then the source. The build context is the repo root; .dockerignore trims it.
COPY . .

ARG VERSION=docker
ARG COMMIT=none
ARG DATE=unknown
ENV CGO_ENABLED=0
RUN go build \
    -ldflags "-X github.com/ristkari-dev/go-training/lessons/27-container/solutions/warmup/buildinfo.Version=${VERSION} \
              -X github.com/ristkari-dev/go-training/lessons/27-container/solutions/warmup/buildinfo.Commit=${COMMIT} \
              -X github.com/ristkari-dev/go-training/lessons/27-container/solutions/warmup/buildinfo.Date=${DATE}" \
    -o /out/logstatsd ./lessons/27-container/solutions/cmd/logstatsd

# ---- final stage: tiny, non-root, no shell ----
FROM gcr.io/distroless/static:nonroot
COPY --from=build /out/logstatsd /logstatsd
USER nonroot:nonroot
EXPOSE 8080 9090
ENTRYPOINT ["/logstatsd"]
