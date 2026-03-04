VERSION := $(shell git describe --tags --always 2>/dev/null || echo "dev")

.PHONY: dev build build-portable clean

dev:
	npm run tauri dev

build:
	npm run tauri build
	@echo "产物: src-tauri/target/release/bundle/nsis/"

build-portable:
	npm run tauri build -- --bundles app
	@echo "产物: src-tauri/target/release/OpenClaw Installer.exe"

clean:
	cargo clean --manifest-path src-tauri/Cargo.toml
	rm -rf dist node_modules
