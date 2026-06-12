package main

import (
	"bufio"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Skill 列表 —— description 对应 SKILL.md / .mdc 的 frontmatter，便于 UI 展示
type Skill struct {
	ID          string   `json:"id"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Triggers    []string `json:"triggers"`  // 触发场景
	Reference   string   `json:"reference"` // 对应 standards/xxx.md（hardcoded/ref-auto）或磁盘绝对路径（gsd-framework）
	Scope       string   `json:"scope"`     // global / conditional
	Source      string   `json:"source"`    // hardcoded / ref-auto / gsd-framework (v1.7.40)
	Group       string   `json:"group"`     // UI 分组标签（v1.7.40）
}

type Demo struct {
	ID          string   `json:"id"`
	File        string   `json:"file"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Keywords    []string `json:"keywords"`
}

type Catalog struct {
	Version   string   `json:"version"`
	Skills    []Skill  `json:"skills"`
	Demos     []Demo   `json:"demos"`
	Glossary  []string `json:"glossary"` // 一级分组标题列表
}

var skills = []Skill{
	{
		ID: "iron-laws", Title: "铁律", Description: "团队无条件执行的 12 条硬约束（密钥/金额/时间/错误/并发/日志/SQL 等）",
		Triggers: []string{"所有 Go 文件", "所有 SQL", "提交 commit 时"},
		Reference: "go-style.md", Scope: "always",
	},
	{
		ID: "go-style", Title: "Go 编码风格", Description: "项目结构（Kratos + Wire）、命名、并发、错误、注释、golangci-lint",
		Triggers: []string{"**/*.go"},
		Reference: "go-style.md", Scope: "conditional",
	},
	{
		ID: "naming-logging", Title: "命名 & 日志", Description: "仓库/包/文件命名 + Kafka topic / Redis key 强制规范 + slog 结构化字段 + W3C trace 上下文",
		Triggers: []string{"所有代码", "定义 topic / key", "打日志"},
		Reference: "naming-logging.md", Scope: "always",
	},
	{
		ID: "error-codes", Title: "错误码体系", Description: "受控 errno 常量 + xerror.New；IDP 注册 → codegen → CI linter 强制",
		Triggers: []string{"return error 时"},
		Reference: "error-codes.md", Scope: "conditional",
	},
	{
		ID: "database", Title: "PostgreSQL 设计", Description: "Schema / 表命名前缀 / 字段类型（DECIMAL/TIMESTAMPTZ/JSONB）/ 索引 / Migration / 审批",
		Triggers: []string{"migrations/**", "**/*.sql"},
		Reference: "database.md", Scope: "conditional",
	},
	{
		ID: "testing", Title: "单元测试", Description: "覆盖率门禁 / 表驱动 / mockgen / 命名 Test{被测}_{场景}_{期望}",
		Triggers: []string{"**/*_test.go"},
		Reference: "testing.md", Scope: "conditional",
	},
	{
		ID: "commit", Title: "Commit Message", Description: "Conventional Commits + scope 枚举 + commitlint CI 校验",
		Triggers: []string{"写 commit message 时"},
		Reference: "commit.md", Scope: "manual",
	},
	{
		ID: "ci-pipeline", Title: "CI Pipeline（参考）", Description: "六阶段 validate→build→test→scan→package→deploy + MR 门禁 + 性能目标",
		Triggers: []string{"需要了解 CI 原理时"},
		Reference: "ci-pipeline.md", Scope: "manual",
	},
	{
		ID: "common-lib", Title: "优先用 common-lib", Description: "禁止绕过 mask-go-common-lib 直连底层库（sarama/go-redis/otel/…），一张替换表",
		Triggers: []string{"**/*.go"},
		Reference: "glossary.md", Scope: "conditional",
	},
	{
		ID: "security", Title: "Cursor / AI 安全红线", Description: ".cursorignore 隔离、输入脱敏、Privacy Mode、禁用自动终端、资金双重 Review",
		Triggers: []string{"所有场景"},
		Reference: "cursor-usage.md", Scope: "always",
	},
	{
		ID: "feature-flags", Title: "特性开关 & 分支管理",
		Description: "FF SDK 集成 + IDP 注册 + GitOps 分支策略 + 5-to-4 预览环境 + Conventional Commits + 治理审计",
		Triggers: []string{"特性开关 / feature flag", "灰度 / 发布 / 回滚", "PR / MR / 分支策略"},
		Reference: "feature-flags.md", Scope: "conditional",
	},
}

// ============================================================
// 动态扫描 references/ 找未覆盖的 .md，自动生成 ☁️ 卡片
// ============================================================
//
// 设计目的（v1.7.36）：解决「云同步推新规范但 UI 卡片不增加」的问题。
//
// 工作流：
//  1. 启动时（或每次 /api/catalog 请求）扫 ~/.claude/skills/go-team-standards/references/
//  2. 找出**没在硬编码 skills slice 里**的 .md 文件
//  3. 跳过明显的 example/template 文件
//  4. 读 frontmatter title 自动生成卡片，加 ☁️ 前缀让用户识别是云端来源
//
// 副作用：每次 catalog 请求都做磁盘 I/O（references 文件少，开销忽略）

// 已被 hardcoded skills 覆盖的 reference 文件
func hardcodedReferenceFiles() map[string]bool {
	m := map[string]bool{}
	for _, s := range skills {
		m[s.Reference] = true
	}
	return m
}

// 扫 references/ 找未覆盖的 .md，生成 fallback Skill 条目
func scanExtraReferences() []Skill {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil
	}
	refDir := filepath.Join(home, ".claude", "skills", "go-team-standards", "references")
	entries, err := os.ReadDir(refDir)
	if err != nil {
		return nil // skill 没装 → 没 extras，返回空
	}
	covered := hardcodedReferenceFiles()
	var extras []Skill

	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if !strings.HasSuffix(name, ".md") {
			continue
		}
		if covered[name] {
			continue // 已被 hardcoded 覆盖
		}
		// 跳过用户自定义文件（custom-* 由用户自行管理，不作为云端规范卡片展示）
		lowerName := strings.ToLower(name)
		if strings.HasPrefix(lowerName, "custom-") {
			continue
		}

		fp := filepath.Join(refDir, name)
		title := extractFrontmatterTitle(fp)
		if title == "" {
			// fallback：文件名去 .md 后缀
			title = strings.TrimSuffix(name, ".md")
		}
		id := strings.TrimSuffix(name, ".md")

		extras = append(extras, Skill{
			ID:          id,
			Title:       "☁️ " + title,
			Description: "云同步规范（自动发现 from references/）—— 触发由 AI 按文件 frontmatter / 关键字判断。点卡片看完整原文。",
			Triggers:    []string{"按文件 frontmatter / 关键字匹配"},
			Reference:   name,
			Scope:       "conditional",
		})
	}
	return extras
}

