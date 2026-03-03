# ─────────────────────────────────────────────
# OpenClaw Manager — Makefile
# ─────────────────────────────────────────────

# 从 git tag 自动获取版本号，失败时回退为 dev
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME := $(shell powershell -NoProfile -Command "Get-Date -Format 'yyyyMMddHHmmss'" 2>/dev/null || date +%Y%m%d%H%M%S)
LDFLAGS := -ldflags "-s -w -X main.Version=$(VERSION)"
OUT := dist\openclaw-manager-amd64.exe

.PHONY: all setup build clean run help

## help: 显示此帮助
help:
	@echo.
	@echo   OpenClaw Manager 构建指令
	@echo   ─────────────────────────
	@echo   make setup   首次开发环境初始化（下载依赖）
	@echo   make build   编译 Windows exe
	@echo   make run     本地运行（需要 Windows）
	@echo   make clean   清理构建产物
	@echo.

## setup: 初始化依赖（首次使用时运行）
setup:
	go mod tidy
	go mod download

## build: 编译 Windows amd64 exe
build:
	@if not exist dist mkdir dist
	set GOOS=windows&& set GOARCH=amd64&& go build $(LDFLAGS) -o $(OUT) .
	@echo.
	@echo   构建完成: $(OUT)
	@echo   版本: $(VERSION)

## build-linux: 在 Linux/macOS CI 上交叉编译（用于 GitHub Actions）
build-linux:
	mkdir -p dist
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o dist/openclaw-manager-amd64.exe .

## run: 直接运行（仅限 Windows 开发调试）
run:
	go run . $(ARGS)

## clean: 删除构建产物
clean:
	@if exist dist rmdir /s /q dist
	@echo 清理完成

all: setup build
