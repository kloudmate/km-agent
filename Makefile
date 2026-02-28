SHELL := /bin/bash

APP_NAME := kmagent
VERSION ?= 1.0.0
BUILD_DIR := dist
SRC_DIR := ./cmd/$(APP_NAME)
SCRIPT_DIR := ./build/linux/scripts

GOOS_LIST := linux
GOARCH ?= amd64
ifeq ($(GOARCH),arm64)
    RPM_ARCH := aarch64
    DEB_ARCH := arm64
else
    RPM_ARCH := x86_64
    DEB_ARCH := amd64
endif
LD_FLAGS="-s -w -X main.version=$(VERSION)"

ISS_FILE := ./build/windows/installer.iss
WINDOWS_BUILD_DIR := $(BUILD_DIR)/win
ISS_FILE_PATH := $(WINDOWS_BUILD_DIR)/installer.iss
WINDOWS_EXE_HOST_PATH := $(WINDOWS_BUILD_DIR)/$(APP_NAME).exe

# --- Docker Specific ---
INNO_IMAGE := amake/innosetup:64bit-bookworm
CONTAINER_WORKDIR := /work

# FIX 1: Define a variable for the colon character to prevent 'make' from misinterpreting it.
COLON := :

# Get current user/group ID for Docker volume permissions (Linux/macOS)
CURRENT_UID := $(shell id -u)
CURRENT_GID := $(shell id -g)

DOCKER_USER_FLAG := --user $(CURRENT_UID):$(CURRENT_GID)

DOCKER_RUN_INNO_ARGS := --rm -v $(PWD):$(CONTAINER_WORKDIR) -w $(CONTAINER_WORKDIR) $(INNO_IMAGE)


.PHONY: clean build build-linux-amd64 build-windows package-linux-deb package-linux-rpm package-windows build-installer

clean:
	rm -rf $(BUILD_DIR)

build: build-linux

build-linux:
	@echo ">>> Building $(APP_NAME) for Linux $(GOARCH)..."
	mkdir -p $(BUILD_DIR)/linux/$(GOARCH)
	GOOS=linux GOARCH=$(GOARCH) CGO_ENABLED=0 go build -tags linux -ldflags=${LD_FLAGS} -o $(BUILD_DIR)/linux/$(GOARCH)/$(APP_NAME) $(SRC_DIR)

build-windows:
	@echo ">>> Building $(APP_NAME) for Windows..."
	mkdir -p $(BUILD_DIR)/win
	GOOS=windows CGO_ENABLED=0 go build -tags windows -ldflags=${LD_FLAGS} -o $(BUILD_DIR)/win/$(APP_NAME).exe $(SRC_DIR)
	@echo ">>> Windows executable built at $(WINDOWS_EXE_HOST_PATH)"

