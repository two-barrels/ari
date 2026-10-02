# Modified by two-barrels in 2026 for ARI v6 modernization.
# SPDX-License-Identifier: Apache-2.0

GO111MODULE = on

SHELL = /usr/bin/env bash

# Pinned Asterisk rest-api/api-docs/events.json at 97d55b3306ca4aa26c0136c67b79470ca4b2b785.
EVENT_SPEC_FILE = internal/eventgen/json/events-23.json

all: dep check api clients contributors extensions test

ci: check api clients extensions test

contributors:
	write_mailmap > CONTRIBUTORS

protobuf: ari.proto
	protoc -I. -I./vendor -I$(GOPATH)/src --gogofast_out=Mgoogle/protobuf/timestamp.proto=github.com/gogo/protobuf/types,plugins=grpc:. ari.proto
	@go run ./tools/file-notices --apply

dep:
	go mod tidy

api:
	go build ./
	go build ./stdbus
	go build ./rid

test:
	go test `go list ./... | grep -v /vendor/`

check:
	go mod verify
	golangci-lint run
	#gometalinter --disable=gotype client/native ext/...

clients:
	go build ./client/native
	go build ./client/arimocks

extensions:
	go build ./ext/audiouri
	go build ./ext/bridgemon
	go build ./ext/keyfilter
	go build ./ext/play
	go build ./ext/record

events:
	@go run ./internal/eventgen internal/eventgen/template.tmpl ${EVENT_SPEC_FILE} > events_gen.go.tmp
	@gofmt -w events_gen.go.tmp
	@mv events_gen.go.tmp events_gen.go
	@go run ./tools/file-notices --apply
	
mock:
	go install github.com/vektra/mockery/v3@latest
	rm -Rf client/arimocks
	mockery
	@go run ./tools/file-notices --apply
