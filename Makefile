MODULE_NAME=pam-werbot
VERSION=$(shell git describe --tags --always)

build:
	@echo "building ${MODULE_NAME} ${MODULE_NAME}"
	go build -buildmode=c-shared -o ${MODULE_NAME}.so

install:
	sudo cp ${MODULE_NAME}.so /lib/security/
	sudo chmod +x /lib/security/${MODULE_NAME}.so
	sudo systemctl restart sshd

clean:
	sudo rm -f ${MODULE_NAME}.so ${MODULE_NAME}.h

test-install:
	sudo sh -c "echo 'auth required ${PWD}/${MODULE_NAME}.so' > /etc/pam.d/${MODULE_NAME}"

test-uninstall:
	sudo rm -f /etc/pam.d/${MODULE_NAME}

test:
	pamtester ${MODULE_NAME} test authenticate

.PHONY: build install clean