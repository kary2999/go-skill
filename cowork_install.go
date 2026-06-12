// Claude Desktop / Cowork 直装支持（v1.7.35）
//
// 背景：Claude Code / Cursor 走标准路径 ~/.claude/skills/ / ~/.cursor/skills-cursor/
// 但 Claude Desktop（claude.ai 桌面 App）和 Cowork 各自有不同的 skill 目录。
// 用户之前要手工 zip 导出 → 浏览器上传，繁琐。
//
// 本模块做两件事：
//   1. 探测系统中可能的 Claude Desktop / Cowork skill 目录（多候选 + 扫描发现）
//   2. 把内置的 5 个 skill 直接写入用户选中的目录（覆盖安装）
//
// ⚠️ 路径准确性：作者**没有**在已装 Cowork 的机器上验证过具体路径。
// 候选路径基于：macOS 标准 App Data 位置 + Anthropic 命名习惯推测。
// 用户应优先选 probe 结果里 `parent_exists=true` 的候选，或基于 discovered_dirs 自定义。

package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// 候选路径（按可能性降序）
type pathCandidate struct {
	Name     string // UI 显示
	Relative string // 相对 $HOME
	Note     string // 提示信息
	Source   string // claude-desktop / cowork / shared
}

var coworkPathCandidates = []pathCandidate{
	// === Claude Desktop（已知度较高） ===
	{
		Name:     "Claude Desktop · Skills（首字母大写）",
		Relative: "Library/Application Support/Claude/Skills",
		Note:     "Claude.ai 官方桌面 App 的标准位置",
		Source:   "claude-desktop",
	},
	{
		Name:     "Claude Desktop · skills（小写）",
		Relative: "Library/Application Support/Claude/skills",
		Note:     "另一个常见命名",
		Source:   "claude-desktop",
	},
	{
		Name:     "Claude Desktop · 沙箱容器路径",
		Relative: "Library/Containers/com.anthropic.claudefordesktop/Data/Library/Application Support/Claude/Skills",
		Note:     "App Sandbox 启用时的真实位置",
		Source:   "claude-desktop",
	},

	// === Cowork（推测） ===
	{
		Name:     "Cowork · 标准 App Support",
		Relative: "Library/Application Support/Cowork/skills",
		Note:     "如果 Cowork 是独立 App",
		Source:   "cowork",
	},
	{
		Name:     "Cowork · Anthropic 子目录",
		Relative: "Library/Application Support/Anthropic/Cowork/skills",
		Note:     "如果 Anthropic 把所有产品放统一子目录",
		Source:   "cowork",
	},
	{
		Name:     "Cowork · Claude 子模块",
		Relative: "Library/Application Support/Claude/cowork/skills",
		Note:     "如果 Cowork 是 Claude 客户端的内嵌功能",
		Source:   "cowork",
	},
	{
		Name:     "Cowork · 沙箱容器",
		Relative: "Library/Containers/com.anthropic.cowork/Data/Library/Application Support/Cowork/Skills",
		Note:     "App Sandbox 启用时",
		Source:   "cowork",
	},
	{
		Name:     "Cowork · CLI 风格",
		Relative: ".cowork/skills",
		Note:     "如果有 CLI 工具版",
		Source:   "cowork",
	},
}

// 探测结果（每条候选）
type pathProbeResult struct {
	Name         string `json:"name"`
	Path         string `json:"path"`
	Note         string `json:"note"`
	Source       string `json:"source"`
	Exists       bool   `json:"exists"`        // skills 目录本身存在
	ParentExists bool   `json:"parent_exists"` // 父目录存在（暗示对应 App 已装）
	Writable     bool   `json:"writable"`      // 用户可写
}

// 扫到的 Anthropic 相关目录（给用户辅助识别）
type discoveredDir struct {
	Path     string `json:"path"`
	HasSkill bool   `json:"has_skill_subdir"` // 是否含 skills/ 子目录
}

