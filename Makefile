BIN_NAME=pam-werbot
VERSION=$(shell git describe --tags --always)

build:
	@echo "building ${BIN_NAME} ${VERSION}"
	go build -buildmode=c-shared -o ${BIN_NAME}.so

install:
	sudo cp ${BIN_NAME}.so /lib/security/
	sudo systemctl restart sshd