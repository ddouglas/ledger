SHELL := /bin/bash

gqlgen:
	gqlgen generate --config .config/gqlgen.yml

build:
	go mod tidy
	go build -o ./.build/ledger ./cmd/ledger
	clear

server: build
	# aws-vault exec ledger-api-admin -- chamber exec ledger-api/development -- 
	./.build/ledger server

worker: build
	# aws-vault exec ledger-api-admin -- chamber exec ledger-api/development -- 
	./.build/ledger worker

# dbuild:
# 	docker build . -t ledger:latest

# dcupd:
# 	docker-compose up -d

# dcdown:
# 	docker-compose down

# dcdownv:
# 	docker-compose down -v

# dclogsf:
# 	docker-compose logs -f

# dcstart: dcupd dclogsf

# dcrestart: dcdown dcupd dclogsf