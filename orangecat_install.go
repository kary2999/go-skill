// OrangeCat 提测文档 Skill —— 独立安装 / 卸载 / 状态 / 自定义模板（QA 版 + 开发版双模板）
//
// 双写：
//   ~/.claude/skills/orangecat/
//   ~/.cursor/skills-cursor/orangecat/
//
// 自定义模板（分 QA / 开发两份）：
//   ~/Library/Application Support/TeamStandards/orangecat-template-qa.md
//   ~/Library/Application Support/TeamStandards/orangecat-template-dev.md
// 若存在 → 安装时分别覆盖对应内置模板；否则用内置。
//
// 安装时会自动清理旧版本 tixuebj-template 目录（一次性迁移）。

package main

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
)

const (
	orangecatSkillName  = "orangecat"
	oldTixuebjSkillName = "tixuebj-template" // 旧名，装新版时清理
	orangecatTplRelQA   = "references/提测报告模板_QA版.md"
	orangecatTplRelDev  = "references/提测报告模板_开发版.md"
)

// templateKind: "qa" | "dev"
func orangecatUserTemplatePath(kind string) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	var base string
	switch runtime.GOOS {
	case "darwin":
		base = filepath.Join(home, "Library", "Application Support", "TeamStandards")
	case "windows":
		base = filepath.Join(os.Getenv("APPDATA"), "TeamStandards")
	default:
		base = filepath.Join(home, ".config", "teamstandards")
	}
	name := "orangecat-template-qa.md"
	if kind == "dev" {
		name = "orangecat-template-dev.md"
	}
	return filepath.Join(base, name), nil
}

func orangecatTplRel(kind string) string {
	if kind == "dev" {
		return orangecatTplRelDev
	}
	return orangecatTplRelQA
}

func orangecatUserTemplateExists(kind string) bool {
	p, err := orangecatUserTemplatePath(kind)
	if err != nil {
		return false
	}
	_, err = os.Stat(p)
	return err == nil
}

func handleOrangecatStatus(w http.ResponseWriter, r *http.Request) {
	home, _ := os.UserHomeDir()
	claudeDir := filepath.Join(home, ".claude", "skills", orangecatSkillName)
	cursorDir := filepath.Join(home, ".cursor", "skills-cursor", orangecatSkillName)

	version := ""
	if b, err := os.ReadFile(filepath.Join(claudeDir, ".installed-version")); err == nil {
		version = string(b)
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"claude_installed":        dirExists(claudeDir),
		"cursor_installed":        dirExists(cursorDir),
		"claude_dir":              claudeDir,
		"cursor_dir":              cursorDir,
		"version":                 version,
		"custom_template_qa":      orangecatUserTemplateExists("qa"),
		"custom_template_dev":     orangecatUserTemplateExists("dev"),
	})
}

func handleOrangecatInstall(w http.ResponseWriter, r *http.Request) {
	home, err := os.UserHomeDir()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "home dir", err)
		return
	}

	version := "dev"
	if b, err := readEmbedFile("VERSION"); err == nil {
		version = string(b)
	}

	var installed []string
	var warnings []string

	// 1. 迁移：删除旧 tixuebj-template
	oldClaude := filepath.Join(home, ".claude", "skills", oldTixuebjSkillName)
	oldCursor := filepath.Join(home, ".cursor", "skills-cursor", oldTixuebjSkillName)
	for _, p := range []string{oldClaude, oldCursor} {
		if dirExists(p) {
			if err := os.RemoveAll(p); err == nil {
				installed = append(installed, "已清理旧版 → "+p)
			}
		}
	}

	claudeDir := filepath.Join(home, ".claude", "skills", orangecatSkillName)
	cursorDir := filepath.Join(home, ".cursor", "skills-cursor", orangecatSkillName)

	for label, dir := range map[string]string{"Claude": claudeDir, "Cursor": cursorDir} {
		if err := installEmbeddedSkill("claude/"+orangecatSkillName, dir); err != nil {
			warnings = append(warnings, label+" 安装失败："+err.Error())
			continue
		}
		// 覆盖 QA / Dev 自定义模板（若存在）
		for _, kind := range []string{"qa", "dev"} {
			if orangecatUserTemplateExists(kind) {
				userTpl, _ := orangecatUserTemplatePath(kind)
				if b, err := os.ReadFile(userTpl); err == nil {
					dstTpl := filepath.Join(dir, orangecatTplRel(kind))
					_ = os.MkdirAll(filepath.Dir(dstTpl), 0755)
					if err := os.WriteFile(dstTpl, b, 0644); err != nil {
						warnings = append(warnings, label+" 自定义模板("+kind+")写入失败："+err.Error())
					} else {
						installed = append(installed, label+" 应用自定义 "+kind+" 版 → "+dstTpl)
					}
				}
			}
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

func handleOrangecatUninstall(w http.ResponseWriter, r *http.Request) {
	home, _ := os.UserHomeDir()
	dirs := []string{
		filepath.Join(home, ".claude", "skills", orangecatSkillName),
		filepath.Join(home, ".cursor", "skills-cursor", orangecatSkillName),
		filepath.Join(home, ".claude", "skills", oldTixuebjSkillName),
		filepath.Join(home, ".cursor", "skills-cursor", oldTixuebjSkillName),
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

// --- 模板编辑器 API（统一 ?which=qa|dev）---

func parseTplKind(r *http.Request) string {
	k := r.URL.Query().Get("which")
	if k == "dev" {
		return "dev"
	}
	return "qa" // 默认 QA 版
}

// GET /api/orangecat/template?which=qa|dev
func handleOrangecatTemplateGet(w http.ResponseWriter, r *http.Request) {
	kind := parseTplKind(r)
	if orangecatUserTemplateExists(kind) {
		p, _ := orangecatUserTemplatePath(kind)
		b, err := os.ReadFile(p)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "read custom template", err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"kind": kind, "source": "custom", "path": p, "content": string(b),
		})
		return
	}
	b, err := readEmbedFile("claude/" + orangecatSkillName + "/" + orangecatTplRel(kind))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "read embed template", err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"kind": kind, "source": "builtin", "path": "", "content": string(b),
	})
}

// POST /api/orangecat/template?which=qa|dev
func handleOrangecatTemplateSave(w http.ResponseWriter, r *http.Request) {
	kind := parseTplKind(r)
	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeError(w, http.StatusBadRequest, "read body", err)
		return
	}
	var req struct {
		Content string `json:"content"`
	}
	if err := json.Unmarshal(body, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json", err)
		return
	}
	if len(req.Content) < 10 {
		writeError(w, http.StatusBadRequest, "template too short", nil)
		return
	}
	p, err := orangecatUserTemplatePath(kind)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "path", err)
		return
	}
	if err := os.MkdirAll(filepath.Dir(p), 0755); err != nil {
		writeError(w, http.StatusInternalServerError, "mkdir", err)
		return
	}
	if err := os.WriteFile(p, []byte(req.Content), 0644); err != nil {
		writeError(w, http.StatusInternalServerError, "write", err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "kind": kind, "path": p})
}

// DELETE /api/orangecat/template?which=qa|dev
func handleOrangecatTemplateReset(w http.ResponseWriter, r *http.Request) {
	kind := parseTplKind(r)
	p, err := orangecatUserTemplatePath(kind)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "path", err)
		return
	}
	_ = os.Remove(p)
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "kind": kind})
}

// GET /api/orangecat/template/demo —— 保留（空响应兼容老前端）
func handleOrangecatTemplateDemo(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"content": ""})
}
