TEST?=./...
PKG_NAME=ldap

default: install

build:
	go build
	mv terraform-provider-ldap ~/.terraform.d/plugins/

install: fmtcheck lint
	go install

# test: fmtcheck
# 	go test $(TEST) -timeout=30s -parallel=4

# testacc: fmtcheck
# 	TF_ACC=1 go test $(TEST) -v -count 1 -parallel 20 $(TESTARGS) -timeout 120m

fmt:
	@echo "==> Fixing source code with gofmt..."
	@gofmt -s -w ./$(PKG_NAME)

fmtcheck:
	@sh -c "'$(CURDIR)/scripts/gofmtcheck.sh'"

lint:
	@echo "==> Checking source code against linters..."
	@golangci-lint run ./$(PKG_NAME)/...
	@tfproviderlint \
		-c 1 \
		-AT001 \
		-AT002 \
		-AT003 \
		-AT004 \
		-R001 \
		-R002 \
		-R003 \
		-R004 \
		-S001 \
		-S002 \
		-S003 \
		-S004 \
		-S005 \
		-S006 \
		-S007 \
		-S008 \
		-S009 \
		-S010 \
		-S011 \
		-S012 \
		-S013 \
		-S014 \
		-S015 \
		-S016 \
		-S017 \
		-S018 \
		-S019 \
		./$(PKG_NAME)

tools:
	GO111MODULE=on go install github.com/bflad/tfproviderlint/cmd/tfproviderlint
	GO111MODULE=on go install github.com/golangci/golangci-lint/cmd/golangci-lint

.PHONY: default build fmt lint tools