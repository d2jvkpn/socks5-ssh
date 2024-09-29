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
	ls -al target/

app:
	mkdir -p target
	go build -o target/socks5-proxy main.go
	ls -al target/

ssh:
	make build
	./target/main ssh -config=configs/local.yaml -addr=127.0.0.1:1080

noauth:
	make build
	./target/main ssh -config=configs/local.yaml -subkey=noauth -addr=127.0.0.1:1080

test:
	curl -k -x 'socks5://hello:world@127.0.0.1:1080' https://icanhazip.com
