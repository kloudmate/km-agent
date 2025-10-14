# --- Build Stage ---
FROM golang:1.25.3-alpine3.22 AS buildstage

# Set the working directory
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .

ARG VERSION=dev
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -a -ldflags "-w -s -X 'main.version=${VERSION}'" -o /kmagent ./cmd/kmagent/...

FROM gcr.io/distroless/static-debian11

COPY --from=buildstage /kmagent /kmagent
COPY ./configs/docker-col-config.yaml /config.yaml

ENTRYPOINT ["/kmagent", "--docker-mode", "--config", "/config.yaml", "start"]