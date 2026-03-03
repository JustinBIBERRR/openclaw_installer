---
name: OpenClaw Intelligent Manager (OIM)
version: "1.0.0"
target_platform: "Windows 11 (优先), Windows 10 兼容"
network_strategy: "在线安装（在线用户优先，离线支持列入 v1.1）"
overview: >
  构建单一 Go 编译的 Windows exe（~10MB），集「安装器 + 管理器」双模式于一体。
  核心目标：在线下载 Node.js（国内镜像优先）、影子路径环境隔离、
  原子化检查点事务（含回滚）、代理感知 API 校验、0 全局污染。

todos:
  - id: scaffold
    content: "Scaffold Go project: go.mod, main.go, internal/ 目录结构，添加依赖 (charmbracelet/huh, lipgloss, bubbles, errgroup)"
    status: pending

  - id: state-machine
    content: "internal/state/: 实现安装器↔管理器双模式状态机，读写 %APPDATA%/OpenClaw/install_manifest.json，含 schema version + failed 态恢复逻辑"
    status: pending

  - id: syscheck
    content: "internal/steps/syscheck.go: 磁盘空间(>200MB)、端口 18789 占用、现有安装检测、Windows 版本检测、网络连通性预检"
    status: pending

  - id: runtime-download
    content: "internal/steps/runtime.go: 探测国内/海外镜像可达性 → 选最快源下载 node-win-x64.zip → 实时进度条 → SHA256 验证 → 解压到 %APPDATA%/OpenClaw/runtime/"
    status: pending

  - id: shadow-env
    content: "internal/env/shadow.go: ShadowEnv() 注入 runtime/node + npm-global/bin 到 PATH，禁止任何注册表写入"
    status: pending

  - id: openclaw-install
    content: "internal/steps/openclaw.go: ShadowEnv() 执行 npm install -g openclaw，Spinner + 流式日志输出，写入 manifest 检查点"
    status: pending

  - id: proxy-detect
    content: "internal/proxy/detect.go: 三层代理探测（HTTPS_PROXY 环境变量 → WinINet 注册表 → 直连），返回 *url.URL"
    status: pending

  - id: apikey
    content: "internal/steps/apikey.go: provider 选择、masked 输入、Regex 前置校验、HTTP 连通性测试（HTTP 状态码精确分类）、最多重试 3 次"
    status: pending

  - id: config
    content: "internal/steps/config.go: 写入 %APPDATA%/OpenClaw/openclaw.json，model + apiKey，写入 manifest 检查点"
    status: pending

  - id: gateway
    content: "internal/steps/gateway.go: 启动 Gateway、健康轮询 GET /health (30s)、端口冲突自动切换、打开浏览器、创建桌面快捷方式"
    status: pending

  - id: rollback
    content: "internal/state/rollback.go: 按步骤逆序清理，支持 --uninstall flag 和安装失败自动触发"
    status: pending

  - id: ui
    content: "internal/ui/: banner.go、progress.go(下载进度 + Spinner)、styles.go、error.go(分场景错误引导文案)"
    status: pending

  - id: manager-mode
    content: "main.go: 状态机入口，failed 态显示恢复引导，complete 态进入动态菜单（按 Gateway 运行状态渲染选项）"
    status: pending

  - id: build-pipeline
    content: "Makefile(本地构建) + .github/workflows/release.yml(CI: windows-latest runner, GOOS=windows, ldflags 版本注入, osslsigncode 或 signtool 签名)"
    status: pending

isProject: true
---

# OpenClaw Intelligent Manager — 完整计划书 v2

## 一、核心设计原则

| 原则 | 含义 |
|------|------|
| **在线优先** | Node.js 运行时在首次安装时从网络下载，exe 本体保持 ~10MB |
| **国内镜像感知** | 自动探测镜像可达性，优先 npmmirror，失败回退 nodejs.org |
| **0 全局污染** | 不写注册表、不改全局 PATH、不影响用户现有 Node 环境 |
| **原子事务** | 每步完成写检查点；失败态重启后显示恢复引导，可精确续传 |
| **双模式感知** | 同一个 exe：未安装 → 安装器；已安装 → 管理器 |
| **诚实 UX** | 只在有真实数据时显示百分比（HTTP 下载）；无法量化时用 Spinner |
| **Win11 优先** | API 选型以 Win11 为基准；Win10 做兼容性测试，不做 Win10 专项优化 |

