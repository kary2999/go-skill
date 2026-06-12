// 自定义 Skill：用户通过 UI 添加团队规范之外的个人 / 小组规则。
// 存储：单一 JSON 源 ~/.team-standards/custom-rules.json
// 分发：保存时立即同步到 ~/.cursor/rules/custom-<id>.mdc
//       和 ~/.claude/skills/go-team-standards/references/custom-<id>.md
package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

type CustomRule struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Reason      string `json:"reason"`
	Content     string `json:"content"`
	ApplyTo     string `json:"apply_to"` // "always" | "globs"
	Globs       string `json:"globs"`
	CreatedAt   string `json:"created_at"`
}

func customDataFile() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(home, ".team-standards")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", err
	}
	return filepath.Join(dir, "custom-rules.json"), nil
}

func loadCustomRules() ([]CustomRule, error) {
	path, err := customDataFile()
	if err != nil {
		return nil, err
	}
	b, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return []CustomRule{}, nil
	}
	if err != nil {
		return nil, err
	}
	var out []CustomRule
	if err := json.Unmarshal(b, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func saveCustomRules(rules []CustomRule) error {
	path, err := customDataFile()
	if err != nil {
		return err
	}
	b, err := json.MarshalIndent(rules, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, b, 0644)
}

var slugRe = regexp.MustCompile(`[^a-z0-9\-]+`)

func slugify(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = slugRe.ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")
	if s == "" {
		s = fmt.Sprintf("rule-%d", time.Now().Unix())
	}
	if len(s) > 40 {
		s = s[:40]
	}
	return s
}

func (r CustomRule) renderMDC() string {
	var sb strings.Builder
	sb.WriteString("---\n")
	sb.WriteString("description: " + singleLine(r.Description) + "\n")
	if r.ApplyTo == "always" {
		sb.WriteString("alwaysApply: true\n")
	} else {
		globs := strings.TrimSpace(r.Globs)
		if globs == "" {
			globs = "**/*.go"
		}
		sb.WriteString("globs: " + globs + "\n")
		sb.WriteString("alwaysApply: false\n")
	}
	sb.WriteString("---\n\n")
	sb.WriteString("# " + r.Title + "\n\n")
	if strings.TrimSpace(r.Reason) != "" {
		sb.WriteString("## 为什么存在这条规则\n\n")
		sb.WriteString(r.Reason + "\n\n")
	}
	sb.WriteString("## 规则\n\n")
	sb.WriteString(r.Content + "\n\n")
	sb.WriteString("## 反馈要求\n\n")
	sb.WriteString("当代码违反此规则时，必须：\n")
	sb.WriteString("1. 指出违反了本规则 (" + r.Title + ")\n")
	sb.WriteString("2. 说明后果（参见上方 Why 小节）\n")
	sb.WriteString("3. 给出修正代码\n")
	return sb.String()
}

// renderCursorSkill 生成 Cursor skills-cursor/<id>/SKILL.md
// 与 Claude SKILL.md 同构：frontmatter 含 name + description
func (r CustomRule) renderCursorSkill() string {
	var sb strings.Builder
	sb.WriteString("---\n")
	sb.WriteString("name: custom-" + r.ID + "\n")
	sb.WriteString("description: " + singleLine(r.Description) + "\n")
	sb.WriteString("---\n\n")
	sb.WriteString("# " + r.Title + "\n\n")
	if strings.TrimSpace(r.Reason) != "" {
		sb.WriteString("## 为什么\n\n" + r.Reason + "\n\n")
	}
	sb.WriteString("## 规则\n\n" + r.Content + "\n")
	return sb.String()
}

func (r CustomRule) renderMD() string {
	var sb strings.Builder
	sb.WriteString("# " + r.Title + " (自定义规则)\n\n")
	sb.WriteString("> " + r.Description + "\n\n")
	if strings.TrimSpace(r.Reason) != "" {
		sb.WriteString("## 为什么\n\n" + r.Reason + "\n\n")
	}
	sb.WriteString("## 规则内容\n\n" + r.Content + "\n\n")
	sb.WriteString("## 适用范围\n\n")
	if r.ApplyTo == "always" {
		sb.WriteString("所有场景（alwaysApply）\n")
	} else {
		sb.WriteString("文件匹配：`" + r.Globs + "`\n")
	}
	return sb.String()
}

func singleLine(s string) string {
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.ReplaceAll(s, "\r", " ")
	return strings.TrimSpace(s)
}

// syncCustomToInstalled 把所有自定义规则落盘到已安装的 Claude / Cursor 目录
// 写 3 处：
//   1. ~/.cursor/rules/custom-<id>.mdc           （老 rules 格式）
//   2. ~/.cursor/skills-cursor/custom-<id>/SKILL.md （Cursor 3.x 官方 skill 格式）
//   3. ~/.claude/skills/go-team-standards/references/custom-<id>.md
func syncCustomToInstalled(rules []CustomRule) ([]string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	cursorRulesDir := filepath.Join(home, ".cursor", "rules")
	cursorSkillsDir := filepath.Join(home, ".cursor", "skills-cursor")
	claudeRefs := filepath.Join(home, ".claude", "skills", "go-team-standards", "references")

	var logs []string

	// 1. 清理旧的 custom-*.mdc 和 custom-*.md 和 skills-cursor/custom-* 目录
	if entries, err := os.ReadDir(cursorRulesDir); err == nil {
		for _, e := range entries {
			if strings.HasPrefix(e.Name(), "custom-") && strings.HasSuffix(e.Name(), ".mdc") {
				_ = os.Remove(filepath.Join(cursorRulesDir, e.Name()))
			}
		}
	}
	if entries, err := os.ReadDir(cursorSkillsDir); err == nil {
		for _, e := range entries {
			if e.IsDir() && strings.HasPrefix(e.Name(), "custom-") {
				_ = os.RemoveAll(filepath.Join(cursorSkillsDir, e.Name()))
			}
		}
	}
	if entries, err := os.ReadDir(claudeRefs); err == nil {
		for _, e := range entries {
			if strings.HasPrefix(e.Name(), "custom-") && strings.HasSuffix(e.Name(), ".md") {
				_ = os.Remove(filepath.Join(claudeRefs, e.Name()))
			}
		}
	}

	// 2. 写入最新规则
	for _, r := range rules {
		filename := "custom-" + r.ID

		// Cursor rules 格式 (.mdc)
		if _, err := os.Stat(cursorRulesDir); err == nil {
			target := filepath.Join(cursorRulesDir, filename+".mdc")
			if err := os.WriteFile(target, []byte(r.renderMDC()), 0644); err == nil {
				logs = append(logs, "Cursor rules → "+target)
			}
		}

		// Cursor skills-cursor 格式 (SKILL.md 目录)
		skillDir := filepath.Join(cursorSkillsDir, filename)
		if err := os.MkdirAll(skillDir, 0755); err == nil {
			target := filepath.Join(skillDir, "SKILL.md")
			if err := os.WriteFile(target, []byte(r.renderCursorSkill()), 0644); err == nil {
				logs = append(logs, "Cursor skill → "+target)
			}
		}

		// Claude references
		if _, err := os.Stat(claudeRefs); err == nil {
			target := filepath.Join(claudeRefs, filename+".md")
			if err := os.WriteFile(target, []byte(r.renderMD()), 0644); err == nil {
				logs = append(logs, "Claude → "+target)
			}
		}
	}
	return logs, nil
}

// ---------- HTTP handlers ----------

func handleCustomList(w http.ResponseWriter, r *http.Request) {
	rules, err := loadCustomRules()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "load custom rules", err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"rules": rules})
}