// ============================================================
// 扫 ~/.claude/skills/gsd-* 自动加载 GSD 框架 skill 为卡片（v1.7.40）
// ============================================================

func scanGSDFrameworkSkills() []Skill {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil
	}
	skillsDir := filepath.Join(home, ".claude", "skills")
	entries, err := os.ReadDir(skillsDir)
	if err != nil {
		return nil
	}

	var out []Skill
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		name := e.Name()
		if !strings.HasPrefix(name, "gsd-") {
			continue
		}
		skillFile := filepath.Join(skillsDir, name, "SKILL.md")
		if _, err := os.Stat(skillFile); err != nil {
			continue
		}
		// 解析 frontmatter `name` 和 `description`（fallback）
		fmName, fmDesc := extractFrontmatterNameDesc(skillFile)
		if fmName == "" {
			fmName = name
		}

		// 中文翻译优先（v1.7.41 加）
		var title, desc string
		if zh, ok := gsdTranslations[name]; ok {
			title = "💪 " + zh.Title
			desc = zh.Desc
		} else {
			// 没翻译就 fallback：gsd-plan-phase → "💪 Plan Phase"
			shortName := strings.TrimPrefix(name, "gsd-")
			shortName = strings.ReplaceAll(shortName, "-", " ")
			words := strings.Fields(shortName)
			for i, w := range words {
				if len(w) > 0 {
					words[i] = strings.ToUpper(w[:1]) + w[1:]
				}
			}
			title = "💪 " + strings.Join(words, " ")
			desc = fmDesc
		}

		out = append(out, Skill{
			ID:          name,
			Title:       title,
			Description: desc,
			Triggers:    []string{"GSD 框架 · 上游 gsd-build/get-shit-done"},
			Reference:   skillFile, // 磁盘绝对路径
			Scope:       "conditional",
			Source:      "gsd-framework",
			Group:       "GSD 框架",
		})
	}

	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}

