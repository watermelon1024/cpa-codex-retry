GOOS ?= $(shell go env GOOS)
GOARCH ?= $(shell go env GOARCH)
EXT := so

ifeq ($(GOOS),darwin)
EXT := dylib
endif
ifeq ($(GOOS),windows)
EXT := dll
endif

OUT_DIR := dist/$(GOOS)/$(GOARCH)
OUT := $(OUT_DIR)/codex-retry-gateway.$(EXT)
HEADER := $(OUT_DIR)/codex-retry-gateway.h

.PHONY: build test clean

build:
	@mkdir -p $(OUT_DIR)
	go build -buildmode=c-shared -o $(OUT) ./cmd/codex-retry-gateway-plugin
	@rm -f $(HEADER)

test:
	go test ./...

clean:
	rm -rf dist
