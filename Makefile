NAME := lgtm-backend
BUILD_DIR := ./bin
CMD_DIR := ./cmd/server
GOLANGCI_LINT := golangci-lint

.PHONY: build run test lint tidy clean

build:
	@mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(NAME) $(CMD_DIR)

run:
	go run $(CMD_DIR)/main.go

format:
	gofmt -w .

lint:
	$(GOLANGCI_LINT) run ./...

tidy: 
	go mod tidy && go mod verify # download deps + clean go.mod

clean:
	rm -rf $(BUILD_DIR) coverage.out

