.PHONY: all build package

PACKAGE_NAME:="go-nss-cache"
PACKAGE_VERSION:="0.0.1"
PACKAGE_DESCRIPTION:="NSS cache sync daemon"
MAINTAINER:="Xuanzhong Wei"

BUILD_DIR:="pkg/"
SOURCES:=\
	${GOPATH}/bin/go-nss-cache=/usr/sbin/ \
	authorized-keys-command=/usr/sbin/

FPM_FLAGS:=\
	-n ${PACKAGE_NAME} \
	-v ${PACKAGE_VERSION} \
	-m ${MAINTAINER} \
	--description ${PACKAGE_DESCRIPTION} \
	-p ${BUILD_DIR} \
	-a native -s dir \
	-d "libnss-cache" \
	-d "openssh-server"

all: build

build:
	@go get ./...

deb: build
	@mkdir -p ${BUILD_DIR}
	@fpm -t deb ${FPM_FLAGS} ${SOURCES}

rpm: build
	@mkdir -p ${BUILD_DIR}
	@fpm -t rpm ${FPM_FLAGS} ${SOURCES} etc/systemd/system/go-nss-cache.service

clean:
	@rm -rf ${BUILD_DIR}
	@go clean -i ./...
