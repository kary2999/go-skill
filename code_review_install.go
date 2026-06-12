// code-review Skill —— 自动评审，被动激活
//
// 与 /review slash command 互补：
//   - /review：用户主动 / 调用，扫 git diff
//   - code-review skill：AI 写完代码 / 用户粘代码 时自动激活，逐条按规范检查 + TODO 注释
//
// 双写：
//   ~/.claude/skills/code-review/
//   ~/.cursor/skills-cursor/code-review/

package main

import (
	"net/http"
	"os"
	"path/filepath"
)

const codeReviewSkillName = "code-review"

func handleCodeReviewStatus(w http.ResponseWriter, r *http.Request) {
	home, _ := os.UserHomeDir()
	claudeDir := filepath.Join(home, ".claude", "skills", codeReviewSkillName)
	cursorDir := filepath.Join(home, ".cursor", "skills-cursor", codeReviewSkillName)

	version := ""
	if b, err := os.ReadFile(filepath.Join(claudeDir, ".installed-version")); err == nil {
		version = string(b)
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"claude_installed": dirExists(claudeDir),
		"cursor_installed": dirExists(cursorDir),
		"claude_dir":       claudeDir,
		"cursor_dir":       cursorDir,
		"version":          version,
	})
}

func handleCodeReviewInstall(w http.ResponseWriter, r *http.Request) {
	home, err := os.UserHomeDir()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "home", err)
		return
	}
	version := "dev"
	if b, err := readEmbedFile("VERSION"); err == nil {
		version = string(b)
	}

	var installed []string
	var warnings []string

	for label, dir := range map[string]string{
		"Claude": filepath.Join(home, ".claude", "skills", codeReviewSkillName),
		"Cursor": filepath.Join(home, ".cursor", "skills-cursor", codeReviewSkillName),
	} {
		if err := installEmbeddedSkill("claude/"+codeReviewSkillName, dir); err != nil {
			warnings = append(warnings, label+" 安装失败："+err.Error())
			continue
		}
		_ = os.WriteFile(filepath.Join(dir, ".installed-version"), []byte(version+"\n"), 0644)
		installed = append(installed, label+" → "+dir)
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"ok":        len(installed) > 0,
		"installed": installed,
		"warnings":  warnings,
		"version":   version,
	})
}

func handleCodeReviewUninstall(w http.ResponseWriter, r *http.Request) {
	home, _ := os.UserHomeDir()
	dirs := []string{
		filepath.Join(home, ".claude", "skills", codeReviewSkillName),
		filepath.Join(home, ".cursor", "skills-cursor", codeReviewSkillName),
	}
	var removed []string
	for _, d := range dirs {
		if dirExists(d) {
			if err := os.RemoveAll(d); err == nil {
				removed = append(removed, d)
			}
		}
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "removed": removed})
}
