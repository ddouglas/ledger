SHELL := /bin/bash

ui:
	source ./frontend/.env && npm --prefix ./frontend run build --watch

# frontend:
# 	source ./frontend/.env && npm --prefix ./frontend run serve

build-backend:
	go mod tidy
	go build -o ./.build/ledger ./cmd/ledger
	clear

server: build-backend
	aws-vault exec ledger-api-admin -- chamber exec ledger-api/development -- ./.build/ledger server

worker: build-backend
	aws-vault exec ledger-api-admin -- chamber exec ledger-api/development -- ./.build/ledger worker

dbuild:
	docker build . -t ledger:latest

dcupd:
	docker-compose up -d

dcdown:
	docker-compose down

dcdownv:
	docker-compose down -v

dclogsf:
	docker-compose logs -f

dcstart: dcupd dclogsf

dcrestart: dcdown dcupd dclogsf