---

## 二、目录结构

```
openclaw-manager/
├── main.go                          # 入口：状态机分发 installer/manager 模式
├── go.mod
├── go.sum
├── Makefile                         # build / release / clean
├── .gitignore                       # 注意：无需忽略大型二进制资产（在线方案）
├── internal/
│   ├── state/
│   │   ├── manifest.go              # InstallManifest 结构体 + 读写 + schema migration
│   │   └── rollback.go              # 按步骤逆序回滚 + --uninstall
│   ├── env/
│   │   └── shadow.go                # ShadowEnv(): 注入 runtime/node + npm-global/bin
│   ├── proxy/
│   │   └── detect.go                # 三层代理探测（env → WinINet → 直连）
│   ├── mirror/
│   │   └── select.go                # 并发探测镜像延迟，选最快可达源
│   ├── steps/
│   │   ├── syscheck.go              # 系统预检（含网络连通性）
│   │   ├── runtime.go               # 下载 + SHA256 验证 + 解压 Node.js
│   │   ├── openclaw.go              # npm install -g openclaw（Spinner + 流式日志）
│   │   ├── apikey.go                # API Key 录入与精确状态码校验
│   │   ├── config.go                # 写入 openclaw.json
│   │   └── gateway.go               # 启动 + 健康轮询 + 桌面快捷方式
│   └── ui/
│       ├── banner.go                # ASCII Art 横幅
│       ├── progress.go              # 下载进度条（真实 %）+ Spinner（不确定阶段）
│       ├── styles.go                # Lipgloss 颜色主题
│       └── error.go                 # 分场景错误引导文案
└── .github/
    └── workflows/
        └── release.yml              # CI/CD（见第七节）
```

> **相比离线方案的简化**：移除了 `embed/` 和 `assets/` 目录，git 仓库无大型二进制文件，`go:embed` 相关的路径问题全部消失，exe 从 ~60MB 缩至 ~10MB。

---

## 三、技术栈

| 职责 | 选型 | 理由 |
|------|------|------|
| TUI 表单 | `charmbracelet/huh` | 声明式表单，Masked input 原生支持 |
| TUI 样式 | `charmbracelet/lipgloss` | 颜色/边框/布局，无 CGo |
| 下载进度条 | `charmbracelet/bubbles/progress` | 真实 HTTP Content-Length 驱动 |
| 不确定性等待 | `charmbracelet/bubbles/spinner` | 用于 npm install 等无进度 API 的步骤 |
| 实时日志流 | `charmbracelet/bubbles/viewport` | npm 输出滚动窗口 |
| 并发编排 | `golang.org/x/sync/errgroup` | 并发镜像探测 + 并发系统预检 |
| 桌面快捷方式 | `go-ole` + `IShellLink` | Win32 COM，纯 Go，无 CGo |
| HTTP 下载 | 标准库 `net/http` | 流式读取 + Content-Length 进度 |
| JSON 读写 | 标准库 `encoding/json` | 无额外依赖 |
| 构建 | `Makefile` + `GOOS=windows` | Linux/macOS CI 可交叉编译 |

---

## 四、核心模块详解

### A. 状态机：双模式感知 + 失败态恢复

`main.go` 启动时读取 manifest，按 `phase` 分发：

```go
func main() {
    m, err := state.LoadManifest()
    switch {
    case err != nil:
        // 全新安装
        runInstaller(state.NewManifest())
    case m.Phase == state.PhaseComplete:
        // 管理器模式
        runManager(m)
    case m.Phase == state.PhaseFailed:
        // 失败恢复引导（关键：不静默续传，先询问用户）
        if ui.AskResume(m) {
            runInstaller(m) // 从最后成功的检查点续传
        } else {
            state.Rollback(m)
        }
    default:
        // installing 态（程序异常退出后重启）
        runInstaller(m)
    }
}
```

