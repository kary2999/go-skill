// gsd-help-zh 中文 help 安装（v1.7.42）
//
// 把 embed 的 assets/gsd-help-zh/ 装到用户环境：
//   ~/.claude/skills/gsd-help-zh/SKILL.md
//   ~/.claude/get-shit-done/workflows/help.zh.md
//
// 装完用户敲 /gsd-help-zh 就能看到中文版命令参考。
//
// 上游升级 npx 不会动 help.zh.md（上游只管 help.md），独立维护。

package main

import (
	"net/http"
	"os"
	"path/filepath"
)

const (
	gsdHelpZhSkillName = "gsd-help-zh"
	gsdHelpZhFile      = "help.zh.md"
)

func gsdHelpZhTargets() (skillDir, skillFile, helpFile string, err error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", "", "", err
	}
	skillDir = filepath.Join(home, ".claude", "skills", gsdHelpZhSkillName)
	skillFile = filepath.Join(skillDir, "SKILL.md")
	helpFile = filepath.Join(home, ".claude", "get-shit-done", "workflows", gsdHelpZhFile)
	return
}

// GET /api/gsd-help-zh/status
func handleGSDHelpZhStatus(w http.ResponseWriter, r *http.Request) {
	skillDir, skillFile, helpFile, err := gsdHelpZhTargets()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "home", err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"skill_installed": fileExists(skillFile),
		"help_installed":  fileExists(helpFile),
		"skill_dir":       skillDir,
		"skill_file":      skillFile,
		"help_file":       helpFile,
		"slash_command":   "/gsd-help-zh",
	})
}

// POST /api/gsd-help-zh/install
//
// 装两个文件：
//   1. ~/.claude/skills/gsd-help-zh/SKILL.md    (从 embed assets/gsd-help-zh/SKILL.md 拷)
//   2. ~/.claude/get-shit-done/workflows/help.zh.md  (从 embed assets/gsd-help-zh/help.zh.md 拷)
func handleGSDHelpZhInstall(w http.ResponseWriter, r *http.Request) {
	skillDir, skillFile, helpFile, err := gsdHelpZhTargets()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "home", err)
		return
	}

	// 读 embed
	skillBytes, err := readEmbedFile("assets/gsd-help-zh/SKILL.md")
	if err != nil {
		writeError(w, http.StatusInternalServerError, "读 embed SKILL.md", err)
		return
	}
	helpBytes, err := readEmbedFile("assets/gsd-help-zh/help.zh.md")
	if err != nil {
		writeError(w, http.StatusInternalServerError, "读 embed help.zh.md", err)
		return
	}

	// 写 SKILL.md
	if err := os.MkdirAll(skillDir, 0755); err != nil {
		writeError(w, http.StatusInternalServerError, "mkdir skill dir", err)
		return
	}
	if err := os.WriteFile(skillFile, skillBytes, 0644); err != nil {
		writeError(w, http.StatusInternalServerError, "写 SKILL.md", err)
		return
	}

	// 写 help.zh.md（确保父目录存在）
	helpDir := filepath.Dir(helpFile)
	if err := os.MkdirAll(helpDir, 0755); err != nil {
		writeError(w, http.StatusInternalServerError, "mkdir help dir", err)
		return
	}
	if err := os.WriteFile(helpFile, helpBytes, 0644); err != nil {
		writeError(w, http.StatusInternalServerError, "写 help.zh.md", err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"ok":            true,
		"skill_file":    skillFile,
		"help_file":     helpFile,
		"slash_command": "/gsd-help-zh",
		"hint":          "Cmd+Q 重启 Claude Code 后敲 /gsd-help-zh 看中文版",
	})
}

// POST /api/gsd-help-zh/uninstall
func handleGSDHelpZhUninstall(w http.ResponseWriter, r *http.Request) {
	skillDir, _, helpFile, err := gsdHelpZhTargets()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "home", err)
		return
	}

	var removed []string
	if err := os.RemoveAll(skillDir); err == nil {
		removed = append(removed, skillDir)
	}
	if err := os.Remove(helpFile); err == nil {
		removed = append(removed, helpFile)
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"ok":      true,
		"removed": removed,
		"note":    "已删本 skill 和 help.zh.md；不影响上游 gsd-help（英文版）",
	})
}
