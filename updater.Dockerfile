FROM golang:1.25.3-alpine3.22 AS buildstage

# Declare build platform arguments (automatically populated by BuildKit)
ARG TARGETARCH
ARG TARGETOS=linux

WORKDIR /app

# Copy dependency files
COPY go.mod go.sum ./

# Copy source code and vendor directory
COPY . .

# Vendor dependencies and patch kube-openapi
RUN go mod vendor && \
    if [ -f "vendor/k8s.io/kube-openapi/pkg/util/proto/document_v3.go" ]; then \
        sed -i 's|"gopkg.in/yaml.v3"|"go.yaml.in/yaml/v3"|g' vendor/k8s.io/kube-openapi/pkg/util/proto/document_v3.go; \
        echo "Patched kube-openapi yaml import"; \
    fi

ARG VERSION=dev
ARG COMMIT_SHA=unknown

# Use mount cache for build cache only (vendor is already in place)
RUN --mount=type=cache,target=/root/.cache/go-build \
    GOOS=${TARGETOS} GOARCH=${TARGETARCH} CGO_ENABLED=0 go build -mod=vendor -ldflags="-w -s -X 'main.version=$VERSION' -X 'main.commit=$COMMIT_SHA'" -o configupdater ./cmd/configupdater/main.go

FROM alpine:latest
COPY --from=buildstage /app/configupdater ./configupdater
ENTRYPOINT ["./configupdater", "run"]