`install_manifest.json` 结构（含 schema version）：

```json
{
  "schema_version": 1,
  "app_version": "1.0.0",
  "phase": "complete",
  "install_dir": "C:\\Users\\Alice\\AppData\\Roaming\\OpenClaw",
  "node_version": "20.11.0",
  "gateway_port": 18789,
  "steps": {
    "runtime_downloaded": { "status": "done", "completed_at": "2026-03-03T10:00:00Z", "hash": "sha256:abc…" },
    "cli_installed":      { "status": "done", "completed_at": "2026-03-03T10:00:20Z" },
    "api_key_saved":      { "status": "done", "completed_at": "2026-03-03T10:01:00Z", "provider": "anthropic", "verified": true },
    "config_written":     { "status": "done", "completed_at": "2026-03-03T10:01:01Z" },
    "gateway_started":    { "status": "done", "completed_at": "2026-03-03T10:01:05Z" }
  }
}
```

`phase` 取值：`"installing"` | `"complete"` | `"failed"` | `"uninstalled"`

**Schema Migration**：`schema_version` 字段用于未来兼容。LoadManifest() 发现版本不匹配时：
- 能迁移（已知升级路径）→ 静默迁移并保存
- 无法迁移（跨大版本）→ 提示用户"检测到旧版安装，建议重新安装以获得最佳体验"

---

### B. 镜像选择（国内用户优先）

并发探测各镜像延迟，选最快可达源：

```go
// internal/mirror/select.go
var candidates = []Mirror{
    {Name: "npmmirror (国内)",  NodeURL: "https://npmmirror.com/mirrors/node/"},
    {Name: "nodejs.org (官方)", NodeURL: "https://nodejs.org/dist/"},
    {Name: "腾讯云镜像",        NodeURL: "https://mirrors.cloud.tencent.com/nodejs-release/"},
}

func SelectFastest(ctx context.Context) Mirror {
    // 对每个候选并发发送 HEAD 请求，取第一个成功响应的
    // 超时 3s，全部失败则返回 nodejs.org 并记录警告
}
```

选定镜像后，Node.js 下载 URL 格式：
`{mirrorURL}/v{version}/node-v{version}-win-x64.zip`

---

### C. 影子路径环境隔离（修正版）

**修正上版遗漏**：同时注入 `runtime/node` 和 `npm-global/bin`，否则 `openclaw` 命令找不到。

```go
// internal/env/shadow.go
func ShadowEnv(m *state.Manifest) []string {
    nodeDir    := filepath.Join(m.InstallDir, "runtime", "node")
    npmBinDir  := filepath.Join(m.InstallDir, "npm-global", "bin")
    sep        := string(os.PathListSeparator)
    newPath    := nodeDir + sep + npmBinDir + sep + os.Getenv("PATH")

    // 使用新切片，避免污染原始 os.Environ() 返回值
    env := make([]string, len(os.Environ()), len(os.Environ())+3)
    copy(env, os.Environ())
    return append(env,
        "PATH="+newPath,
        "OPENCLAW_HOME="+m.InstallDir,
        "NPM_CONFIG_PREFIX="+filepath.Join(m.InstallDir, "npm-global"),
    )
}
```

---

### D. 在线下载 Node.js（含真实进度）

