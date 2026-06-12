// 提交规范化 · 给指定 git 项目装 pre-commit hook（v1.7.37）
//
// 路径：把 embedded scripts/* 写到 <project>/scripts/，再写 .git/hooks/pre-commit
//
// 安全检查：
//   - 项目路径必须在 $HOME 内（拒绝写入系统目录）
//   - 路径必须存在 .git 目录（确认是 git 仓库）
//   - 不覆盖用户原有 pre-commit hook（如有则备份）

package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

const commitGuardScriptName = "team-standards-check.sh"

// 检查路径合法性：在 $HOME 内 + 是 git 仓库
func validateCommitGuardPath(p string) (string, error) {
	if p == "" {
		return "", fmt.Errorf("项目路径必填")
	}
	abs, err := filepath.Abs(strings.TrimSpace(p))
	if err != nil {
		return "", fmt.Errorf("路径解析: %w", err)
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	homeAbs, _ := filepath.Abs(home)
	if !strings.HasPrefix(abs, homeAbs+string(filepath.Separator)) && abs != homeAbs {
		return "", fmt.Errorf("拒绝写入 $HOME 之外：%s", abs)
	}
	if _, err := os.Stat(filepath.Join(abs, ".git")); err != nil {
		return "", fmt.Errorf("不是 git 仓库（缺 .git 目录）：%s", abs)
	}
	return abs, nil
}

// GET /api/commit-guard/status?path=...
//
// 返回该项目当前 hook 安装状态
func handleCommitGuardStatus(w http.ResponseWriter, r *http.Request) {
	p := r.URL.Query().Get("path")
	abs, err := validateCommitGuardPath(p)
	if err != nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"path":       p,
			"valid":      false,
			"error":      err.Error(),
			"is_git_dir": false,
		})
		return
	}
	hookPath := filepath.Join(abs, ".git", "hooks", "pre-commit")
	scriptPath := filepath.Join(abs, "scripts", commitGuardScriptName)
	hookExists := fileExists(hookPath)
	scriptExists := fileExists(scriptPath)
	isTeamHook := false
	if hookExists {
		b, _ := os.ReadFile(hookPath)
		if strings.Contains(string(b), "team-standards-check.sh") ||
			strings.Contains(string(b), "Team Standards App") {
			isTeamHook = true
		}
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"path":          abs,
		"valid":         true,
		"is_git_dir":    true,
		"hook_exists":   hookExists,
		"script_exists": scriptExists,
		"is_team_hook":  isTeamHook,
		"hook_path":     hookPath,
		"script_path":   scriptPath,
	})
}

