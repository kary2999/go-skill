package main

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// InstallRequest 安装请求
//
//	scope  = "global" | "project"
//	target = "claude" | "cursor" | "codex" | "all"
//	path   = 项目根目录（scope=project 时必填）
type InstallRequest struct {
	Scope  string `json:"scope"`
	Target string `json:"target"`
	Path   string `json:"path,omitempty"`
}

type InstallResult struct {
	OK         bool     `json:"ok"`
	Version    string   `json:"version"`
	Installed  []string `json:"installed"`
	ClaudeDir  string   `json:"claude_dir,omitempty"`
	CursorDir  string   `json:"cursor_dir,omitempty"`
	CodexDir   string   `json:"codex_dir,omitempty"`
	Warnings   []string `json:"warnings,omitempty"`
	Error      string   `json:"error,omitempty"`
}

func handleInstall(w http.ResponseWriter, r *http.Request) {
	var req InstallRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request", err)
		return
	}

	if req.Target == "" {
		req.Target = "all"
	}

	result := InstallResult{OK: true}
	verBytes, _ := readEmbedFile("VERSION")
	result.Version = strings.TrimSpace(string(verBytes))

	home, err := os.UserHomeDir()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "find home dir", err)
		return
	}

	var claudeBase, cursorBase, codexBase string
	switch req.Scope {
	case "global":
		claudeBase = filepath.Join(home, ".claude", "skills")
		cursorBase = filepath.Join(home, ".cursor", "rules")
		codexBase  = filepath.Join(home, ".codex", "skills")
	case "project":
		if req.Path == "" {
			writeError(w, http.StatusBadRequest, "path required for project scope", nil)
			return
		}
		abs, err := filepath.Abs(req.Path)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid path", err)
			return
		}
		if _, err := os.Stat(abs); err != nil {
			writeError(w, http.StatusBadRequest, "path does not exist", err)
			return
		}
		claudeBase = filepath.Join(abs, ".claude", "skills")
		cursorBase = filepath.Join(abs, ".cursor", "rules")
		codexBase  = filepath.Join(abs, ".codex", "skills")
	default:
		writeError(w, http.StatusBadRequest, "scope must be global or project", nil)
		return
	}

	if req.Target == "claude" || req.Target == "all" {
		dir := filepath.Join(claudeBase, "go-team-standards")
		if err := installClaude(dir); err != nil {
			result.OK = false
			result.Error = fmt.Sprintf("claude install failed: %v", err)
			writeJSON(w, http.StatusInternalServerError, result)
			return
		}
		_ = os.WriteFile(filepath.Join(dir, ".installed-version"), []byte(result.Version+"\n"), 0644)
		result.ClaudeDir = dir
		result.Installed = append(result.Installed, "Claude Skill → "+dir)

		// 清理老的 tixuebj-template 目录（一次性迁移到 OrangeCat）
		oldTixuebj := filepath.Join(claudeBase, "tixuebj-template")
		if dirExists(oldTixuebj) {
			if err := os.RemoveAll(oldTixuebj); err == nil {
				result.Installed = append(result.Installed, "已清理旧版 → "+oldTixuebj)
			}
		}
		// DevDefender skill
		ddDir := filepath.Join(claudeBase, "DevDefender")
		if err := installEmbeddedSkill("claude/DevDefender", ddDir); err != nil {
			result.Warnings = append(result.Warnings, "DevDefender skill 安装失败："+err.Error())
		} else {
			result.Installed = append(result.Installed, "DevDefender Skill → "+ddDir)
		}

		// DevDefender hook（仅 global scope 安装到 ~/.claude/hooks/）
		if req.Scope == "global" {
			hooksDir := filepath.Join(home, ".claude", "hooks")
			if err := os.MkdirAll(hooksDir, 0755); err == nil {
				hookDst := filepath.Join(hooksDir, "devdefender-guard.sh")
				if src, err := readEmbedFile("hooks/devdefender-guard.sh"); err == nil {
					if err := os.WriteFile(hookDst, src, 0755); err == nil {
						result.Installed = append(result.Installed, "DevDefender Hook → "+hookDst)
					}
				}
			}
		}

		// 注意：OrangeCat 与 go-unit-test 走独立安装接口（/api/orangecat/install 和 /api/unit-test/install），
		// 不再随主安装一起装，用户在 UI 里按需勾选。
	}

	if req.Target == "cursor" || req.Target == "all" {
		if err := installCursor(cursorBase); err != nil {
			result.OK = false
			result.Error = fmt.Sprintf("cursor install failed: %v", err)
			writeJSON(w, http.StatusInternalServerError, result)
			return
		}
		_ = os.WriteFile(filepath.Join(cursorBase, ".installed-version"), []byte(result.Version+"\n"), 0644)
		result.CursorDir = cursorBase
		result.Installed = append(result.Installed, "Cursor rules → "+cursorBase)

		// 额外：Cursor 官方 skill 路径
		// 项目 scope：<项目>/.cursor/skills-cursor/
		// 全局 scope：~/.cursor/skills-cursor/
		var cursorSkillsBase string
		if req.Scope == "global" {
			cursorSkillsBase = filepath.Join(home, ".cursor", "skills-cursor")
		} else {
			cursorSkillsBase = filepath.Join(req.Path, ".cursor", "skills-cursor")
		}
		cursorSkillDir := filepath.Join(cursorSkillsBase, "go-team-standards")
		if err := installClaude(cursorSkillDir); err != nil {
			// 非致命：记 warning 但不中断
			result.Warnings = append(result.Warnings,
				"Cursor skills-cursor 写入失败（.mdc 已装可用）："+err.Error())
		} else {
			result.Installed = append(result.Installed,
				"Cursor Skill (skills-cursor) → "+cursorSkillDir)
		}

		// 清理老的 Cursor 侧 tixuebj-template 目录
		oldCursorTixuebj := filepath.Join(cursorSkillsBase, "tixuebj-template")
		if dirExists(oldCursorTixuebj) {
			if err := os.RemoveAll(oldCursorTixuebj); err == nil {
				result.Installed = append(result.Installed, "已清理旧版 → "+oldCursorTixuebj)
			}
		}

		// Project 模式下顺手下发 .cursorignore（如已存在则警告跳过）
		if req.Scope == "project" {
			ignoreDst := filepath.Join(req.Path, ".cursorignore")
			if _, err := os.Stat(ignoreDst); err == nil {
				result.Warnings = append(result.Warnings,
					".cursorignore 已存在，未覆盖。请对照 cursor/.cursorignore.template 手动合并")
			} else {
				src, err := readEmbedFile("cursor/.cursorignore.template")
				if err == nil {
					if werr := os.WriteFile(ignoreDst, src, 0644); werr == nil {
						result.Installed = append(result.Installed, ".cursorignore → "+ignoreDst)
					}
				}
			}
		}
	}

	if req.Target == "codex" || req.Target == "all" {
		dir := filepath.Join(codexBase, "go-team-standards")
		if err := installEmbeddedSkill("codex/go-team-standards", dir); err != nil {
			result.Warnings = append(result.Warnings, "Codex skill 安装失败："+err.Error())
		} else {
			_ = os.WriteFile(filepath.Join(dir, ".installed-version"), []byte(result.Version+"\n"), 0644)
			result.CodexDir = dir
			result.Installed = append(result.Installed, "Codex Skill → "+dir)
		}
	}

	// 关键：重装结束后把用户的自定义 Skill 再写回去
	// 避免 installClaude 的 rm -rf 把 custom-*.md 一起清掉
	if req.Scope == "global" {
		if customs, err := loadCustomRules(); err == nil && len(customs) > 0 {
			if logs, err := syncCustomToInstalled(customs); err == nil && len(logs) > 0 {
				result.Installed = append(result.Installed,
					fmt.Sprintf("自定义 Skill 已同步 %d 条", len(customs)))
			}
		}
	}

	writeJSON(w, http.StatusOK, result)
}

