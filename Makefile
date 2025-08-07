SHELL := /bin/bash

APP_NAME := kmagent
VERSION ?= 1.0.0
BUILD_DIR := dist
SRC_DIR := ./cmd/$(APP_NAME)
SCRIPT_DIR := ./build/linux/scripts

GOOS_LIST := linux
GOARCH := amd64
LD_FLAGS="-s -w"

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

build: build-linux-amd64

build-linux-amd64:
	@echo ">>> Building $(APP_NAME) for Linux AMD64..."
	mkdir -p $(BUILD_DIR)/linux/amd64
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags=${LD_FLAGS} -o $(BUILD_DIR)/linux/amd64/$(APP_NAME) $(SRC_DIR)

build-windows:
	@echo ">>> Building $(APP_NAME) for Windows..."
	mkdir -p $(BUILD_DIR)/win
	GOOS=windows CGO_ENABLED=0 go build -ldflags=${LD_FLAGS} -o $(BUILD_DIR)/win/$(APP_NAME).exe $(SRC_DIR)
	@echo ">>> Windows executable built at $(WINDOWS_EXE_HOST_PATH)"

package-linux-deb: build-linux-amd64
	@echo "Packaging .deb..."
	mkdir -p $(BUILD_DIR)/linux/deb/DEBIAN
	mkdir -p $(BUILD_DIR)/linux/deb/usr/bin
	mkdir -p $(BUILD_DIR)/linux/deb/lib/systemd/system
	mkdir -p $(BUILD_DIR)/linux/deb/etc/$(APP_NAME)

	cp $(BUILD_DIR)/linux/$(GOARCH)/$(APP_NAME) $(BUILD_DIR)/linux/deb/usr/bin/
	cp build/linux/kmagent.service $(BUILD_DIR)/linux/deb/lib/systemd/system/
	cp configs/host-col-config.yaml $(BUILD_DIR)/linux/deb/etc/$(APP_NAME)/config.yaml

	# Replace version and copy modified control file
	sed "s|^Version:.*|Version: ${VERSION}|" build/linux/deb/control > "$(BUILD_DIR)/linux/deb/DEBIAN/control"

	# Copy DEBIAN control files
	cp $(SCRIPT_DIR)/preinst $(BUILD_DIR)/linux/deb/DEBIAN/
	cp $(SCRIPT_DIR)/postinst $(BUILD_DIR)/linux/deb/DEBIAN/
	cp $(SCRIPT_DIR)/prerm $(BUILD_DIR)/linux/deb/DEBIAN/
	cp $(SCRIPT_DIR)/postrm $(BUILD_DIR)/linux/deb/DEBIAN/

	# Ensure scripts are executable
	chmod 755 $(BUILD_DIR)/linux/deb/DEBIAN/preinst
	chmod 755 $(BUILD_DIR)/linux/deb/DEBIAN/postinst
	chmod 755 $(BUILD_DIR)/linux/deb/DEBIAN/prerm
	chmod 755 $(BUILD_DIR)/linux/deb/DEBIAN/postrm

	dpkg-deb --build $(BUILD_DIR)/linux/deb $(BUILD_DIR)/$(APP_NAME)_$(VERSION)_amd64.deb

package-linux-rpm: build-linux-amd64
	@echo "Packaging .rpm (requires rpmbuild)..."
	mkdir -p $(BUILD_DIR)/rpm/{BUILD,RPMS,SOURCES,SPECS,SRPMS}
	cp $(BUILD_DIR)/linux/$(GOARCH)/$(APP_NAME) $(BUILD_DIR)/rpm/SOURCES/
	cp build/linux/kmagent.service $(BUILD_DIR)/rpm/SOURCES/
	cp build/linux/rpm/kmagent.spec $(BUILD_DIR)/rpm/SPECS/

	cp $(SCRIPT_DIR)/postinst $(BUILD_DIR)/rpm/SOURCES/postinst
	cp configs/host-col-config.yaml $(BUILD_DIR)/rpm/SOURCES/config.yaml

	rpmbuild --define "_topdir $(PWD)/$(BUILD_DIR)/rpm" \
			 --define "version $(VERSION)" \
			 -bb $(BUILD_DIR)/rpm/SPECS/kmagent.spec

# --- Windows Installer Packaging ---

build-installer: build-windows
	mkdir -p $(WINDOWS_BUILD_DIR)
	cp $(ISS_FILE) $(ISS_FILE_PATH)
	@echo ">>> Compiling Windows Installer using Docker ($(INNO_IMAGE))..."
	# Run the Inno Setup compiler (iscc) inside the container
	# NOTE: We do NOT specify 'iscc' here, as it's the container's entrypoint.
	docker run $(DOCKER_RUN_INNO_ARGS) \
        		/dMyAppVersion=$(VERSION) "$(ISS_FILE_PATH)"
	@echo ">>> Installer compilation finished."


# FIX 5: Removed the redundant `package-windows: build-windows` definition.
# This target now correctly depends on `build-installer`.
package-windows: build-installer
	@echo ">>> Windows installer package created."