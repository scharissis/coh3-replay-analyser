.PHONY: build build-rust build-go build-web clean install test update-data

# Default target
build: build-rust build-go build-web

# Build the Rust wrapper library
build-rust:
	cd vault-wrapper && cargo build --release
	mkdir -p vault/lib
	cp vault-wrapper/target/release/libvault_wrapper.a vault/lib/ || \
	cp vault-wrapper/target/release/vault_wrapper.lib vault/lib/

# Build the Go CLI
build-go:
	CGO_ENABLED=1 go build -ldflags '-extldflags "-static"' -o coh3-build-order ./cmd/coh3-build-order

# Build the web server
build-web:
	CGO_ENABLED=1 go build -ldflags '-extldflags "-static"' -o coh3-web-server ./cmd/coh3-web-server

# Clean build artifacts
clean:
	cd vault-wrapper && cargo clean
	rm -rf vault/lib
	rm -f coh3-build-order coh3-web-server

# Install dependencies
install:
	cd vault-wrapper && cargo fetch
	go mod download

# Run tests
test:
	go test -v ./...
	cd tests && go test -v .

# Update game data from coh3-data repository
update-data:
	@echo "Fetching latest game data from coh3-data repository..."
	./scripts/fetch-data.sh
	@echo "Data update complete! Raw data files are now available in data/coh3-data/"
	@echo "The Go application will automatically use these files for PBGID lookups."