func handleCustomUpsert(w http.ResponseWriter, r *http.Request) {
	var in CustomRule
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		writeError(w, http.StatusBadRequest, "invalid body", err)
		return
	}
	if strings.TrimSpace(in.Title) == "" {
		writeError(w, http.StatusBadRequest, "title required", nil)
		return
	}
	if strings.TrimSpace(in.Content) == "" {
		writeError(w, http.StatusBadRequest, "content required", nil)
		return
	}
	if in.ApplyTo != "always" && in.ApplyTo != "globs" {
		in.ApplyTo = "globs"
	}
	if in.ID == "" {
		in.ID = slugify(in.Title)
		in.CreatedAt = time.Now().UTC().Format(time.RFC3339)
	}

	rules, err := loadCustomRules()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "load", err)
		return
	}
	replaced := false
	for i, r := range rules {
		if r.ID == in.ID {
			rules[i] = in
			replaced = true
			break
		}
	}
	if !replaced {
		rules = append(rules, in)
	}
	if err := saveCustomRules(rules); err != nil {
		writeError(w, http.StatusInternalServerError, "save", err)
		return
	}
	logs, _ := syncCustomToInstalled(rules)
	writeJSON(w, http.StatusOK, map[string]any{
		"ok":     true,
		"rule":   in,
		"synced": logs,
	})
}

