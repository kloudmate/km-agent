FROM golang:alpine AS buildstage
RUN mkdir /app
COPY . /app
WORKDIR /app

RUN apk add --no-cache ca-certificates openssl && \
    rm -rf /var/lib/apk/lists/*
    # OPTIONAL: Modify if you need to ensure CA certificates are fully up-to-date,
    # update-ca-certificates is still relevant, but often handled by install.

RUN GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -tags k8s -o configupdater cmd/configupdater/main.go

FROM alpine:latest

# for handling ssl/tls connections
COPY --from=buildstage /etc/ssl/certs /etc/ssl/certs
COPY --from=buildstage /app/configupdater ./configupdater


RUN chmod +x configupdater
ENTRYPOINT ["./configupdater", "run"]