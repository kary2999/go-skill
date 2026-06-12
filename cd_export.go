// Claude Desktop 导出 —— 给 claude.ai 网页 / Claude Desktop App 用
//
// Claude Desktop 的 Customize → Skills → + 按钮支持两种导入方式：
//   1. 上传 zip（完整格式：SKILL.md + references + demos + assets）
//   2. 在 Editor 里粘贴 SKILL.md 文本（轻量版，丢失 references 里的细节）
//
// 端点：
//   GET  /api/claude-desktop/export-zip   → 流式返回 skill 目录的 zip（旧版，浏览器 ~/Downloads）
//   POST /api/claude-desktop/save-zip     → 保存到固定位置 + 返回路径（v1.7.18 加，可在 Finder 直接定位）
//   GET  /api/claude-desktop/skill-md     → 返回 SKILL.md 纯文本（供前端复制剪贴板）
package main

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// buildSkillZip 把 ~/.claude/skills/<skillName>/ 整目录打 zip 写到任意 writer
// 给 export-zip（流到 HTTP）和 save-zip（落盘）共用
func buildSkillZip(skillName string, out io.Writer) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	skillDir := filepath.Join(home, ".claude", "skills", skillName)
	if _, err := os.Stat(skillDir); err != nil {
		return fmt.Errorf("未找到 skill，请先到 ⚡ 安装 Tab 装一次：%s", skillDir)
	}

	zw := zip.NewWriter(out)
	defer zw.Close()

	// v1.7.22 新增：zip 根写一个 INSTALL.md（手动安装说明），接收方解压即看到
	addEntry := func(name string, data []byte) error {
		h := &zip.FileHeader{Name: name, Method: zip.Deflate}
		h.SetMode(0644)
		h.NonUTF8 = false
		h.Flags |= 0x800
		fh, err := zw.CreateHeader(h)
		if err != nil {
			return err
		}
		_, err = fh.Write(data)
		return err
	}
	if err := addEntry("INSTALL.md", []byte(buildInstallReadme(skillName))); err != nil {
		return err
	}

	parent := filepath.Dir(skillDir) // 让 zip 里顶层是 "<skillName>/..."
	return filepath.Walk(skillDir, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() {
			return nil
		}
		// 只跳过我们自己的 metadata 和 OS 垃圾文件
		base := filepath.Base(p)
		switch base {
		case ".installed-version", ".DS_Store", "Thumbs.db":
			return nil
		}
		if strings.HasPrefix(base, "._") {
			return nil
		}
		rel, _ := filepath.Rel(parent, p)
		rel = filepath.ToSlash(rel)
		b, rerr := os.ReadFile(p)
		if rerr != nil {
			return nil
		}
		h := &zip.FileHeader{Name: rel, Method: zip.Deflate}
		h.SetMode(info.Mode())
		// UTF-8 flag：保证中文文件名（如 orangecat 模板）在解压工具里不变 ???
		h.NonUTF8 = false
		h.Flags |= 0x800
		fh, cerr := zw.CreateHeader(h)
		if cerr != nil {
			return nil
		}
		_, _ = fh.Write(b)
		return nil
	})
}

// buildInstallReadme v1.7.22: 给 zip 根加一份手动安装说明，接收方不依赖我们的 App 就能装
func buildInstallReadme(skillName string) string {
	return `# ` + skillName + ` Skill · 手动安装说明

> 接收方收到这份 zip 后按下面步骤就能装上 Skill。
> **不需要** Team Standards App，**不需要**管理员权限。

---

## 一、Claude Code（最常用）

` + "```bash" + `
# 1. 解压（如果还没解压）
unzip ` + skillName + `.zip

# 2. 把整个 skill 目录拷到 Claude Code 的 skill 库
mkdir -p ~/.claude/skills/
cp -R ` + skillName + ` ~/.claude/skills/

# 3. 验证
ls ~/.claude/skills/` + skillName + `/
# 应该看到 SKILL.md + references/ 等

# 4. 彻底重启 Claude Code（Cmd+Q 后再打开，不是 reload）
` + "```" + `

### 验证 Skill 装上了

在 Claude Code 里问相关话题（按 SKILL.md 里描述的触发词）→ 期望模型回复末尾出现：

| Skill | 触发标记 |
|---|---|
| go-team-standards | 🌟 |
| orangecat | 🐱 |
| dev-dna | 🧬 |

没看到标记 = Skill 没装上或 Claude Code 没重启。

---

## 二、Cursor

` + "```bash" + `
mkdir -p ~/.cursor/skills-cursor/
cp -R ` + skillName + ` ~/.cursor/skills-cursor/
` + "```" + `

然后 Cmd+Q 重启 Cursor。

---

## 三、Claude Desktop（claude.ai 网页 / 桌面 App）

1. Settings → Skills → **+** 新建
2. 选 **Upload zip** → 把 ` + skillName + `.zip 直接拖进去
3. Claude Desktop 自动识别 ` + skillName + `/SKILL.md 为入口

---

## 四、卸载

` + "```bash" + `
rm -rf ~/.claude/skills/` + skillName + `
rm -rf ~/.cursor/skills-cursor/` + skillName + `
# Cmd+Q 重启
` + "```" + `

---

## 五、想用图形化工具

跟提供这份 zip 的同事拿 **Team Standards App DMG**，提供安装/更新/扫描/覆盖检查的可视化操作。
`
}

