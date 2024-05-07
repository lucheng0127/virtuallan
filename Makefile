.DEFAULT_GOAL := build
IMG ?= virtuallan:latest
CONTAINER_TOOL ?= docker

.PHONY: gen
gen:
	go generate pkg/cipher/cipher.go

.PHONY: build
build: gen
	go build -o virtuallan main.go

.PHONY: clean
clean:
	rm -rf virtuallan

.PHONY: build-docker
build-docker: gen
	$(CONTAINER_TOOL) build -t ${IMG} .