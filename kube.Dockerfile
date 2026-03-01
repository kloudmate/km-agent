FROM golang:alpine AS buildstage
ARG TARGETARCH
ARG TARGETOS=linux
WORKDIR /app
COPY go.mod go.sum ./
COPY . .
# go.mod replace directives point to /tmp/km-ebpf and /tmp/km-ebpf/collector
COPY --from=km-ebpf . /tmp/km-ebpf
ARG VERSION=dev
ARG COMMIT_SHA=unknown
RUN --mount=type=cache,target=/root/.cache/go-build \
    GOOS=${TARGETOS} GOARCH=${TARGETARCH} CGO_ENABLED=0 go build -tags kubernetes -ldflags="-w -s -X 'main.version=$VERSION' -X 'main.commit=$COMMIT_SHA'" -o kmagent ./cmd/kubeagent/main.go

FROM alpine:latest
COPY --from=buildstage /app/kmagent ./kmagent
EXPOSE 4317 4318
RUN chmod +x kmagent
ENTRYPOINT ["./kmagent"]
