// Shell 自解压安装脚本 —— App 用不了时的终极兼容方案
//
// 产物：一个单独的 .sh 文件（约 700KB），内嵌所有 skill 文件的 base64 tar.gz
// 对方只需：chmod +x install-skills-*.sh && ./install-skills-*.sh
// 纯 bash，不依赖 Go / Xcode / App / 任何二进制
package main

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func handleSaveShell(w http.ResponseWriter, r *http.Request) {
	version := "dev"
	if b, err := readEmbedFile("VERSION"); err == nil {
		version = strings.TrimSpace(string(b))
	}

	// 1. 打包 skill 树到内存 tar.gz
	tarBuf, err := buildSkillTarGz()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "build tarball", err)
		return
	}

	// 2. base64 编码 payload
	payload := base64.StdEncoding.EncodeToString(tarBuf)
	// 每 76 字符换行，便于 tail/base64 正确处理
	payload = wrap76(payload)

	// 3. 合成 shell 脚本
	script := buildShellScript(version, payload)

	// 4. 保存到 version/
	dstDir := "~/skills-version"
	if err := os.MkdirAll(dstDir, 0755); err != nil {
		home, _ := os.UserHomeDir()
		dstDir = filepath.Join(home, "Downloads")
	}
	filename := fmt.Sprintf("install-skills-v%s-%s.sh", version, time.Now().Format("20060102-1504"))
	dstPath := filepath.Join(dstDir, filename)
	if err := os.WriteFile(dstPath, []byte(script), 0755); err != nil {
		writeError(w, http.StatusInternalServerError, "write", err)
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
		"name": filename,
		"size": size,
		"dir":  dstDir,
	})
}

