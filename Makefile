#!/bin/make
# include envfile
# export $(shell sed 's/=.*//' envfile)

working_dir = $(shell pwd)


lint:
	go mod tidy
	if [ -d vendor ]; then go mod vendor; fi
	go fmt ./...
	go vet ./...

	app_name=swagger bash bin/swagger-go/swag.sh false > /dev/null

build:
	mkdir -p target
	go build -o target/main main.go

run:
	make build
	./target/main --config=configs/local.yaml -debug
