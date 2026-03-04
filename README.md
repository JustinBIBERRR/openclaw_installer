# OpenClaw Quick Installer for Windows

> 针对 Windows 用户的 OpenClaw **一键图形化安装器**。
> 点击 `.exe` 弹出安装向导，全程无需手动配置环境。

[![License: MIT](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![Platform: Windows](https://img.shields.io/badge/Platform-Windows%2010%2B-blue.svg)]()
[![Built with Tauri](https://img.shields.io/badge/Built%20with-Tauri%202.0-orange.svg)](https://tauri.app)

---

## 目录

- [功能概览](#功能概览)
- [安装原理](#安装原理)
- [快速验证（浏览器预览）](#快速验证浏览器预览)
- [完整开发环境搭建](#完整开发环境搭建)
- [构建生产 exe](#构建生产-exe)
- [项目结构](#项目结构)
- [常见问题](#常见问题)

---

## 功能概览

| 步骤 | 描述 |
|------|------|
| **① 系统预检** | 自动检测管理员权限、WebView2、磁盘空间、端口可用性、网络连通性 |
| **② 安装 OpenClaw** | 解压内置 Node.js v22 便携版，通过 npmmirror 加速执行 `npm install -g openclaw` |
| **③ 配置 AI 模型** | 支持 Anthropic / OpenAI / DeepSeek / 自定义四种服务商，含 API Key 格式校验与连通性验证 |
| **④ 启动 Gateway** | 启动本地 AI 网关（默认端口 18789），成功后弹出浏览器打开聊天页面 |
| **Manager 界面** | 安装完成后进入管理界面，可启动/停止 Gateway、修改 API Key、卸载 |

**核心设计原则：**

- 🔒 **零污染**：内置 Node.js 便携版，不修改系统 PATH，不影响已有 Node 环境
- ⚡ **国内加速**：自动使用 `registry.npmmirror.com` 镜像
- 🛠️ **自愈逻辑**：自动修复 Windows 长路径限制、npm SSL、Windows Defender 误拦截
- 🔁 **断点重试**：每一步失败均可单独重试，不重复已完成步骤

---

## 快速验证（浏览器预览）

> 无需安装 Rust 工具链，5 分钟即可在浏览器中看到完整 UI。

### 前置条件

- Node.js 16+（[下载](https://nodejs.org)）
- Git

### 步骤

```bash
# 1. 克隆仓库
git clone https://github.com/JustinBIBERRR/openclaw_quick_installer.git
cd openclaw_quick_installer/openclaw_installer_windows

# 2. 安装前端依赖
npm install

# 3. 启动 Vite 开发服务器（纯前端，无需 Rust）
npm run dev
```

浏览器访问 **http://localhost:1420** 即可看到完整安装向导。

> **说明**：浏览器预览模式下，所有 Tauri 后端调用（系统检测、文件安装、Gateway 启动）均使用 **Mock 模拟数据**，可完整走通 4 步向导，不执行任何真实系统操作。

### 预览模式测试清单

| 测试项 | 操作 | 预期结果 |
|--------|------|----------|
| 系统预检 | 打开页面等待约 1 秒 | 6 项检测全部显示绿色 ✓，"开始安装"按钮亮起 |
| 安装流程 | 点击"开始安装 →" | 4 步进度动画，日志逐行输出，约 3 秒完成 |
| API Key 格式校验 | 输入错误格式 Key | 黄色警告提示，保存按钮保持禁用 |
| API Key 验证 | 选 DeepSeek，输入 `sk-` 开头的 Key，点击验证 | 模拟验证通过，绿色图标 |
| 跳过 Key 配置 | 点击"跳过，稍后配置" | 直接进入 Gateway 启动步骤 |
| Gateway 启动 | 自动进入步骤 4 | 显示"Gateway 启动中..."旋转动画，日志流输出 |
| Manager 界面 | Gateway 完成后 | 进入 Manager，显示端口、AI 服务、操作按钮 |
| 异常重试 | （Tauri 环境下才触发真实错误）| 红色日志 + "重试"按钮 |

---

## 完整开发环境搭建

> 运行真实的 Tauri 桌面窗口（有 Rust 后端逻辑），需要以下前置条件。

### 系统要求

- Windows 10/11 64 位
- Node.js 18+
- **Rust 工具链**（stable-x86_64-pc-windows-msvc）
- **Visual Studio Build Tools 2022**（需勾选"使用 C++ 的桌面开发"工作负载）
- **WebView2 Runtime**（Windows 11 内置，Windows 10 需单独[下载](https://go.microsoft.com/fwlink/p/?LinkId=2124703)）

### 安装 Rust

```powershell
# 下载并运行 rustup 安装程序
Invoke-WebRequest -Uri https://win.rustup.rs -OutFile rustup-init.exe
.\rustup-init.exe

# 安装完成后验证
rustc --version   # 应输出 rustc 1.7x.x
cargo --version
```

> 如在国内网络环境遇到下载慢，可在 PowerShell 中设置代理后再运行：
> ```powershell
> $env:HTTPS_PROXY = "http://127.0.0.1:7890"  # 替换为你的代理地址
> $env:HTTP_PROXY  = "http://127.0.0.1:7890"
> ```

### 安装 VS Build Tools

前往 [Visual Studio 下载页](https://visualstudio.microsoft.com/downloads/#build-tools-for-visual-studio-2022) 下载 **Build Tools for Visual Studio 2022**，安装时勾选：

- ✅ 使用 C++ 的桌面开发

### 启动 Tauri 开发模式

```powershell
cd openclaw_installer_windows

# 安装 npm 依赖
npm install

# 启动（首次会编译 Rust，约 5-10 分钟）
npm run tauri dev
# 或者
make dev
```

首次编译完成后会弹出原生 Windows 窗口，后续热重载速度很快。

---

## 构建生产 exe

```powershell
cd openclaw_installer_windows

# 方式 1：NSIS 安装包（推荐，有安装向导）
make build
# 产物：src-tauri/target/release/bundle/nsis/OpenClaw Installer_1.0.0_x64-setup.exe

# 方式 2：单文件便携 exe
make build-portable
# 产物：src-tauri/target/release/OpenClaw Installer.exe
```

> **注意**：构建前需要将 Node.js v22 便携包放到 `src-tauri/resources/node-v22-win-x64.zip`
>
> ```powershell
> # 下载 Node.js v22 Windows x64 zip（国内镜像）
> Invoke-WebRequest -Uri "https://npmmirror.com/mirrors/node/v22.11.0/node-v22.11.0-win-x64.zip" `
>   -OutFile "src-tauri/resources/node-v22-win-x64.zip"
> ```

---

## 项目结构

```
openclaw_installer_windows/
├── src/                          # React 前端
│   ├── App.tsx                   # 主组件：路由 + 全局状态
│   ├── types.ts                  # TypeScript 类型定义
│   ├── components/
│   │   ├── TitleBar.tsx          # 自定义标题栏
│   │   ├── StepBar.tsx           # 步骤进度条
│   │   ├── LogScroller.tsx       # 终端风格日志组件
│   │   └── StatusDot.tsx         # Gateway 状态指示灯
│   └── pages/
│       ├── SysCheck.tsx          # 步骤1：系统预检
│       ├── Installing.tsx        # 步骤2：安装 OpenClaw
│       ├── ApiKeySetup.tsx       # 步骤3：配置 AI 模型
│       ├── Launching.tsx         # 步骤4：启动 Gateway
│       └── Manager.tsx           # 安装后管理界面
│
├── src-tauri/                    # Rust 后端
│   ├── src/
│   │   ├── main.rs               # 入口
│   │   ├── lib.rs                # Tauri 应用初始化 + 命令注册
│   │   └── commands.rs           # 所有 Tauri 命令实现
│   ├── scripts/
│   │   ├── syscheck.ps1          # 系统预检 PowerShell 脚本
│   │   ├── install.ps1           # 安装逻辑脚本
│   │   └── gateway.ps1           # Gateway 管理脚本
│   ├── resources/
│   │   └── node-v22-win-x64.zip  # 内置 Node.js（构建时需手动下载）
│   ├── icons/                    # 应用图标
│   ├── Cargo.toml
│   └── tauri.conf.json           # Tauri 配置
│
├── package.json
├── vite.config.ts
├── tailwind.config.js
└── Makefile                      # 快捷构建命令
```

---

## 常见问题

**Q: 为什么选择 Tauri 而不是 Electron？**

Tauri 产物体积约 3-8 MB（vs Electron 的 100+ MB），使用系统自带的 WebView2，无需捆绑 Chromium。

**Q: 浏览器预览和真实 Tauri 窗口的区别？**

浏览器预览中所有 `invoke()` 调用（系统检测、文件操作、进程管理）均被 Mock 替代，UI 交互完全相同，但不执行任何真实系统操作。真实 Tauri 窗口会调用 Rust 后端，通过 PowerShell 脚本执行实际安装。

**Q: 支持哪些 Windows 版本？**

Windows 10（1903+）和 Windows 11。WebView2 在 Win11 上已内置，Win10 需额外安装。

**Q: 安装的 OpenClaw 在哪个目录？**

默认安装到 `C:\OpenClaw`，安装时可自定义路径（建议使用纯英文无空格路径）。

**Q: 如何完全卸载？**

在 Manager 界面点击"卸载 OpenClaw"，将删除安装目录及配置文件。或手动删除 `C:\OpenClaw` 目录。

**Q: 国内网络下 npm install 很慢怎么办？**

安装器已自动配置 `registry.npmmirror.com` 镜像，无需手动设置。

---

## 技术栈

| 层 | 技术 |
|----|------|
| 桌面框架 | [Tauri 2.0](https://tauri.app) |
| 前端 | React 18 + TypeScript + TailwindCSS |
| 后端 | Rust（tokio 异步运行时） |
| 安装逻辑 | PowerShell 5.1+ |
| 打包 | NSIS（Windows 安装包） |
| 图标 | Lucide React |

---

## License

MIT
