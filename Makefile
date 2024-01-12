all: build

build:
	@echo "Building..."
	@go build -o tailscale-tui cmd/tailscale-tui/main.go

run:
	@go run cmd/tailscale-tui/main.go

clean:
	@echo "Cleaning..."
	@rm -f tailscale-tui

.PHONY: all build run clean