// installEmbeddedSkill 通用：把 embed 里的任意 skill 目录展开到 dst
// 比 installClaude 更通用（后者只处理 go-team-standards + references + assets + demos）
func installEmbeddedSkill(embedDir, dst string) error {
	if err := os.RemoveAll(dst); err != nil {
		return err
	}
	if err := os.MkdirAll(dst, 0755); err != nil {
		return err
	}
	return fs.WalkDir(embeddedFS, embedDir, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(embedDir, p)
		target := filepath.Join(dst, rel)
		if d.IsDir() {
			return os.MkdirAll(target, 0755)
		}
		return copyEmbedFile(p, target)
	})
}

// installClaude 把嵌入的 claude/go-team-standards 整棵树展开到目标目录。
// references/ 需要展开成真实 md（embed 已把符号链接解析为真实文件）。
func installClaude(dst string) error {
	if err := os.RemoveAll(dst); err != nil {
		return err
	}
	if err := os.MkdirAll(dst, 0755); err != nil {
		return err
	}

	// 1. SKILL.md
	if err := copyEmbedFile("claude/go-team-standards/SKILL.md", filepath.Join(dst, "SKILL.md")); err != nil {
		return err
	}

	// 2. references/ 展开 standards/ 下所有文件（动态扫描，不依赖硬编码列表）
	// 好处：standards/ 加 / 删文件时 installClaude 自动跟进，不会因列表过期导致安装失败。
	refDir := filepath.Join(dst, "references")
	if err := os.MkdirAll(refDir, 0755); err != nil {
		return err
	}
	if err := copyEmbedDir("standards", refDir); err != nil {
		return err
	}

	// 3. assets/
	assetDir := filepath.Join(dst, "assets")
	if err := os.MkdirAll(assetDir, 0755); err != nil {
		return err
	}
	if err := copyEmbedFile("assets/.golangci.yml", filepath.Join(assetDir, ".golangci.yml")); err != nil {
		return err
	}

	// 4. demos/
	demoDst := filepath.Join(dst, "demos")
	return copyEmbedDir("demos", demoDst)
}

