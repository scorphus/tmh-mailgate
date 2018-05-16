GIT ?= git
GO_VARS ?=
GO ?= go
REVISION := $(shell $(GIT) rev-parse HEAD)
VERSION ?= $(shell $(GIT) describe --tags ${REVISION} 2> /dev/null || echo "$(REVISION)")
DATE_TIME := $(shell LANG=en_US date +%Y-%m-%dT%H:%M:%S%z)
MESSAGE := $(shell git show -s --format=%B | head -1)
LD_FLAGS := -X main.Version=$(VERSION) \
	-X main.Revision=$(REVISION) \
	-X main.DateTime=$(DATE_TIME) \
	-X "'"main.Message=$(MESSAGE)"'"

install-dep:
	@env GOBIN=$(PWD)/bin go get -u github.com/golang/dep/cmd/dep

dep:
	@env GOBIN=$(PWD)/bin PATH=$(PWD)/bin:$(PATH) dep ensure

setup: install-dep dep

run:
	@$(GO) run $$(ls *.go | grep -v _test) serve

build:
	@$(GO_VARS) $(GO) build -o="tmh-mailgate" -ldflags="$(LD_FLAGS)"

reload:
	@pkill -hup -F tmh-mailgate.pid

.PHONY: build dep install-dep reload run setup
