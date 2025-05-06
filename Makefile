SHELL := /bin/bash

APP_NAME := kmagent
VERSION := 1.0.0
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
INNO_IMAGE := amake/innosetup:64bit-bookworm # Use the desired amake/innosetup tag
CONTAINER_WORKDIR := /work

# Get current user/group ID for Docker volume permissions (Linux/macOS)
# For Windows (Git Bash/WSL), this usually works. For native Windows Docker, permissions might differ.
CURRENT_UID := $(shell id -u)
CURRENT_GID := $(shell id -g)

# Define user flag, default to host user
DOCKER_USER_FLAG := --user $(CURRENT_UID):$(CURRENT_GID)

# Docker run arguments for Inno Setup build
DOCKER_RUN_INNO_ARGS := --rm -v $(PWD):$(CONTAINER_WORKDIR) -w $(CONTAINER_WORKDIR) $(INNO_IMAGE)


.PHONY: clean build

clean:
	rm -rf $(BUILD_DIR)


build-linux-amd64:
	@echo ">>> Building $(APP_NAME) for Linux AMD64..."
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags=${LD_FLAGS} -o $(BUILD_DIR)/linux/amd64/$(APP_NAME) $(SRC_DIR);

build-windows:
	@echo ">>> Building $(APP_NAME) for Windows (native)..."
	GOOS=windows CGO_ENABLED=0 go build -ldflags=${LD_FLAGS} -o $(BUILD_DIR)/win/$(APP_NAME).exe $(SRC_DIR);
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

package-linux-rpm: build-linux-amd64
	@echo "Packaging .rpm (requires rpmbuild)..."
	mkdir -p $(BUILD_DIR)/rpm/{BUILD,RPMS,SOURCES,SPECS,SRPMS}
	cp $(BUILD_DIR)/linux/$(GOARCH)/$(APP_NAME) $(BUILD_DIR)/rpm/SOURCES/
	cp build/linux/kmagent.service $(BUILD_DIR)/rpm/SOURCES/
	cp build/linux/rpm/kmagent.spec $(BUILD_DIR)/rpm/SPECS/

	cp $(SCRIPT_DIR)/postinst $(BUILD_DIR)/rpm/SOURCES/postinst
	cp configs/host-col-config.yaml $(BUILD_DIR)/rpm/SOURCES/config.yaml


	rpmbuild --define "_topdir $(PWD)/$(BUILD_DIR)/rpm" -bb $(BUILD_DIR)/rpm/SPECS/kmagent.spec

# --- Windows Installer Packaging ---
package-windows: build-windows

build-installer:build-windows
	cp $(ISS_FILE) $(WINDOWS_BUILD_DIR)/installer.iss
	mkdir -p $(WINDOWS_BUILD_DIR)
	chmod 777 $(WINDOWS_BUILD_DIR)
	@echo ">>> Compiling Windows Installer using Docker ($(INNO_IMAGE))..."
	# Run the Inno Setup compiler (iscc) inside the amake/innosetup container
	docker run $(DOCKER_RUN_INNO_ARGS) \
    		"$(ISS_FILE_PATH)"

		# Add flags to iscc if needed, e.g., /OOutputdir for output path
		# Example: iscc /O$(BUILD_DIR)/win $(ISS_FILE)
	@echo ">>> Installer compilation finished."
	@echo ">>> NOTE: Output setup file is likely in ./Output directory (or as specified in ISS file)."

package-windows: build-installer
	@echo ">>> Windows installer package created in $(WINDOWS_BUILD_DIR)."