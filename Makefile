.PHONY: client
client:
	go generate ./...

build:
	go install ./...