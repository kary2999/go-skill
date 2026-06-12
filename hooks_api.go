package main

// 强制约束 Hooks 管理
//
// 设计：
//   - App 内嵌一批 hook 脚本（hooks/ 目录），每个 hook 有元数据
//   - 安装 = 写到 ~/.claude/hooks/<name>.sh（可执行）
//   - 禁用 = 重命名为 <name>.sh.disabled
//   - 启用 = 重命名回 <name>.sh
//   - 卸载 = 删文件（.sh 和 .disabled 都删）
//
// API：
//   GET  /api/hooks          → 列出所有可用 hook + 状态
//   POST /api/hooks/install  → 安装 hook
//   POST /api/hooks/uninstall→ 卸载 hook
//   POST /api/hooks/toggle   → 切换启用/禁用

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
)

// hookMeta 描述一个可分发的 hook
type hookMeta struct {
	ID          string `json:"id"`           // 文件名不含 .sh，如 "devdefender-guard"
	Name        string `json:"name"`         // 显示名
	Description string `json:"description"`  // 一行说明
	HookType    string `json:"hook_type"`    // UserPromptSubmit / PostToolUse / PreToolUse
	EmbedPath   string `json:"-"`            // embed 里的路径
}

// 所有可分发的 hook（与 hooks/ 目录对应）
var availableHooks = []hookMeta{
	// ── 一体化守卫（推荐，替代下方所有单项）──
	{
		ID:          "team-guard",
		Name:        "🛡 团队规范一体化守卫",
		Description: "一个 hook 覆盖全部时机：UserPromptSubmit=DevDefender需求防御 / PreToolUse=提交前全量铁律扫描+commit格式 / PostToolUse=写文件后铁律检测（Go/SQL/Proto）",
		HookType:    "UserPromptSubmit+PreToolUse+PostToolUse",
		EmbedPath:   "hooks/team-guard.sh",
	},
	// ── 单项 hook（可单独安装，与 team-guard 二选一）──
	{
		ID:          "devdefender-guard",
		Name:        "🛡️ DevDefender 需求防御",
		Description: "用户输入含产品需求时注入 PRD 硬约束，防止 AI 脑补需求、要技术方案、写前端细节",
		HookType:    "UserPromptSubmit",
		EmbedPath:   "hooks/devdefender-guard.sh",
	},
	{
		ID:          "go-panic-guard",
		Name:        "🚨 Go 禁裸 panic",
		Description: "写 .go 文件后检测业务代码中的裸 panic()，要求改用 xerror.New / xerror.Wrap",
		HookType:    "PostToolUse",
		EmbedPath:   "hooks/go-panic-guard.sh",
	},
	{
		ID:          "camelcase-guard",
		Name:        "🔤 JSON 禁驼峰字段",
		Description: "写 .go / .json / .md 文件后检测 JSON key 驼峰命名，要求改为 snake_case",
		HookType:    "PostToolUse",
		EmbedPath:   "hooks/camelcase-guard.sh",
	},
	{
		ID:          "commit-format-guard",
		Name:        "📝 Commit Message 格式",
		Description: "拦截 git commit，校验是否符合 Conventional Commits 规范（feat/fix/docs…）",
		HookType:    "PreToolUse",
		EmbedPath:   "hooks/commit-format-guard.sh",
	},
	// ── 铁律衍生 Hook（v1.8.9）──
	{
		ID:          "user-id-guard",
		Name:        "🪪 禁用 user_id 字段",
		Description: "写 .go/.sql/.proto 后检测 user_id，全平台统一用 uid（field-naming §1.2）",
		HookType:    "PostToolUse",
		EmbedPath:   "hooks/user-id-guard.sh",
	},
	{
		ID:          "float-amount-guard",
		Name:        "💰 金额禁 float",
		Description: "写 .go 后检测 float32/float64 存金额/价格字段，必须用 decimal（database §字段规范）",
		HookType:    "PostToolUse",
		EmbedPath:   "hooks/float-amount-guard.sh",
	},
	{
		ID:          "fmt-println-guard",
		Name:        "🪵 禁用 fmt.Println",
		Description: "写 .go 后检测业务代码中 fmt.Println/Printf，必须改用 slog 结构化日志（go-style §7）",
		HookType:    "PostToolUse",
		EmbedPath:   "hooks/fmt-println-guard.sh",
	},
	{
		ID:          "error-ignore-guard",
		Name:        "⚠️ 禁止忽略 error",
		Description: "写 .go 后检测 _, _ := someFunc() 的 error 丢弃模式（go-style §4.1）",
		HookType:    "PostToolUse",
		EmbedPath:   "hooks/error-ignore-guard.sh",
	},
	{
		ID:          "soft-delete-guard",
		Name:        "🗑 软删除用 deleted_at",
		Description: "写 .sql/.go/.proto 后检测 is_deleted，软删除统一用 deleted_at TIMESTAMPTZ（database §字段规范）",
		HookType:    "PostToolUse",
		EmbedPath:   "hooks/soft-delete-guard.sh",
	},
	{
		ID:          "goroutine-naked-guard",
		Name:        "🔀 禁裸启 goroutine",
		Description: "写 .go 后检测裸 go func()，必须通过 errgroup 管理生命周期（go-style §5）",
		HookType:    "PostToolUse",
		EmbedPath:   "hooks/goroutine-naked-guard.sh",
	},
	// ── 提交前全量守卫（v1.8.11）──
	{
		ID:          "git-commit-guard",
		Name:        "🔍 提交前全量守卫",
		Description: "git commit 前一次性检测：Commit Message 格式 + 全部铁律（Go/SQL/Proto） + SQL 表规范",
		HookType:    "PreToolUse",
		EmbedPath:   "hooks/git-commit-guard.sh",
	},
}

