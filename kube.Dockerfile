FROM golang:alpine AS buildstage
RUN mkdir /app
COPY . /app
WORKDIR /app
RUN GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -tags k8s -o kmagent cmd/kmagent/main_k8s.go

FROM alpine:latest
COPY --from=buildstage /app/kmagent ./kmagent

EXPOSE 4317 4318

RUN chmod +x kmagent
ENTRYPOINT ["./kmagent"]