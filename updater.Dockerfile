FROM golang:alpine AS buildstage

# Declare build platform arguments (automatically populated by BuildKit)
ARG TARGETARCH
ARG TARGETOS=linux

WORKDIR /app

# Copy dependency files first for better caching
COPY go.mod go.sum ./

# Use mount cache for go modules to speed up downloads
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

# Copy source code
COPY . .

# build arguments for version information
ARG VERSION=dev
ARG COMMIT_SHA=unknown

# Use mount cache for both go modules and build cache
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    GOOS=${TARGETOS} GOARCH=${TARGETARCH} CGO_ENABLED=0 go build -ldflags="-X 'main.version=$VERSION' -X 'main.commit=$COMMIT_SHA'" -o configupdater ./cmd/configupdater/main.go

FROM alpine:latest
COPY --from=buildstage /app/configupdater ./configupdater
ENTRYPOINT ["./configupdater", "run"]