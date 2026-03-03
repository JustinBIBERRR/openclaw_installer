# OpenClaw Manager

一键部署本地 AI 网关的 Windows 管理工具。用户**只需双击 exe**，即可完成从零到运行的全部流程。

## 使用方式（最终用户）

1. 下载 `openclaw-manager-amd64.exe`
2. 双击运行
3. 按屏幕提示操作（约 2-5 分钟）

> 如果 Windows SmartScreen 弹出提示，点击「更多信息」→「仍要运行」。

## 功能

| 场景 | 行为 |
|------|------|
| 首次运行 | 引导安装流程（系统预检 → 下载 Node.js → 安装 CLI → 配置 API Key → 启动 Gateway） |
| 安装后运行 | 进入管理器模式（启动/停止/重启 Gateway、更换 API Key、更新、卸载） |
| 安装中断后运行 | 显示恢复向导，从断点继续 |

## 开发者构建

### 前提条件

- [Go 1.22+](https://go.dev/dl/)
- Windows 10/11（或 Linux/macOS 用于交叉编译）

### 首次初始化

```powershell
make setup   # 或: go mod tidy
```

### 编译

```powershell
# Windows 本地编译
make build

# Linux/macOS 交叉编译
make build-linux
```

输出文件：`dist/openclaw-manager-amd64.exe`

### 运行（本地调试）

```powershell
make run
# 或: go run .
```

## 架构说明

```
openclaw-manager/
├── main.go                    # 入口：状态机分发 installer/manager 模式
├── internal/
│   ├── state/                 # 安装状态清单（manifest.json）+ 回滚逻辑
│   ├── env/                   # 影子路径环境隔离（不写注册表）
│   ├── proxy/                 # 三层代理探测（env → WinINet 注册表 → 直连）
│   ├── mirror/                # 并发镜像探测（npmmirror / 腾讯云 / nodejs.org）
│   ├── steps/                 # 安装步骤：预检、下载、安装、API Key、配置、Gateway
│   └── ui/                    # Banner、进度条、Spinner、样式
```

## 核心设计原则

- **离线优先**（安装完成后）：除 API Key 校验，Gateway 本体无需网络
- **0 全局污染**：不写注册表，不改全局 PATH，不影响用户现有 Node 环境
- **原子事务**：每步完成写检查点，支持断点续传和精确回滚
- **国内镜像感知**：自动选择最快可达的 Node.js 下载源

## 配置文件位置

| 文件 | 路径 |
|------|------|
| 安装状态清单 | `%APPDATA%\OpenClaw\install_manifest.json` |
| AI 服务配置 | `%APPDATA%\OpenClaw\openclaw.json` |
| Node.js 运行时 | `%APPDATA%\OpenClaw\runtime\node\` |
| npm 全局包 | `%APPDATA%\OpenClaw\npm-global\` |

## 卸载

```powershell
openclaw-manager-amd64.exe --uninstall
```

或在程序管理器模式中选择「卸载 OpenClaw」。
