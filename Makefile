SHELL := /bin/bash

run:
	$(MAKE) -j2 ui backend


ui:
	source ./frontend/.env && npm --prefix ./frontend run build --watch

# frontend:
# 	source ./frontend/.env && npm --prefix ./frontend run serve

build-backend:
	go mod tidy
	go build -o ./.build/ledger ./cmd/ledger
	clear

server: build-backend
	./.build/ledger server

importer: build-backend
	./.build/ledger importer

tunnel:
	ngrok http 9000