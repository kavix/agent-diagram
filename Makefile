BINARY_NAME=agent-diagram

.PHONY: all build test clean fmt lint install run-example

all: test build

build:
	go build -o $(BINARY_NAME) ./cmd/agent-diagram

test:
	go test -v -race ./...

fmt:
	go fmt ./...

clean:
	rm -f $(BINARY_NAME)

install:
	go install ./cmd/agent-diagram

run-example: build
	./$(BINARY_NAME) render examples/kueue_resize.mmd --width 120
	@echo ""
	./$(BINARY_NAME) render examples/kueue_resize.mmd --width 70
	@echo ""
	./$(BINARY_NAME) render examples/kueue_reconcile_flow.mmd --width 90
