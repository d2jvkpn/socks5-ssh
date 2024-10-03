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

app:
	mkdir -p target

	# -w -s -X main.build_hostname=$(build_hostname)
	@go build -ldflags="\
	  -X main.build_time=$(build_time) \
	  -X main.git_repository=$(git_repository) \
	  -X main.git_branch=$(git_branch) \
	  -X main.git_commit_id=$(git_commit_id) \
	  -X main.git_commit_time=$(git_commit_time)" \
	  -o target/socks5-proxy main.go

	ls -al target/

ssh:
	make build
	./target/main ssh -config=configs/local.yaml -addr=127.0.0.1:1080

noauth:
	make build
	./target/main ssh -config=configs/local.yaml -subkey=noauth -addr=127.0.0.1:1080

test:
	curl -k -x 'socks5://hello:world@127.0.0.1:1080' https://icanhazip.com
