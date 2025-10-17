# --- Build Stage ---
FROM golang:1.25.3-alpine3.22 AS buildstage

# Set the working directory
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
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -mod=vendor -a -ldflags "-w -s -X 'main.version=${VERSION}'" -o /kmagent ./cmd/kmagent/...

FROM gcr.io/distroless/static-debian11

COPY --from=buildstage /kmagent /kmagent
COPY ./configs/docker-col-config.yaml /config.yaml

ENTRYPOINT ["/kmagent", "--docker-mode", "--config", "/config.yaml", "start"]