FROM golang:alpine as BuildStage
RUN mkdir /app
COPY . /app
WORKDIR /app
RUN GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o kmagent cmd/kmagent/kmagent.go

FROM alpine:latest
COPY --from=BuildStage /app/kmagent ./kmagent
COPY ./configs/agent-config.yaml /var/kloudmate/agent-config.yaml
COPY ./configs/host-col-config.yaml /var/kloudmate/host-col-config.yaml
COPY ./configs/docker-col-config.yaml /var/kloudmate/docker-col-config.yaml

RUN chmod +x kmagent
ENTRYPOINT sh -c './kmagent -mode=docker start'