SHELL := /bin/bash

APP_NAME := kmagent
VERSION := 1.0.0
BUILD_DIR := dist
SRC_DIR := ./cmd/$(APP_NAME)
SCRIPT_DIR := ./build/linux/scripts

GOOS_LIST := linux
GOARCH := amd64

.PHONY: clean build build-debian build-debian-binary run setup-config

clean:
	rm -rf $(BUILD_DIR)

build:
	@echo "Building for all platforms..."
	@for GOOS in $(GOOS_LIST); do \
		OUT_DIR=$(BUILD_DIR)/$$GOOS/$(GOARCH); \
		mkdir -p $$OUT_DIR; \
		GOOS=$$GOOS GOARCH=$(GOARCH) CGO_ENABLED=0 go build -o $$OUT_DIR/$(APP_NAME) $(SRC_DIR); \
	done

setup-config:
	@echo Setting Up default configuration.
	@mkdir -p /etc/kloudmate
	@rsync -a configs/agent-config.yaml /etc/kmagent/agent.yaml
	@rsync -a configs/host-col-config.yaml /etc/kmagent/config.yaml


package-linux-deb: build
	@echo "Packaging .deb..."
	mkdir -p $(BUILD_DIR)/linux/deb/DEBIAN
	mkdir -p $(BUILD_DIR)/linux/deb/usr/bin
	mkdir -p $(BUILD_DIR)/linux/deb/lib/systemd/system
	mkdir -p $(BUILD_DIR)/linux/deb/etc/$(APP_NAME)

	cp $(BUILD_DIR)/linux/$(GOARCH)/$(APP_NAME) $(BUILD_DIR)/linux/deb/usr/bin/
	cp build/linux/kmagent.service $(BUILD_DIR)/linux/deb/lib/systemd/system/
	cp configs/host-col-config.yaml $(BUILD_DIR)/linux/deb/etc/$(APP_NAME)/config.yaml
	cp build/linux/deb/control $(BUILD_DIR)/linux/deb/DEBIAN/

	# Copy DEBIAN control files
	cp build/linux/deb/control $(BUILD_DIR)/linux/deb/DEBIAN/
	cp $(SCRIPT_DIR)/preinst $(BUILD_DIR)/linux/deb/DEBIAN/
	cp $(SCRIPT_DIR)/postinst $(BUILD_DIR)/linux/deb/DEBIAN/
	cp $(SCRIPT_DIR)/prerm $(BUILD_DIR)/linux/deb/DEBIAN/
	cp $(SCRIPT_DIR)/postrm $(BUILD_DIR)/linux/deb/DEBIAN/

	# Ensure scripts are executable
	chmod 755 $(BUILD_DIR)/linux/deb/DEBIAN/preinst || true
	chmod 755 $(BUILD_DIR)/linux/deb/DEBIAN/postinst || true
	chmod 755 $(BUILD_DIR)/linux/deb/DEBIAN/prerm || true
	chmod 755 $(BUILD_DIR)/linux/deb/DEBIAN/postrm || true

	dpkg-deb --build $(BUILD_DIR)/linux/deb $(BUILD_DIR)/$(APP_NAME)_$(VERSION)_amd64.deb

package-linux-rpm: build
	@echo "Packaging .rpm (requires rpmbuild)..."
	mkdir -p $(BUILD_DIR)/rpm/{BUILD,RPMS,SOURCES,SPECS,SRPMS}
	cp $(BUILD_DIR)/linux/$(GOARCH)/$(APP_NAME) $(BUILD_DIR)/rpm/SOURCES/
	cp build/linux/kmagent.service $(BUILD_DIR)/rpm/SOURCES/
	cp build/linux/rpm/kmagent.spec $(BUILD_DIR)/rpm/SPECS/


#	cp $(SCRIPT_DIR)/preinst  $(BUILD_DIR)/rpm/SOURCES/preinst
	cp $(SCRIPT_DIR)/postinst $(BUILD_DIR)/rpm/SOURCES/postinst
	cp configs/host-col-config.yaml $(BUILD_DIR)/rpm/SOURCES/config.yaml
#	cp $(SCRIPT_DIR)/prerm   $(BUILD_DIR)/rpm/SOURCES/prerm
#	cp $(SCRIPT_DIR)/postrm  $(BUILD_DIR)/rpm/SOURCES/postrm

	rpmbuild --define "_topdir $(PWD)/$(BUILD_DIR)/rpm" -bb $(BUILD_DIR)/rpm/SPECS/kmagent.spec
