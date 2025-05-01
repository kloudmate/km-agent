FROM golang:alpine AS buildstage
RUN mkdir /app
COPY . /app
WORKDIR /app
RUN GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o kmagent cmd/kmagent/main.go

FROM alpine:latest
COPY --from=buildstage /app/kmagent ./kmagent
COPY ./configs/agent-config.yaml /var/kloudmate/agent-config.yaml
COPY ./configs/docker-col-config.yaml /var/kloudmate/docker-col-config.yaml

RUN chmod +x kmagent
ENTRYPOINT ["./kmagent", "--docker-mode", "--config", "/var/kloudmate/docker-col-config.yaml", "run"]