// team-standards-installer —— 团队规范一键安装 App
//
// 原生窗口（macOS WKWebView），无浏览器弹窗，无终端窗口。
// 双击 .app 启动，关闭窗口即退出进程。
//
// 内部架构：背景启动 HTTP loopback 服务（127.0.0.1:随机端口）提供 API，
// 前端 UI 通过 WKWebView 加载；对用户完全透明。
package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	webview "github.com/webview/webview_go"
)

//go:embed all:assets all:standards all:demos all:claude all:cursor all:codex all:scripts all:hooks VERSION CHANGELOG.md
var embeddedFS embed.FS

//go:embed web/index.html
var indexHTML []byte

//go:embed web/app.js
var appJS []byte

//go:embed web/style.css
var appCSS []byte

//go:embed all:web/vendor
var webVendor embed.FS

// newMux 组装全部 HTTP 路由。main() 和静态导出模式（export_static.go）共用。
func newMux() *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write(indexHTML)
	})
	mux.HandleFunc("GET /app.js", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/javascript; charset=utf-8")
		_, _ = w.Write(appJS)
	})
	mux.HandleFunc("GET /style.css", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/css; charset=utf-8")
		_, _ = w.Write(appCSS)
	})
	vendorSub, err := fs.Sub(webVendor, "web/vendor")
	if err != nil {
		log.Fatalf("web/vendor embed: %v", err)
	}
	mux.Handle("GET /vendor/", http.StripPrefix("/vendor/", http.FileServerFS(vendorSub)))

	// 所有 /api/* 端点用 logMiddleware 包一层，自动记录 HTTP 日志
	api := func(pattern string, h http.HandlerFunc) {
		mux.HandleFunc(pattern, logMiddleware(h))
	}
	api("GET /api/catalog", handleCatalog)
	api("GET /api/reference", handleReference)
	api("GET /api/skill-disk-file", handleSkillDiskFile)        // v1.7.40 读 ~/.claude/skills/gsd-*/SKILL.md
	// v1.8.7 · Superpowers skill
	api("GET /api/superpowers/status", handleSuperpowersStatus)
	api("POST /api/superpowers/install", handleSuperpowersInstall)
	api("POST /api/superpowers/uninstall", handleSuperpowersUninstall)
	api("POST /api/skill-uninstall", handleSkillUninstall)              // v1.7.41 卸载单个 skill (by id+source)
	api("POST /api/installed-skill-uninstall", handleInstalledSkillUninstall) // v1.7.43 按路径删（已装清单用）

	// v1.7.42 · 中文 gsd-help skill 装入用户环境
	api("GET /api/gsd-help-zh/status", handleGSDHelpZhStatus)
	api("POST /api/gsd-help-zh/install", handleGSDHelpZhInstall)
	api("POST /api/gsd-help-zh/uninstall", handleGSDHelpZhUninstall)
	api("POST /api/install", handleInstall)
	// 自定义 Skill
	api("GET /api/custom", handleCustomList)
	api("POST /api/custom", handleCustomUpsert)
	api("DELETE /api/custom", handleCustomDelete)
	api("GET /api/custom/presets", handleCustomPresets)
	// 分发
	api("GET /api/download-zip", handleDownloadZip)
	api("POST /api/save-zip", handleSaveZip)
	api("POST /api/save-dmg", handleSaveDMG)
	api("POST /api/save-shell", handleSaveShell)
	api("POST /api/reveal", handleReveal)
	// 脚手架
	api("GET /api/scaffold/check-template", handleScaffoldCheckTemplate)
	api("POST /api/scaffold/pick-folder", handleScaffoldPickFolder)
	api("POST /api/scaffold/create", handleScaffoldCreate)
	api("GET /api/scaffold/derive-remote", handleScaffoldDeriveRemote)
	// 已装清单
	api("GET /api/installed", handleInstalled)
	api("POST /api/agents-init", handleAgentsInit)
	api("POST /api/universal/install", handleUniversalInstall) // 通用 Agent 适配（Copilot/Cline/Roo/Windsurf/Continue/Gemini…）
	// 日志
	mux.HandleFunc("GET /api/logs", handleLogs)
	api("POST /api/logs/clear", handleLogsClear)
	// AI 人格（全局态度 prompt，唯一保留）
	api("GET /api/persona", handlePersonaGet)
	api("POST /api/persona", handlePersonaSave)
	api("DELETE /api/persona", handlePersonaRemove)
	// Claude Desktop 导出
	api("GET /api/claude-desktop/export-zip", handleCDExportZip)
	api("POST /api/claude-desktop/save-zip", handleCDSaveZip) // v1.7.18: 落盘 + 返路径
	api("GET /api/claude-desktop/skill-md", handleCDSkillMD)
	api("POST /api/open-app", handleOpenApp) // v1.7.18: 启动 Claude Desktop / Cursor 等
	// OrangeCat 提测文档 Skill 独立管理 + 模板编辑器
	api("GET /api/orangecat/status", handleOrangecatStatus)
	api("POST /api/orangecat/install", handleOrangecatInstall)
	api("POST /api/orangecat/uninstall", handleOrangecatUninstall)
	api("GET /api/orangecat/template", handleOrangecatTemplateGet)
	api("POST /api/orangecat/template", handleOrangecatTemplateSave)
	api("DELETE /api/orangecat/template", handleOrangecatTemplateReset)
	api("GET /api/orangecat/template/demo", handleOrangecatTemplateDemo)

	// Go Unit Test Skill 模块化安装
	api("GET /api/unit-test/manifest", handleUnittestManifest)
	api("GET /api/unit-test/status", handleUnittestStatus)
	api("POST /api/unit-test/install", handleUnittestInstall)
	api("POST /api/unit-test/uninstall", handleUnittestUninstall)

	// 规范覆盖检查
	api("GET /api/coverage/check", handleCoverageCheck)

	// 项目级 skill 扫描 + 同步（v1.7.21）
	api("GET /api/project-skills/scan", handleProjectSkillsScan)
	api("POST /api/project-skills/sync", handleProjectSkillsSync)

	// Slash Commands 安装（v1.7.23）
	api("GET /api/commands/list", handleCommandsList)
	api("POST /api/commands/install", handleCommandsInstall)
	api("POST /api/commands/uninstall", handleCommandsUninstall)

	// code-review Skill（v1.7.25 自动评审）
	api("GET /api/code-review/status", handleCodeReviewStatus)
	api("POST /api/code-review/install", handleCodeReviewInstall)
	api("POST /api/code-review/uninstall", handleCodeReviewUninstall)

	// dev-dna 个人开发档案 Skill (v1.7.16)
	api("GET /api/dev-dna/status", handleDevDNAStatus)
	api("POST /api/dev-dna/install", handleDevDNAInstall)
	api("POST /api/dev-dna/uninstall", handleDevDNAUninstall)
	api("GET /api/dev-dna/profile", handleDevDNAProfileGet)
	api("POST /api/dev-dna/profile", handleDevDNAProfileSave)
	api("DELETE /api/dev-dna/profile", handleDevDNAProfileReset)

	// 规范云端同步（v1.7.30：从公开 GitLab mirror 拉最新 standards/*.md）
	api("GET /api/standards-sync/config", handleStandardsSyncConfigGet)
	api("POST /api/standards-sync/config", handleStandardsSyncConfigSave)
	api("GET /api/standards-sync/check", handleStandardsSyncCheck)
	api("POST /api/standards-sync/pull", handleStandardsSyncPull)

	// GSD · Getting Shit Done · 个人 GTD 流程 (v1.7.32) → 后改为 gsd-build 框架壳 (v1.7.39)
	api("GET /api/gsd/status", handleGSDStatus)
	api("POST /api/gsd/install", handleGSDInstall)
	api("GET /api/gsd/install/progress", handleGSDInstallProgress) // v1.7.45 异步 install 进度
	api("POST /api/gsd/uninstall", handleGSDUninstall)
	api("GET /api/gsd/list", handleGSDListGet)

	// Claude Desktop / Cowork 直装支持（v1.7.35）
	api("GET /api/cowork/probe", handleCoworkProbe)
	api("POST /api/cowork/install", handleCoworkInstall)

	// 提交规范化（v1.7.37：pre-commit hook + CI 配置）
	api("GET /api/commit-guard/status", handleCommitGuardStatus)
	api("POST /api/commit-guard/install", handleCommitGuardInstall)
	api("POST /api/commit-guard/uninstall", handleCommitGuardUninstall)
	api("POST /api/commit-guard/check", handleCommitGuardCheck)
	api("GET /api/commit-guard/scripts", handleCommitGuardScripts)

	// App 自动更新（v1.7.34：从公开 GitLab Releases 检测新版，L1 通知 + 引导）
	api("GET /api/app-update/check", handleAppUpdateCheck)
	api("POST /api/app-update/skip", handleAppUpdateSkip)
	api("POST /api/app-update/unskip", handleAppUpdateUnskip)
	api("POST /api/app-update/open-download", handleAppUpdateOpenDownload)
	// v1.7.38 · L2 一键自升级
	api("POST /api/app-update/auto-install", handleAppUpdateAutoInstall)
	api("GET /api/app-update/progress", handleAppUpdateProgress)
	// v1.7.38 · 全局代理配置（影响所有 HTTP 出站）
	api("GET /api/proxy/config", handleProxyConfigGet)
	api("POST /api/proxy/config", handleProxyConfigSave)
	// v1.8.5 · 强制约束 Hooks 管理
	api("GET /api/hooks", handleHooksList)
	api("POST /api/hooks/install", handleHooksInstall)
	api("POST /api/hooks/uninstall", handleHooksUninstall)
	api("POST /api/hooks/toggle", handleHooksToggle)

	return mux
}

