package main

// 静态 API 快照导出 —— GitHub Pages（PWA 静态版）的数据来源。
//
// 用法（在干净 HOME 下跑，避免把本机 ~/.claude 状态/路径写进公开站点）：
//
//	HOME=$(mktemp -d) TS_EXPORT_STATIC=web/api go run .
//
// URL → 文件名映射规则（必须与 web/app.js 里的 staticApiKey() 一致）：
//
//	/api/catalog                     → catalog.json
//	/api/reference?file=go-style.md  → reference__file_go-style.md.json
//
// 即：去掉 /api/ 前缀；有 query 时追加 "__" + query（非 [A-Za-z0-9._-] 字符替换为 "_"）。

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"unicode/utf8"
)

var staticKeyUnsafe = regexp.MustCompile(`[^A-Za-z0-9._-]`)

func staticExportTargets() []string {
	urls := []string{
		"/api/catalog",
		"/api/custom",
		"/api/custom/presets",
		"/api/commands/list",
		"/api/unit-test/manifest",
		"/api/unit-test/status",
		"/api/hooks",
		"/api/installed",
		"/api/persona",
		"/api/orangecat/status",
		"/api/orangecat/template?which=qa",
		"/api/orangecat/template?which=dev",
		"/api/orangecat/template/demo",
		"/api/claude-desktop/skill-md?name=orangecat",
		"/api/coverage/check",
		"/api/standards-sync/config",
		"/api/proxy/config",
		"/api/gsd/status",
		"/api/gsd/list",
		"/api/gsd-framework/list",
		"/api/superpowers/status",
		"/api/gsd-help-zh/status",
		"/api/code-review/status",
		"/api/dev-dna/status",
		"/api/commit-guard/status",
		"/api/commit-guard/scripts",
		"/api/eval/cases",
		"/api/eval/config",
		"/api/logs",
	}
	// 规范原文：standards/*.md 全量快照（卡片点开时按 file= 读取）
	if entries, err := os.ReadDir("standards"); err == nil {
		for _, e := range entries {
			if !e.IsDir() && strings.HasSuffix(e.Name(), ".md") {
				urls = append(urls, "/api/reference?file="+e.Name())
			}
		}
	}
	return urls
}

func staticExportKey(rawURL string) string {
	path, query, _ := strings.Cut(rawURL, "?")
	key := strings.TrimPrefix(path, "/api/")
	if query != "" {
		key += "__" + staticKeyUnsafe.ReplaceAllString(query, "_")
	}
	return key + ".json"
}

// payloadFile 安装负载里的单个文件。文本直接存 content；二进制（png 等）base64。
type payloadFile struct {
	Path     string `json:"path"`     // 目标相对路径（claude: go-team-standards/ 下；cursor: rules/ 下）
	Content  string `json:"content"`
	Encoding string `json:"encoding"` // "utf8" | "base64"
}

func newPayloadFile(path string, b []byte) payloadFile {
	if utf8.Valid(b) {
		return payloadFile{Path: path, Content: string(b), Encoding: "utf8"}
	}
	return payloadFile{Path: path, Content: base64.StdEncoding.EncodeToString(b), Encoding: "base64"}
}

// buildWebInstallPayload 生成网页版（File System Access API 模式）一键安装所需的
// 文件清单，结构与 installClaude / installCursor / installEmbeddedSkill 完全一致：
//   claude: SKILL.md + references/<standards/*> + assets/.golangci.yml + demos/*
//   cursor: cursor/rules/*.mdc
//   codex:  codex/go-team-standards 整棵树
func buildWebInstallPayload() (map[string]any, error) {
	var claude, cursor, codex []payloadFile

	addEmbed := func(list *[]payloadFile, embedPath, dstRel string) error {
		b, err := embeddedFS.ReadFile(embedPath)
		if err != nil {
			return err
		}
		*list = append(*list, newPayloadFile(dstRel, b))
		return nil
	}
	addEmbedDir := func(list *[]payloadFile, embedDir, dstPrefix string) error {
		entries, err := embeddedFS.ReadDir(embedDir)
		if err != nil {
			return err
		}
		for _, e := range entries {
			if e.IsDir() {
				continue
			}
			if err := addEmbed(list, embedDir+"/"+e.Name(), dstPrefix+e.Name()); err != nil {
				return err
			}
		}
		return nil
	}

	if err := addEmbed(&claude, "claude/go-team-standards/SKILL.md", "go-team-standards/SKILL.md"); err != nil {
		return nil, err
	}
	if err := addEmbedDir(&claude, "standards", "go-team-standards/references/"); err != nil {
		return nil, err
	}
	if err := addEmbed(&claude, "assets/.golangci.yml", "go-team-standards/assets/.golangci.yml"); err != nil {
		return nil, err
	}
	if err := addEmbedDir(&claude, "demos", "go-team-standards/demos/"); err != nil {
		return nil, err
	}
	if err := addEmbedDir(&cursor, "cursor/rules", ""); err != nil {
		return nil, err
	}

	// codex：整棵树递归（含 references/ 子目录），目标 ~/.codex/skills/go-team-standards/
	if err := fs.WalkDir(embeddedFS, "codex/go-team-standards", func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		rel := strings.TrimPrefix(p, "codex/")
		b, err := embeddedFS.ReadFile(p)
		if err != nil {
			return err
		}
		codex = append(codex, newPayloadFile(rel, b))
		return nil
	}); err != nil {
		return nil, err
	}

	return map[string]any{"claude": claude, "cursor": cursor, "codex": codex}, nil
}

