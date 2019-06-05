# -*- mode: Makefile-gmake -*-

SHELL := bash

BUILD_DIR := build

clean:
	rm -rf $(BUILD_DIR)

fmt:
	go fmt ./...

vet:
	go vet ./...

build:
	go build -o $(BUILD_DIR)/sidevault .

.PHONY: clean vet fmt build
