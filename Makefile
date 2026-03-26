BINARY := qlog
BUILD_DIR := bin
CMD := ./cmd/qlog

.PHONY: build install clean

build:
	go build -o $(BUILD_DIR)/$(BINARY) $(CMD)

install:
	GOBIN=$(HOME)/go/bin go install $(CMD)

clean:
	rm -rf $(BUILD_DIR)
