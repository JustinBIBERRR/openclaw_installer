package proxy

import (
	"net/http"
	"net/url"
	"time"
)

// Detect 三层代理探测，返回探测到的代理 URL（未找到返回 nil）
// 探测顺序：HTTPS_PROXY 环境变量 → WinINet 注册表 → 直连
func Detect() *url.URL {
	// 第一层：标准环境变量（Go net/http 原生支持）
	if u := fromEnv(); u != nil {
		return u
	}
	// 第二层：Windows WinINet 注册表（Clash/V2Ray 等工具写入的系统代理）
	if u := fromWinINet(); u != nil {
		return u
	}
	return nil
}

// NewHTTPClient 返回一个已配置代理的 http.Client
func NewHTTPClient(timeout time.Duration) *http.Client {
	proxy := Detect()
	transport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
	}
	if proxy != nil {
		transport.Proxy = http.ProxyURL(proxy)
	}
	return &http.Client{
		Transport: transport,
		Timeout:   timeout,
	}
}
