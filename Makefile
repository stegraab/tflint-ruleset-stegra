default: build

GOCACHE ?= $(CURDIR)/.gocache

test:
	GOCACHE=$(GOCACHE) go test ./...

build:
	go build

install: build
	mkdir -p ~/.tflint.d/plugins
	mv ./tflint-ruleset-stegra ~/.tflint.d/plugins
