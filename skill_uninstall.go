// 单 skill 卸载 API（v1.7.41）
//
// 允许卸载：
//   - source=gsd-framework  → rm -rf ~/.claude/skills/<id>/ + ~/.cursor/skills-cursor/<id>/
//   - source=ref-auto       → rm ~/.claude/skills/go-team-standards/references/<id>.md
//                             + ~/.cursor/skills-cursor/go-team-standards/references/<id>.md
//
// 禁止卸载：
//   - source=hardcoded（团队规范 12 条 + feature-flags）
//     —— 这些 embed 在 DMG 里，要卸只能不装 DMG / 装更新 DMG
//
// 所有路径强制在 $HOME 内（防穿越）。

package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

type skillUninstallReq struct {
	ID     string `json:"id"`     // 如 gsd-plan-phase / code-review
	Source string `json:"source"` // gsd-framework / ref-auto
}

// POST /api/installed-skill-uninstall
//
// 路径直删版（v1.7.43）：从「已装清单」tab 用，输入完整 skill 目录路径。
// body: {path: "~/.claude/skills/skill-name"}
//
// 安全：
//   - 必须在 ~/.claude/skills/ 或 ~/.cursor/skills-cursor/ 下
//   - 必须是目录（不是文件）
//   - 不允许 ..
//   - 团队 4 个核心 skill 拒绝（必须走专门卸载入口）
func handleInstalledSkillUninstall(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Path string `json:"path"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "bad json", err)
		return
	}
	req.Path = strings.TrimSpace(req.Path)
	if req.Path == "" {
		writeError(w, http.StatusBadRequest, "path 必填", nil)
		return
	}
	if strings.Contains(req.Path, "..") {
		writeError(w, http.StatusForbidden, "路径含 ..", nil)
		return
	}

	home, err := os.UserHomeDir()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "home", err)
		return
	}
	abs, err := filepath.Abs(req.Path)
	if err != nil {
		writeError(w, http.StatusBadRequest, "路径解析失败", err)
		return
	}

	// 允许的根：~/.claude/skills/ 和 ~/.cursor/rules/（v1.7.44 修：cursor 实际是 rules/）
	// 也保留 ~/.cursor/skills-cursor/（备用，某些版本用）
	allowedRoots := []string{
		filepath.Join(home, ".claude", "skills") + string(filepath.Separator),
		filepath.Join(home, ".cursor", "rules") + string(filepath.Separator),
		filepath.Join(home, ".cursor", "skills-cursor") + string(filepath.Separator),
	}
	inAllowed := false
	for _, root := range allowedRoots {
		if strings.HasPrefix(abs, root) {
			inAllowed = true
			break
		}
	}
	if !inAllowed {
		writeError(w, http.StatusForbidden, "路径不在允许的 skill 目录内: "+abs, nil)
		return
	}

	// 团队 4 个核心 skill / Cursor 团队 rules 拒绝
	// Claude skill 名（无后缀）：go-team-standards / orangecat / dev-dna / code-review
	// Cursor rule 文件名（含 .mdc）：00-iron-laws.mdc / 01-go-style.mdc 等（数字前缀）
	skillName := filepath.Base(abs)
	skillNameNoExt := strings.TrimSuffix(skillName, ".mdc")
	teamHardcoded := map[string]bool{
		"go-team-standards": true,
		"orangecat":         true,
		"dev-dna":           true,
		"code-review":       true,
	}
	if teamHardcoded[skillName] || teamHardcoded[skillNameNoExt] {
		writeError(w, http.StatusBadRequest, "团队核心 skill 请走「⚡ 安装」对应卡片卸载", nil)
		return
	}
	// Cursor 团队 rule 文件：数字前缀的 .mdc 不让删
	if strings.HasSuffix(skillName, ".mdc") && len(skillName) > 3 &&
		skillName[0] >= '0' && skillName[0] <= '9' &&
		skillName[1] >= '0' && skillName[1] <= '9' &&
		skillName[2] == '-' {
		writeError(w, http.StatusBadRequest, "团队 cursor rule 请走「⚡ 安装」卸载/覆盖", nil)
		return
	}

	// 检查存在 + 区分文件还是目录（cursor rules 是文件）
	info, err := os.Stat(abs)
	if err != nil {
		writeError(w, http.StatusNotFound, "路径不存在", err)
		return
	}

	if info.IsDir() {
		if err := os.RemoveAll(abs); err != nil {
			writeError(w, http.StatusInternalServerError, "删目录失败", err)
			return
		}
	} else {
		if err := os.Remove(abs); err != nil {
			writeError(w, http.StatusInternalServerError, "删文件失败", err)
			return
		}
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"ok":      true,
		"removed": abs,
		"is_dir":  info.IsDir(),
		"note":    "Cmd+Q 重启 Claude Code / Cursor 后 skill 列表更新",
	})
}

func handleSkillUninstall(w http.ResponseWriter, r *http.Request) {
	var req skillUninstallReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "bad json", err)
		return
	}
	req.ID = strings.TrimSpace(req.ID)
	req.Source = strings.TrimSpace(req.Source)
	if req.ID == "" || req.Source == "" {
		writeError(w, http.StatusBadRequest, "id 和 source 必填", nil)
		return
	}
	// 严防 .. 和 / 出现在 ID
	if strings.Contains(req.ID, "..") || strings.ContainsAny(req.ID, "/\\") {
		writeError(w, http.StatusForbidden, "非法 id（含 .. 或路径分隔符）", nil)
		return
	}

	home, err := os.UserHomeDir()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "home", err)
		return
	}

	var targets []string
	switch req.Source {
	case "gsd-framework":
		// gsd-* 整目录卸载
		if !strings.HasPrefix(req.ID, "gsd-") {
			writeError(w, http.StatusBadRequest, "gsd-framework source 的 id 必须以 gsd- 开头", nil)
			return
		}
		targets = []string{
			filepath.Join(home, ".claude", "skills", req.ID),
			filepath.Join(home, ".cursor", "skills-cursor", req.ID),
		}

	case "ref-auto":
		// 单个 .md 文件
		filename := req.ID
		if !strings.HasSuffix(filename, ".md") {
			filename += ".md"
		}
		targets = []string{
			filepath.Join(home, ".claude", "skills", "go-team-standards", "references", filename),
			filepath.Join(home, ".cursor", "skills-cursor", "go-team-standards", "references", filename),
		}

	case "hardcoded":
		writeError(w, http.StatusBadRequest, "团队规范是 embed 在 DMG 里的，无法卸载", nil)
		return

	default:
		writeError(w, http.StatusBadRequest, "未知 source: "+req.Source, nil)
		return
	}

	// 安全检查：所有 target 必须在 $HOME 内
	homeAbs, _ := filepath.Abs(home)
	var removed []string
	var skipped []string
	var failed []string
	for _, t := range targets {
		abs, err := filepath.Abs(t)
		if err != nil {
			failed = append(failed, t+": "+err.Error())
			continue
		}
		if !strings.HasPrefix(abs, homeAbs+string(filepath.Separator)) {
			failed = append(failed, abs+": 不在 $HOME 内")
			continue
		}
		if _, err := os.Stat(abs); err != nil {
			skipped = append(skipped, abs+"（不存在）")
			continue
		}
		if err := os.RemoveAll(abs); err != nil {
			failed = append(failed, abs+": "+err.Error())
			continue
		}
		removed = append(removed, abs)
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"ok":      len(failed) == 0,
		"removed": removed,
		"skipped": skipped,
		"failed":  failed,
		"hint":    fmt.Sprintf("Cmd+Q 重启 Claude Code / Cursor 后 skill 生效消失"),
	})
}
