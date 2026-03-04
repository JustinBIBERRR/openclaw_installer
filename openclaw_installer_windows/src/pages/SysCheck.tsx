import { useEffect, useState } from "react";
import { invoke } from "@tauri-apps/api/core";

const isTauri = typeof window !== "undefined" && "__TAURI__" in window;
import { CheckCircle, XCircle, AlertCircle, Loader, FolderOpen } from "lucide-react";
import type { SysCheckItem } from "../types";

interface Props {
  onDone: (installDir: string) => void;
}

const DEFAULT_INSTALL_DIR = "C:\\OpenClaw";

export default function SysCheck({ onDone }: Props) {
  const [checks, setChecks] = useState<SysCheckItem[]>([
    { key: "admin",   label: "管理员权限",       status: "checking", detail: "检测中..." },
    { key: "webview2",label: "WebView2 运行时",   status: "checking", detail: "检测中..." },
    { key: "disk",    label: "磁盘空间 (≥ 2GB)", status: "checking", detail: "检测中..." },
    { key: "port",    label: "端口 18789 可用",   status: "checking", detail: "检测中..." },
    { key: "path",    label: "安装路径合法",       status: "checking", detail: "检测中..." },
    { key: "network", label: "网络连通性",         status: "checking", detail: "检测中..." },
  ]);
  const [installDir, setInstallDir] = useState(DEFAULT_INSTALL_DIR);
  const [done, setDone] = useState(false);
  const [hasError, setHasError] = useState(false);

  const updateCheck = (key: string, update: Partial<SysCheckItem>) => {
    setChecks((prev) =>
      prev.map((c) => (c.key === key ? { ...c, ...update } : c))
    );
  };

  useEffect(() => {
    runChecks(installDir);
  }, []);

  async function runChecks(dir: string) {
    setDone(false);
    setHasError(false);
    setChecks((prev) => prev.map((c) => ({ ...c, status: "checking", detail: "检测中..." })));

    if (!isTauri) {
      // 浏览器预览模式：模拟全部通过
      setTimeout(() => {
        setChecks((prev) => prev.map((c) => ({ ...c, status: "ok", detail: "预览模式（模拟通过）" })));
        setDone(true);
      }, 800);
      return;
    }

    try {
      const result = await invoke<{
        admin: boolean;
        webview2: boolean;
        disk_gb: number;
        port: number;
        path_valid: boolean;
        path_issue: string;
        network_ok: boolean;
        suggested_dir: string;
      }>("run_syscheck", { installDir: dir });

      updateCheck("admin", {
        status: result.admin ? "ok" : "error",
        detail: result.admin ? "已获得管理员权限" : "需要以管理员身份运行",
      });
      updateCheck("webview2", {
        status: result.webview2 ? "ok" : "warn",
        detail: result.webview2 ? "已安装" : "未检测到，将在安装时自动处理",
      });
      updateCheck("disk", {
        status: result.disk_gb >= 2 ? "ok" : "warn",
        detail: `可用空间: ${result.disk_gb.toFixed(1)} GB`,
      });
      updateCheck("port", {
        status: "ok",
        detail: result.port === 18789 ? "端口 18789 可用" : `端口 18789 已占用，将使用 ${result.port}`,
      });
      updateCheck("path", {
        status: result.path_valid ? "ok" : "warn",
        detail: result.path_valid ? dir : result.path_issue,
      });
      if (!result.path_valid && result.suggested_dir) {
        setInstallDir(result.suggested_dir);
      }
      updateCheck("network", {
        status: result.network_ok ? "ok" : "warn",
        detail: result.network_ok
          ? "可以访问 npmmirror.com"
          : "网络受限，安装可能较慢，建议检查代理设置",
      });

      const errorExists = !result.admin;
      setHasError(errorExists);
      setDone(true);
    } catch (e) {
      setHasError(true);
      setDone(true);
    }
  }

  const statusIcon = (status: SysCheckItem["status"]) => {
    if (status === "checking") return <Loader size={16} className="text-gray-400 animate-spin" />;
    if (status === "ok")       return <CheckCircle size={16} className="text-brand-400" />;
    if (status === "warn")     return <AlertCircle size={16} className="text-yellow-400" />;
    return <XCircle size={16} className="text-red-400" />;
  };

  return (
    <div className="h-full flex flex-col px-6 py-4 gap-4 overflow-y-auto">
      <div>
        <h2 className="text-lg font-semibold text-gray-100">系统预检</h2>
        <p className="text-sm text-gray-400 mt-0.5">确认运行环境满足安装要求</p>
      </div>

      {/* 安装目录 */}
      <div className="bg-gray-900 rounded-lg border border-gray-700 p-3">
        <label className="block text-xs text-gray-400 mb-1.5">安装目录</label>
        <div className="flex gap-2">
          <div className="flex-1 flex items-center gap-2 bg-gray-800 rounded border border-gray-600 px-3 py-1.5">
            <FolderOpen size={14} className="text-gray-400 flex-shrink-0" />
            <input
              type="text"
              value={installDir}
              onChange={(e) => setInstallDir(e.target.value)}
              className="flex-1 bg-transparent text-sm text-gray-200 outline-none"
              style={{ userSelect: "text" }}
            />
          </div>
          <button
            onClick={() => runChecks(installDir)}
            className="px-3 py-1.5 text-xs bg-gray-700 hover:bg-gray-600 rounded border border-gray-600 text-gray-300 transition-colors whitespace-nowrap"
          >
            重新检测
          </button>
        </div>
        <p className="text-[10px] text-gray-500 mt-1">
          建议使用英文路径（如 C:\OpenClaw），避免中文和空格
        </p>
      </div>

      {/* 检测列表 */}
      <div className="bg-gray-900 rounded-lg border border-gray-700 divide-y divide-gray-800">
        {checks.map((c) => (
          <div key={c.key} className="flex items-center gap-3 px-4 py-3">
            <div className="flex-shrink-0">{statusIcon(c.status)}</div>
            <div className="flex-1 min-w-0">
              <div className="text-sm text-gray-200">{c.label}</div>
              <div className={`text-xs mt-0.5 truncate
                ${c.status === "ok"   ? "text-gray-500" : ""}
                ${c.status === "warn" ? "text-yellow-500" : ""}
                ${c.status === "error"? "text-red-400" : ""}
                ${c.status === "checking" ? "text-gray-600" : ""}
              `}>{c.detail}</div>
            </div>
          </div>
        ))}
      </div>

      <div className="flex-1" />

      {/* 底部操作 */}
      {done && (
        <div className="flex items-center justify-between">
          {hasError ? (
            <p className="text-sm text-red-400">请先解决上方标红的问题再继续</p>
          ) : (
            <p className="text-sm text-gray-500">
              {checks.some(c => c.status === "warn")
                ? "存在警告项，但不影响安装"
                : "所有检测通过"}
            </p>
          )}
          <button
            disabled={hasError}
            onClick={() => onDone(installDir)}
            className="px-6 py-2 bg-brand-500 hover:bg-brand-600 disabled:bg-gray-700 disabled:text-gray-500
              text-gray-950 font-semibold text-sm rounded-lg transition-colors disabled:cursor-not-allowed"
          >
            开始安装 →
          </button>
        </div>
      )}
    </div>
  );
}
