.PHONY: fmt

SRC=$(shell find . -name "*.go")

bin/ff: $(SRC)
	@go build -o bin/ff main.go

build: bin/ff

run: build
	@bin/imgur

fmt: $(SRC)
	@go fmt ./cmd
	@go fmt ./utils
	@go fmt main.go

clean:
	@rm -rf bin/
