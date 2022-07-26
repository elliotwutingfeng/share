## Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
## Use of this source code is governed by a BSD-style
## license that can be found in the LICENSE file.

COVER_OUT:=cover.out
COVER_HTML:=cover.html
CPU_PROF:=cpu.prof
MEM_PROF:=mem.prof

CIIGO := ${GOBIN}/ciigo

.PHONY: all install lint build docs docs-serve clean distclean
.PHONY: test test.prof bench.lib.websocket coverbrowse

all: test lint build

install:
	go install ./cmd/...

build:
	mkdir -p _bin/linux-amd64
	GOOS=linux GOARCH=amd64 go build -o _bin/linux-amd64/ ./cmd/...

test:
	CGO_ENABLED=1 go test -failfast -race -count=1 -coverprofile=$(COVER_OUT) ./...
	go tool cover -html=$(COVER_OUT) -o $(COVER_HTML)

test.prof:
	go test -race -cpuprofile $(CPU_PROF) -memprofile $(MEM_PROF) ./...

bench.lib.websocket:
	export GORACE=history_size=7 && \
		export CGO_ENABLED=1 && \
		go test -race -run=none -bench -benchmem \
			-cpuprofile=$(CPU_PROF) \
			-memprofile=$(MEM_PROF) \
			. ./lib/websocket

coverbrowse:
	xdg-open $(COVER_HTML)

lint:
	-golangci-lint run ./...

$(CIIGO):
	go install git.sr.ht/~shulhan/ciigo/cmd/ciigo

docs: $(CIIGO)
	ciigo convert _doc

docs-serve: $(CIIGO)
	ciigo -address 127.0.0.1:21019 serve _doc

clean:
	rm -f $(COVER_OUT) $(COVER_HTML)
	rm -f ./**/${CPU_PROF}
	rm -f ./**/${MEM_PROF}
	rm -f ./**/$(COVER_OUT)
	rm -f ./**/$(COVER_HTML)

distclean:
	go clean -i ./...
