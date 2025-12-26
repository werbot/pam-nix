.PHONY: build install clean build-all help

BIN_NAME=wpam
VERSION=$(shell git describe --tags --always 2>/dev/null || echo "dev")

# Detect OS and set PAM directory
UNAME_S := $(shell uname -s 2>/dev/null || echo "Linux")
ifeq ($(UNAME_S),Linux)
	PAM_DIR=/lib/security
	GOOS=linux
endif
ifeq ($(UNAME_S),Darwin)
	PAM_DIR=/usr/lib/pam
	GOOS=darwin
endif
ifeq ($(UNAME_S),FreeBSD)
	PAM_DIR=/usr/lib
	GOOS=freebsd
endif
ifeq ($(UNAME_S),OpenBSD)
	PAM_DIR=/usr/lib
	GOOS=openbsd
endif
ifeq ($(UNAME_S),NetBSD)
	PAM_DIR=/usr/lib/security
	GOOS=netbsd
endif

PAM_MODULE=$(PAM_DIR)/pam_wpam.so

# Build for current platform
build:
	@echo "Building ${BIN_NAME} ${VERSION} for $(GOOS)/$(shell go env GOARCH)"
	@echo "PAM directory: $(PAM_DIR)"
	GOOS=$(GOOS) go build -buildmode=c-shared -o ${BIN_NAME}.so

# Build for specific platform
build-linux:
	@echo "Building ${BIN_NAME} ${VERSION} for linux/amd64"
	@echo "Note: CGO_ENABLED=1 is required for cross-compilation with CGO"
	CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -buildmode=c-shared -o ${BIN_NAME}-linux-amd64.so

build-linux-arm64:
	@echo "Building ${BIN_NAME} ${VERSION} for linux/arm64"
	@echo "Note: CGO_ENABLED=1 is required for cross-compilation with CGO"
	CGO_ENABLED=1 GOOS=linux GOARCH=arm64 go build -buildmode=c-shared -o ${BIN_NAME}-linux-arm64.so

build-freebsd:
	@echo "Building ${BIN_NAME} ${VERSION} for freebsd/amd64"
	@echo "Note: CGO_ENABLED=1 is required for cross-compilation with CGO"
	CGO_ENABLED=1 GOOS=freebsd GOARCH=amd64 go build -buildmode=c-shared -o ${BIN_NAME}-freebsd-amd64.so

build-openbsd:
	@echo "Building ${BIN_NAME} ${VERSION} for openbsd/amd64"
	@echo "Note: CGO_ENABLED=1 is required for cross-compilation with CGO"
	CGO_ENABLED=1 GOOS=openbsd GOARCH=amd64 go build -buildmode=c-shared -o ${BIN_NAME}-openbsd-amd64.so

build-netbsd:
	@echo "Building ${BIN_NAME} ${VERSION} for netbsd/amd64"
	@echo "Note: CGO_ENABLED=1 is required for cross-compilation with CGO"
	CGO_ENABLED=1 GOOS=netbsd GOARCH=amd64 go build -buildmode=c-shared -o ${BIN_NAME}-netbsd-amd64.so

# Build for all supported platforms
# Note: Cross-compilation with CGO requires cross-compilers or Docker
build-all:
	@echo "Building for all supported platforms..."
	@echo "Note: Some platforms may fail if cross-compilers are not available"
	@echo "For reliable builds, use Docker or build on target platforms"
	-@$(MAKE) build-linux || echo "Failed to build for linux/amd64"
	-@$(MAKE) build-linux-arm64 || echo "Failed to build for linux/arm64"
	-@$(MAKE) build-freebsd || echo "Failed to build for freebsd/amd64"
	-@$(MAKE) build-openbsd || echo "Failed to build for openbsd/amd64"
	-@$(MAKE) build-netbsd || echo "Failed to build for netbsd/amd64"
	@echo ""
	@echo "Build summary:"
	@ls -lh ${BIN_NAME}-*.so 2>/dev/null || echo "No cross-compiled binaries found"
	@echo ""
	@echo "To build reliably, use Docker or build on target platforms"

install: build
	@echo "Installing ${BIN_NAME} to ${PAM_MODULE}"
	@test -f ${BIN_NAME}.so || (echo "Error: ${BIN_NAME}.so not found. Run 'make build' first." && exit 1)
	sudo cp ${BIN_NAME}.so ${PAM_MODULE}
	sudo chmod 644 ${PAM_MODULE}
	@echo "Module installed. Restart SSH service: sudo systemctl restart sshd"

clean:
	go clean
	-rm -f ${BIN_NAME}.so ${BIN_NAME}.h
	-rm -f ${BIN_NAME}-*.so

help:
	@echo "Available targets:"
	@echo "  build          - Build for current platform"
	@echo "  build-linux    - Build for Linux amd64"
	@echo "  build-linux-arm64 - Build for Linux arm64"
	@echo "  build-freebsd  - Build for FreeBSD amd64"
	@echo "  build-openbsd  - Build for OpenBSD amd64"
	@echo "  build-netbsd  - Build for NetBSD amd64"
	@echo "  build-all      - Build for all supported platforms"
	@echo "  install        - Build and install for current platform"
	@echo "  clean          - Remove build artifacts"
	@echo "  help           - Show this help message"