// scanKnownExtraSkills 扫描 Team Standards 随主安装分发的非 gsd-* skill（如 DevDefender）。
// 已安装则生成卡片，未安装则跳过。
func scanKnownExtraSkills() []Skill {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil
	}
	skillsDir := filepath.Join(home, ".claude", "skills")
	known := []struct {
		dir   string
		emoji string
		group string
	}{
		{"DevDefender", "🛡️", "需求防御"},
	}

	var out []Skill
	for _, k := range known {
		skillFile := filepath.Join(skillsDir, k.dir, "SKILL.md")
		if _, err := os.Stat(skillFile); err != nil {
			continue // 未安装则不展示
		}
		fmName, fmDesc := extractFrontmatterNameDesc(skillFile)
		if fmName == "" {
			fmName = k.dir
		}
		out = append(out, Skill{
			ID:          strings.ToLower(k.dir),
			Title:       k.emoji + " " + fmName,
			Description: fmDesc,
			Triggers:    []string{"输入 /devdefender 或「启动神盾局」激活"},
			Reference:   skillFile,
			Scope:       "conditional",
			Source:      "team-standards",
			Group:       k.group,
		})
	}
	return out
}

// 解析 frontmatter 拿 name 和 description（支持单行 / 多行 / 引号）
func extractFrontmatterNameDesc(fp string) (name, desc string) {
	f, err := os.Open(fp)
	if err != nil {
		return "", ""
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	inFM := false
	inDescMultiline := false
	var descLines []string

	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		if trimmed == "---" {
			if !inFM {
				inFM = true
				continue
			}
			break
		}
		if !inFM {
			continue
		}

		// 多行 description 收集
		if inDescMultiline {
			// 缩进的行 → 内容；顶格的 key 行 → 结束
			if strings.HasPrefix(line, " ") || strings.HasPrefix(line, "\t") {
				descLines = append(descLines, strings.TrimSpace(line))
				continue
			}
			// 否则 fall through 当新 key 处理
			inDescMultiline = false
		}

		// 解析 key: value
		if i := strings.Index(line, ":"); i > 0 {
			k := strings.TrimSpace(line[:i])
			v := strings.TrimSpace(line[i+1:])
			v = strings.Trim(v, `"'`)

			if k == "name" {
				name = v
			} else if k == "description" {
				if v == "" || v == "|" || v == ">" || v == "|-" || v == ">-" {
					inDescMultiline = true
				} else {
					desc = v
				}
			}
		}
	}

	if desc == "" && len(descLines) > 0 {
		desc = strings.Join(descLines, " ")
	}
	// 截断过长 description（卡片放不下）
	if len(desc) > 200 {
		desc = desc[:200] + "…"
	}
	return name, desc
}

