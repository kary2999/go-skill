// 规范覆盖检查 —— 扫描用户已安装的规范文件，和当前 App 内置清单比对。
//
// 三种状态：
//   ✅ ok          ：文件存在且内容 hash 与内置一致（最新版）
//   🟡 outdated    ：文件存在但 hash 不同（老版本 / 被改过 / 用户手动替换）
//   ❌ missing     ：文件应该存在但不在（需要重装）
//   🔷 custom      ：user 的 custom-*.md 文件（白名单，始终视为 ok）
//
// 扫描范围：
//   - ~/.claude/skills/go-team-standards/（SKILL.md + references/*.md）
//   - ~/.cursor/rules/（*.mdc）
//   - ~/.claude/skills/orangecat/（SKILL.md + references/*）
//   - ~/.claude/skills/go-unit-test/（SKILL.md + 选装模块）
//
// 返回结构化结果给前端渲染。"一键补齐" = 调 go-team-standards install。

package main

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

type coverageItem struct {
	RelPath  string `json:"rel_path"`  // 相对 skill 根（安装后的位置）
	Status   string `json:"status"`    // ok / outdated / missing / custom
	EmbedSrc string `json:"embed_src"` // 内置 embed 路径

	// v1.7.15: 规范模块版本元数据（从 frontmatter 解析）
	InstalledVersion  string `json:"installed_version,omitempty"`  // 用户机器上文件的 version
	InstalledModified string `json:"installed_modified,omitempty"` // 用户机器上 last_modified
	EmbedVersion      string `json:"embed_version,omitempty"`      // App 内置最新 version
	EmbedModified     string `json:"embed_modified,omitempty"`    // App 内置最新 last_modified
}

type coverageBundle struct {
	Label    string         `json:"label"`       // 用户看到的分组名
	BaseDir  string         `json:"base_dir"`    // 实际扫描目录
	Items    []coverageItem `json:"items"`
	Summary  map[string]int `json:"summary"`     // ok/outdated/missing/custom 计数
}

// 每个 bundle 的清单条目：安装后的相对路径 → embed 里真实能读到的路径
type specFile struct {
	InstallRel string // 装到 BaseDir 下的相对路径，如 "references/api-design.md"
	EmbedPath  string // embed 里真实可读的路径，如 "standards/api-design.md"
	//                  注意：不能直接用 "claude/go-team-standards/references/api-design.md"
	//                  因为那些是符号链接，install.go 实际是从 standards/ 读的
}

type bundleSpec struct {
	Label       string
	BaseDirFn   func(home string) string
	Files       []specFile
	CustomGlob  string // custom-*.md 白名单 glob
	DirRecurse  bool
}

