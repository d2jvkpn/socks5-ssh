#!/bin/make
# include envfile
# export $(shell sed 's/=.*//' envfile)

working_dir = $(shell pwd)


lint:
	go mod tidy
	if [ -d vendor ]; then go mod vendor; fi
	go fmt ./...
	go vet ./...

build:
	mkdir -p target
	go build -o target/main main.go

app:
	mkdir -p target
	go build -o target/socks5-ssh main.go

run:
	make build
	./target/main --config=configs/local.yaml -debug
