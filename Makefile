.PHONY: build build-debian build-debian-binary run setup-config

setup-config:
	@echo Setting Up default configuration.
	@mkdir -p /var/kloudmate
	@rsync -a configs/agent-config.yaml /var/kloudmate/agent-config.yaml
	@rsync -a configs/host-col-config.yaml /var/kloudmate/host-col-config.yaml
	@rsync -a configs/docker-col-config.yaml /var/kloudmate/docker-col-config.yaml

build:
	@go build -o builds/bin/kmagent cmd/kmagent/kmagent.go

build-debian-binary:
	@go build -o km-agent/usr/local/bin/kmagent cmd/kmagent/kmagent.go

build-debian: build-debian-binary
	@dpkg-deb --build --nocheck km-agent

run: build
	@./builds/bin/kmagent