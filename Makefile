APP_NAME := glsimulator

.PHONY: all build clean test

all: build

configure:
	@go install

build: configure
	@CGO_ENABLED=0 GOOS=linux go build -o $(APP_NAME) -v -ldflags="-s -w"

clean:
	@rm -f $(APP_NAME)

lint:
	@golangci-lint run
