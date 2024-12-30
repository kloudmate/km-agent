.PHONY: build build-debian build-debian-binary run setup-config create-config-dir

create-config-dir:
	@echo Setting Up Config Directory.
	@mkdir -p /var/kloudmate

setup-config: create-config-dir
	@echo Copying default configuration.
	@rsync -a configs/default.yaml /var/kloudmate/agent-config.yaml

build: setup-config
	@go build cmd/kmagent/main.go

build-debian-binary:
	@go build -o km-agent/usr/local/bin/km-agent cmd/kmagent/main.go

build-debian: build-debian-binary
	@dpkg-deb --build --nocheck km-agent

run: build
	@./main