```go
// internal/steps/runtime.go
func StepDownloadRuntime(m *state.Manifest, prog *ui.ProgressModel) error {
    if m.Steps["runtime_downloaded"].Status == "done" {
        ui.Skip("Node.js 运行时已就绪，跳过")
        return nil
    }

    mirror := mirror.SelectFastest(ctx)
    url    := mirror.NodeZipURL(nodeVersion)

    // 流式下载，通过 Content-Length 驱动真实进度条
    resp, _ := http.Get(url)
    defer resp.Body.Close()
    total := resp.ContentLength   // 用于计算百分比

    dest := filepath.Join(m.InstallDir, "runtime", "node.zip")
    f, _ := os.Create(dest)
    defer f.Close()

    buf := make([]byte, 32*1024)
    var downloaded int64
    for {
        n, err := resp.Body.Read(buf)
        downloaded += int64(n)
        f.Write(buf[:n])
        prog.SetPercent(float64(downloaded) / float64(total)) // 真实百分比
        if err == io.EOF { break }
    }

    // SHA256 验证
    if err := verifySHA256(dest, expectedHash); err != nil {
        os.Remove(dest)
        return fmt.Errorf("文件完整性校验失败，已删除损坏文件: %w", err)
    }

    extractZip(dest, filepath.Join(m.InstallDir, "runtime", "node"))
    return m.MarkDone("runtime_downloaded", map[string]string{"hash": expectedHash})
}
```

---

### E. 代理探测（三层，修正版）

```go
// internal/proxy/detect.go
func Detect() *url.URL {
    // 第一层：标准环境变量（Go net/http 原生支持）
    if v := os.Getenv("HTTPS_PROXY"); v != "" {
        if u, err := url.Parse(v); err == nil { return u }
    }
    // 第二层：WinINet 注册表（覆盖大多数国内代理工具：Clash/V2Ray 等）
    // HKCU\Software\Microsoft\Windows\CurrentVersion\Internet Settings
    // ProxyEnable=1, ProxyServer="127.0.0.1:7890"
    if u := readWinINetProxy(); u != nil { return u }
    // 第三层：直连
    return nil
}

func NewHTTPClient() *http.Client {
    proxy := Detect()
    transport := &http.Transport{}
    if proxy != nil {
        transport.Proxy = http.ProxyURL(proxy)
    }
    return &http.Client{Transport: transport, Timeout: 10 * time.Second}
}
```

---

### F. API Key 校验（HTTP 状态码精确分类）

```
用户输入 API Key
  └─ Regex 前置校验（本地，零网络）
       ├─ 匹配 sk-ant-api03-*   → [✓ 已识别为 Anthropic Claude Key]
       ├─ 匹配 sk-[a-zA-Z0-9]+ → [✓ 已识别为 OpenAI 兼容 Key]
       └─ 不匹配               → [✗ 格式不符，请检查是否完整复制]（留在输入框）

  └─ HTTP 连通性测试（带代理、10s 超时）
       ├─ 200               → Key 有效，继续
       ├─ 401 Unauthorized  → Key 无效（明确拒绝，不放行）→ "Key 验证失败，请确认是否正确"
       ├─ 403 Forbidden     → Key 无权限 → "Key 无此功能权限，建议继续（可稍后更换）"，询问是否放行
       ├─ 429 Too Many Req  → Key 有效但限速 → 放行，提示"使用频率受限，不影响安装"
       ├─ 5xx 服务端故障    → 视为网络问题 → 询问跳过
       ├─ 网络超时/连接失败  → 提示「连接失败，是否跳过验证？」
       │                       ├─ 跳过 → 保存 Key，标记 verified=false，继续
       │                       └─ 重试 → 显示代理配置建议，返回
       └─ 连续失败 3 次     → 「建议检查代理后重启程序」→ phase="failed"，保存进度，退出
```

---

### G. 端口冲突处理

```go
// internal/steps/gateway.go
func resolvePort(preferred int) (int, error) {
    candidates := []int{preferred, preferred+1, preferred+2}
    for _, port := range candidates {
        if isFree(port) { return port, nil }
        proc := getProcessByPort(port) // 调用 netstat + tasklist
        ui.Warn(fmt.Sprintf("端口 %d 已被 %s 占用", port, proc))
    }
    // 所有候选端口均被占用，让用户手动输入
    return ui.AskCustomPort()
}
```

---

### H. 原子化安装事务（含回滚）

每步遵循：**执行 → 验证 → 写检查点**

**回滚顺序**（逆序）：

