#!/bin/make
# include envfile
# export $(shell sed 's/=.*//' envfile)

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

save:
	mkdir -p target
	docker save socks5-proxy:dev -o target/socks5-proxy_dev.tar
	pigz -f target/socks5-proxy_dev.tar
	ls -alh target/socks5-proxy_dev.tar.gz

release:
	release=true deployments/go_build.sh
	ls -al target

ssh:
	make build
	./target/main ssh -config=configs/local.yaml -addr=127.0.0.1:1081

server:
	make build
	./target/main server -addr=127.0.0.1:1091

socks5_proxy:
	config=configs/ssh.conf ./bin/socks5_proxy.sh remote_host 127.0.0.1:1081

noauth:
	make build
	./target/main ssh -config=configs/local.yaml -subkey=noauth -addr=127.0.0.1:1081

test:
	curl -k -x 'socks5h://hello:world@127.0.0.1:1081' https://icanhazip.com

image-local:
	BUILD_Region=cn DOCKER_Pull=false DOCKER_Tag=local DOCKER_Push=false GIT=false \
	  bash deployments/build.sh dev

image-dev:
	BUILD_Region=cn DOCKER_Pull=false DOCKER_Tag=dev DOCKER_Push=false \
	  bash deployments/build.sh dev
