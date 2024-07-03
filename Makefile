.DEFAULT_GOAL := all
IMG ?= virtuallan:latest
CONTAINER_TOOL ?= docker

.PHONY: all
all: build build-windows

.PHONY: gen
gen:
	go generate pkg/cipher/cipher.go

.PHONY: build
build: gen
	go build -o virtuallan main.go

.PHONY: clean
clean:
	rm -rf virtuallan virtuallan.exe

.PHONY: build-docker
build-docker: gen
	$(CONTAINER_TOOL) build -t ${IMG} .

.PHONY: build-windows
build-windows:
	GOOS=windows GOARCH=amd64 go build -o virtuallan.exe main.go