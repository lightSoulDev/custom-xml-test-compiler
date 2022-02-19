build:
	go build -v ./cmd/xmlParser
	
run:
	go run -v ./cmd/xmlParser

.PHONY: build run
.DEFAULT_GOAL := build