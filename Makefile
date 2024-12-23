.PHONY: build run setup-config create-config-dir

create-config-dir:
	@echo Setting Up Config Directory.
	@mkdir -p ${HOME}/.kloudmate

setup-config: create-config-dir
	@echo Copying default configuration.
	@rsync -a configs/default.yaml ${HOME}/.kloudmate/agent-config.yaml

build: setup-config
	@go build cmd/kmagent/main.go

run: build
	@./main