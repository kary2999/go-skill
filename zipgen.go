// 一键打包：把 embedded 规范 + 当前用户自定义 skill 打成一个 zip 流给浏览器下载
package main

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// findAppBundle 从当前运行的二进制路径回溯找到 XXX.app 根目录
// 典型路径：/path/to/Team Standards.app/Contents/MacOS/team-standards-installer
func findAppBundle() (string, bool) {
	exe, err := os.Executable()
	if err != nil {
		return "", false
	}
	exe, _ = filepath.EvalSymlinks(exe)
	dir := filepath.Dir(exe)
	for i := 0; i < 6; i++ {
		if strings.HasSuffix(dir, ".app") {
			return dir, true
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return "", false
}

// handleSaveZip 把规范包写到本地 ~/Downloads/，返回绝对路径
func handleSaveZip(w http.ResponseWriter, r *http.Request) {
	version := "dev"
	if b, err := readEmbedFile("VERSION"); err == nil {
		version = strings.TrimSpace(string(b))
	}
	filename := fmt.Sprintf("TeamStandards-v%s-%s.zip", version, time.Now().Format("20060102-1504"))

	// 固定保存到项目 version/ 目录，作为历史归档
	// 若路径不存在（例如在别人机器上运行），回退 ~/Downloads
	dstDir := "~/skills-version"
	if err := os.MkdirAll(dstDir, 0755); err != nil {
		home, _ := os.UserHomeDir()
		dstDir = filepath.Join(home, "Downloads")
		if !dirExists(dstDir) {
			dstDir = home
		}
	}
	dstPath := filepath.Join(dstDir, filename)

	// 如果当前从 .app 内运行，优先用 ditto 打包 —— 保留签名、权限、符号链接、
	// 避免收件人那边出现"文件已损坏"
	if appPath, ok := findAppBundle(); ok {
		if err := dittoZipApp(appPath, dstPath); err != nil {
			writeError(w, http.StatusInternalServerError, "ditto zip", err)
			return
		}
		// 附带：使用说明 + 修复脚本（追加到 zip 里）
		customs, _ := loadCustomRules()
		if err := appendExtrasToZip(dstPath, version, len(customs)); err != nil {
			// 非致命，记录但不影响主流程
			_ = err
		}
	} else {
		f, err := os.Create(dstPath)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "create file", err)
			return
		}
		if err := writeZipBody(f); err != nil {
			f.Close()
			writeError(w, http.StatusInternalServerError, "write zip", err)
			return
		}
		f.Close()
	}

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

func dirExists(p string) bool {
	st, err := os.Stat(p)
	return err == nil && st.IsDir()
}

// handleSaveDMG 把当前运行的 .app 打成 DMG 保存到 version/
// DMG 是 macOS 原生磁盘镜像，挂载时 macOS 直接 mount，完整保留 ad-hoc 签名、xattr、权限。
// 相比 zip，收件人那端不存在"解压器丢元数据"的风险。
func handleSaveDMG(w http.ResponseWriter, r *http.Request) {
	version := "dev"
	if b, err := readEmbedFile("VERSION"); err == nil {
		version = strings.TrimSpace(string(b))
	}
	appPath, ok := findAppBundle()
	if !ok {
		writeError(w, http.StatusBadRequest,
			"当前不是从 .app 启动的，无法打 DMG（请先 ./build-mac-app.sh 后再从 Team Standards.app 运行本工具）", nil)
		return
	}

	// 准备 DMG 源目录
	srcDir, err := os.MkdirTemp("", "ts-dmg-src-*")
	if err != nil {
		writeError(w, http.StatusInternalServerError, "tmp dir", err)
		return
	}
	defer os.RemoveAll(srcDir)

	// 1. 复制 .app 到临时目录
	cpCmd := exec.Command("cp", "-R", appPath, srcDir+"/")
	if out, err := cpCmd.CombinedOutput(); err != nil {
		writeError(w, http.StatusInternalServerError, "cp app", fmt.Errorf("%v %s", err, string(out)))
		return
	}
	// 2. 做 /Applications 的符号链接 —— 拖拽即安装
	_ = os.Symlink("/Applications", filepath.Join(srcDir, "Applications"))

	// 3. 附带使用说明
	customs, _ := loadCustomRules()
	_ = os.WriteFile(filepath.Join(srcDir, "使用说明.md"),
		[]byte(appZipReadme(version, len(customs))), 0644)

	// 4. 输出路径
	dstDir := "~/skills-version"
	if err := os.MkdirAll(dstDir, 0755); err != nil {
		home, _ := os.UserHomeDir()
		dstDir = filepath.Join(home, "Downloads")
	}
	filename := fmt.Sprintf("TeamStandards-v%s-%s.dmg", version, time.Now().Format("20060102-1504"))
	dstPath := filepath.Join(dstDir, filename)

	// 5. hdiutil 打包（macOS 自带，不需要 CLT）
	cmd := exec.Command("hdiutil", "create",
		"-volname", "Team Standards",
		"-srcfolder", srcDir,
		"-ov", "-format", "UDZO", // UDZO = 压缩，Gatekeeper 友好
		dstPath,
	)
	if out, err := cmd.CombinedOutput(); err != nil {
		writeError(w, http.StatusInternalServerError, "hdiutil create",
			fmt.Errorf("%v\n%s", err, string(out)))
		return
	}

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

// handleReveal 在系统文件管理器中选中并显示该文件
func handleReveal(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Path string `json:"path"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid body", err)
		return
	}
	if body.Path == "" {
		writeError(w, http.StatusBadRequest, "path required", nil)
		return
	}
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", "-R", body.Path)
	case "windows":
		cmd = exec.Command("explorer", "/select,", body.Path)
	default:
		cmd = exec.Command("xdg-open", filepath.Dir(body.Path))
	}
	if err := cmd.Start(); err != nil {
		writeError(w, http.StatusInternalServerError, "reveal failed", err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

// handleDownloadZip 原来的浏览器流式下载（保留兼容）
func handleDownloadZip(w http.ResponseWriter, r *http.Request) {
	version := "dev"
	if b, err := readEmbedFile("VERSION"); err == nil {
		version = strings.TrimSpace(string(b))
	}
	filename := fmt.Sprintf("team-standards-v%s-%s.zip", version, time.Now().Format("20060102-1504"))
	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", `attachment; filename="`+filename+`"`)
	_ = writeZipBody(w)
}

// writeZipBody 把完整 zip 写到 w（文件或 HTTP 响应）
// Zip 内容：
//   - Team Standards.app/   整个 macOS 应用包（双击即安装）
//   - 使用说明.md            中文快速上手
func writeZipBody(out io.Writer) error {
	zw := zip.NewWriter(out)
	defer zw.Close()

	// 1. 如果当前是从 .app 运行，把 .app 整包塞进 zip
	if appPath, ok := findAppBundle(); ok {
		parent := filepath.Dir(appPath) // .app 的父目录
		walkErr := filepath.Walk(appPath, func(p string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			rel, err := filepath.Rel(parent, p)
			if err != nil {
				return err
			}
			if info.IsDir() {
				// 显式加目录条目保证 Finder 正确识别
				_, err := zw.Create(rel + "/")
				return err
			}
			// 保留符号链接（.app bundle 偶尔有）
			if info.Mode()&os.ModeSymlink != 0 {
				target, err := os.Readlink(p)
				if err != nil {
					return err
				}
				h := &zip.FileHeader{Name: rel, Modified: info.ModTime()}
				h.SetMode(info.Mode())
				fh, err := zw.CreateHeader(h)
				if err != nil {
					return err
				}
				_, err = fh.Write([]byte(target))
				return err
			}
			data, err := os.ReadFile(p)
			if err != nil {
				return err
			}
			h := &zip.FileHeader{Name: rel, Method: zip.Deflate, Modified: info.ModTime()}
			h.SetMode(info.Mode())
			fh, err := zw.CreateHeader(h)
			if err != nil {
				return err
			}
			_, err = fh.Write(data)
			return err
		})
		if walkErr != nil {
			return walkErr
		}

		// 2. 使用说明
		customs, _ := loadCustomRules()
		version := "dev"
		if b, err := readEmbedFile("VERSION"); err == nil {
			version = strings.TrimSpace(string(b))
		}
		return writeZipEntry(zw, "使用说明.md", []byte(appZipReadme(version, len(customs))), 0644)
	}

	// 回退方案：如果不是从 .app 运行（比如裸二进制），装 skill 文件 + install.sh

	version := "dev"
	if b, err := readEmbedFile("VERSION"); err == nil {
		version = strings.TrimSpace(string(b))
	}
	topDirs := []string{"standards", "demos", "claude", "cursor", "assets"}
	topFiles := []string{"VERSION", "CHANGELOG.md"}

	for _, d := range topDirs {
		if err := copyEmbedDirToZip(zw, d, d); err != nil {
			return err
		}
	}
	for _, f := range topFiles {
		if err := copyEmbedFileToZip(zw, f, f); err != nil {
			return err
		}
	}
	if err := writeZipEntry(zw, "install.sh", []byte(bundledInstallScript), 0755); err != nil {
		return err
	}
	customs, _ := loadCustomRules()
	for _, cr := range customs {
		_ = writeZipEntry(zw, "cursor/rules/custom-"+cr.ID+".mdc", []byte(cr.renderMDC()), 0644)
		_ = writeZipEntry(zw, "claude/go-team-standards/references/custom-"+cr.ID+".md", []byte(cr.renderMD()), 0644)
	}
	return writeZipEntry(zw, "README.md", []byte(zipReadme(version, len(customs))), 0644)
}

func writeZipEntry(zw *zip.Writer, name string, body []byte, mode fs.FileMode) error {
	h := &zip.FileHeader{Name: name, Method: zip.Deflate, Modified: time.Now()}
	h.SetMode(mode)
	f, err := zw.CreateHeader(h)
	if err != nil {
		return err
	}
	_, err = f.Write(body)
	return err
}

// copyEmbedDirToZip 递归把 embed 子树写入 zip（保持目录结构）
func copyEmbedDirToZip(zw *zip.Writer, srcDir, zipPrefix string) error {
	return fs.WalkDir(embeddedFS, srcDir, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		rel := strings.TrimPrefix(p, srcDir)
		rel = strings.TrimPrefix(rel, "/")
		zipPath := path.Join(zipPrefix, rel)
		b, err := embeddedFS.ReadFile(p)
		if err != nil {
			return err
		}
		return writeZipEntry(zw, zipPath, b, 0644)
	})
}

func copyEmbedFileToZip(zw *zip.Writer, src, dst string) error {
	b, err := embeddedFS.ReadFile(src)
	if err != nil {
		return err
	}
	return writeZipEntry(zw, dst, b, 0644)
}


// 简化版 install.sh，和 app 一起发给对方（对方没有 Go 环境也能装）
const bundledInstallScript = `#!/usr/bin/env bash
# Team Standards 一键安装 - 零参数，默认全局覆盖安装 Claude + Cursor
set -euo pipefail
REPO="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
VER="$(cat "$REPO/VERSION")"

CLAUDE="$HOME/.claude/skills/go-team-standards"
CURSOR="$HOME/.cursor/rules"

echo "→ 安装 Team Standards v$VER 到全局"

rm -rf "$CLAUDE"
mkdir -p "$CLAUDE/references" "$CLAUDE/assets" "$CLAUDE/demos"
cp "$REPO/claude/go-team-standards/SKILL.md" "$CLAUDE/"
cp -R "$REPO/standards/." "$CLAUDE/references/" 2>/dev/null || true
cp "$REPO/assets/.golangci.yml" "$CLAUDE/assets/" 2>/dev/null || true
cp -R "$REPO/demos/." "$CLAUDE/demos/"
# 合并 claude references 中可能存在的 custom-*
cp -f "$REPO/claude/go-team-standards/references"/custom-*.md "$CLAUDE/references/" 2>/dev/null || true
echo "$VER" > "$CLAUDE/.installed-version"
echo "  ✓ Claude Skill → $CLAUDE"

mkdir -p "$CURSOR"
find "$CURSOR" -maxdepth 1 -name '[0-9][0-9]-*.mdc' -delete 2>/dev/null || true
find "$CURSOR" -maxdepth 1 -name 'custom-*.mdc' -delete 2>/dev/null || true
cp "$REPO"/cursor/rules/*.mdc "$CURSOR/" 2>/dev/null || true
echo "$VER" > "$CURSOR/.installed-version"
echo "  ✓ Cursor rules → $CURSOR"

echo "✓ 完成，重启 Claude Code / Cursor 即生效"
`

// dittoZipApp 用 macOS 系统 `ditto` 命令把 .app 打包成 zip。
// ditto 是 Apple 官方推荐的方式，保留签名、权限、符号链接、资源分叉等，
// 避免 archive/zip 产出的 zip 在收件人 Mac 上触发"文件已损坏"。
func dittoZipApp(appPath, dstZip string) error {
	// -c 创建归档; -k 用 PKZip 格式; --sequesterRsrc 保存资源分叉;
	// --keepParent 保留顶层目录名（即 "Team Standards.app" 会是顶层）
	cmd := exec.Command("ditto",
		"-c", "-k",
		"--sequesterRsrc",
		"--keepParent",
		appPath,
		dstZip,
	)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("ditto failed: %v\n%s", err, out)
	}
	return nil
}

// appendExtrasToZip 把使用说明和修复脚本追加进已有 zip
// 因为 ditto 创建的 zip 用 Mac 独有格式，这里要再用 archive/zip 打开 append。
// ditto 产出的 zip 实际上是标准 PKZip + 一些 AppleDouble 元数据，Go 的 zip 库能读写。
func appendExtrasToZip(zipPath, version string, customCount int) error {
	// 读出现有内容
	existing, err := os.ReadFile(zipPath)
	if err != nil {
		return err
	}
	// 追加 mode：打开已有 zip，加新 entry，重写
	// 最简做法是解压到临时目录用 zip 重建 —— 但 ditto zip 里有 AppleDouble（._*）
	// 用 archive/zip 重建会丢失资源分叉。
	// 改为 shell 追加：用 zip 命令行工具（macOS 自带）
	_ = existing

	readmePath := filepath.Join(os.TempDir(), "使用说明.md")
	if err := os.WriteFile(readmePath, []byte(appZipReadme(version, customCount)), 0644); err != nil {
		return err
	}
	defer os.Remove(readmePath)

	fixPath := filepath.Join(os.TempDir(), "首次打开失败-双击修复.command")
	if err := os.WriteFile(fixPath, []byte(fixItScript), 0755); err != nil {
		return err
	}
	defer os.Remove(fixPath)

	// zip 命令的 -j 丢掉路径前缀，只保留文件名；-g 追加到已有 zip
	cmd := exec.Command("zip", "-g", "-j", zipPath, readmePath, fixPath)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("zip append failed: %v\n%s", err, out)
	}
	return nil
}

// fixItScript 是一个双击可运行的修复脚本
// 处理两种场景：
//   - "两按钮" 未签名拦截 → xattr -cr 就够
//   - "一按钮 已损坏" → 要先 --remove-signature 再重签名
// 用 osascript 一次性弹 admin 密码框，避免分多次 sudo
const fixItScript = `#!/bin/bash
# Team Standards · 完全解锁脚本
# 覆盖 macOS Sonoma (14+) 的两按钮 / 一按钮两种错误
set -e

# 查找 .app 所在位置
APP=""
for src in "$(dirname "$0")/Team Standards.app" \
           "$HOME/Downloads/Team Standards.app" \
           "$HOME/Desktop/Team Standards.app" \
           "/Applications/Team Standards.app"; do
    if [[ -d "$src" ]]; then APP="$src"; break; fi
done
if [[ -z "$APP" ]]; then
    osascript -e 'display dialog "找不到 Team Standards.app，请把它和本脚本放一起，或放到 /Applications" buttons {"好"} with icon stop'
    exit 1
fi

# 目标位置：/Applications（避开 macOS 的 App Translocation）
DST="/Applications/Team Standards.app"

# 用 admin 权限一次性搞定所有步骤
osascript <<APPLESCRIPT
do shell script "
  # 1. 搬到 /Applications
  if [ \"$APP\" != \"$DST\" ]; then
    rm -rf '$DST' 2>/dev/null
    cp -R '$APP' '$DST'
  fi

  # 2. 清掉所有扩展属性（quarantine / provenance / 下载记录）
  xattr -cr '$DST'

  # 3. 移除旧 ad-hoc 签名（zip 传输可能损坏它 → 一按钮错误）
  codesign --remove-signature '$DST' 2>/dev/null || true

  # 4. 重签名 ad-hoc（在接收端重建，bytes 对齐肯定过）
  codesign --force --deep --sign - '$DST'

  # 5. Gatekeeper 白名单
  spctl --add '$DST' 2>/dev/null || true

  # 6. LaunchServices 重扫，忘掉'这 app 坏了'的缓存
  /System/Library/Frameworks/CoreServices.framework/Frameworks/LaunchServices.framework/Support/lsregister -kill -r -domain local -domain system -domain user
" with administrator privileges with prompt "解锁 Team Standards.app 需要管理员权限"
APPLESCRIPT

osascript -e 'display dialog "✓ Team Standards.app 已完全解锁\n即将打开" buttons {"好"} default button 1'
open "$DST"
`

// appZipReadme 当 zip 里装了 .app 时用这个中文说明
func appZipReadme(version string, customCount int) string {
	var sb strings.Builder
	sb.WriteString("# Team Standards v" + version + " — 安装指南\n\n")
	sb.WriteString("Go 微服务团队编码规范 · Claude Code + Cursor 双平台 · 一键安装。\n\n")
	if customCount > 0 {
		sb.WriteString(fmt.Sprintf("> 本包含发送者额外添加的 **%d** 条自定义 Skill（SM），已内置在 App 中。\n\n", customCount))
	}
	sb.WriteString("## 安装步骤（30 秒）\n\n")
	sb.WriteString("1. 把 **Team Standards.app** 拖到 `应用程序` 文件夹（或任何你喜欢的位置）\n")
	sb.WriteString("2. 双击打开，会弹出一个 App 窗口\n")
	sb.WriteString("3. 点左侧 **⚡ 安装** → 选择「全局安装」+「两个都装」\n")
	sb.WriteString("4. 点 **⚡ 立即安装（覆盖已有）** 按钮\n")
	sb.WriteString("5. 重启 **Claude Code** 和 **Cursor** 即生效\n\n")
	sb.WriteString("完全不需要终端、不需要任何命令行操作。\n\n")
	sb.WriteString("## 首次打开被拦截？\n\n")
	sb.WriteString("macOS 对非 App Store 下载的应用有 **Gatekeeper 隔离机制**，可能提示：\n\n")
	sb.WriteString("- 「已损坏，无法打开」← ⚠️ 不是真损坏，是 Gatekeeper 阻止\n")
	sb.WriteString("- 「无法验证开发者」\n\n")
	sb.WriteString("### 解锁方法（任选一个）\n\n")
	sb.WriteString("**方法 A（最简单）**：双击 `首次打开失败-双击修复.command`\n")
	sb.WriteString("→ 自动移除隔离标记并打开 App\n\n")
	sb.WriteString("**方法 B**：在 `.app` 上 **右键 → 打开 → 打开**（一次性授权）\n\n")
	sb.WriteString("**方法 C**：打开终端，粘贴：\n")
	sb.WriteString("```bash\n")
	sb.WriteString("xattr -cr /Applications/Team\\ Standards.app\n")
	sb.WriteString("```\n\n")
	sb.WriteString("**方法 D**：系统设置 → 隐私与安全性 → 滚到底看到「仍要打开」\n\n")
	sb.WriteString("## 没装 Claude Code 或 Cursor 怎么办？\n\n")
	sb.WriteString("照样可以装。App 会自动创建目录，以后你装了 Claude / Cursor 就能直接加载团队规则。\n\n")
	sb.WriteString("## 这个 App 能做什么？\n\n")
	sb.WriteString("- 📚 **规范模块**：查看所有团队规则（点卡片看详情）\n")
	sb.WriteString("- ⚡ **安装**：一键装到全局或指定项目\n")
	sb.WriteString("- 🦕 **SM Skill 管理**：添加你自己的规则，保存后自动同步\n")
	sb.WriteString("- 📦 **打包分发**：把你的规则 + 自定义项一起打包发给别人\n\n")
	sb.WriteString("## 安装位置\n\n")
	sb.WriteString("- Claude：`~/.claude/skills/go-team-standards/`\n")
	sb.WriteString("- Cursor：`~/.cursor/rules/`\n\n")
	sb.WriteString("覆盖式安装，重复运行安全。\n")
	return sb.String()
}

func zipReadme(version string, customCount int) string {
	var sb strings.Builder
	sb.WriteString("# Team Standards v" + version + "\n\n")
	sb.WriteString("Go 微服务团队编码规范 + Claude Skill + Cursor Rules 一键安装包。\n\n")
	if customCount > 0 {
		sb.WriteString(fmt.Sprintf("> 本包含发送者添加的 %d 条自定义 Skill，已合并到 cursor/rules/ 和 claude/.../references/。\n\n", customCount))
	}
	sb.WriteString("## 快速安装（收到包后）\n\n")
	sb.WriteString("```bash\n")
	sb.WriteString("unzip team-standards-*.zip -d team-standards && cd team-standards\n")
	sb.WriteString("./install.sh\n")
	sb.WriteString("```\n\n")
	sb.WriteString("## 目录\n\n")
	sb.WriteString("- `standards/` — 规范原文\n")
	sb.WriteString("- `demos/` — 轻量代码模板\n")
	sb.WriteString("- `claude/go-team-standards/` — Claude Skill 结构\n")
	sb.WriteString("- `cursor/rules/` — Cursor .mdc 规则（含 custom-* 自定义）\n")
	sb.WriteString("- `assets/.golangci.yml` — lint 配置，复制到项目根\n")
	sb.WriteString("- `install.sh` — 零参数全局安装\n\n")
	sb.WriteString("## 说明\n\n")
	sb.WriteString("- 安装位置：`~/.claude/skills/go-team-standards/` + `~/.cursor/rules/`\n")
	sb.WriteString("- 幂等覆盖安装，重复运行不会出问题\n")
	sb.WriteString("- 重启 Claude Code / Cursor 后规则自动生效\n")
	return sb.String()
}