// buildSkillTarGz 打包规范内容到 tar.gz（给 Shell 安装脚本用的 payload）
//
// v1.7.8 起（P2 方案）：**优先从本地已装目录读取**，把打包人的所有本地修改
// （含 custom-*.md、手改 references、更新过的 orangecat 模板等）完整带给接收方。
// 如果某个 live 目录不存在（比如没装过 orangecat），该目录才回退到 embed。
// 目的：消除"发送方看到的规范 ≠ 接收方装上后的规范"的不对称。
//
// tar 布局（和旧版一致，确保 Shell 脚本端兼容）：
//   claude/go-team-standards/**    ← 优先从 ~/.claude/skills/go-team-standards/
//   claude/orangecat/**            ← 优先从 ~/.claude/skills/orangecat/（新增）
//   cursor/rules/**                ← 优先从 ~/.cursor/rules/
//   assets/.golangci.yml           ← 始终从 embed（和技术栈相关，不该改）
//   META-source.txt                ← 标记每组来自 live / embed，方便诊断
func buildSkillTarGz() ([]byte, error) {
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gz)

	home, _ := os.UserHomeDir()
	var sourceLog []string

	// 写一个文件到 tar 的辅助
	writeTarFile := func(name string, data []byte) error {
		h := &tar.Header{
			Name: name, Mode: 0644, Size: int64(len(data)),
			ModTime: time.Now(),
		}
		if err := tw.WriteHeader(h); err != nil {
			return err
		}
		_, err := tw.Write(data)
		return err
	}

	// 优先从 live dir 抓；找不到才回退 embed
	packBundle := func(liveDir, embedRoot, tarPrefix string, skipMetadata bool) error {
		if home != "" && dirExists(liveDir) {
			sourceLog = append(sourceLog, tarPrefix+" ← live: "+liveDir)
			return filepath.Walk(liveDir, func(p string, info os.FileInfo, err error) error {
				if err != nil || info.IsDir() {
					return nil
				}
				base := filepath.Base(p)
				if skipMetadata {
					switch base {
					case ".installed-version", ".DS_Store", "Thumbs.db":
						return nil
					}
					if strings.HasPrefix(base, "._") {
						return nil
					}
				}
				rel, _ := filepath.Rel(liveDir, p)
				data, rerr := os.ReadFile(p)
				if rerr != nil {
					return nil
				}
				return writeTarFile(filepath.ToSlash(filepath.Join(tarPrefix, rel)), data)
			})
		}
		// 回退 embed
		sourceLog = append(sourceLog, tarPrefix+" ← embed (live dir 未找到)")
		return fs.WalkDir(embeddedFS, embedRoot, func(p string, d fs.DirEntry, err error) error {
			if err != nil || d.IsDir() {
				return err
			}
			data, rerr := embeddedFS.ReadFile(p)
			if rerr != nil {
				return rerr
			}
			// embed 路径是 "claude/go-team-standards/xxx"，去掉 embedRoot 前缀
			rel := strings.TrimPrefix(p, embedRoot)
			rel = strings.TrimPrefix(rel, "/")
			return writeTarFile(filepath.ToSlash(filepath.Join(tarPrefix, rel)), data)
		})
	}

	// 1. go-team-standards
	if err := packBundle(
		filepath.Join(home, ".claude", "skills", "go-team-standards"),
		"claude/go-team-standards",
		"claude/go-team-standards",
		true,
	); err != nil {
		return nil, err
	}

	// 2. orangecat（新增）
	if err := packBundle(
		filepath.Join(home, ".claude", "skills", "orangecat"),
		"claude/orangecat",
		"claude/orangecat",
		true,
	); err != nil {
		return nil, err
	}

	// 3. Cursor rules
	if err := packBundle(
		filepath.Join(home, ".cursor", "rules"),
		"cursor/rules",
		"cursor/rules",
		true,
	); err != nil {
		return nil, err
	}

	// 4. assets/.golangci.yml —— 跟技术栈绑定，始终用 embed
	if data, err := embeddedFS.ReadFile("assets/.golangci.yml"); err == nil {
		if err := writeTarFile("assets/.golangci.yml", data); err != nil {
			return nil, err
		}
	}
	sourceLog = append(sourceLog, "assets/.golangci.yml ← embed (固定技术栈)")

	// 5. 补漏：如果 live 里因为某些原因没同步 customs，再从 custom-rules.json 补一次
	//    （belt & suspenders —— live 一般已经有了，但防止极端情况）
	customs, _ := loadCustomRules()
	for _, cr := range customs {
		writeTarFile(
			"cursor/rules/custom-"+cr.ID+".mdc",
			[]byte(cr.renderMDC()),
		)
		writeTarFile(
			"claude/go-team-standards/references/custom-"+cr.ID+".md",
			[]byte(cr.renderMD()),
		)
	}

	// 6. 元信息文件
	meta := "Team Standards Shell Payload · 打包时间 " + time.Now().Format("2006-01-02 15:04:05") + "\n\n"
	meta += "打包来源：\n"
	for _, s := range sourceLog {
		meta += "  · " + s + "\n"
	}
	meta += "\n说明：\n"
	meta += "  · live 表示从打包人的 ~/.claude/skills 或 ~/.cursor/rules 读取（含本地修改）\n"
	meta += "  · embed 表示从 App 内置版本读取（= App 编译时的版本）\n"
	_ = writeTarFile("META-source.txt", []byte(meta))

	if err := tw.Close(); err != nil {
		return nil, err
	}
	if err := gz.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func wrap76(s string) string {
	var sb strings.Builder
	for i := 0; i < len(s); i += 76 {
		end := i + 76
		if end > len(s) {
			end = len(s)
		}
		sb.WriteString(s[i:end])
		sb.WriteByte('\n')
	}
	return sb.String()
}

