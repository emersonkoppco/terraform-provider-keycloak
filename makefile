GOFMT_FILES?=$$(find . -name '*.go' |grep -v vendor)
GOOS?=darwin
GOARCH?=arm64

MAKEFLAGS += --silent

TAG := $$(git describe --tags)
VERSION := $(shell echo $(TAG) | cut -c2-)

build:
	echo "TAG=$(TAG)"
	echo "VERSION=$(VERSION)"
	CGO_ENABLED=0 go build -trimpath -ldflags "-s -w -X main.version=$(TAG)" -o terraform-provider-keycloak_$(TAG)

deploy-local: build
	rm -rf ~/.terraform.d/plugins/terraform.local/emersonkoppco/keycloak/$(VERSION)/$(GOOS)_$(GOARCH)
	mkdir -p ~/.terraform.d/plugins/terraform.local/emersonkoppco/keycloak/$(VERSION)/$(GOOS)_$(GOARCH)
	mv terraform-provider-keycloak_$(TAG) ~/.terraform.d/plugins/terraform.local/emersonkoppco/keycloak/$(VERSION)/$(GOOS)_$(GOARCH)
	echo "Use provider = \"terraform.local/emersonkoppco/keycloak\" and version = \"$(VERSION)\" to test"

clean-local:
	rm -rf ~/.terraform.d/plugins/terraform.local/emersonkoppco/keycloak

build-example: build
	mkdir -p example/.terraform/plugins/terraform.local/mrparkers/keycloak/4.0.0/$(GOOS)_$(GOARCH)
	mkdir -p example/terraform.d/plugins/terraform.local/mrparkers/keycloak/4.0.0/$(GOOS)_$(GOARCH)
	cp terraform-provider-keycloak_* example/.terraform/plugins/terraform.local/mrparkers/keycloak/4.0.0/$(GOOS)_$(GOARCH)/
	cp terraform-provider-keycloak_* example/terraform.d/plugins/terraform.local/mrparkers/keycloak/4.0.0/$(GOOS)_$(GOARCH)/

local: deps
	docker compose up --build -d
	./scripts/wait-for-local-keycloak.sh
	./scripts/create-terraform-client.sh

deps:
	./scripts/check-deps.sh

fmt:
	gofmt -w -s $(GOFMT_FILES)

test: fmtcheck vet
	go test $(TEST)

testacc: fmtcheck vet
	go test -v github.com/mrparkers/terraform-provider-keycloak/keycloak
	TF_ACC=1 CHECKPOINT_DISABLE=1 go test -v -timeout 60m -parallel 4 github.com/mrparkers/terraform-provider-keycloak/provider $(TESTARGS)

fmtcheck:
	lineCount=$(shell gofmt -l -s $(GOFMT_FILES) | wc -l | tr -d ' ') && exit $$lineCount

vet:
	go vet ./...

user-federation-example:
	cd custom-user-federation-example && ./gradlew shadowJar