func main() {
	// 静态导出模式：TS_EXPORT_STATIC=<dir> 时把只读 GET 接口快照成 JSON 后退出，
	// 供 GitHub Pages（PWA 静态版）使用，不开窗口、不起后台任务
	if dir := os.Getenv("TS_EXPORT_STATIC"); dir != "" {
		if err := exportStaticAPI(newMux(), dir); err != nil {
			log.Fatalf("export static: %v", err)
		}
		return
	}

	// 启动时一次性清理旧版本（<=1.3.0）在磁盘上留下的遗留片段
	cleanupLegacyArtifacts()

	mux := newMux()

	// v1.7.46: 启动 + 每小时 ticker
	//   - App 更新：仅刷新 LastSeenVersion，前端 banner 自取
	//   - 规范同步：SHA 不同 → 静默自动覆盖 references/（无需重启 Claude/Cursor）
	go appUpdateBackgroundLoop()
	go standardsSyncBackgroundLoop()

	// 1. 启动 HTTP loopback 服务
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		log.Fatalf("listen failed: %v", err)
	}
	url := "http://" + ln.Addr().String()
	log.Printf("internal HTTP on %s", url)
	go func() {
		srv := &http.Server{
			Handler:      mux,
			ReadTimeout:  5 * time.Second,
			WriteTimeout: 30 * time.Second,
		}
		if err := srv.Serve(ln); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()
	time.Sleep(100 * time.Millisecond) // 让 server 起来

	// v1.7.46: 装 macOS 原生菜单栏 → Cmd+C/V/X/A/Z 等 OS 标准 shortcut 走 NSMenu 路由
	// 这比纯靠 JS keydown handler 可靠 10 倍（JS 在某些 WKWebView 配置下收不到事件）
	//
	// ⚠️ 必须在 webview.New 之前调用：webview_go 创建窗口时若发现 NSApp 没装 mainMenu，
	// 会自己临时挂一个空的，导致后装的 Edit 菜单被遮蔽，Cmd+C/V 不生效（v1.7.45 的 bug）。
	installNativeAppMenu()

	// 2. 打开原生 WebView 窗口
	w := webview.New(false)
	defer w.Destroy()
	w.SetTitle("Team Standards")
	w.SetSize(1100, 760, webview.HintNone)

	// v1.7.14: 修 Cmd+Q 不生效。
	// webview_go 用 WKWebView，不会自动装 macOS 标准 NSApp 菜单（含 Quit/Cmd+Q），
	// 所以系统级 Cmd+Q 信号根本到不了 App。
	// 我们 bind 一个 JS 调用 Go 的桥，前端监听 keydown，命中 Cmd+Q 就 terminate window。
	_ = w.Bind("appQuit", func() {
		go func() {
			time.Sleep(50 * time.Millisecond) // 让 JS handler 返回
			w.Terminate()
		}()
	})

	w.Navigate(url)
	w.Run() // 阻塞到窗口关闭
}

// ---- 响应辅助 ----

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

func writeError(w http.ResponseWriter, status int, msg string, err error) {
	log.Printf("API error: %s: %v", msg, err)
	body := map[string]string{"error": msg}
	if err != nil {
		body["detail"] = err.Error()
	}
	writeJSON(w, status, body)
}

func readEmbedFile(path string) ([]byte, error) {
	b, err := embeddedFS.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}
	return b, nil
}