// GET /api/cowork/probe
//
// 返回所有候选路径的探测结果 + 扫到的 Anthropic 相关 App 目录
func handleCoworkProbe(w http.ResponseWriter, r *http.Request) {
	home, err := os.UserHomeDir()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "home", err)
		return
	}

	var results []pathProbeResult
	for _, c := range coworkPathCandidates {
		full := filepath.Join(home, c.Relative)
		parent := filepath.Dir(full)
		results = append(results, pathProbeResult{
			Name:         c.Name,
			Path:         full,
			Note:         c.Note,
			Source:       c.Source,
			Exists:       dirExists(full),
			ParentExists: dirExists(parent),
			Writable:     isDirWritable(parent),
		})
	}

	// 扫 ~/Library/Application Support/ + ~/Library/Containers/ 找所有 anthropic-related dir
	var discovered []discoveredDir
	for _, scanRoot := range []string{
		filepath.Join(home, "Library", "Application Support"),
		filepath.Join(home, "Library", "Containers"),
	} {
		entries, err := os.ReadDir(scanRoot)
		if err != nil {
			continue
		}
		for _, e := range entries {
			n := strings.ToLower(e.Name())
			if !strings.Contains(n, "claude") &&
				!strings.Contains(n, "cowork") &&
				!strings.Contains(n, "anthropic") {
				continue
			}
			dirPath := filepath.Join(scanRoot, e.Name())
			// 检查它下面有没 skills 子目录
			hasSkill := false
			if subEntries, err := os.ReadDir(dirPath); err == nil {
				for _, se := range subEntries {
					if !se.IsDir() {
						continue
					}
					sn := strings.ToLower(se.Name())
					if sn == "skills" {
						hasSkill = true
						break
					}
				}
			}
			discovered = append(discovered, discoveredDir{
				Path:     dirPath,
				HasSkill: hasSkill,
			})
		}
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"candidates":      results,
		"discovered_dirs": discovered,
		"home":            home,
	})
}

// POST /api/cowork/install
//
// body: {paths: ["...", "..."], skills: ["go-team-standards", ...] (可空=全部)}
//
// 把选中的 skill 覆盖装到 paths 列出的每个目录。
// 安全检查：所有 path 必须在 $HOME 内。
func handleCoworkInstall(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Paths  []string `json:"paths"`
		Skills []string `json:"skills"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "bad json", err)
		return
	}
	if len(req.Paths) == 0 {
		writeError(w, http.StatusBadRequest, "paths 必填", nil)
		return
	}
	if len(req.Skills) == 0 {
		// 默认装全部 5 个 skill
		req.Skills = []string{"go-team-standards", "orangecat", "dev-dna", "code-review", "gsd"}
	}

	home, err := os.UserHomeDir()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "home", err)
		return
	}
	homeAbs, _ := filepath.Abs(home)

	var installed []string
	var warnings []string

	for _, p := range req.Paths {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		absP, err := filepath.Abs(p)
		if err != nil {
			warnings = append(warnings, "路径解析失败: "+p)
			continue
		}
		// 安全：必须在 $HOME 内
		if !strings.HasPrefix(absP, homeAbs+string(filepath.Separator)) && absP != homeAbs {
			warnings = append(warnings, "拒绝写入 $HOME 之外: "+absP)
			continue
		}
		// 父目录必须存在或可创建
		if err := os.MkdirAll(absP, 0755); err != nil {
			warnings = append(warnings, "创建目录失败 "+absP+": "+err.Error())
			continue
		}

		for _, skillName := range req.Skills {
			dst := filepath.Join(absP, skillName)
			if err := installEmbeddedSkill("claude/"+skillName, dst); err != nil {
				warnings = append(warnings, fmt.Sprintf("%s → %s 失败: %v", skillName, dst, err))
				continue
			}
			installed = append(installed, fmt.Sprintf("%s → %s", skillName, dst))
		}
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"ok":        len(installed) > 0,
		"installed": installed,
		"warnings":  warnings,
		"note":      "若 Claude Desktop / Cowork 启动后看不到 skill，必须 Cmd+Q 彻底退出再开（skill 仅启动时加载一次）。",
	})
}

// 检查目录是否可写（尝试创建一个临时文件）
func isDirWritable(dir string) bool {
	if !dirExists(dir) {
		return false
	}
	tmp, err := os.CreateTemp(dir, ".tsi-writable-check-*")
	if err != nil {
		return false
	}
	_ = tmp.Close()
	_ = os.Remove(tmp.Name())
	return true
}
