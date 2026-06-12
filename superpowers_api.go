// Superpowers skill 管理（v1.8.7）
//
// https://github.com/obra/superpowers
// 8 个覆盖完整开发工作流的 Claude Code skill。
//
// 安装方式：复制安装命令到 Claude Code / Cursor 对话框执行，
// 或直接调用此 API（后端通过 claude --dangerously-skip-permissions 执行命令）。
//
// 检测方式：扫 ~/.claude/skills/superpowers* 目录是否存在。

package main

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// superpowersTarget 解析安装目标 + 路径
func superpowersTarget(target, projectPath string) (skillsDir string, err error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	switch target {
	case "cursor":
		return filepath.Join(home, ".cursor", "skills"), nil
	case "project":
		if projectPath == "" {
			return "", errors.New("project path required for project scope")
		}
		return filepath.Join(projectPath, ".claude", "skills"), nil
	default: // "claude"
		return filepath.Join(home, ".claude", "skills"), nil
	}
}

// isSuperpowersInstalled 检查 skillsDir 下是否有 superpowers* 目录
func isSuperpowersInstalled(skillsDir string) (installed bool, count int, installPath string) {
	entries, err := os.ReadDir(skillsDir)
	if err != nil {
		return false, 0, ""
	}
	for _, e := range entries {
		if e.IsDir() && strings.HasPrefix(e.Name(), "superpowers") {
			count++
			if installPath == "" {
				installPath = filepath.Join(skillsDir, e.Name())
			}
		}
	}
	return count > 0, count, skillsDir
}

// handleSuperpowersStatus GET /api/superpowers/status?target=claude|cursor|project&path=...
func handleSuperpowersStatus(w http.ResponseWriter, r *http.Request) {
	target := r.URL.Query().Get("target")
	if target == "" {
		target = "claude"
	}
	projectPath := r.URL.Query().Get("path")

	skillsDir, err := superpowersTarget(target, projectPath)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": err.Error()})
		return
	}

	installed, count, installPath := isSuperpowersInstalled(skillsDir)
	writeJSON(w, http.StatusOK, map[string]any{
		"ok":           true,
		"installed":    installed,
		"skill_count":  count,
		"install_path": installPath,
		"target":       target,
	})
}

type superpowersReq struct {
	Target string `json:"target"`
	Path   string `json:"path"`
}

// handleSuperpowersInstall POST /api/superpowers/install
// 通过 git clone 安装 superpowers 到指定目标目录
func handleSuperpowersInstall(w http.ResponseWriter, r *http.Request) {
	var req superpowersReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": "bad json: " + err.Error()})
		return
	}
	if req.Target == "" {
		req.Target = "claude"
	}

	skillsDir, err := superpowersTarget(req.Target, req.Path)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": err.Error()})
		return
	}

	if err := os.MkdirAll(skillsDir, 0755); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"ok": false, "error": "mkdir: " + err.Error()})
		return
	}

	destDir := filepath.Join(skillsDir, "superpowers")
	// 若已存在则先删除再重装
	if _, err := os.Stat(destDir); err == nil {
		_ = os.RemoveAll(destDir)
	}

	cmd := exec.Command("git", "clone", "--depth=1", "https://github.com/obra/superpowers.git", destDir)
	out, runErr := cmd.CombinedOutput()
	detail := strings.TrimSpace(string(out))

	if runErr != nil {
		log.Printf("superpowers install failed: %v\n%s", runErr, detail)
		writeJSON(w, http.StatusOK, map[string]any{
			"ok":     false,
			"error":  runErr.Error(),
			"detail": detail + "\n\n💡 提示：请确保能访问 github.com，或在 Claude Code 里手动执行安装命令。",
		})
		return
	}

	log.Printf("superpowers installed to %s", destDir)
	writeJSON(w, http.StatusOK, map[string]any{
		"ok":     true,
		"detail": "✓ clone 到 " + destDir + "\n" + detail,
	})
}

// handleSuperpowersUninstall POST /api/superpowers/uninstall
func handleSuperpowersUninstall(w http.ResponseWriter, r *http.Request) {
	var req superpowersReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": "bad json: " + err.Error()})
		return
	}
	if req.Target == "" {
		req.Target = "claude"
	}

	skillsDir, err := superpowersTarget(req.Target, req.Path)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": err.Error()})
		return
	}

	entries, err := os.ReadDir(skillsDir)
	if err != nil {
		writeJSON(w, http.StatusOK, map[string]any{"ok": true, "removed": []string{}, "note": "目录不存在，无需卸载"})
		return
	}

	var removed []string
	for _, e := range entries {
		if e.IsDir() && strings.HasPrefix(e.Name(), "superpowers") {
			p := filepath.Join(skillsDir, e.Name())
			if err := os.RemoveAll(p); err == nil {
				removed = append(removed, p)
				log.Printf("superpowers uninstalled: %s", p)
			}
		}
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"ok":      true,
		"removed": removed,
	})
}