func handleCustomDelete(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "id required", nil)
		return
	}
	rules, err := loadCustomRules()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "load", err)
		return
	}
	kept := rules[:0]
	for _, r := range rules {
		if r.ID != id {
			kept = append(kept, r)
		}
	}
	if err := saveCustomRules(kept); err != nil {
		writeError(w, http.StatusInternalServerError, "save", err)
		return
	}
	logs, _ := syncCustomToInstalled(kept)
	writeJSON(w, http.StatusOK, map[string]any{
		"ok":     true,
		"synced": logs,
	})
}

func handleCustomPresets(w http.ResponseWriter, r *http.Request) {
	presets := []CustomRule{
		{
			Title:       "Go 代码必带解释性注释",
			Description: "All Go functions and non-trivial blocks must have Chinese comments explaining WHY, not WHAT",
			Reason:      "团队成员背景差异大（PHP / Java / Python 转 Go），代码行为本身的 What 可以通过命名和类型看懂，但设计取舍的 Why（为什么选这个算法 / 为什么这里需要重试 / 为什么不用标准做法）只能靠注释留存。没有 Why 的代码 3 个月后无人敢改。",
			Content:     "- 所有导出函数（大写开头）必须有 godoc 风格注释，以函数名开头\n- 非平凡的私有函数（含分支 / 重试 / 缓存 / 锁）必须有中文注释说明设计 Why\n- 复杂表达式 / 反直觉写法 / 性能 hack / 兼容性 workaround 必须加 // 注释\n- 禁止复述代码（不要写 `// 设置 name 为 foo` 这种废话）\n- 禁止提交被注释掉的代码（用 Git 历史追溯）",
			ApplyTo:     "globs",
			Globs:       "**/*.go",
		},
		{
			Title:       "所有对外写接口必须幂等",
			Description: "All external-facing write APIs must be idempotent with idempotency_key",
			Reason:      "网络层重试、客户端重放、MQ at-least-once 投递都会造成同一请求多次到达。没有幂等设计 = 用户多次扣款 / 多次发货。交易所场景下几乎等于事故。",
			Content:     "- 所有 POST / PUT / PATCH handler 必须接受 idempotency_key（Header 或 body 字段）\n- 数据库对应表必须有 idempotency_key VARCHAR(64) UNIQUE 字段\n- 幂等命中时返回首次的结果（2xx），不返回 409\n- 参考 demos/redis-idempotency.go 和 demos/pg-gorm-repo.go",
			ApplyTo:     "globs",
			Globs:       "internal/service/**,internal/server/**,**/handler/**",
		},
		{
			Title:       "禁止在代码中硬编码 sleep 时长",
			Description: "No magic sleep/timeout values in code; all durations must come from config",
			Reason:      "硬编码 time.Sleep(500 * time.Millisecond) 一旦线上环境延迟变化，调整就需要重新发版。所有等待时长走 config 可让运维秒级调整。",
			Content:     "- 禁止在业务代码中出现 time.Sleep / time.After 的魔法数字\n- 所有 timeout / retry interval / polling 间隔必须从 config 读取\n- 测试代码豁免\n- 允许的例外：指数退避的倍数常量（如 retry 的 base 间隔）可定义为包级 const 并写清 Why",
			ApplyTo:     "globs",
			Globs:       "**/*.go",
		},
	}
	writeJSON(w, http.StatusOK, map[string]any{"presets": presets})
}
