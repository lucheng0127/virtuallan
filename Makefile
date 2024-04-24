.DEFAULT_GOAL := build

.PHONY: gen
gen:
	go generate pkg/cipher/cipher.go

.PHONY: build
build: gen
	go build -o virtuallan main.go

.PHONY: clean
clean:
	rm -rf virtuallan