```
gateway_started    → openclaw gateway stop（通过 ShadowEnv 执行）
config_written     → 删除 %APPDATA%\OpenClaw\openclaw.json
api_key_saved      → （无文件产物，跳过）
cli_installed      → 删除 npm-global/ 目录
runtime_downloaded → 删除 runtime/ 目录
最终               → 删除 %APPDATA%\OpenClaw\install_manifest.json
                   → 删除桌面快捷方式
                   （保留 %APPDATA%\OpenClaw\ 目录，供用户检查日志）
```

---

## 五、用户体验路径

### 阶段一：启动与预检（~3秒）

```
╔══════════════════════════════════╗
║   OpenClaw Manager  v1.0.0      ║
║   本地 AI 网关一键部署工具       ║
╚══════════════════════════════════╝

  正在检查运行环境...

  [✓] Windows 11 (26100)
  [✓] 可用磁盘空间: 127 GB
  [✓] 端口 18789: 可用
  [✓] 网络连通: npmmirror (延迟 23ms) ← 自动选定镜像

  首次安装，开始引导流程。
```

### 阶段二：下载并部署运行时（真实进度）

```
  [1/4] 下载 Node.js v20.11.0
        来源: npmmirror.com  大小: 28.4 MB

        ████████████████░░░░  79%  22.4 MB / 28.4 MB  ↓ 4.2 MB/s
```

### 阶段三：安装 CLI（诚实的不确定性）

```
  [2/4] 安装 OpenClaw CLI
        ⠸ 正在安装依赖包...

        > openclaw@2.1.0 postinstall
        > node scripts/setup.js
        ✓ 已安装 12 个依赖包
```

### 阶段四：API Key 录入

```
  [3/4] 配置 AI 服务

  请选择 AI 服务商:
  ❯ Anthropic (Claude)
    OpenAI (GPT-4)
    其他 OpenAI 兼容服务（需提供 Base URL）

  请输入 API Key: sk-ant-api03-**********************
  [✓ 已识别为 Anthropic Claude Key]

  正在验证连通性（通过系统代理）... ✓ 验证通过
```

### 阶段五：启动与完成

```
  [4/4] 启动本地 Gateway

  [✓] Gateway 已启动 → http://localhost:18789
  [✓] 桌面快捷方式已创建 → OpenClaw Manager

  ══════════════════════════════════════════
   全部完成！正在打开浏览器...
   本窗口将在 5 秒后自动关闭。  [立即关闭]
  ══════════════════════════════════════════
```

### 失败后重启（恢复引导）

```
  ⚠ 检测到上次安装未完成（在「安装 CLI」步骤中断）

  已完成: [✓] 下载 Node.js
  中断处: [✗] 安装 OpenClaw CLI

  请选择:
  ❯ 从中断处继续安装
    重新开始（将清理已下载的文件）
    退出
```

### 管理器模式（状态感知菜单）

```
╔══════════════════════════════════╗
║   OpenClaw Manager  v1.0.0      ║
║   ● Gateway 运行中 (:18789)     ║  ← 动态状态
╚══════════════════════════════════╝

  请选择操作:
    重启 Gateway          ← 运行中时显示"重启"，停止时显示"启动"
  ❯ 停止 Gateway
    查看运行状态
    更新 OpenClaw CLI
    更换 API Key
    ───────────────
    卸载 OpenClaw
    退出
```

---

## 六、代码签名与分发策略

| 场景 | 策略 |
|------|------|
| **正式发布** | EV 代码签名证书（Sectigo/DigiCert），SmartScreen 信任立即生效 |
| **开发/测试** | README 说明点击步骤 + 发布页附 SHA256 哈希 |
| **CI 自动签名** | GitHub Actions `windows-latest` runner + `signtool.exe`（见第七节）|
| **Defender 误报** | syscheck 阶段检测 Defender 实时保护状态，若开启则在完成界面显示白名单添加指引 |

---

## 七、构建与发布流水线

### Makefile（本地开发）

