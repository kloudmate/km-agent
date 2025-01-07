# -----Primary Stage----------
FROM golang:alpine as BuildStage
RUN mkdir /app
COPY . /app
WORKDIR /app
RUN GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o kmagent cmd/kmagent/kmagent.go

# -----Secondary Stage-----
FROM alpine:latest
COPY --from=BuildStage /app/kmagent .
COPY ./configs/agent-config.yaml /var/kloudmate/agent-config.yaml

CMD ["./kmagent"]