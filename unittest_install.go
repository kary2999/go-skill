// Go Unit Test Skill —— 模块化安装
//
// 基础层：SKILL.md（必装）
// 可选层：references/*.md（8 个）+ assets/skeleton-*.go（3 个），默认全不装
//
// 双写：
//   ~/.claude/skills/go-unit-test/
//   ~/.cursor/skills-cursor/go-unit-test/
//
// 选择保存在（macOS 约定）：
//   ~/Library/Application Support/TeamStandards/unittest-modules.json

package main

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

const unittestSkillName = "go-unit-test"

// 模块定义
type unittestModule struct {
	ID       string `json:"id"`       // 模块标识，等同文件 basename（无扩展名）
	Title    string `json:"title"`    // UI 标题
	Desc     string `json:"desc"`     // UI 描述
	Group    string `json:"group"`    // "references" | "assets"
	EmbedRel string `json:"-"`        // embed 相对路径（相对于 claude/go-unit-test/）
	DstRel   string `json:"-"`        // 安装目标相对路径（相对 skill 根）
}

var unittestManifest = []unittestModule{
	// references 组
	{ID: "test-structure", Title: "测试文件结构", Desc: "测试文件放哪 + 包名规则", Group: "references",
		EmbedRel: "references/test-structure.md", DstRel: "references/test-structure.md"},
	{ID: "table-driven", Title: "表驱动完整套路", Desc: "subtests / t.Parallel / 闭包陷阱", Group: "references",
		EmbedRel: "references/table-driven.md", DstRel: "references/table-driven.md"},
	{ID: "mock-gomock", Title: "mockgen + gomock 详解", Desc: "EXPECT/Return/Times/InOrder/自定义匹配器", Group: "references",
		EmbedRel: "references/mock-gomock.md", DstRel: "references/mock-gomock.md"},
	{ID: "assertions", Title: "testify assert vs require", Desc: "断言决策树 + 错误/金额/时间断言", Group: "references",
		EmbedRel: "references/assertions.md", DstRel: "references/assertions.md"},
	{ID: "anti-patterns", Title: "反模式清单", Desc: "time.Now / rand / 全局态 / sleep / 真 IO 等 12 种", Group: "references",
		EmbedRel: "references/anti-patterns.md", DstRel: "references/anti-patterns.md"},
	{ID: "coverage", Title: "覆盖率", Desc: "命令 + 目标 + exclude 规则", Group: "references",
		EmbedRel: "references/coverage.md", DstRel: "references/coverage.md"},
	{ID: "fixtures", Title: "fixtures / testdata / golden", Desc: "-update flag / go-cmp diff", Group: "references",
		EmbedRel: "references/fixtures.md", DstRel: "references/fixtures.md"},
	{ID: "integration", Title: "单元 vs 集成", Desc: "build tag + dockertest + sqlmock 选型", Group: "references",
		EmbedRel: "references/integration.md", DstRel: "references/integration.md"},

	// assets 组
	{ID: "skeleton-usecase", Title: "usecase 测试骨架", Desc: "表驱动 + gomock 完整骨架，复制即改", Group: "assets",
		EmbedRel: "assets/skeleton-usecase.go", DstRel: "assets/skeleton-usecase.go"},
	{ID: "skeleton-http-handler", Title: "HTTP handler 测试骨架", Desc: "httptest.NewRequest/Recorder + mock usecase", Group: "assets",
		EmbedRel: "assets/skeleton-http-handler.go", DstRel: "assets/skeleton-http-handler.go"},
	{ID: "skeleton-repo-sqlmock", Title: "repo 测试骨架（sqlmock）", Desc: "sqlmock.ExpectBegin/Query/Commit 套路", Group: "assets",
		EmbedRel: "assets/skeleton-repo-sqlmock.go", DstRel: "assets/skeleton-repo-sqlmock.go"},
}

// 选择文件路径
func unittestSelectionPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	var base string
	switch runtime.GOOS {
	case "darwin":
		base = filepath.Join(home, "Library", "Application Support", "TeamStandards")
	case "windows":
		base = filepath.Join(os.Getenv("APPDATA"), "TeamStandards")
	default:
		base = filepath.Join(home, ".config", "teamstandards")
	}
	return filepath.Join(base, "unittest-modules.json"), nil
}

func loadUnittestSelection() []string {
	p, err := unittestSelectionPath()
	if err != nil {
		return nil
	}
	b, err := os.ReadFile(p)
	if err != nil {
		return nil
	}
	var ids []string
	_ = json.Unmarshal(b, &ids)
	return ids
}

func saveUnittestSelection(ids []string) error {
	p, err := unittestSelectionPath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(p), 0755); err != nil {
		return err
	}
	b, _ := json.Marshal(ids)
	return os.WriteFile(p, b, 0644)
}

func handleUnittestManifest(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"modules":  unittestManifest,
		"selected": loadUnittestSelection(),
	})
}