```makefile
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS := -ldflags "-s -w -X main.Version=$(VERSION) -X main.BuildTime=$(shell date -u +%Y%m%d%H%M%S)"
OUT     := dist/openclaw-manager-amd64.exe

.PHONY: build clean

build:
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o $(OUT) .

clean:
	rm -rf dist/
```

> **无需 `prepare-assets`**：在线方案不需要在构建时打包 Node.js，`go build` 可直接执行。

### `.github/workflows/release.yml`

```yaml
name: Release

on:
  push:
    tags: ['v*']

jobs:
  build-and-sign:
    runs-on: windows-latest          # Windows runner：编译 + 签名合二为一

    steps:
      - uses: actions/checkout@v4

      - uses: actions/setup-go@v5
        with:
          go-version: '1.22'

      - name: Run tests
        run: go test ./...

      - name: Lint
        run: |
          go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
          golangci-lint run

      - name: Build
        run: |
          $version = git describe --tags --always
          go build -ldflags "-s -w -X main.Version=$version" `
            -o dist/openclaw-manager-amd64.exe .

      - name: Sign executable
        if: ${{ secrets.CODESIGN_PFX != '' }}
        run: |
          $pfxBytes = [Convert]::FromBase64String("${{ secrets.CODESIGN_PFX }}")
          [IO.File]::WriteAllBytes("cert.pfx", $pfxBytes)
          & "C:\Program Files (x86)\Windows Kits\10\bin\10.0.22621.0\x64\signtool.exe" `
            sign /f cert.pfx /p "${{ secrets.CODESIGN_PASS }}" `
            /tr http://timestamp.digicert.com /td sha256 /fd sha256 `
            dist/openclaw-manager-amd64.exe
          Remove-Item cert.pfx

      - name: Compute SHA256
        run: |
          $hash = (Get-FileHash dist/openclaw-manager-amd64.exe -Algorithm SHA256).Hash.ToLower()
          "$hash  openclaw-manager-amd64.exe" | Out-File -Encoding ascii dist/SHA256SUMS.txt

      - name: Upload GitHub Release
        uses: softprops/action-gh-release@v2
        with:
          files: |
            dist/openclaw-manager-amd64.exe
            dist/SHA256SUMS.txt
```

---

## 八、v1.1 路线图（范围外，明确列出）

| 功能 | 说明 | 前提条件 |
|------|------|---------|
| **离线安装包** | 提供含内置 Node.js 的胖版 exe（~60MB），`go:embed` 方案重启 | v1.0 在线版稳定后 |
| **`openclaw://` 协议** | 写 `HKCU\Software\Classes\openclaw`，从网页唤起 Gateway | v1.0 稳定后 |
| **系统托盘** | 常驻后台，需重构为 Wails/WebView 架构 | 架构评审后 |
| **自动更新** | 启动时对比 GitHub Releases，提示可用更新 | 需稳定分发渠道 |
| **多 Gateway 实例** | 多 API Key 配置，manifest 结构需重新设计 | v1.1 schema 升级 |

---

## 九、关键风险与应对

| 风险 | 概率 | 应对 |
|------|------|------|
| npmmirror 下载失败 | 低 | 自动回退 nodejs.org；两者均失败则提示用户检查网络，退出保留已下载进度 |
| SmartScreen 拦截 | 高（无签名时） | 分发页面附操作截图 + SHA256；长期申请 EV 证书 |
| Windows Defender 误报 | 中 | syscheck 检测 Defender 状态，完成界面提供白名单指引 |
| API 服务商端点变更 | 低 | 校验 URL 写入 `openclaw.json` 可配置字段，非硬编码 |
| manifest 被手动删除 | 中 | 启动时无 manifest 但 runtime 目录存在 → "检测到不完整安装，是否修复？" |
| Win10 兼容性问题 | 低（Win11 优先） | CI 矩阵测试 Win10 22H2；不阻断发布，记录为 known issue |
| 用户 Node 环境冲突 | 低（影子路径隔离） | ShadowEnv 完全隔离，理论上零冲突 |
