// 扫描本地已安装的 skill / cursor rules，按来源分类展示
package main

import (
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type InstalledItem struct {
	Name     string    `json:"name"`
	Path     string    `json:"path"`
	Size     int64     `json:"size"`
	Modified time.Time `json:"modified"`
	Source   string    `json:"source"` // "team" | "custom" | "other"
	Summary  string    `json:"summary,omitempty"`
}

type InstalledResp struct {
	ClaudeRoot   string          `json:"claude_root"`
	CursorRoot   string          `json:"cursor_root"`
	CodexRoot    string          `json:"codex_root"`
	ClaudeSkills []InstalledItem `json:"claude_skills"`
	CursorRules  []InstalledItem `json:"cursor_rules"`
	CodexSkills  []InstalledItem `json:"codex_skills"`
}

func handleInstalled(w http.ResponseWriter, r *http.Request) {
	home, err := os.UserHomeDir()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "home dir", err)
		return
	}
	claudeRoot := filepath.Join(home, ".claude", "skills")
	cursorRoot := filepath.Join(home, ".cursor", "rules")
	codexRoot  := filepath.Join(home, ".codex", "skills")

	resp := InstalledResp{ClaudeRoot: claudeRoot, CursorRoot: cursorRoot, CodexRoot: codexRoot}

	// Claude：列 ~/.claude/skills/ 下每个一级目录
	if entries, err := os.ReadDir(claudeRoot); err == nil {
		for _, e := range entries {
			if !e.IsDir() || strings.HasPrefix(e.Name(), ".") {
				continue
			}
			info, _ := e.Info()
			p := filepath.Join(claudeRoot, e.Name())
			item := InstalledItem{
				Name:     e.Name(),
				Path:     p,
				Size:     dirTotalSize(p),
				Modified: info.ModTime(),
				Source:   classifyClaudeSkill(e.Name()),
				Summary:  readSkillDescription(filepath.Join(p, "SKILL.md")),
			}
			resp.ClaudeSkills = append(resp.ClaudeSkills, item)
		}
	}
	sort.Slice(resp.ClaudeSkills, func(i, j int) bool {
		// team 排前面，其次 custom，其次字母序
		oi := sortKey(resp.ClaudeSkills[i].Source) + resp.ClaudeSkills[i].Name
		oj := sortKey(resp.ClaudeSkills[j].Source) + resp.ClaudeSkills[j].Name
		return oi < oj
	})

	// Cursor：列 ~/.cursor/rules/ 下的 *.mdc
	if entries, err := os.ReadDir(cursorRoot); err == nil {
		for _, e := range entries {
			if e.IsDir() || !strings.HasSuffix(e.Name(), ".mdc") {
				continue
			}
			info, _ := e.Info()
			p := filepath.Join(cursorRoot, e.Name())
			item := InstalledItem{
				Name:     e.Name(),
				Path:     p,
				Size:     info.Size(),
				Modified: info.ModTime(),
				Source:   classifyCursorRule(e.Name()),
				Summary:  readMdcDescription(p),
			}
			resp.CursorRules = append(resp.CursorRules, item)
		}
	}
	sort.Slice(resp.CursorRules, func(i, j int) bool {
		oi := sortKey(resp.CursorRules[i].Source) + resp.CursorRules[i].Name
		oj := sortKey(resp.CursorRules[j].Source) + resp.CursorRules[j].Name
		return oi < oj
	})

	// Codex：列 ~/.codex/skills/ 下每个一级目录（跳过 .system）
	if entries, err := os.ReadDir(codexRoot); err == nil {
		for _, e := range entries {
			if !e.IsDir() || strings.HasPrefix(e.Name(), ".") {
				continue
			}
			info, _ := e.Info()
			p := filepath.Join(codexRoot, e.Name())
			item := InstalledItem{
				Name:     e.Name(),
				Path:     p,
				Size:     dirTotalSize(p),
				Modified: info.ModTime(),
				Source:   classifyCodexSkill(e.Name()),
				Summary:  readSkillDescription(filepath.Join(p, "SKILL.md")),
			}
			resp.CodexSkills = append(resp.CodexSkills, item)
		}
	}
	sort.Slice(resp.CodexSkills, func(i, j int) bool {
		oi := sortKey(resp.CodexSkills[i].Source) + resp.CodexSkills[i].Name
		oj := sortKey(resp.CodexSkills[j].Source) + resp.CodexSkills[j].Name
		return oi < oj
	})

	writeJSON(w, http.StatusOK, resp)
}

func classifyCodexSkill(name string) string {
	if name == "go-team-standards" {
		return "team"
	}
	return "other"
}

// sortKey：team=0 / custom=1 / other=2
func sortKey(src string) string {
	switch src {
	case "team":
		return "0"
	case "custom":
		return "1"
	}
	return "2"
}

func classifyClaudeSkill(name string) string {
	if name == "go-team-standards" {
		return "team"
	}
	return "other"
}

func classifyCursorRule(name string) string {
	if strings.HasPrefix(name, "custom-") {
		return "custom"
	}
	// 团队规则：NN-xxx.mdc 两位数字前缀
	if isTeamRuleFile(name) {
		return "team"
	}
	return "other"
}

func dirTotalSize(root string) int64 {
	var total int64
	_ = filepath.Walk(root, func(p string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() {
			total += info.Size()
		}
		return nil
	})
	return total
}

// readSkillDescription 读 SKILL.md frontmatter 的 description 字段
func readSkillDescription(p string) string {
	b, err := os.ReadFile(p)
	if err != nil {
		return ""
	}
	return parseFrontmatterValue(string(b), "description")
}

// readMdcDescription 读 .mdc 文件的 description
func readMdcDescription(p string) string {
	b, err := os.ReadFile(p)
	if err != nil {
		return ""
	}
	return parseFrontmatterValue(string(b), "description")
}

func parseFrontmatterValue(content, key string) string {
	lines := strings.Split(content, "\n")
	if len(lines) < 2 || strings.TrimSpace(lines[0]) != "---" {
		return ""
	}
	for i := 1; i < len(lines); i++ {
		line := lines[i]
		if strings.TrimSpace(line) == "---" {
			break
		}
		if strings.HasPrefix(line, key+":") {
			val := strings.TrimSpace(line[len(key)+1:])
			// 截短到 200 字符
			if len(val) > 200 {
				val = val[:200] + "…"
			}
			return val
		}
	}
	return ""
}
