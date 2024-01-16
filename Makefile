all: build

build:
	@echo "Building..."
	@go build -o tailscale-tui main.go

run:
	@go run main.go

clean:
	@echo "Cleaning..."
	@rm -f tailscale-tui

.PHONY: all build run clean