package-linux-deb: build-linux
	@echo "Packaging .deb for $(DEB_ARCH)..."
	mkdir -p $(BUILD_DIR)/linux/deb-$(DEB_ARCH)/DEBIAN
	mkdir -p $(BUILD_DIR)/linux/deb-$(DEB_ARCH)/usr/bin
	mkdir -p $(BUILD_DIR)/linux/deb-$(DEB_ARCH)/lib/systemd/system
	mkdir -p $(BUILD_DIR)/linux/deb-$(DEB_ARCH)/etc/$(APP_NAME)

	cp $(BUILD_DIR)/linux/$(GOARCH)/$(APP_NAME) $(BUILD_DIR)/linux/deb-$(DEB_ARCH)/usr/bin/
	cp build/linux/kmagent.service $(BUILD_DIR)/linux/deb-$(DEB_ARCH)/lib/systemd/system/
	cp configs/host-col-config.yaml $(BUILD_DIR)/linux/deb-$(DEB_ARCH)/etc/$(APP_NAME)/config.yaml

	# Replace version and copy modified control file
	sed -e "s|^Version:.*|Version: ${VERSION}|" -e "s|^Architecture:.*|Architecture: ${DEB_ARCH}|" build/linux/deb/control > "$(BUILD_DIR)/linux/deb-$(DEB_ARCH)/DEBIAN/control"

	# Copy DEBIAN control files
	cp $(SCRIPT_DIR)/preinst $(BUILD_DIR)/linux/deb-$(DEB_ARCH)/DEBIAN/
	cp $(SCRIPT_DIR)/postinst $(BUILD_DIR)/linux/deb-$(DEB_ARCH)/DEBIAN/
	cp $(SCRIPT_DIR)/prerm $(BUILD_DIR)/linux/deb-$(DEB_ARCH)/DEBIAN/
	cp $(SCRIPT_DIR)/postrm $(BUILD_DIR)/linux/deb-$(DEB_ARCH)/DEBIAN/

	# Ensure scripts are executable
	chmod 755 $(BUILD_DIR)/linux/deb-$(DEB_ARCH)/DEBIAN/preinst
	chmod 755 $(BUILD_DIR)/linux/deb-$(DEB_ARCH)/DEBIAN/postinst
	chmod 755 $(BUILD_DIR)/linux/deb-$(DEB_ARCH)/DEBIAN/prerm
	chmod 755 $(BUILD_DIR)/linux/deb-$(DEB_ARCH)/DEBIAN/postrm

	dpkg-deb --build $(BUILD_DIR)/linux/deb-$(DEB_ARCH) $(BUILD_DIR)/$(APP_NAME)_$(VERSION)_$(DEB_ARCH).deb

package-linux-rpm: build-linux
	@echo "Packaging .rpm for $(RPM_ARCH) (requires rpmbuild)..."
	mkdir -p $(BUILD_DIR)/rpm-$(RPM_ARCH)/{BUILD,RPMS,SOURCES,SPECS,SRPMS}
	cp $(BUILD_DIR)/linux/$(GOARCH)/$(APP_NAME) $(BUILD_DIR)/rpm-$(RPM_ARCH)/SOURCES/
	cp build/linux/kmagent.service $(BUILD_DIR)/rpm-$(RPM_ARCH)/SOURCES/
	RPM_SAFE_VER=$$(echo "$(VERSION)" | tr '-' '_'); \
	sed -e "s|^BuildArch:.*|BuildArch: $(RPM_ARCH)|" \
	    -e "s|^Version:.*|Version: $$RPM_SAFE_VER|" \
	    build/linux/rpm/kmagent.spec > $(BUILD_DIR)/rpm-$(RPM_ARCH)/SPECS/kmagent.spec

	cp $(SCRIPT_DIR)/postinst $(BUILD_DIR)/rpm-$(RPM_ARCH)/SOURCES/postinst
	cp configs/host-col-config.yaml $(BUILD_DIR)/rpm-$(RPM_ARCH)/SOURCES/config.yaml

	docker run --rm \
		--platform linux/amd64 \
		-v $(PWD)/$(BUILD_DIR)/rpm-$(RPM_ARCH):/rpm \
		fedora:latest \
		bash -c "dnf install -y rpm-build 2>&1 | tail -1 && \
		         rpmbuild --define '_topdir /rpm' \
		                  --target $(RPM_ARCH) \
		                  -bb /rpm/SPECS/kmagent.spec"

build-installer: build-windows
	mkdir -p $(WINDOWS_BUILD_DIR)
	chmod 777 $(WINDOWS_BUILD_DIR)
	cp $(ISS_FILE) $(ISS_FILE_PATH)
	cp -r ./build/windows/assets $(BUILD_DIR)/win/assets
	cp ./configs/host-col-config.yaml $(BUILD_DIR)/win/host-col-config.yaml
	@echo ">>> Compiling Windows Installer using Docker ($(INNO_IMAGE))..."
	# Run the Inno Setup compiler (iscc) inside the container
	# NOTE: We do NOT specify 'iscc' here, as it's the container's entrypoint.
	docker run $(DOCKER_RUN_INNO_ARGS) \
        		/dMyAppVersion=$(VERSION) "$(ISS_FILE_PATH)"
	@echo ">>> Installer compilation finished."

package-windows: build-installer
	@echo ">>> Windows installer package created."