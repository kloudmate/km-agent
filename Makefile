.PHONY: build build-debian build-debian-binary run setup-config

setup-config:
	@echo Setting Up default configuration.
	@mkdir -p /var/kloudmate
	@rsync -a configs/agent-config.yaml /var/kloudmate/agent-config.yaml
	@rsync -a configs/agent-docker-config.yaml /var/kloudmate/agent-docker-config.yaml

build:
	@go build -o builds/bin/kmagent cmd/kmagent/kmagent.go

build-debian-binary:
	@go build -o km-agent/usr/local/bin/kmagent cmd/kmagent/kmagent.go

build-debian: build-debian-binary
	@dpkg-deb --build --nocheck km-agent

run: build
	@./builds/bin/kmagent