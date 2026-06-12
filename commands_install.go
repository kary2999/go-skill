// Slash Commands 安装（v1.7.23）
//
// 把 5 个 .md slash command 装到：
//   ~/.claude/commands/<name>.md   ← Claude Code 用 /<name> 调用
//   ~/.cursor/commands/<name>.md   ← Cursor 同样支持
//
// 与 Skill 不同：Skill 是被动激活（模型读 description 决定），
// Slash Command 是用户**主动** /<name> 显式调用，意图明确。
//
// 适合"用户主动想做某件事"的场景：
//   /tech-design / /design-table / /api-doc / /review / /tixuebj

package main

import (
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// CommandMeta UI 列表用
type CommandMeta struct {
	ID          string `json:"id"`          // 文件名（无扩展），用作 /<id>
	Title       string `json:"title"`       // UI 显示名
	Description string `json:"description"` // 一句话用途（从 frontmatter 读）
	UsageHint   string `json:"usage_hint"`  // /id <参数提示>
}

var slashCommands = []CommandMeta{
	{
		ID: "tech-design", Title: "📐 写技术方案",
		Description: "按 tech-design-example.md 7 段范例展开技术方案",
		UsageHint:   "/tech-design 用户中台 v2 改造",
	},
	{
		ID: "design-table", Title: "🗄️ 设计 SQL 表",
		Description: "按 database.md + 命名规范产出完整 CREATE TABLE",
		UsageHint:   "/design-table orders 订单表，含币种、价格、状态",
	},
	{
		ID: "api-doc", Title: "🔌 写接口文档",
		Description: "按 api-doc-example.md 格式产出接口文档",
		UsageHint:   "/api-doc POST /api/v1/orders 创建订单",
	},
	{
		ID: "review", Title: "🔍 团队规范 review",
		Description: "按 14 条铁律 + dev-dna 偏好 review git diff / 当前文件",
		UsageHint:   "/review     /review diff     /review path/to/file.go",
	},
	{
		ID: "tixuebj", Title: "🐱 主动触发提测（OrangeCat）",
		Description: "显式调用 OrangeCat 生成提测报告（比模糊触发更可靠）",
		UsageHint:   "/tixuebj v1.0.0",
	},
	// 注：v1.7.39 起 /gsd 命令交给 gsd-build/get-shit-done 框架（npx 装），
	// 这里不再注册我们简版的 /gsd
}

func handleCommandsList(w http.ResponseWriter, r *http.Request) {
	home, _ := os.UserHomeDir()
	claudeDir := filepath.Join(home, ".claude", "commands")
	cursorDir := filepath.Join(home, ".cursor", "commands")

	// 探测每个命令是否已装
	type cmdStatus struct {
		CommandMeta
		ClaudeInstalled bool `json:"claude_installed"`
		CursorInstalled bool `json:"cursor_installed"`
	}
	var out []cmdStatus
	for _, m := range slashCommands {
		s := cmdStatus{CommandMeta: m}
		if _, err := os.Stat(filepath.Join(claudeDir, m.ID+".md")); err == nil {
			s.ClaudeInstalled = true
		}
		if _, err := os.Stat(filepath.Join(cursorDir, m.ID+".md")); err == nil {
			s.CursorInstalled = true
		}
		out = append(out, s)
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"commands":   out,
		"claude_dir": claudeDir,
		"cursor_dir": cursorDir,
	})
}

func handleCommandsInstall(w http.ResponseWriter, r *http.Request) {
	home, err := os.UserHomeDir()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "home", err)
		return
	}
	claudeDir := filepath.Join(home, ".claude", "commands")
	cursorDir := filepath.Join(home, ".cursor", "commands")
	for _, d := range []string{claudeDir, cursorDir} {
		if err := os.MkdirAll(d, 0755); err != nil {
			writeError(w, http.StatusInternalServerError, "mkdir "+d, err)
			return
		}
	}

	var installed []string
	var warnings []string

	// 从 embed 拷贝所有 claude/commands/*.md 到两个目录
	_ = fs.WalkDir(embeddedFS, "claude/commands", func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		if !strings.HasSuffix(p, ".md") {
			return nil
		}
		base := filepath.Base(p)
		data, err := embeddedFS.ReadFile(p)
		if err != nil {
			warnings = append(warnings, "读 "+p+" 失败："+err.Error())
			return nil
		}
		for label, dst := range map[string]string{"Claude": claudeDir, "Cursor": cursorDir} {
			target := filepath.Join(dst, base)
			if err := os.WriteFile(target, data, 0644); err != nil {
				warnings = append(warnings, label+" 写 "+target+" 失败："+err.Error())
				continue
			}
			installed = append(installed, label+" → "+target)
		}
		return nil
	})

	writeJSON(w, http.StatusOK, map[string]any{
		"ok":        len(installed) > 0,
		"installed": installed,
		"warnings":  warnings,
		"count":     len(installed),
	})
}

func handleCommandsUninstall(w http.ResponseWriter, r *http.Request) {
	home, _ := os.UserHomeDir()
	dirs := []string{
		filepath.Join(home, ".claude", "commands"),
		filepath.Join(home, ".cursor", "commands"),
	}
	var removed []string
	for _, dir := range dirs {
		for _, m := range slashCommands {
			p := filepath.Join(dir, m.ID+".md")
			if _, err := os.Stat(p); err == nil {
				if err := os.Remove(p); err == nil {
					removed = append(removed, p)
				}
			}
		}
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "removed": removed})
}
