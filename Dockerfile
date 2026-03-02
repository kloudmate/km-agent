FROM golang:1.25.3-alpine3.22 AS buildstage
WORKDIR /app
COPY go.mod go.sum ./
COPY . .
# go.mod replace directives point to /tmp/km-ebpf and /tmp/km-ebpf/collector
COPY --from=km-ebpf . /tmp/km-ebpf
ARG VERSION=dev
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -a -tags linux -ldflags "-w -s -X 'main.version=${VERSION}'" -o /kmagent ./cmd/kmagent/...

FROM gcr.io/distroless/static-debian11
COPY --from=buildstage /kmagent /kmagent
COPY ./configs/docker-col-config.yaml /config.yaml
EXPOSE 4317 4318
ENTRYPOINT ["/kmagent", "--docker-mode", "--config", "/config.yaml", "start"]