func buildShellScript(version, payload string) string {
	// bash 自解压脚本模板
	// 用 __PAYLOAD_BELOW__ 标记分割，awk 找到分界行，tail 截取 + base64 decode + tar 解压
	// base64 decode：macOS 用 -D，Linux 用 -d → 用 || 试两次，兼容两端
	return `#!/usr/bin/env bash
# Team Standards · 独立安装脚本（无需 App，无需 Go / Xcode）
# Version: ` + version + `
#
# 用法：
#   chmod +x ` + `$(basename "$0")` + `
#   ./` + `$(basename "$0")` + `
#
# 会把 Claude Skill 和 Cursor Rules 装到：
#   ~/.claude/skills/go-team-standards/
#   ~/.cursor/rules/
set -euo pipefail

if [[ "${EUID:-1}" == "0" ]]; then
    echo "✗ 不要用 sudo/root 跑，应该以你自己的用户身份执行" >&2
    exit 1
fi

echo "→ Team Standards Skills 独立安装脚本 v` + version + `"
echo ""

CLAUDE="$HOME/.claude/skills/go-team-standards"
ORANGECAT="$HOME/.claude/skills/orangecat"
CURSOR="$HOME/.cursor/rules"

# 1. 提取 payload 到临时目录
TMP=$(mktemp -d)
trap 'rm -rf "$TMP"' EXIT

PAYLOAD_LINE=$(awk '/^__PAYLOAD_BELOW__$/{print NR+1; exit}' "$0")
if [[ -z "$PAYLOAD_LINE" ]]; then
    echo "✗ 脚本损坏：找不到 __PAYLOAD_BELOW__ 标记" >&2
    exit 1
fi

# macOS base64 用 -D，Linux 用 -d，先 -D 不行再 -d
decode_base64() {
    if base64 -D </dev/null 2>/dev/null; then
        base64 -D
    else
        base64 -d
    fi
}

echo "→ 解压 skill 文件..."
tail -n +"$PAYLOAD_LINE" "$0" | decode_base64 | tar xz -C "$TMP"

# 打印打包来源元信息（如果有），帮接收方知道来自 live 还是 embed
if [[ -f "$TMP/META-source.txt" ]]; then
    echo ""
    cat "$TMP/META-source.txt"
    echo ""
fi

# 2. 安装 Claude Skill · go-team-standards
echo "→ 安装 Claude Skill 到 $CLAUDE"
rm -rf "$CLAUDE"
mkdir -p "$CLAUDE/references" "$CLAUDE/assets" "$CLAUDE/demos"
if [[ -d "$TMP/claude/go-team-standards" ]]; then
    cp -R "$TMP/claude/go-team-standards/." "$CLAUDE/"
fi
if [[ -f "$TMP/assets/.golangci.yml" ]]; then
    cp "$TMP/assets/.golangci.yml" "$CLAUDE/assets/"
fi
echo "` + version + `" > "$CLAUDE/.installed-version"

# 2b. 安装 Claude Skill · orangecat（v1.7.8 新增）
if [[ -d "$TMP/claude/orangecat" ]]; then
    echo "→ 安装 OrangeCat 提测文档 Skill 到 $ORANGECAT"
    rm -rf "$ORANGECAT"
    mkdir -p "$ORANGECAT/references"
    cp -R "$TMP/claude/orangecat/." "$ORANGECAT/"
    echo "` + version + `" > "$ORANGECAT/.installed-version"
fi

# 3. 安装 Cursor Rules
echo "→ 安装 Cursor Rules 到 $CURSOR"
mkdir -p "$CURSOR"
find "$CURSOR" -maxdepth 1 -name '[0-9][0-9]-*.mdc' -delete 2>/dev/null || true
find "$CURSOR" -maxdepth 1 -name 'custom-*.mdc' -delete 2>/dev/null || true
if [[ -d "$TMP/cursor/rules" ]]; then
    cp "$TMP/cursor/rules/"*.mdc "$CURSOR/" 2>/dev/null || true
fi
echo "` + version + `" > "$CURSOR/.installed-version"

echo ""
echo "✓ 安装完成（v` + version + `）"
echo ""
echo "统计："
echo "  Claude references:  $(ls "$CLAUDE/references" 2>/dev/null | wc -l | tr -d ' ') 个"
echo "  Claude orangecat:   $(ls "$ORANGECAT/references" 2>/dev/null | wc -l | tr -d ' ') 个 template"
echo "  Cursor .mdc:        $(ls "$CURSOR"/*.mdc 2>/dev/null | wc -l | tr -d ' ') 个"
echo ""
echo "下一步："
echo "  · 彻底重启 Claude Code（Cmd+Q 再打开，不是 reload）"
echo "  · 彻底重启 Cursor"
echo "  · 之后写 Go 代码时规则自动生效"
exit 0

__PAYLOAD_BELOW__
` + payload
}
