FROM golang:alpine AS buildstage
RUN mkdir /app
COPY . /app
WORKDIR /app

# build arguments for version information
ARG VERSION=dev
ARG COMMIT_SHA=unknown
ARG TARGETARCH
RUN GOOS=linux GOARCH=${TARGETARCH} CGO_ENABLED=0 go build -ldflags="-X 'main.version=$VERSION' -X 'main.commit=$COMMIT_SHA'" -o configupdater cmd/configupdater/main.go

FROM alpine:latest
COPY --from=buildstage /app/configupdater ./configupdater
ENTRYPOINT ["./configupdater", "run"]