// 从 .md frontmatter 提取 title 字段
func extractFrontmatterTitle(fp string) string {
	f, err := os.Open(fp)
	if err != nil {
		return ""
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	inFM := false
	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "---" {
			if !inFM {
				inFM = true
				continue
			}
			return "" // 走出 frontmatter 未找到 title
		}
		if !inFM {
			continue
		}
		if strings.HasPrefix(line, "title:") {
			t := strings.TrimSpace(strings.TrimPrefix(line, "title:"))
			t = strings.Trim(t, `"'`)
			return t
		}
	}
	return ""
}

var demos = []Demo{
	{ID: "kratos-service", File: "kratos-service-min.go", Title: "Kratos 服务骨架", Description: "最小 main.go：config + logging + tracing + alarm + wire", Keywords: []string{"kratos", "新服务", "main", "骨架", "启动"}},
	{ID: "wire-di", File: "wire-providerset.go", Title: "Wire 依赖注入", Description: "ProviderSet 声明 + wire.Build + 生成约定", Keywords: []string{"wire", "di", "依赖注入", "providerset"}},
	{ID: "kafka-producer", File: "kafka-producer.go", Title: "Kafka 生产者", Description: "mq.NewProducer + naming.TopicInProject 命名规范", Keywords: []string{"kafka", "生产者", "publish", "发消息", "topic"}},
	{ID: "kafka-consumer", File: "kafka-consumer.go", Title: "Kafka 消费者", Description: "mq.NewConsumer + 错误分流（协议错误不重试）+ 优雅退出", Keywords: []string{"kafka", "consumer", "消费", "订阅"}},
	{ID: "pg-migration", File: "pg-migration.sql", Title: "PG Migration 模板", Description: "goose up/down + 必备字段 + 索引 + COMMENT", Keywords: []string{"建表", "migration", "goose", "schema", "pg"}},
	{ID: "pg-gorm-repo", File: "pg-gorm-repo.go", Title: "GORM Repo 模式", Description: "context timeout + 软删 + 游标分页 + 显式列字段", Keywords: []string{"gorm", "repo", "crud", "查询", "游标分页"}},
	{ID: "redis-idempotency", File: "redis-idempotency.go", Title: "Redis 幂等 & 分布式锁", Description: "redisx.Store + SETNX 幂等 + Lua 解锁", Keywords: []string{"redis", "幂等", "锁", "idempotency", "lock"}},
	{ID: "errno-xerror", File: "errno-xerror.go", Title: "errno / xerror", Description: "错误码常量 + xerror.New + %w 包装", Keywords: []string{"errno", "xerror", "错误码", "return error"}},
	{ID: "slog-trace", File: "slog-trace.go", Title: "slog + OTel 日志", Description: "结构化日志 + ctx 提取 trace_id/span_id + error_code 标签", Keywords: []string{"slog", "日志", "trace_id", "logging"}},
	{ID: "table-test", File: "table-driven-test.go", Title: "表驱动单测", Description: "testify/assert + mockgen + 命名约定", Keywords: []string{"单测", "表驱动", "mockgen", "mock", "测试"}},
}

var glossarySections = []string{
	"Go 语言生态", "构建 / 依赖 / 代码生成", "团队自建组件（mask-go-common-lib）",
	"可观测性", "云原生 / Kubernetes / 服务网格", "CI / CD 与质量门禁",
	"安全 / 合规", "消息队列特有术语", "常见命令 / 工具",
}

