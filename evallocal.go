// 零 API key 测试方案：
//   - handleEvalLintSkill:   静态扫描本地 SKILL.md + references，检查结构完整性
//   - handleEvalExportPrompts: 导出 20 个测试用例为一份 markdown，用户自己粘到 Cursor/Claude Code
package main

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// ---------- 结构自检 ----------

type StructCheck struct {
	Name   string `json:"name"`
	Pass   bool   `json:"pass"`
	Detail string `json:"detail"`
}

func handleEvalLintSkill(w http.ResponseWriter, r *http.Request) {
	home, err := os.UserHomeDir()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "home dir", err)
		return
	}
	skillRoot := filepath.Join(home, ".claude", "skills", "go-team-standards")

	var checks []StructCheck

	// 1. SKILL.md 文件存在
	skillPath := filepath.Join(skillRoot, "SKILL.md")
	skillBytes, err := os.ReadFile(skillPath)
	if err != nil {
		checks = append(checks, StructCheck{
			Name: "SKILL.md 文件", Pass: false,
			Detail: "未找到：" + skillPath + "（先去 ⚡ 安装 Tab 装一次）",
		})
		writeJSON(w, http.StatusOK, map[string]any{"checks": checks})
		return
	}
	skill := string(skillBytes)
	checks = append(checks, StructCheck{
		Name: "SKILL.md 存在", Pass: true,
		Detail: fmt.Sprintf("%d 字节 · %s", len(skillBytes), skillPath),
	})

	// 2. frontmatter 字段
	checks = append(checks, hasSub(skill, "name: go-team-standards", "frontmatter name = go-team-standards"))
	checks = append(checks, hasSub(skill, "description:", "frontmatter 含 description"))

	// 3. 12 条铁律
	var missing []string
	for i := 1; i <= 12; i++ {
		markers := []string{fmt.Sprintf("铁律 %d", i), fmt.Sprintf("| %d ", i), fmt.Sprintf("%d · ", i)}
		hit := false
		for _, m := range markers {
			if strings.Contains(skill, m) {
				hit = true
				break
			}
		}
		if !hit {
			missing = append(missing, fmt.Sprintf("#%d", i))
		}
	}
	if len(missing) == 0 {
		checks = append(checks, StructCheck{Name: "12 条铁律齐全", Pass: true, Detail: "全部存在"})
	} else {
		checks = append(checks, StructCheck{
			Name: "12 条铁律齐全", Pass: false,
			Detail: "SKILL.md 里找不到: " + strings.Join(missing, ", "),
		})
	}

	// 4. 技术栈前提 / common-lib 替换表
	checks = append(checks, hasSub(skill, "common-lib", "mask-go-common-lib 替换表"))
	checks = append(checks, hasSub(skill, "xerror", "xerror/errno 约束"))
	checks = append(checks, hasSub(skill, "decimal", "金额 decimal.Decimal"))

	// 5. 路由表
	checks = append(checks, hasSub(skill, "references/", "路由表指向 references/"))
	checks = append(checks, hasSubAny(skill, []string{"demos/", "demo 路由"}, "Demo 路由表"))

	// 6. 违规三段式
	checks = append(checks, hasSubAny(skill,
		[]string{"三段式", "为什么", "怎么改"},
		"违规反馈三段式原则（哪条规则 / 为什么 / 怎么改）",
	))

	// 7. 强制加载 custom
	checks = append(checks, hasSubAny(skill,
		[]string{"custom-*.md", "自定义规则"},
		"自定义规则强制读取指令",
	))

	// 8. references 目录齐全
	refDir := filepath.Join(skillRoot, "references")
	entries, err := os.ReadDir(refDir)
	if err != nil {
		checks = append(checks, StructCheck{
			Name: "references 目录", Pass: false,
			Detail: "不存在：" + refDir,
		})
	} else {
		expected := []string{
			"go-style.md", "naming-logging.md", "error-codes.md",
			"api-design.md", "database.md", "testing.md",
			"commit.md", "glossary.md",
		}
		existing := map[string]bool{}
		for _, e := range entries {
			existing[e.Name()] = true
		}
		var miss []string
		for _, r := range expected {
			if !existing[r] {
				miss = append(miss, r)
			}
		}
		if len(miss) == 0 {
			checks = append(checks, StructCheck{
				Name: "references 文件齐全", Pass: true,
				Detail: fmt.Sprintf("%d 个文件（预期 ≥ %d）", len(entries), len(expected)),
			})
		} else {
			checks = append(checks, StructCheck{
				Name: "references 文件齐全", Pass: false,
				Detail: "缺: " + strings.Join(miss, ", "),
			})
		}

		// 9. 自定义规则计数
		customCount := 0
		for _, e := range entries {
			if strings.HasPrefix(e.Name(), "custom-") && strings.HasSuffix(e.Name(), ".md") {
				customCount++
			}
		}
		jsonCustoms, _ := loadCustomRules()
		jsonCount := len(jsonCustoms)
		pass := customCount == jsonCount
		detail := fmt.Sprintf("安装目录 %d 条 · JSON 源 %d 条", customCount, jsonCount)
		if !pass {
			detail += "（不一致 → 在 SM Tab 任意点一次保存触发同步）"
		}
		checks = append(checks, StructCheck{Name: "自定义规则同步状态", Pass: pass, Detail: detail})
	}

	// 10. demos 目录
	demoDir := filepath.Join(skillRoot, "demos")
	if demoEntries, err := os.ReadDir(demoDir); err == nil {
		checks = append(checks, StructCheck{
			Name: "demos/ 模板目录", Pass: true,
			Detail: fmt.Sprintf("%d 个 demo 文件", len(demoEntries)),
		})
	} else {
		checks = append(checks, StructCheck{Name: "demos/ 模板目录", Pass: false, Detail: "不存在"})
	}

	// 汇总
	passCount := 0
	for _, c := range checks {
		if c.Pass {
			passCount++
		}
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"checks": checks,
		"total":  len(checks),
		"passed": passCount,
	})
}

