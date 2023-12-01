BIN := obu 
GOBIN ?= $(shell go env GOPATH)/bin

.PHONY: all
all: build

.PHONY: build
build:
	go build -o $(BIN) main.go


.PHONY: test
test: build
	go test -v -race ./...

.PHONY: lint
lint:   $(GOBIN)/staticcheck
	go vet ./...
	staticcheck ./...

$(GOBIN)/staticcheck:
	go install honnef.co/go/tools/cmd/staticcheck@latest