func handleCatalog(w http.ResponseWriter, r *http.Request) {
	versionBytes, err := readEmbedFile("VERSION")
	version := "unknown"
	if err == nil {
		version = strings.TrimSpace(string(versionBytes))
	}

	// v1.7.43 调整：规范模块 = 开发规范，纯 16 条
	//   1. hardcoded 团队规范 12 条
	//   2. ☁️ ref-auto（references/ 里未覆盖的 .md，~4 张）
	// GSD 框架 66 张卡**不在此处**——它是独立 skill 体系，走独立 sidebar tab（见 /api/gsd-framework/list）
	allSkills := append([]Skill{}, skills...)
	for i := range allSkills {
		if allSkills[i].Source == "" {
			allSkills[i].Source = "hardcoded"
			allSkills[i].Group = "团队规范"
		}
	}
	extras := scanExtraReferences()
	for i := range extras {
		extras[i].Source = "ref-auto"
		extras[i].Group = "云同步规范"
	}
	allSkills = append(allSkills, extras...)

	// 扫 Team Standards 分发的已安装额外 skill（DevDefender 等）
	extraKnown := scanKnownExtraSkills()
	allSkills = append(allSkills, extraKnown...)

	writeJSON(w, http.StatusOK, Catalog{
		Version: version, Skills: allSkills, Demos: demos, Glossary: glossarySections,
	})
}

func handleChangelog(w http.ResponseWriter, r *http.Request) {
	b, err := readEmbedFile("CHANGELOG.md")
	if err != nil {
		writeError(w, http.StatusInternalServerError, "read changelog", err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"content": string(b)})
}

func handleDemo(w http.ResponseWriter, r *http.Request) {
	file := r.URL.Query().Get("file")
	if file == "" || strings.Contains(file, "..") || strings.Contains(file, "/") {
		writeError(w, http.StatusBadRequest, "invalid file", nil)
		return
	}
	b, err := readEmbedFile("demos/" + file)
	if err != nil {
		writeError(w, http.StatusNotFound, "demo not found", err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{
		"file": file, "content": string(b),
	})
}

// GET /api/gsd-framework/list
//
// 独立返回 GSD 框架的 66 张卡（不再混进 /api/catalog）。
// v1.7.43 调整：规范模块归规范模块，GSD 归 GSD，互不污染。
func handleGSDFrameworkList(w http.ResponseWriter, r *http.Request) {
	cards := scanGSDFrameworkSkills()
	writeJSON(w, http.StatusOK, map[string]any{
		"cards": cards,
		"count": len(cards),
		"hint":  "GSD 框架独立于团队规范——通过 npx 安装，~/.claude/skills/gsd-* 自动扫描",
	})
}

// GET /api/skill-disk-file?path=<abs path>
// 读 ~/.claude/skills/ 下的任意文件（gsd-framework skill 内容预览用，v1.7.40）
// 安全：路径必须在 ~/.claude/skills/ 内
func handleSkillDiskFile(w http.ResponseWriter, r *http.Request) {
	pathQ := r.URL.Query().Get("path")
	if pathQ == "" {
		writeError(w, http.StatusBadRequest, "path required", nil)
		return
	}
	home, err := os.UserHomeDir()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "home", err)
		return
	}
	abs, err := filepath.Abs(pathQ)
	if err != nil {
		writeError(w, http.StatusBadRequest, "bad path", err)
		return
	}
	allowedRoot := filepath.Join(home, ".claude", "skills") + string(filepath.Separator)
	if !strings.HasPrefix(abs, allowedRoot) {
		writeError(w, http.StatusForbidden, "path outside ~/.claude/skills/", nil)
		return
	}
	if strings.Contains(abs, "/..") {
		writeError(w, http.StatusForbidden, "path contains ..", nil)
		return
	}
	b, err := os.ReadFile(abs)
	if err != nil {
		writeError(w, http.StatusNotFound, "read file", err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"path":    abs,
		"content": string(b),
	})
}

func handleReference(w http.ResponseWriter, r *http.Request) {
	file := r.URL.Query().Get("file")
	if file == "" || strings.Contains(file, "..") || strings.Contains(file, "/") {
		writeError(w, http.StatusBadRequest, "invalid file", nil)
		return
	}
	b, err := readEmbedFile("standards/" + file)
	if err != nil {
		writeError(w, http.StatusNotFound, "reference not found", err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{
		"file": file, "content": string(b),
	})
}