func hasSub(s, needle, name string) StructCheck {
	if strings.Contains(s, needle) {
		return StructCheck{Name: name, Pass: true, Detail: "命中 `" + needle + "`"}
	}
	return StructCheck{Name: name, Pass: false, Detail: "未找到 `" + needle + "`"}
}

func hasSubAny(s string, needles []string, name string) StructCheck {
	for _, n := range needles {
		if strings.Contains(s, n) {
			return StructCheck{Name: name, Pass: true, Detail: "命中 `" + n + "`"}
		}
	}
	return StructCheck{Name: name, Pass: false, Detail: "未命中任一: " + strings.Join(needles, ", ")}
}

// ---------- 导出测试 prompt ----------

func handleEvalExportPrompts(w http.ResponseWriter, r *http.Request) {
	var sb strings.Builder
	sb.WriteString("# Team Standards · Skill 测试用例集\n\n")
	sb.WriteString("> 把下面每一个用例粘进 Cursor / Claude Code 聊天框，看 AI 是否按团队规范的**三段式**反馈：\n")
	sb.WriteString("> 1. 引用了哪条规则（铁律 X / 规则名）\n")
	sb.WriteString("> 2. 解释了为什么（违反的后果）\n")
	sb.WriteString("> 3. 给出修正代码\n\n")
	sb.WriteString("这种测试**完全不需要 API key** —— 你日常工作用的那个 Cursor / Claude Code 就够了。\n\n")
	sb.WriteString(fmt.Sprintf("共 %d 个用例 · 生成于 %s\n\n", len(testCases), time.Now().Format("2006-01-02 15:04")))
	sb.WriteString("---\n\n")

	for i, tc := range testCases {
		sb.WriteString(fmt.Sprintf("## 用例 %d / %d · %s\n\n", i+1, len(testCases), tc.Rule))
		sb.WriteString("**粘下面这段到聊天框，追问一句**：\n")
		sb.WriteString("> 请按团队 Go 规范 review 以下代码，三段式回复（规则 / 为什么 / 怎么改）\n\n")
		sb.WriteString("```go\n")
		sb.WriteString(tc.ViolationCode)
		sb.WriteString("\n```\n\n")
		sb.WriteString(fmt.Sprintf("**期望 AI 命中**（至少 %d 个关键词）：", tc.MinMatch))
		sb.WriteString("`" + strings.Join(tc.Keywords, "` · `") + "`\n\n")
		sb.WriteString("**你主观评分**：\n")
		sb.WriteString("- [ ] ✓ AI 按三段式反馈，命中 ≥ " + fmt.Sprintf("%d", tc.MinMatch) + " 个关键词\n")
		sb.WriteString("- [ ] ✗ AI 没触发规则 / 反馈不够具体\n\n")
		sb.WriteString("---\n\n")
	}

	// 保存到 version/ 目录
	dstDir := "~/skills-version"
	if err := os.MkdirAll(dstDir, 0755); err != nil {
		home, _ := os.UserHomeDir()
		dstDir = filepath.Join(home, "Downloads")
	}
	version := "dev"
	if b, err := readEmbedFile("VERSION"); err == nil {
		version = strings.TrimSpace(string(b))
	}
	filename := fmt.Sprintf("skill-test-prompts-v%s-%s.md", version, time.Now().Format("20060102-1504"))
	dstPath := filepath.Join(dstDir, filename)

	if err := os.WriteFile(dstPath, []byte(sb.String()), 0644); err != nil {
		writeError(w, http.StatusInternalServerError, "write file", err)
		return
	}

	stat, _ := os.Stat(dstPath)
	size := int64(0)
	if stat != nil {
		size = stat.Size()
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"ok":    true,
		"path":  dstPath,
		"name":  filename,
		"size":  size,
		"cases": len(testCases),
	})
}
