# OpenCV / ArUco integration (contrib)
PKG_CFLAGS   := $(shell pkg-config --cflags opencv4)
PKG_LIBS     := $(shell pkg-config --libs   opencv4) -lopencv_aruco

export CGO_CXXFLAGS := $(PKG_CFLAGS)
export CGO_LDFLAGS  := $(PKG_LIBS)
export CGO_ENABLED  := 1

# Standard build settings
BIN_OUTPUT_PATH = bin
TOOL_BIN        = bin/gotools/$(shell uname -s)-$(shell uname -m)
UNAME_S ?= $(shell uname -s)
GOPATH    = $(HOME)/go/bin
export PATH := ${PATH}:$(GOPATH)

build: format update-rdk
	@rm -f $(BIN_OUTPUT_PATH)/placeholder
	@go build -tags opencvstatic $(LDFLAGS) -o $(BIN_OUTPUT_PATH)/placeholder main.go

module.tar.gz: build
	@rm -f $(BIN_OUTPUT_PATH)/module.tar.gz
	@tar czf $(BIN_OUTPUT_PATH)/module.tar.gz $(BIN_OUTPUT_PATH)/placeholder

setup:
	@if [ "$(UNAME_S)" = "Linux" ]; then \
		sudo apt-get install -y apt-utils coreutils tar libnlopt-dev libjpeg-dev pkg-config; \
	fi
	# remove unused imports
	@go install golang.org/x/tools/cmd/goimports@latest
	@find . -name '*.go' -exec $(GOPATH)/goimports -w {} +

clean:
	@rm -rf $(BIN_OUTPUT_PATH)/placeholder $(BIN_OUTPUT_PATH)/module.tar.gz placeholder

format:
	@gofmt -w -s .

update-rdk:
	@go get go.viam.com/rdk@latest
	@go mod tidy
