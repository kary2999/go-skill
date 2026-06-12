// dev-dna Skill —— 个人开发档案，跨电脑跨账号无缝迁移
//
// 双写：
//   ~/.claude/skills/dev-dna/
//   ~/.cursor/skills-cursor/dev-dna/
//
// 用户 profile（textarea 编辑）保存到：
//   ~/Library/Application Support/TeamStandards/dev-dna-profile.md
// 安装时若存在 → 覆盖 references/profile.md；否则用内置默认模板。

package main

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

const (
	devDNASkillName    = "dev-dna"
	devDNAProfileRel   = "references/profile.md"
	devDNAProfileLocal = "dev-dna-profile.md"
)

func devDNAUserProfilePath() (string, error) {
	// 复用 orangecat 的目录约定（macOS Application Support / Win APPDATA / linux .config）
	dir, err := teamStandardsAppDataDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, devDNAProfileLocal), nil
}

// teamStandardsAppDataDir 抽出公共的应用配置目录函数（也被 orangecat / eval-config 用）
// 注意：如果未来重构，把 orangecatUserTemplatePath 也迁过来用这个
func teamStandardsAppDataDir() (string, error) {
	// 复用 evalConfigPath 同样的逻辑（macOS Library/AppSupport/TeamStandards）
	p, err := evalConfigPath()
	if err != nil {
		return "", err
	}
	return filepath.Dir(p), nil
}

func devDNAUserProfileExists() bool {
	p, err := devDNAUserProfilePath()
	if err != nil {
		return false
	}
	_, err = os.Stat(p)
	return err == nil
}

func handleDevDNAStatus(w http.ResponseWriter, r *http.Request) {
	home, _ := os.UserHomeDir()
	claudeDir := filepath.Join(home, ".claude", "skills", devDNASkillName)
	cursorDir := filepath.Join(home, ".cursor", "skills-cursor", devDNASkillName)

	version := ""
	if b, err := os.ReadFile(filepath.Join(claudeDir, ".installed-version")); err == nil {
		version = string(b)
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"claude_installed":     dirExists(claudeDir),
		"cursor_installed":     dirExists(cursorDir),
		"claude_dir":           claudeDir,
		"cursor_dir":           cursorDir,
		"version":              version,
		"custom_profile_used":  devDNAUserProfileExists(),
	})
}

func handleDevDNAInstall(w http.ResponseWriter, r *http.Request) {
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

	claudeDir := filepath.Join(home, ".claude", "skills", devDNASkillName)
	cursorDir := filepath.Join(home, ".cursor", "skills-cursor", devDNASkillName)

	for label, dir := range map[string]string{"Claude": claudeDir, "Cursor": cursorDir} {
		if err := installEmbeddedSkill("claude/"+devDNASkillName, dir); err != nil {
			warnings = append(warnings, label+" 安装失败："+err.Error())
			continue
		}
		// 若用户保存过 profile → 覆盖默认
		if devDNAUserProfileExists() {
			userProf, _ := devDNAUserProfilePath()
			if b, err := os.ReadFile(userProf); err == nil {
				dst := filepath.Join(dir, devDNAProfileRel)
				_ = os.MkdirAll(filepath.Dir(dst), 0755)
				if err := os.WriteFile(dst, b, 0644); err != nil {
					warnings = append(warnings, label+" 自定义 profile 写入失败："+err.Error())
				} else {
					installed = append(installed, label+" 应用自定义 profile → "+dst)
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

func handleDevDNAUninstall(w http.ResponseWriter, r *http.Request) {
	home, _ := os.UserHomeDir()
	dirs := []string{
		filepath.Join(home, ".claude", "skills", devDNASkillName),
		filepath.Join(home, ".cursor", "skills-cursor", devDNASkillName),
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

// GET /api/dev-dna/profile
func handleDevDNAProfileGet(w http.ResponseWriter, r *http.Request) {
	if devDNAUserProfileExists() {
		p, _ := devDNAUserProfilePath()
		b, err := os.ReadFile(p)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "read custom profile", err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"source":  "custom",
			"path":    p,
			"content": string(b),
		})
		return
	}
	b, err := readEmbedFile("claude/" + devDNASkillName + "/" + devDNAProfileRel)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "read embed profile", err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"source":  "builtin",
		"path":    "",
		"content": string(b),
	})
}

// POST /api/dev-dna/profile
func handleDevDNAProfileSave(w http.ResponseWriter, r *http.Request) {
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
	if len(req.Content) < 50 {
		writeError(w, http.StatusBadRequest, "profile too short (< 50 字符)", nil)
		return
	}
	p, err := devDNAUserProfilePath()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "path", err)
		return
	}
	if err := os.MkdirAll(filepath.Dir(p), 0700); err != nil {
		writeError(w, http.StatusInternalServerError, "mkdir", err)
		return
	}
	// 0600：只有本人可读，避免敏感偏好泄露
	if err := os.WriteFile(p, []byte(req.Content), 0600); err != nil {
		writeError(w, http.StatusInternalServerError, "write", err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "path": p})
}

// DELETE /api/dev-dna/profile —— 清除自定义，回到内置默认
func handleDevDNAProfileReset(w http.ResponseWriter, r *http.Request) {
	p, err := devDNAUserProfilePath()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "path", err)
		return
	}
	_ = os.Remove(p)
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}