func exportStaticAPI(mux *http.ServeMux, dir string) error {
	// 快照里不能出现机器本地路径（公开站点）：HOME 一律替换为 ~
	home, _ := os.UserHomeDir()

	for _, u := range staticExportTargets() {
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, u, nil))
		if rec.Code != http.StatusOK {
			log.Printf("⚠ skip %s: HTTP %d", u, rec.Code)
			continue
		}
		body := rec.Body.Bytes()
		if home != "" {
			body = []byte(strings.ReplaceAll(string(body), home, "~"))
		}
		out := filepath.Join(dir, staticExportKey(u))
		if err := os.MkdirAll(filepath.Dir(out), 0o755); err != nil {
			return fmt.Errorf("mkdir %s: %w", filepath.Dir(out), err)
		}
		if err := os.WriteFile(out, body, 0o644); err != nil {
			return fmt.Errorf("write %s: %w", out, err)
		}
		log.Printf("✓ %s → %s (%d bytes)", u, out, len(body))
	}

	// 网页版一键安装负载
	payload, err := buildWebInstallPayload()
	if err != nil {
		return fmt.Errorf("build web install payload: %w", err)
	}
	pb, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	out := filepath.Join(dir, "web-install-payload.json")
	if err := os.WriteFile(out, pb, 0o644); err != nil {
		return err
	}
	log.Printf("✓ web-install-payload → %s (%d bytes)", out, len(pb))

	// 一键安装脚本（curl | bash）：Chrome 禁止网页授权点开头的隐藏目录，
	// 终端脚本是网页版安装的可靠兜底。tar 内嵌 base64，无外部依赖。
	script, err := buildInstallScript(payload)
	if err != nil {
		return fmt.Errorf("build install.sh: %w", err)
	}
	shOut := filepath.Join(filepath.Dir(dir), "install.sh") // web/install.sh
	if err := os.WriteFile(shOut, script, 0o755); err != nil {
		return err
	}
	log.Printf("✓ install.sh → %s (%d bytes)", shOut, len(script))
	return nil
}

// buildInstallScript 把安装负载打成 tar.gz + base64 内嵌 shell 脚本。
// 解包路径以 $HOME 为根：.claude/skills/… / .cursor/rules/… / .codex/skills/…
func buildInstallScript(payload map[string]any) ([]byte, error) {
	var tgz bytes.Buffer
	gw := gzip.NewWriter(&tgz)
	tw := tar.NewWriter(gw)

	addTree := func(prefix string, files []payloadFile) error {
		for _, pf := range files {
			var data []byte
			if pf.Encoding == "base64" {
				b, err := base64.StdEncoding.DecodeString(pf.Content)
				if err != nil {
					return err
				}
				data = b
			} else {
				data = []byte(pf.Content)
			}
			hdr := &tar.Header{Name: prefix + pf.Path, Mode: 0o644, Size: int64(len(data))}
			if err := tw.WriteHeader(hdr); err != nil {
				return err
			}
			if _, err := tw.Write(data); err != nil {
				return err
			}
		}
		return nil
	}
	if err := addTree(".claude/skills/", payload["claude"].([]payloadFile)); err != nil {
		return nil, err
	}
	if err := addTree(".cursor/rules/", payload["cursor"].([]payloadFile)); err != nil {
		return nil, err
	}
	if err := addTree(".codex/skills/", payload["codex"].([]payloadFile)); err != nil {
		return nil, err
	}
	if err := tw.Close(); err != nil {
		return nil, err
	}
	if err := gw.Close(); err != nil {
		return nil, err
	}

	b64 := base64.StdEncoding.EncodeToString(tgz.Bytes())
	var sb strings.Builder
	sb.WriteString(`#!/bin/sh
# GoSkill 团队规范一键安装（由构建流程自动生成，勿手改）
# 用法：curl -fsSL https://kary2999.github.io/go-skill/install.sh | sh
set -e
echo "→ 安装 GoSkill 团队规范到 ~/.claude + ~/.cursor + ~/.codex"
rm -rf "$HOME/.claude/skills/go-team-standards" "$HOME/.codex/skills/go-team-standards"
find "$HOME/.cursor/rules" -maxdepth 1 -name '[0-9][0-9]-*.mdc' -delete 2>/dev/null || true
mkdir -p "$HOME/.claude/skills" "$HOME/.cursor/rules" "$HOME/.codex/skills"
base64 -d <<'GOSKILL_EOF' | tar xz -C "$HOME"
`)
	// 76 列换行，避免超长单行
	for i := 0; i < len(b64); i += 76 {
		end := i + 76
		if end > len(b64) {
			end = len(b64)
		}
		sb.WriteString(b64[i:end])
		sb.WriteByte('\n')
	}
	sb.WriteString(`GOSKILL_EOF
echo "✅ 安装完成：Claude Code / Cursor / Codex 三端规范已就位"
echo "   重启 Claude Code（Cmd+Q）/ Cursor 后生效"
`)
	return []byte(sb.String()), nil
}