func handleUnittestStatus(w http.ResponseWriter, r *http.Request) {
	home, _ := os.UserHomeDir()
	claudeDir := filepath.Join(home, ".claude", "skills", unittestSkillName)
	cursorDir := filepath.Join(home, ".cursor", "skills-cursor", unittestSkillName)

	version := ""
	if b, err := os.ReadFile(filepath.Join(claudeDir, ".installed-version")); err == nil {
		version = strings.TrimSpace(string(b))
	}

	// 探测已装模块
	installed := map[string]bool{}
	for _, m := range unittestManifest {
		p := filepath.Join(claudeDir, m.DstRel)
		if _, err := os.Stat(p); err == nil {
			installed[m.ID] = true
		}
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"claude_installed":   dirExists(claudeDir),
		"cursor_installed":   dirExists(cursorDir),
		"claude_dir":         claudeDir,
		"cursor_dir":         cursorDir,
		"version":            version,
		"installed_modules":  installed,
		"selected_persisted": loadUnittestSelection(),
	})
}

func handleUnittestInstall(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeError(w, http.StatusBadRequest, "read body", err)
		return
	}
	var req struct {
		Modules []string `json:"modules"` // 被勾选的模块 id
	}
	if len(body) > 0 {
		_ = json.Unmarshal(body, &req)
	}
	// 去重 + 合法性过滤
	validIDs := map[string]bool{}
	for _, m := range unittestManifest {
		validIDs[m.ID] = true
	}
	selected := map[string]bool{}
	var cleanIDs []string
	for _, id := range req.Modules {
		if validIDs[id] && !selected[id] {
			selected[id] = true
			cleanIDs = append(cleanIDs, id)
		}
	}
	_ = saveUnittestSelection(cleanIDs)

	home, err := os.UserHomeDir()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "home", err)
		return
	}
	version := "dev"
	if b, err := readEmbedFile("VERSION"); err == nil {
		version = strings.TrimSpace(string(b))
	}

	var installed []string
	var warnings []string

	for _, label := range []string{"Claude", "Cursor"} {
		var dir string
		if label == "Claude" {
			dir = filepath.Join(home, ".claude", "skills", unittestSkillName)
		} else {
			dir = filepath.Join(home, ".cursor", "skills-cursor", unittestSkillName)
		}

		// 清空重装（保证选项精确反映当前选择）
		if err := os.RemoveAll(dir); err != nil {
			warnings = append(warnings, label+" 清理失败："+err.Error())
			continue
		}
		if err := os.MkdirAll(dir, 0755); err != nil {
			warnings = append(warnings, label+" 目录创建失败："+err.Error())
			continue
		}

		// 必装：SKILL.md
		skillMD, err := readEmbedFile("claude/" + unittestSkillName + "/SKILL.md")
		if err != nil {
			warnings = append(warnings, label+" SKILL.md 读取失败："+err.Error())
			continue
		}
		if err := os.WriteFile(filepath.Join(dir, "SKILL.md"), skillMD, 0644); err != nil {
			warnings = append(warnings, label+" SKILL.md 写入失败："+err.Error())
			continue
		}

		// 可选模块：按 selected 装
		for _, m := range unittestManifest {
			if !selected[m.ID] {
				continue
			}
			src, err := readEmbedFile("claude/" + unittestSkillName + "/" + m.EmbedRel)
			if err != nil {
				warnings = append(warnings, label+" "+m.ID+" 读取失败："+err.Error())
				continue
			}
			dst := filepath.Join(dir, m.DstRel)
			_ = os.MkdirAll(filepath.Dir(dst), 0755)
			if err := os.WriteFile(dst, src, 0644); err != nil {
				warnings = append(warnings, label+" "+m.ID+" 写入失败："+err.Error())
				continue
			}
		}

		_ = os.WriteFile(filepath.Join(dir, ".installed-version"), []byte(version+"\n"), 0644)
		installed = append(installed, label+" → "+dir+" ("+intToStr(len(cleanIDs))+" 个可选模块)")
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"ok":                len(installed) > 0,
		"installed":         installed,
		"warnings":          warnings,
		"version":           version,
		"modules_installed": cleanIDs,
	})
}

func handleUnittestUninstall(w http.ResponseWriter, r *http.Request) {
	home, _ := os.UserHomeDir()
	dirs := []string{
		filepath.Join(home, ".claude", "skills", unittestSkillName),
		filepath.Join(home, ".cursor", "skills-cursor", unittestSkillName),
	}
	var removed []string
	for _, d := range dirs {
		if dirExists(d) {
			if err := os.RemoveAll(d); err == nil {
				removed = append(removed, d)
			}
		}
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "removed": removed})
}

func intToStr(n int) string {
	if n == 0 {
		return "0"
	}
	s := ""
	neg := n < 0
	if neg {
		n = -n
	}
	for n > 0 {
		s = string(rune('0'+n%10)) + s
		n /= 10
	}
	if neg {
		s = "-" + s
	}
	return s
}
