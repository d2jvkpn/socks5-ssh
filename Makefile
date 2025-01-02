#!/bin/make
# include envfile
# export $(shell sed 's/=.*//' envfile)

SHELL := /bin/bash

working_dir = $(shell pwd)
build_hostname = $(shell hostname)

build_time = $(shell date +'%FT%T.%N%:z')
git_repository = $(shell git config --get remote.origin.url)
git_branch = $(shell git rev-parse --abbrev-ref HEAD)
git_commit_id = $(shell git rev-parse --verify HEAD)
git_commit_time = $(shell git log -1 --format="%at" | xargs -I{} date -d @{} +%FT%T%:z)

lint:
	go mod tidy
	if [ -d vendor ]; then go mod vendor; fi
	go fmt ./...
	go vet ./...

build:
	mkdir -p target
	go build -o target/main main.go
	ls -al target/

release:
	release=true containers/go_build.sh
	ls -al target

ssh:
	make build
	./target/main ssh -config=configs/local.yaml -addr=127.0.0.1:1081

server:
	make build
	./target/main server -addr=127.0.0.1:1091

socks5_proxy:
	config=configs/ssh.conf ./bin/socks5_ssh.sh remote_host 127.0.0.1:1081

noauth:
	make build
	./target/main ssh -config=configs/local.yaml -subkey=noauth -addr=127.0.0.1:1081

test:
	curl -k -x 'socks5h://hello:world@127.0.0.1:1081' https://icanhazip.com

image-local:
	region=cn DOCKER_Pull=false DOCKER_Tag=local DOCKER_Push=false GIT_Pull=false \
	  bash containers/build.sh dev

image-dev:
	region=cn DOCKER_Pull=false DOCKER_Tag=dev DOCKER_Push=false \
	  bash containers/build.sh dev