func coverageBundles() []bundleSpec {
	// go-team-standards references 实际是 standards/*.md 的符号链接，
	// install.go 装的时候直接 copyEmbedFile("standards/xxx.md", ...) —— 所以 embed 源头在 standards/
	gtsRefs := []string{
		"api-design.md", "ci-pipeline.md", "commit.md", "cursor-usage.md",
		"database.md", "error-codes.md", "glossary.md", "go-style.md",
		"naming-logging.md", "testing.md",
		// v1.7.14 新增
		"code-review.md", "deployment-checklist.md",
		"feature-flags.md", "api-doc-example.md",
		// v1.7.17 新增
		"meeting-minutes.md", "tech-design-example.md", "tixuebj-template-simple.md",
		// v1.7.20 新增
		"field-naming.md",
	}
	gtsFiles := []specFile{
		{InstallRel: "SKILL.md", EmbedPath: "claude/go-team-standards/SKILL.md"},
		{InstallRel: "assets/.golangci.yml", EmbedPath: "assets/.golangci.yml"},
	}
	for _, n := range gtsRefs {
		gtsFiles = append(gtsFiles, specFile{
			InstallRel: "references/" + n,
			EmbedPath:  "standards/" + n,
		})
	}
	// demos 也是 copyEmbedDir("demos", ...) 从顶级 demos/
	for _, n := range []string{
		"README.md", "errno-xerror.go", "kafka-consumer.go", "kafka-producer.go",
		"kratos-service-min.go", "pg-gorm-repo.go", "pg-migration.sql",
		"redis-idempotency.go", "slog-trace.go", "table-driven-test.go",
		"wire-providerset.go",
	} {
		gtsFiles = append(gtsFiles, specFile{
			InstallRel: "demos/" + n,
			EmbedPath:  "demos/" + n,
		})
	}

	// Cursor rules：install.go 从 cursor/rules/ 直接展开，路径一致
	cursorFiles := []specFile{}
	for _, n := range []string{
		"00-iron-laws.mdc", "01-go-style.mdc", "02-naming-logging.mdc",
		"03-error-codes.mdc", "04-api-design.mdc", "05-database.mdc",
		"06-testing.mdc", "07-commit.mdc", "08-ci-pipeline.mdc",
		"10-common-lib.mdc", "11-demos.mdc",
		"15-doc-trigger.mdc", // v1.7.22 新增：globs 强行按 .md 触发
		"99-cursor-security.mdc",
	} {
		cursorFiles = append(cursorFiles, specFile{
			InstallRel: n,
			EmbedPath:  "cursor/rules/" + n,
		})
	}

	// orangecat：路径一致，不是符号链接
	orangeFiles := []specFile{
		{InstallRel: "SKILL.md", EmbedPath: "claude/orangecat/SKILL.md"},
		{InstallRel: "references/提测报告模板_QA版.md",
			EmbedPath: "claude/orangecat/references/提测报告模板_QA版.md"},
		{InstallRel: "references/提测报告模板_开发版.md",
			EmbedPath: "claude/orangecat/references/提测报告模板_开发版.md"},
	}

	// dev-dna：v1.7.16 新增，与 orangecat 同结构
	devDnaFiles := []specFile{
		{InstallRel: "SKILL.md", EmbedPath: "claude/dev-dna/SKILL.md"},
		{InstallRel: "references/profile.md",
			EmbedPath: "claude/dev-dna/references/profile.md"},
		{InstallRel: "references/anti-distillation-policy.md",
			EmbedPath: "claude/dev-dna/references/anti-distillation-policy.md"},
	}

	// code-review：v1.7.25 新增，自动评审
	codeReviewFiles := []specFile{
		{InstallRel: "SKILL.md", EmbedPath: "claude/code-review/SKILL.md"},
		{InstallRel: "references/fix-examples.md",
			EmbedPath: "claude/code-review/references/fix-examples.md"},
	}

	return []bundleSpec{
		{
			Label:      "Claude · go-team-standards",
			BaseDirFn:  func(home string) string { return filepath.Join(home, ".claude", "skills", "go-team-standards") },
			Files:      gtsFiles,
			CustomGlob: "custom-*.md",
			DirRecurse: true,
		},
		{
			Label:      "Cursor · rules",
			BaseDirFn:  func(home string) string { return filepath.Join(home, ".cursor", "rules") },
			Files:      cursorFiles,
			CustomGlob: "99-custom-*.mdc",
			DirRecurse: true,
		},
		{
			Label:      "Claude · orangecat",
			BaseDirFn:  func(home string) string { return filepath.Join(home, ".claude", "skills", "orangecat") },
			Files:      orangeFiles,
			CustomGlob: "custom-*.md",
			DirRecurse: true,
		},
		{
			Label:      "Claude · dev-dna（个人开发档案）",
			BaseDirFn:  func(home string) string { return filepath.Join(home, ".claude", "skills", "dev-dna") },
			Files:      devDnaFiles,
			CustomGlob: "custom-*.md",
			DirRecurse: true,
		},
		{
			Label:      "Claude · code-review（自动评审）",
			BaseDirFn:  func(home string) string { return filepath.Join(home, ".claude", "skills", "code-review") },
			Files:      codeReviewFiles,
			CustomGlob: "custom-*.md",
			DirRecurse: true,
		},
	}
}

func sha256Hex(b []byte) string {
	h := sha256.Sum256(b)
	return hex.EncodeToString(h[:])
}