// installCursor 把嵌入的 cursor/rules 同步到目标目录。保留用户自定义规则（不按前缀匹配的文件）。
func installCursor(dst string) error {
	if err := os.MkdirAll(dst, 0755); err != nil {
		return err
	}
	// 清理旧版团队规则（前缀匹配），保留用户自定义 *.mdc
	entries, _ := os.ReadDir(dst)
	for _, e := range entries {
		name := e.Name()
		if isTeamRuleFile(name) {
			_ = os.Remove(filepath.Join(dst, name))
		}
	}
	// 写入当前嵌入的规则
	return fs.WalkDir(embeddedFS, "cursor/rules", func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		name := filepath.Base(path)
		target := filepath.Join(dst, name)
		return copyEmbedFile(path, target)
	})
}

func isTeamRuleFile(name string) bool {
	// team 规则命名：NN-xxx.mdc（2 位数字前缀）
	if !strings.HasSuffix(name, ".mdc") {
		return false
	}
	if len(name) < 4 {
		return false
	}
	return isDigit(name[0]) && isDigit(name[1]) && name[2] == '-'
}

func isDigit(c byte) bool { return c >= '0' && c <= '9' }

// copyEmbedFile 从 embed 读取并写到目标路径
func copyEmbedFile(src, dst string) error {
	b, err := embeddedFS.ReadFile(src)
	if err != nil {
		return fmt.Errorf("read embed %s: %w", src, err)
	}
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}
	return os.WriteFile(dst, b, 0644)
}

// copyEmbedDir 递归复制 embed 目录
func copyEmbedDir(srcDir, dstDir string) error {
	return fs.WalkDir(embeddedFS, srcDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(srcDir, path)
		target := filepath.Join(dstDir, rel)
		if d.IsDir() {
			return os.MkdirAll(target, 0755)
		}
		return copyEmbedFile(path, target)
	})
}
