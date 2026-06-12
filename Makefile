.PHONY: all build build-linux build-windows clean

all: build

build: build-linux build-windows

build-linux:
	@echo "Building for Linux (amd64)..."
	@mkdir -p bin/linux-amd64
	GOOS=linux GOARCH=amd64 go build -o bin/linux-amd64/secnotes ./cmd/secnotes
	GOOS=linux GOARCH=amd64 go build -o bin/linux-amd64/secnotes-server ./cmd/server

build-windows:
	@echo "Building for Windows (amd64)..."
	@mkdir -p bin/windows-amd64
	GOOS=windows GOARCH=amd64 go build -o bin/windows-amd64/secnotes.exe ./cmd/secnotes
	GOOS=windows GOARCH=amd64 go build -o bin/windows-amd64/secnotes-server.exe ./cmd/server

clean:
	@echo "Cleaning up..."
	rm -rf bin/
