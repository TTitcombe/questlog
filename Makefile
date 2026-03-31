BINARY := qlog
BUILD_DIR := bin
CMD := ./cmd/qlog
VERSION := $(shell git describe --tags --dirty 2>/dev/null || echo "dev")
LDFLAGS := -ldflags "-X github.com/TTitcombe/questlog/internal/cli.Version=$(VERSION)"

.PHONY: build install clean

build:
	go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY) $(CMD)

install:
	GOBIN=$(HOME)/go/bin go install $(LDFLAGS) $(CMD)

clean:
	rm -rf $(BUILD_DIR)