// POST /api/commit-guard/install
// body: {path: "..."}
//
// 流程：
//  1. 校验路径
//  2. embed scripts/team-standards-check.sh → 写到 <path>/scripts/team-standards-check.sh
//  3. 如果 .git/hooks/pre-commit 已存在且非 team 写的 → 备份到 pre-commit.bak-<timestamp>
//  4. 写新 pre-commit hook
//  5. chmod +x
func handleCommitGuardInstall(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Path string `json:"path"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "bad json", err)
		return
	}
	abs, err := validateCommitGuardPath(req.Path)
	if err != nil {
		writeError(w, http.StatusBadRequest, "路径校验失败", err)
		return
	}

	checkScript, err := readEmbedFile("scripts/" + commitGuardScriptName)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "读 embed script", err)
		return
	}

	// 1. 写 scripts/team-standards-check.sh
	projScriptsDir := filepath.Join(abs, "scripts")
	if err := os.MkdirAll(projScriptsDir, 0755); err != nil {
		writeError(w, http.StatusInternalServerError, "mkdir scripts", err)
		return
	}
	projScript := filepath.Join(projScriptsDir, commitGuardScriptName)
	if err := os.WriteFile(projScript, checkScript, 0755); err != nil {
		writeError(w, http.StatusInternalServerError, "写 check.sh", err)
		return
	}

	// 2. 备份原 hook（如果存在且非 team 写的）
	hookDir := filepath.Join(abs, ".git", "hooks")
	if err := os.MkdirAll(hookDir, 0755); err != nil {
		writeError(w, http.StatusInternalServerError, "mkdir hooks", err)
		return
	}
	hookPath := filepath.Join(hookDir, "pre-commit")
	var backupPath string
	if old, err := os.ReadFile(hookPath); err == nil {
		if !strings.Contains(string(old), "team-standards-check.sh") {
			backupPath = hookPath + ".bak-" + time.Now().Format("20060102-150405")
			if err := os.WriteFile(backupPath, old, 0755); err != nil {
				writeError(w, http.StatusInternalServerError, "备份原 hook", err)
				return
			}
		}
	}

	// 3. 写新 hook
	hookContent := `#!/usr/bin/env bash
# 团队规范 pre-commit hook（v1.7.37）· 由 Team Standards App 装
# 跳过单次检查：git commit --no-verify

set -uo pipefail

SCRIPT="scripts/team-standards-check.sh"
if [ ! -x "${SCRIPT}" ]; then
    echo "⚠ ${SCRIPT} 不存在或不可执行，跳过团队规范检查"
    exit 0
fi

bash "${SCRIPT}"
`
	if err := os.WriteFile(hookPath, []byte(hookContent), 0755); err != nil {
		writeError(w, http.StatusInternalServerError, "写 hook", err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"ok":           true,
		"hook_path":    hookPath,
		"script_path":  projScript,
		"backup_path":  backupPath, // 空字符串 = 没备份
		"message":      "已装。下次 git commit 时自动跑规范检查。",
	})
}

// POST /api/commit-guard/uninstall
// body: {path: "..."}
//
// 删 pre-commit hook + scripts/team-standards-check.sh（如果是 team 装的）
func handleCommitGuardUninstall(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Path string `json:"path"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "bad json", err)
		return
	}
	abs, err := validateCommitGuardPath(req.Path)
	if err != nil {
		writeError(w, http.StatusBadRequest, "路径校验失败", err)
		return
	}

	var removed []string

	hookPath := filepath.Join(abs, ".git", "hooks", "pre-commit")
	if b, err := os.ReadFile(hookPath); err == nil {
		if strings.Contains(string(b), "team-standards-check.sh") {
			if err := os.Remove(hookPath); err == nil {
				removed = append(removed, hookPath)
			}
		}
	}

	scriptPath := filepath.Join(abs, "scripts", commitGuardScriptName)
	if _, err := os.Stat(scriptPath); err == nil {
		if err := os.Remove(scriptPath); err == nil {
			removed = append(removed, scriptPath)
		}
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"ok":      true,
		"removed": removed,
		"note":    "scripts/ 目录如果是空的，未删（防止误删用户其他脚本）",
	})
}

// POST /api/commit-guard/check
// body: {path: "..."}
//
// 在指定项目里跑一次检查（--all 模式），返回输出 + 退出码
func handleCommitGuardCheck(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Path string `json:"path"`
		All  bool   `json:"all"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "bad json", err)
		return
	}
	abs, err := validateCommitGuardPath(req.Path)
	if err != nil {
		writeError(w, http.StatusBadRequest, "路径校验失败", err)
		return
	}

	scriptPath := filepath.Join(abs, "scripts", commitGuardScriptName)
	if !fileExists(scriptPath) {
		writeError(w, http.StatusBadRequest, "脚本未装", nil)
		return
	}

	args := []string{scriptPath}
	if req.All {
		args = append(args, "--all")
	}
	cmd := exec.Command("bash", args...)
	cmd.Dir = abs
	// 强制无色输出（不在终端显示，前端 textarea 渲染）
	cmd.Env = append(os.Environ(), "TERM=dumb", "NO_COLOR=1")
	out, runErr := cmd.CombinedOutput()
	exitCode := 0
	if cmd.ProcessState != nil {
		exitCode = cmd.ProcessState.ExitCode()
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"ok":        runErr == nil || exitCode == 1, // 退出码 1（有违规）也算"跑成功"
		"output":    string(out),
		"exit_code": exitCode,
		"path":      abs,
	})
}

// GET /api/commit-guard/scripts
//
// 返回 embedded scripts 的内容（前端预览 / 给用户拷贝用）
func handleCommitGuardScripts(w http.ResponseWriter, r *http.Request) {
	files := []string{
		"scripts/team-standards-check.sh",
		"scripts/install-precommit.sh",
		"scripts/gitlab-ci-snippet.yml",
		"scripts/README.md",
	}
	result := map[string]string{}
	for _, f := range files {
		if b, err := readEmbedFile(f); err == nil {
			result[filepath.Base(f)] = string(b)
		}
	}
	writeJSON(w, http.StatusOK, result)
}

// 工具
func fileExists(p string) bool {
	info, err := os.Stat(p)
	return err == nil && !info.IsDir()
}
