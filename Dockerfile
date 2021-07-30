FROM golang:1.16.3 as builder
WORKDIR /app
COPY . .
RUN CGO_ENABLED=0 go build -o ./.build/ledger ./cmd/ledger

FROM alpine:latest AS release
WORKDIR /app

RUN apk --no-cache add tzdata ca-certificates

COPY --from=builder /app/.build/ledger .

LABEL maintainer="David Douglas <david@onetwentyseven.dev>"