type hookStatus struct {
	hookMeta
	Installed     bool   `json:"installed"`
	Enabled       bool   `json:"enabled"`         // installed && not .disabled
	InstalledPath string `json:"installed_path,omitempty"`
	Scope         string `json:"scope"`           // global / project
}

// resolveHooksDir 根据 scope + projectPath 返回 hooks 目录
//   scope="global"  → ~/.claude/hooks/
//   scope="project" → <projectPath>/.claude/hooks/
func resolveHooksDir(scope, projectPath string) (string, error) {
	switch scope {
	case "project":
		if projectPath == "" {
			return "", fmt.Errorf("scope=project 时 path 必填")
		}
		abs, err := filepath.Abs(projectPath)
		if err != nil {
			return "", err
		}
		return filepath.Join(abs, ".claude", "hooks"), nil
	default: // global
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(home, ".claude", "hooks"), nil
	}
}

func getHookStatus(h hookMeta, scope, projectPath string) hookStatus {
	dir, err := resolveHooksDir(scope, projectPath)
	if err != nil {
		return hookStatus{hookMeta: h, Scope: scope}
	}
	activePath   := filepath.Join(dir, h.ID+".sh")
	disabledPath := filepath.Join(dir, h.ID+".sh.disabled")

	if _, err := os.Stat(activePath); err == nil {
		return hookStatus{hookMeta: h, Installed: true, Enabled: true, InstalledPath: activePath, Scope: scope}
	}
	if _, err := os.Stat(disabledPath); err == nil {
		return hookStatus{hookMeta: h, Installed: true, Enabled: false, InstalledPath: disabledPath, Scope: scope}
	}
	return hookStatus{hookMeta: h, Installed: false, Enabled: false, Scope: scope}
}

// hookBaseReq 所有 hook 请求的公共字段
type hookBaseReq struct {
	ID    string `json:"id"`
	Scope string `json:"scope"`       // global（默认）| project
	Path  string `json:"path"`        // scope=project 时必填：项目根目录
}

func findHookMeta(id string) *hookMeta {
	for i, h := range availableHooks {
		if h.ID == id {
			return &availableHooks[i]
		}
	}
	return nil
}

// GET /api/hooks?scope=global|project&path=<dir>
func handleHooksList(w http.ResponseWriter, r *http.Request) {
	scope := r.URL.Query().Get("scope")
	path  := r.URL.Query().Get("path")
	if scope == "" {
		scope = "global"
	}
	var list []hookStatus
	for _, h := range availableHooks {
		list = append(list, getHookStatus(h, scope, path))
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "hooks": list, "scope": scope})
}

