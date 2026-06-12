// 项目级 skill 扫描 + 一键同步（v1.7.21）
//
// 用户在多个 Cursor / Claude Code 项目下做了 skill / rules，
// 全局「🔄 一键更新规范到最新」只覆盖 ~/.claude/skills/ 和 ~/.cursor/rules/，
// 不动项目级路径。本模块提供：
//
//   GET /api/project-skills/scan?root=<目录>
//     扫描该目录及其直接子目录（深度 2）下的：
//       <project>/.claude/skills/
//       <project>/.cursor/skills-cursor/
//       <project>/.cursor/rules/
//     列出所有发现的"项目"及它们包含的 skill / rules 文件。
//
//   POST /api/project-skills/sync
//     body: {projects: [<project_path>...], targets: ["claude","cursor"]}
//     用 embed 内置版本覆盖每个项目里的对应路径。
//     复用现有 installClaude / installCursor 函数。

package main

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

type projectSkillInfo struct {
	ProjectPath        string   `json:"project_path"`
	HasClaudeSkills    bool     `json:"has_claude_skills"`
	HasCursorSkills    bool     `json:"has_cursor_skills"`
	HasCursorRules     bool     `json:"has_cursor_rules"`
	ClaudeSkillNames   []string `json:"claude_skill_names,omitempty"`
	CursorSkillNames   []string `json:"cursor_skill_names,omitempty"`
	CursorRulesCount   int      `json:"cursor_rules_count"`
	CursorRulesFiles   []string `json:"cursor_rules_files,omitempty"`
}

// handleProjectSkillsScan 扫描指定根目录下的项目级 skill
// 算法：从 root 起 walk，深度最多 5 层，遇到 .git 视为项目根
// 在每个项目根上检查 .claude/skills/ 和 .cursor/{rules,skills-cursor}/
func handleProjectSkillsScan(w http.ResponseWriter, r *http.Request) {
	root := r.URL.Query().Get("root")
	if root == "" {
		// 默认扫用户 Code 目录（你机器约定）
		home, _ := os.UserHomeDir()
		root = filepath.Join(home, "Desktop", "work", "Code")
	}
	if !dirExists(root) {
		writeError(w, http.StatusBadRequest, "root 目录不存在: "+root, nil)
		return
	}

	var projects []projectSkillInfo
	rootDepth := strings.Count(root, string(filepath.Separator))

	_ = filepath.Walk(root, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !info.IsDir() {
			return nil
		}
		// 限深 5 层（避免扫太深）
		depth := strings.Count(p, string(filepath.Separator)) - rootDepth
		if depth > 5 {
			return filepath.SkipDir
		}
		// 跳过常见无关目录
		base := filepath.Base(p)
		if base == "node_modules" || base == "vendor" || base == ".git" ||
			base == "dist" || base == "build" || base == ".idea" ||
			strings.HasPrefix(base, ".") && base != ".cursor" && base != ".claude" {
			return filepath.SkipDir
		}

		// 看这个目录是不是"项目根"——含 .git / go.mod / package.json / .cursor / .claude
		isProj := false
		for _, marker := range []string{".git", "go.mod", "package.json", "Cargo.toml", ".cursor", ".claude"} {
			if _, err := os.Stat(filepath.Join(p, marker)); err == nil {
				isProj = true
				break
			}
		}
		if !isProj {
			return nil
		}

		info0 := projectSkillInfo{ProjectPath: p}

		// .claude/skills/
		claudeDir := filepath.Join(p, ".claude", "skills")
		if dirExists(claudeDir) {
			info0.HasClaudeSkills = true
			if entries, err := os.ReadDir(claudeDir); err == nil {
				for _, e := range entries {
					if e.IsDir() {
						info0.ClaudeSkillNames = append(info0.ClaudeSkillNames, e.Name())
					}
				}
			}
		}
		// .cursor/skills-cursor/
		cursorSkillsDir := filepath.Join(p, ".cursor", "skills-cursor")
		if dirExists(cursorSkillsDir) {
			info0.HasCursorSkills = true
			if entries, err := os.ReadDir(cursorSkillsDir); err == nil {
				for _, e := range entries {
					if e.IsDir() {
						info0.CursorSkillNames = append(info0.CursorSkillNames, e.Name())
					}
				}
			}
		}
		// .cursor/rules/
		cursorRulesDir := filepath.Join(p, ".cursor", "rules")
		if dirExists(cursorRulesDir) {
			info0.HasCursorRules = true
			if entries, err := os.ReadDir(cursorRulesDir); err == nil {
				for _, e := range entries {
					if !e.IsDir() && strings.HasSuffix(e.Name(), ".mdc") {
						info0.CursorRulesCount++
						info0.CursorRulesFiles = append(info0.CursorRulesFiles, e.Name())
					}
				}
			}
		}

		// 只列含任一 skill 路径的项目
		if info0.HasClaudeSkills || info0.HasCursorSkills || info0.HasCursorRules {
			projects = append(projects, info0)
		}
		// 不再深入项目内部
		return filepath.SkipDir
	})

	writeJSON(w, http.StatusOK, map[string]any{
		"root":     root,
		"projects": projects,
		"count":    len(projects),
	})
}

