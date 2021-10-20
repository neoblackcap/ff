.PHONY: fmt

SRC=$(shell find . -name "*.go")

build: bin/ff

build_linux: bin/ff_linux

bin/ff: $(SRC)
	@go build -o bin/ff main.go

bin/ff_linux: $(SRC)
	@GOOS=linux GOARCH=amd64 go build -o bin/ff_linux main.go

run: build
	@bin/imgur

fmt: $(SRC)
	@go fmt ./cmd
	@go fmt ./utils
	@go fmt main.go

clean:
	@rm -rf bin/