// POST /api/hooks/install  body: {"id":"...", "scope":"global|project", "path":"..."}
func handleHooksInstall(w http.ResponseWriter, r *http.Request) {
	var req hookBaseReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.ID == "" {
		writeError(w, http.StatusBadRequest, "id 必填", nil)
		return
	}
	if req.Scope == "" {
		req.Scope = "global"
	}
	meta := findHookMeta(req.ID)
	if meta == nil {
		writeError(w, http.StatusBadRequest, "未知 hook id: "+req.ID, nil)
		return
	}
	dir, err := resolveHooksDir(req.Scope, req.Path)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error(), nil)
		return
	}
	if err := os.MkdirAll(dir, 0755); err != nil {
		writeError(w, http.StatusInternalServerError, "创建 hooks 目录失败", err)
		return
	}
	content, err := readEmbedFile(meta.EmbedPath)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "读 hook 脚本失败", err)
		return
	}
	_ = os.Remove(filepath.Join(dir, meta.ID+".sh.disabled"))
	dst := filepath.Join(dir, meta.ID+".sh")
	if err := os.WriteFile(dst, content, 0755); err != nil {
		writeError(w, http.StatusInternalServerError, "写 hook 文件失败", err)
		return
	}

	// team-guard 额外写入 settings.json（三个事件注册）
	var settingsNote string
	if meta.ID == "team-guard" {
		if err := injectTeamGuardSettings(req.Scope, req.Path, dst); err != nil {
			settingsNote = "脚本已安装，但写 settings.json 失败：" + err.Error()
		} else {
			settingsNote = "已自动写入 settings.json（UserPromptSubmit + PreToolUse + PostToolUse）"
		}
	}

	resp := map[string]any{"ok": true, "path": dst, "status": getHookStatus(*meta, req.Scope, req.Path)}
	if settingsNote != "" {
		resp["settings_note"] = settingsNote
	}
	writeJSON(w, http.StatusOK, resp)
}

// POST /api/hooks/uninstall  body: {"id":"...", "scope":"...", "path":"..."}
func handleHooksUninstall(w http.ResponseWriter, r *http.Request) {
	var req hookBaseReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.ID == "" {
		writeError(w, http.StatusBadRequest, "id 必填", nil)
		return
	}
	if req.Scope == "" {
		req.Scope = "global"
	}
	dir, err := resolveHooksDir(req.Scope, req.Path)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error(), nil)
		return
	}
	scriptPath := filepath.Join(dir, req.ID+".sh")
	_ = os.Remove(scriptPath)
	_ = os.Remove(filepath.Join(dir, req.ID+".sh.disabled"))

	// team-guard 卸载时同步清理 settings.json
	if req.ID == "team-guard" {
		_ = removeTeamGuardSettings(req.Scope, req.Path, scriptPath)
	}

	meta := findHookMeta(req.ID)
	if meta == nil {
		meta = &hookMeta{ID: req.ID}
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "status": getHookStatus(*meta, req.Scope, req.Path)})
}

// POST /api/hooks/toggle  body: {"id":"...", "scope":"...", "path":"..."}
func handleHooksToggle(w http.ResponseWriter, r *http.Request) {
	var req hookBaseReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.ID == "" {
		writeError(w, http.StatusBadRequest, "id 必填", nil)
		return
	}
	if req.Scope == "" {
		req.Scope = "global"
	}
	dir, err := resolveHooksDir(req.Scope, req.Path)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error(), nil)
		return
	}
	activePath   := filepath.Join(dir, req.ID+".sh")
	disabledPath := filepath.Join(dir, req.ID+".sh.disabled")

	meta := findHookMeta(req.ID)
	if meta == nil {
		meta = &hookMeta{ID: req.ID}
	}

	if _, err := os.Stat(activePath); err == nil {
		if err := os.Rename(activePath, disabledPath); err != nil {
			writeError(w, http.StatusInternalServerError, "禁用失败", err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"ok": true, "enabled": false, "status": getHookStatus(*meta, req.Scope, req.Path)})
		return
	}
	if _, err := os.Stat(disabledPath); err == nil {
		if err := os.Rename(disabledPath, activePath); err != nil {
			writeError(w, http.StatusInternalServerError, "启用失败", err)
			return
		}
		_ = os.Chmod(activePath, 0755)
		writeJSON(w, http.StatusOK, map[string]any{"ok": true, "enabled": true, "status": getHookStatus(*meta, req.Scope, req.Path)})
		return
	}
	writeError(w, http.StatusBadRequest, "hook 未安装，请先安装再切换", nil)
}