// parseSpecFrontmatter 提取 YAML frontmatter 里的 version + last_modified（容错：没装就返回空）
// 不引第三方 yaml 库，自己解 —— frontmatter 我们自己写，格式固定，简单字符串切就行
func parseSpecFrontmatter(content []byte) (version, modified string) {
	s := string(content)
	if !strings.HasPrefix(s, "---\n") {
		return "", ""
	}
	end := strings.Index(s[4:], "\n---")
	if end < 0 {
		return "", ""
	}
	fm := s[4 : 4+end]
	for _, line := range strings.Split(fm, "\n") {
		// 简单 key: "value" 解析
		if strings.HasPrefix(line, "version:") {
			version = strings.Trim(strings.TrimSpace(strings.TrimPrefix(line, "version:")), `"'`)
		} else if strings.HasPrefix(line, "last_modified:") {
			modified = strings.Trim(strings.TrimSpace(strings.TrimPrefix(line, "last_modified:")), `"'`)
		}
	}
	return
}

func parseFrontmatterFromFile(path string) (version, modified string) {
	b, err := os.ReadFile(path)
	if err != nil {
		return "", ""
	}
	return parseSpecFrontmatter(b)
}

func parseFrontmatterFromEmbed(embedPath string) (version, modified string) {
	b, err := readEmbedFile(embedPath)
	if err != nil {
		return "", ""
	}
	return parseSpecFrontmatter(b)
}

func hashFile(path string) (string, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return sha256Hex(b), nil
}

func hashEmbed(embedPath string) (string, error) {
	b, err := readEmbedFile(embedPath)
	if err != nil {
		return "", err
	}
	return sha256Hex(b), nil
}

func runCoverage() ([]coverageBundle, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	var out []coverageBundle
	for _, spec := range coverageBundles() {
		b := coverageBundle{
			Label:   spec.Label,
			BaseDir: spec.BaseDirFn(home),
			Items:   []coverageItem{},
			Summary: map[string]int{},
		}

		seen := map[string]bool{}

		// 1. 清单里的文件必须在
		for _, sf := range spec.Files {
			seen[sf.InstallRel] = true
			installed := filepath.Join(b.BaseDir, sf.InstallRel)

			item := coverageItem{RelPath: sf.InstallRel, EmbedSrc: sf.EmbedPath}

			// 始终读 embed 的 frontmatter（"最新版"参考）
			if strings.HasSuffix(sf.EmbedPath, ".md") {
				item.EmbedVersion, item.EmbedModified = parseFrontmatterFromEmbed(sf.EmbedPath)
			}

			if _, err := os.Stat(installed); err != nil {
				item.Status = "missing"
			} else {
				if strings.HasSuffix(installed, ".md") {
					item.InstalledVersion, item.InstalledModified = parseFrontmatterFromFile(installed)
				}
				instHash, _ := hashFile(installed)
				embedHash, eh := hashEmbed(sf.EmbedPath)
				if eh != nil {
					item.Status = "outdated"
				} else if instHash == embedHash {
					item.Status = "ok"
				} else {
					item.Status = "outdated"
				}
			}
			b.Items = append(b.Items, item)
			b.Summary[item.Status]++
		}

		// 2. 递归扫 BaseDir 找额外文件（custom-*.md 白名单）
		if spec.DirRecurse && dirExists(b.BaseDir) {
			_ = filepath.Walk(b.BaseDir, func(p string, info os.FileInfo, err error) error {
				if err != nil || info.IsDir() {
					return nil
				}
				rel, _ := filepath.Rel(b.BaseDir, p)
				if seen[rel] {
					return nil
				}
				base := filepath.Base(rel)
				if strings.HasPrefix(base, ".") {
					return nil
				}
				// 默认当 custom
				status := "custom"
				b.Items = append(b.Items, coverageItem{RelPath: rel, Status: status})
				b.Summary[status]++
				return nil
			})
		}

		out = append(out, b)
	}
	return out, nil
}

func handleCoverageCheck(w http.ResponseWriter, r *http.Request) {
	bundles, err := runCoverage()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "coverage", err)
		return
	}

	// 顶层汇总
	total := map[string]int{}
	for _, b := range bundles {
		for k, v := range b.Summary {
			total[k] += v
		}
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"bundles": bundles,
		"total":   total,
	})
}
