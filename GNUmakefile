TEST?=$$(go list ./... | grep -v 'vendor')
NAME=activedirectory
BINARY=terraform-provider-${NAME}
OS_ARCH=linux_amd64
VERSION=0.7.0
HOST=registry.terraform.io
NAMESPACE=hashicorp
ROOT_DIR:=$(shell dirname $(realpath $(firstword $(MAKEFILE_LIST))))
GO_CACHE= GOCACHE=$(ROOT_DIR)/.gocache


ifneq ("$(wildcard ./testacc.env)","")
	include testacc.env
	export $(shell sed 's/=.*//' testacc.env)
endif

default: install

build:
	$(GO_CACHE) go build -o ${BINARY}

release:
	$(GO_CACHE) GOOS=darwin GOARCH=amd64 go build -o ./bin/${BINARY}_${VERSION}_darwin_amd64
	$(GO_CACHE) GOOS=freebsd GOARCH=386 go build -o ./bin/${BINARY}_${VERSION}_freebsd_386
	$(GO_CACHE) GOOS=freebsd GOARCH=amd64 go build -o ./bin/${BINARY}_${VERSION}_freebsd_amd64
	$(GO_CACHE) GOOS=freebsd GOARCH=arm go build -o ./bin/${BINARY}_${VERSION}_freebsd_arm
	$(GO_CACHE) GOOS=linux GOARCH=386 go build -o ./bin/${BINARY}_${VERSION}_linux_386
	$(GO_CACHE) GOOS=linux GOARCH=amd64 go build -o ./bin/${BINARY}_${VERSION}_linux_amd64
	$(GO_CACHE) GOOS=linux GOARCH=arm go build -o ./bin/${BINARY}_${VERSION}_linux_arm
	$(GO_CACHE) GOOS=openbsd GOARCH=386 go build -o ./bin/${BINARY}_${VERSION}_openbsd_386
	$(GO_CACHE) GOOS=openbsd GOARCH=amd64 go build -o ./bin/${BINARY}_${VERSION}_openbsd_amd64
	$(GO_CACHE) GOOS=solaris GOARCH=amd64 go build -o ./bin/${BINARY}_${VERSION}_solaris_amd64
	$(GO_CACHE) GOOS=windows GOARCH=386 go build -o ./bin/${BINARY}_${VERSION}_windows_386
	$(GO_CACHE) GOOS=windows GOARCH=amd64 go build -o ./bin/${BINARY}_${VERSION}_windows_amd64

install: build
	mkdir -p ~/.terraform.d/plugins/${HOST}/${NAMESPACE}/${NAME}/${VERSION}/${OS_ARCH}
	mv ${BINARY} ~/.terraform.d/plugins/${HOST}/${NAMESPACE}/${NAME}/${VERSION}/${OS_ARCH}

test:
	$(GO_CACHE) go test -i $(TEST) || exit 1
	echo $(TEST) | xargs -t -n4 go test $(TESTARGS) -timeout=30s -parallel=4

testacc:
	TF_ACC=1 \
	$(GO_CACHE) go test $(TEST) -v $(TESTARGS) -timeout 120m
