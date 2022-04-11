
VERSION := 0.1.0
BUILD := $(shell git rev-parse --short HEAD)
BRANCH ?= `git rev-parse --abbrev-ref HEAD`
DATE := `date "+%Y%m%d%H%M%S"`
PROJECT = k8sync
GROUP := lichen
TARGETS := k8sync
GO = go
GO_SRC = $(shell find . -type f -name '*.go')
PROTO_SRC = $(shell find . -type f -name '*.proto')
IMAGE_NAME := k8sync
REGISTRY_ADDRESS ?= registry-hz.rubikstack.com
IMAGE_VERSION ?= $(IMAGE_NAME):$(BRANCH)-$(BUILD)-$(DATE)
IMAGE_FULLNAME ?= $(REGISTRY_ADDRESS)/$(GROUP)/$(IMAGE_VERSION)

VER_PKG = $(PROJECT)/cmd
LDFLAGS += -X "$(VER_PKG).Version=$(VERSION)"
LDFLAGS += -X "$(VER_PKG).BuildTS=$(shell date -u '+%Y-%m-%d %I:%M:%S')"
LDFLAGS += -X "$(VER_PKG).GitHash=$(shell git rev-parse HEAD)"
LDFLAGS += -X "$(VER_PKG).GitBranch=$(shell git rev-parse --abbrev-ref HEAD)"

.PHONY: all build image gen image-fullname image-version swagger buf-lint go-lint test clean
all: swagger go-lint build

build: $(TARGETS)

$(TARGETS): $(GO_SRC) $(PROTO_SRC)
	$(GO) build -o $@ -ldflags '$(LDFLAGS)' $(PROJECT)

swagger: buf-lint gen
	@cp api/k8sync/v1/k8sync.swagger.json third_party/OpenAPI/

gen:
	buf generate --path api

buf-lint:
	buf lint

go-lint:
	golangci-lint run --deadline=5m

image:
	docker build . -t $(IMAGE_FULLNAME)
#	docker build --build-arg https_proxy=http://192.168.20.61:7890 --build-arg http_proxy=http://192.168.20.61:7890 . -t $(IMAGE_FULLNAME)

image-fullname:
	@echo $(IMAGE_FULLNAME)

image-version:
	@echo $(IMAGE_VERSION)

clean:
	rm -f $(TARGETS)
	rm -f api/k8sync/v1/*.go