// handleProjectSkillsSync 把 embed 版本的规范覆盖到选中的项目级路径
// 不动用户私人 custom-* 文件 / 用户自加的非清单 skill
func handleProjectSkillsSync(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeError(w, http.StatusBadRequest, "read body", err)
		return
	}
	var req struct {
		Projects []string `json:"projects"` // 项目根目录列表
	}
	if err := json.Unmarshal(body, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json", err)
		return
	}
	if len(req.Projects) == 0 {
		writeError(w, http.StatusBadRequest, "projects 列表不能为空", nil)
		return
	}

	type projectResult struct {
		ProjectPath string   `json:"project_path"`
		OK          bool     `json:"ok"`
		Installed   []string `json:"installed,omitempty"`
		Warnings    []string `json:"warnings,omitempty"`
	}

	var results []projectResult
	for _, proj := range req.Projects {
		// 防路径穿越：必须是绝对路径 + 真实存在
		if !filepath.IsAbs(proj) || !dirExists(proj) {
			results = append(results, projectResult{
				ProjectPath: proj,
				OK:          false,
				Warnings:    []string{"路径无效或不存在"},
			})
			continue
		}

		res := projectResult{ProjectPath: proj, OK: true}

		// 1. <proj>/.claude/skills/go-team-standards/
		claudeDir := filepath.Join(proj, ".claude", "skills", "go-team-standards")
		if err := installClaude(claudeDir); err != nil {
			res.Warnings = append(res.Warnings, "Claude Skill 失败: "+err.Error())
			res.OK = false
		} else {
			res.Installed = append(res.Installed, "→ "+claudeDir)
		}
		// 2. <proj>/.cursor/rules/
		cursorRulesDir := filepath.Join(proj, ".cursor", "rules")
		if err := installCursor(cursorRulesDir); err != nil {
			res.Warnings = append(res.Warnings, "Cursor rules 失败: "+err.Error())
			res.OK = false
		} else {
			res.Installed = append(res.Installed, "→ "+cursorRulesDir)
		}
		// 3. <proj>/.cursor/skills-cursor/go-team-standards/
		cursorSkillDir := filepath.Join(proj, ".cursor", "skills-cursor", "go-team-standards")
		if err := installClaude(cursorSkillDir); err != nil {
			res.Warnings = append(res.Warnings, "Cursor Skill 失败: "+err.Error())
		} else {
			res.Installed = append(res.Installed, "→ "+cursorSkillDir)
		}
		// 4. orangecat 也同步装到项目（如果项目已有 orangecat 目录则覆盖；没有就跳过）
		ocClaude := filepath.Join(proj, ".claude", "skills", "orangecat")
		if dirExists(ocClaude) {
			if err := installEmbeddedSkill("claude/orangecat", ocClaude); err == nil {
				res.Installed = append(res.Installed, "→ "+ocClaude)
			}
		}
		// 5. dev-dna 同上
		dnaClaude := filepath.Join(proj, ".claude", "skills", "dev-dna")
		if dirExists(dnaClaude) {
			if err := installEmbeddedSkill("claude/dev-dna", dnaClaude); err == nil {
				res.Installed = append(res.Installed, "→ "+dnaClaude)
			}
		}

		results = append(results, res)
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"ok":      true,
		"results": results,
		"count":   len(results),
	})
}
