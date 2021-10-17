FROM golang:1.16.3 as builder
WORKDIR /app
COPY . .
RUN CGO_ENABLED=0 go build -o ./.build/ledger ./cmd/ledger

FROM wait-for-it:latest as waiter

FROM alpine:latest AS release
WORKDIR /app

RUN apk --no-cache add tzdata ca-certificates

COPY --from=builder /app/.build/ledger .
COPY --from=waiter /go/bin/wait-for-it .

LABEL maintainer="David Douglas <david@onetwentyseven.dev>"