func handleCDExportZip(w http.ResponseWriter, r *http.Request) {
	skillName := r.URL.Query().Get("name")
	if skillName == "" {
		skillName = "go-team-standards"
	}
	if !safeSkillName(skillName) {
		writeError(w, http.StatusBadRequest, "invalid skill name", nil)
		return
	}

	filename := skillName + ".zip"
	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", `attachment; filename="`+filename+`"`)
	if err := buildSkillZip(skillName, w); err != nil {
		// 注意：header 已发，HTTP 层只能写错误内容到 body
		fmt.Fprintf(w, "ERROR: %v", err)
	}
}

// handleCDSaveZip 把 skill zip 保存到固定位置（version/exported-skills/）+ 返回路径
// v1.7.18: 解决"用户不知道 zip 下载到哪"的问题
func handleCDSaveZip(w http.ResponseWriter, r *http.Request) {
	skillName := r.URL.Query().Get("name")
	if skillName == "" {
		skillName = "go-team-standards"
	}
	if !safeSkillName(skillName) {
		writeError(w, http.StatusBadRequest, "invalid skill name", nil)
		return
	}

	// 决定保存目录：优先项目 version/exported-skills/（开发机）
	// 否则 ~/Downloads/ExportedSkills/（接收方机器）
	dstDir := "~/skills-version/exported-skills"
	if err := os.MkdirAll(dstDir, 0755); err != nil {
		home, _ := os.UserHomeDir()
		dstDir = filepath.Join(home, "Downloads", "ExportedSkills")
		if err := os.MkdirAll(dstDir, 0755); err != nil {
			writeError(w, http.StatusInternalServerError, "create dir", err)
			return
		}
	}

	stamp := time.Now().Format("20060102-1504")
	filename := fmt.Sprintf("%s-%s.zip", skillName, stamp)
	dstPath := filepath.Join(dstDir, filename)

	f, err := os.Create(dstPath)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "create file", err)
		return
	}
	if err := buildSkillZip(skillName, f); err != nil {
		f.Close()
		os.Remove(dstPath)
		writeError(w, http.StatusInternalServerError, "build zip: "+err.Error(), nil)
		return
	}
	f.Close()

	stat, _ := os.Stat(dstPath)
	size := int64(0)
	if stat != nil {
		size = stat.Size()
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"ok":   true,
		"path": dstPath,
		"dir":  dstDir,
		"name": filename,
		"size": size,
	})
}

func handleCDSkillMD(w http.ResponseWriter, r *http.Request) {
	skillName := r.URL.Query().Get("name")
	if skillName == "" {
		skillName = "go-team-standards"
	}
	if !safeSkillName(skillName) {
		writeError(w, http.StatusBadRequest, "invalid skill name", nil)
		return
	}
	home, _ := os.UserHomeDir()
	skillMD := filepath.Join(home, ".claude", "skills", skillName, "SKILL.md")
	b, err := os.ReadFile(skillMD)
	if err != nil {
		writeError(w, http.StatusNotFound, "未找到 SKILL.md（先去 ⚡ 安装）", err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"name":    skillName,
		"content": string(b),
		"bytes":   len(b),
	})
}

// handleOpenApp 启动一个 macOS 应用（用于打开 Claude Desktop / Cursor）
// v1.7.18 新增。仅 macOS 支持 `open -a`。
func handleOpenApp(w http.ResponseWriter, r *http.Request) {
	var body struct {
		App string `json:"app"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid body", err)
		return
	}
	// 白名单：只允许我们认识的 App 名（防命令注入）
	allowed := map[string]bool{
		"Claude": true, "Cursor": true, "Visual Studio Code": true,
	}
	if !allowed[body.App] {
		writeError(w, http.StatusBadRequest, "app not in allowlist", nil)
		return
	}
	if runtime.GOOS != "darwin" {
		writeError(w, http.StatusNotImplemented, "目前仅支持 macOS", nil)
		return
	}
	cmd := exec.Command("open", "-a", body.App)
	if err := cmd.Start(); err != nil {
		writeError(w, http.StatusInternalServerError,
			"打不开 "+body.App+"（可能未安装）："+err.Error(), nil)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

func safeSkillName(s string) bool {
	if s == "" || len(s) > 64 {
		return false
	}
	for _, c := range s {
		ok := (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '-' || c == '_'
		if !ok {
			return false
		}
	}
	return true
}
