// 规范云端同步 —— 从公开 GitLab 仓库拉取最新 standards/*.md 覆盖本地
//
// 设计：
//   - 只更 standards 源文件，不动 SKILL.md / commands / 二进制
//   - 公开 readonly mirror，无需认证
//   - 落盘到 ~/.claude/skills/go-team-standards/references/ + ~/.cursor/skills-cursor/.../references/
//   - 离线兼容：检查/拉取失败不阻断 App 启动，回退到内嵌
//
// 配置：~/Library/Application Support/TeamStandards/standards-sync.json
// （mode 0600，含 repo_url / project_id / last_synced_sha）

package main

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// 从 GitHub commits Atom feed 里提取 40 位 commit SHA
var reCommitSHA = regexp.MustCompile(`[Cc]ommit/([0-9a-f]{40})`)

// ---------- 配置存储 ----------

type StandardsSyncConfig struct {
	RepoURL        string    `json:"repo_url"`         // 形如 https://github.com/team/standards
	Branch         string    `json:"branch"`           // 默认 main
	LastSyncedSHA  string    `json:"last_synced_sha"`  // 上次成功拉取的 commit SHA
	LastSyncedAt   time.Time `json:"last_synced_at"`   // 上次成功拉取时间
	LastCheckedAt  time.Time `json:"last_checked_at"`  // 上次检查时间（无论成败）
	LastCheckError string    `json:"last_check_error"` // 上次检查失败原因
}

func standardsSyncConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(home, "Library", "Application Support", "TeamStandards")
	if err := os.MkdirAll(dir, 0700); err != nil {
		return "", err
	}
	return filepath.Join(dir, "standards-sync.json"), nil
}

func loadStandardsSyncConfig() (*StandardsSyncConfig, error) {
	p, err := standardsSyncConfigPath()
	if err != nil {
		return nil, err
	}
	b, err := os.ReadFile(p)
	if err != nil {
		if os.IsNotExist(err) {
			return &StandardsSyncConfig{Branch: "main"}, nil
		}
		return nil, err
	}
	var c StandardsSyncConfig
	if err := json.Unmarshal(b, &c); err != nil {
		return nil, err
	}
	if c.Branch == "" {
		c.Branch = "main"
	}
	// 迁移：旧版本保存的私有 git 地址 → GitHub 公开仓库。
	// 老配置文件不会因为代码改了默认值而更新，这里读到非 GitHub 地址就地替换并落盘。
	if c.RepoURL != "" && !strings.Contains(c.RepoURL, "github.com") {
		c.RepoURL = DefaultStandardsRepoURL
		c.LastSyncedSHA = "" // 换仓库后强制重新同步
		_ = saveStandardsSyncConfig(&c)
	}
	return &c, nil
}

func saveStandardsSyncConfig(c *StandardsSyncConfig) error {
	p, err := standardsSyncConfigPath()
	if err != nil {
		return err
	}
	b, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(p, b, 0600)
}

// ---------- HTTP 客户端（统一超时 + UA + 代理支持） ----------
//
// 代理优先级（v1.7.38）：
//  1. ~/Library/Application Support/TeamStandards/proxy.json 里 explicit proxy_url（最高）
//  2. HTTPS_PROXY / HTTP_PROXY 环境变量（系统默认行为）
//  3. 无代理（直连）

// AppProxyConfig 统一代理配置（影响所有 HTTP 出站）
type AppProxyConfig struct {
	ProxyURL string `json:"proxy_url"` // e.g. http://127.0.0.1:1087
	Enabled  bool   `json:"enabled"`
}

func appProxyConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(home, "Library", "Application Support", "TeamStandards")
	if err := os.MkdirAll(dir, 0700); err != nil {
		return "", err
	}
	return filepath.Join(dir, "proxy.json"), nil
}

func loadAppProxyConfig() *AppProxyConfig {
	p, err := appProxyConfigPath()
	if err != nil {
		return &AppProxyConfig{}
	}
	b, err := os.ReadFile(p)
	if err != nil {
		return &AppProxyConfig{}
	}
	var c AppProxyConfig
	if err := json.Unmarshal(b, &c); err != nil {
		return &AppProxyConfig{}
	}
	return &c
}

func saveAppProxyConfig(c *AppProxyConfig) error {
	p, err := appProxyConfigPath()
	if err != nil {
		return err
	}
	b, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(p, b, 0600)
}

// 构建带代理的 HTTP client
//
// v1.7.46: 走 Clash/V2Ray 这类本地 HTTP 代理时，Go 默认 Transport 会偶发
// "unexpected EOF" / 中途断流——因为 Go 默认尝试 HTTP/2 + 长连接复用，
// 与本地代理的 keepalive 时机经常对不上。做几件事缓解：
//   1) 强制 HTTP/1.1（关掉 H2 探测）
//   2) 显式 TLSHandshakeTimeout / ResponseHeaderTimeout（默认 0 = 永不超时太松）
//   3) DisableKeepAlives = true：每次请求新连接，宁可多握几次手也不要被代理
//      偷偷关掉的旧 conn 坑到（规范同步一小时才一次，性能损失忽略不计）
func newProxyAwareHTTPClient(timeout time.Duration) *http.Client {
	tr := &http.Transport{
		Proxy:                 http.ProxyFromEnvironment, // 默认走 env var
		ForceAttemptHTTP2:     false,
		TLSHandshakeTimeout:   10 * time.Second,
		ResponseHeaderTimeout: 15 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		IdleConnTimeout:       30 * time.Second,
		DisableKeepAlives:     true,
	}
	// 如果显式配了 proxy，覆盖 env
	c := loadAppProxyConfig()
	if c.Enabled && c.ProxyURL != "" {
		if u, err := url.Parse(c.ProxyURL); err == nil {
			tr.Proxy = http.ProxyURL(u)
		}
	}
	return &http.Client{Timeout: timeout, Transport: tr}
}

// v1.7.46: 15s → 30s。走代理 + GitLab API 的链路在内网偶发 10s+，
// 之前的 15s 紧贴边界，偶发 deadline。
var standardsHTTPClient = newProxyAwareHTTPClient(30 * time.Second)

// GET /api/proxy/config
func handleProxyConfigGet(w http.ResponseWriter, r *http.Request) {
	c := loadAppProxyConfig()
	writeJSON(w, http.StatusOK, c)
}

// POST /api/proxy/config
// body: {proxy_url: "http://127.0.0.1:1087", enabled: true}
func handleProxyConfigSave(w http.ResponseWriter, r *http.Request) {
	var req AppProxyConfig
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "bad json", err)
		return
	}
	req.ProxyURL = strings.TrimSpace(req.ProxyURL)
	if req.Enabled && req.ProxyURL == "" {
		writeError(w, http.StatusBadRequest, "启用代理必须填 proxy_url", nil)
		return
	}
	if req.Enabled {
		if u, err := url.Parse(req.ProxyURL); err != nil || u.Host == "" {
			writeError(w, http.StatusBadRequest, "proxy_url 格式错", err)
			return
		}
	}
	if err := saveAppProxyConfig(&req); err != nil {
		writeError(w, http.StatusInternalServerError, "save", err)
		return
	}
	// 刷新全局 client
	standardsHTTPClient = newProxyAwareHTTPClient(30 * time.Second)
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "note": "代理已生效（影响所有出站 HTTP 请求）"})
}

func standardsHTTPGet(rawURL string) ([]byte, int, error) {
	req, err := http.NewRequest("GET", rawURL, nil)
	if err != nil {
		return nil, 0, err
	}
	req.Header.Set("User-Agent", "TeamStandards-App/standards-sync")
	resp, err := standardsHTTPClient.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 50<<20)) // 50MB hard cap
	if err != nil {
		return nil, resp.StatusCode, err
	}
	return body, resp.StatusCode, nil
}

// ---------- GitLab API helpers ----------

// 从 repo_url 解析出 base host 和 namespace/project（用于 GitLab API 调用）
//
// 输入示例：https://github.com/team/standards
// 返回：host=https://github.com, project="team/standards"
func parseGitLabRepoURL(repoURL string) (host, project string, err error) {
	u, err := url.Parse(strings.TrimSuffix(repoURL, "/"))
	if err != nil {
		return "", "", fmt.Errorf("repo_url 格式错: %w", err)
	}
	if u.Scheme == "" || u.Host == "" {
		return "", "", fmt.Errorf("repo_url 缺 scheme 或 host：%s", repoURL)
	}
	host = u.Scheme + "://" + u.Host
	project = strings.TrimPrefix(u.Path, "/")
	project = strings.TrimSuffix(project, ".git")
	if project == "" {
		return "", "", fmt.Errorf("repo_url 解析不出 namespace/project：%s", repoURL)
	}
	return host, project, nil
}

// 拉远端最新 commit SHA（公开 repo 不需 token）
//
// GitHub：GET https://api.github.com/repos/<owner>/<repo>/commits?sha=<branch>&per_page=1
// GitLab：GET <host>/api/v4/projects/<urlencoded path>/repository/commits?ref_name=<branch>&per_page=1
func fetchLatestCommitSHA(repoURL, branch string) (string, error) {
	host, project, err := parseGitLabRepoURL(repoURL)
	if err != nil {
		return "", err
	}
	if strings.Contains(host, "github.com") {
		// 用 commits 的 Atom feed（走 github.com 主站），避免 api.github.com
		// 未认证 60 次/小时的严格限流。feed 第一个 entry 即最新 commit。
		feedURL := fmt.Sprintf("https://github.com/%s/commits/%s.atom",
			project, url.PathEscape(branch))
		body, status, err := standardsHTTPGet(feedURL)
		if err != nil {
			return "", err
		}
		if status != 200 {
			return "", fmt.Errorf("GitHub 返回 %d (仓库不存在或私有？)：%s", status, truncate(string(body), 200))
		}
		// entry id 形如 tag:github.com,2008:Grit::Commit/<40位sha>，
		// 或 link href .../commit/<40位sha>
		m := reCommitSHA.FindSubmatch(body)
		if m == nil {
			return "", fmt.Errorf("无法从 atom feed 解析 commit（仓库空或格式变化）")
		}
		return string(m[1]), nil
	}
	encProject := url.PathEscape(project)
	apiURL := fmt.Sprintf("%s/api/v4/projects/%s/repository/commits?ref_name=%s&per_page=1",
		host, encProject, url.QueryEscape(branch))
	body, status, err := standardsHTTPGet(apiURL)
	if err != nil {
		return "", err
	}
	if status != 200 {
		return "", fmt.Errorf("GitLab API 返回 %d (公开仓库未配置或 URL 错？)：%s", status, truncate(string(body), 200))
	}
	var commits []struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(body, &commits); err != nil {
		return "", fmt.Errorf("解析 commits 失败: %w", err)
	}
	if len(commits) == 0 {
		return "", fmt.Errorf("仓库 %s 分支 %s 没 commit", project, branch)
	}
	return commits[0].ID, nil
}

// 下载 archive zip
// GitHub：<host>/<project>/archive/refs/heads/<branch>.zip
// GitLab：<host>/<project>/-/archive/<branch>/<repo>-<branch>.zip
func downloadArchiveZip(repoURL, branch string) ([]byte, error) {
	host, project, err := parseGitLabRepoURL(repoURL)
	if err != nil {
		return nil, err
	}
	if strings.Contains(host, "github.com") {
		zipURL := fmt.Sprintf("%s/%s/archive/refs/heads/%s.zip", host, project, url.PathEscape(branch))
		body, status, err := standardsHTTPGet(zipURL)
		if err != nil {
			return nil, err
		}
		if status != 200 {
			return nil, fmt.Errorf("下载 zip 返回 %d (URL=%s)", status, zipURL)
		}
		return body, nil
	}
	parts := strings.Split(project, "/")
	repoName := parts[len(parts)-1]
	zipURL := fmt.Sprintf("%s/%s/-/archive/%s/%s-%s.zip",
		host, project, url.PathEscape(branch), repoName, branch)
	body, status, err := standardsHTTPGet(zipURL)
	if err != nil {
		return nil, err
	}
	if status != 200 {
		return nil, fmt.Errorf("下载 zip 返回 %d (URL=%s)", status, zipURL)
	}
	return body, nil
}

// 解压 zip 到内存 map[文件名]内容
// 只保留 .md 和 .png（规范文档），跳过 README 和 .gitignore
func extractStandardsFromZip(data []byte) (map[string][]byte, error) {
	r, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return nil, fmt.Errorf("解 zip 失败: %w", err)
	}
	out := map[string][]byte{}
	for _, f := range r.File {
		// zip 内路径形如 "team-standards-main/database.md"
		_, base := filepath.Split(f.Name)
		if base == "" {
			continue
		}
		if !strings.HasSuffix(base, ".md") && !strings.HasSuffix(base, ".png") {
			continue
		}
		if strings.EqualFold(base, "README.md") {
			continue
		}
		rc, err := f.Open()
		if err != nil {
			return nil, fmt.Errorf("读 %s 失败: %w", f.Name, err)
		}
		content, err := io.ReadAll(io.LimitReader(rc, 10<<20))
		_ = rc.Close()
		if err != nil {
			return nil, fmt.Errorf("读 %s 失败: %w", f.Name, err)
		}
		out[base] = content
	}
	return out, nil
}

// 落盘到 references 目录（同时覆盖 Claude + Cursor）
func writeStandardsToReferences(files map[string][]byte) (changed []string, err error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	targets := []string{
		filepath.Join(home, ".claude", "skills", "go-team-standards", "references"),
		filepath.Join(home, ".cursor", "skills-cursor", "go-team-standards", "references"),
	}
	for _, dir := range targets {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("mkdir %s: %w", dir, err)
		}
		for name, content := range files {
			fp := filepath.Join(dir, name)
			old, _ := os.ReadFile(fp)
			if bytes.Equal(old, content) {
				continue // 内容相同，跳过
			}
			if err := os.WriteFile(fp, content, 0644); err != nil {
				return nil, fmt.Errorf("写 %s: %w", fp, err)
			}
			rel := name + " → " + dir
			changed = append(changed, rel)
		}

		// 镜像清理：删除远端已移除的遗留 .md/.png，避免规范模块列出磁盘上
		// 已废弃的旧文件（catalog 扫磁盘 → 点开却读不到 → 报错）。
		// 保留用户自定义的 custom-* 文件。
		if entries, e := os.ReadDir(dir); e == nil {
			for _, ent := range entries {
				if ent.IsDir() {
					continue
				}
				n := ent.Name()
				if strings.HasPrefix(n, "custom-") {
					continue
				}
				if !strings.HasSuffix(n, ".md") && !strings.HasSuffix(n, ".png") {
					continue
				}
				if _, ok := files[n]; ok {
					continue // 远端仍有此文件
				}
				if err := os.Remove(filepath.Join(dir, n)); err == nil {
					changed = append(changed, "🗑 清理遗留 "+n+" ← "+dir)
				}
			}
		}
	}
	return changed, nil
}

// ---------- HTTP API handlers ----------

// GET /api/standards-sync/config
// 返回当前配置（不含敏感字段，因为公开仓无 token）
func handleStandardsSyncConfigGet(w http.ResponseWriter, r *http.Request) {
	c, err := loadStandardsSyncConfig()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "load config", err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"repo_url":         c.RepoURL,
		"branch":           c.Branch,
		"last_synced_sha":  c.LastSyncedSHA,
		"last_synced_at":   c.LastSyncedAt,
		"last_checked_at":  c.LastCheckedAt,
		"last_check_error": c.LastCheckError,
		"configured":       c.RepoURL != "",
	})
}

// POST /api/standards-sync/config
// body: {repo_url: string, branch?: string}
// 验证 URL + 测试可达性（拉一次 commit SHA），通过才存盘
func handleStandardsSyncConfigSave(w http.ResponseWriter, r *http.Request) {
	var req struct {
		RepoURL string `json:"repo_url"`
		Branch  string `json:"branch"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "bad json", err)
		return
	}
	req.RepoURL = strings.TrimSpace(req.RepoURL)
	if req.RepoURL == "" {
		writeError(w, http.StatusBadRequest, "repo_url 必填", nil)
		return
	}
	if req.Branch == "" {
		req.Branch = "main"
	}
	// 测试可达性
	if _, _, err := parseGitLabRepoURL(req.RepoURL); err != nil {
		writeError(w, http.StatusBadRequest, "URL 格式错", err)
		return
	}
	sha, err := fetchLatestCommitSHA(req.RepoURL, req.Branch)
	if err != nil {
		writeError(w, http.StatusBadRequest, "连不上仓库（公开访问失败）", err)
		return
	}
	c, _ := loadStandardsSyncConfig()
	c.RepoURL = req.RepoURL
	c.Branch = req.Branch
	c.LastCheckedAt = time.Now()
	c.LastCheckError = ""
	if err := saveStandardsSyncConfig(c); err != nil {
		writeError(w, http.StatusInternalServerError, "save config", err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"ok":         true,
		"latest_sha": sha,
		"message":    "配置已保存，仓库可达",
	})
}

// GET /api/standards-sync/check
// 拉远端最新 SHA，与 last_synced_sha 比对
func handleStandardsSyncCheck(w http.ResponseWriter, r *http.Request) {
	c, err := loadStandardsSyncConfig()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "load config", err)
		return
	}
	if c.RepoURL == "" {
		writeJSON(w, http.StatusOK, map[string]any{
			"configured": false,
			"message":    "尚未配置规范仓库",
		})
		return
	}
	sha, err := fetchLatestCommitSHA(c.RepoURL, c.Branch)
	c.LastCheckedAt = time.Now()
	if err != nil {
		c.LastCheckError = err.Error()
		_ = saveStandardsSyncConfig(c)
		writeJSON(w, http.StatusOK, map[string]any{
			"configured": true,
			"reachable":  false,
			"error":      err.Error(),
		})
		return
	}
	c.LastCheckError = ""
	_ = saveStandardsSyncConfig(c)
	hasUpdate := sha != c.LastSyncedSHA
	writeJSON(w, http.StatusOK, map[string]any{
		"configured":   true,
		"reachable":    true,
		"has_update":   hasUpdate,
		"current_sha":  c.LastSyncedSHA,
		"latest_sha":   sha,
		"last_synced":  c.LastSyncedAt,
		"last_checked": c.LastCheckedAt,
	})
}

// POST /api/standards-sync/pull
// 下载最新 zip → 解压 → 写盘 → 更新 last_synced_sha
func handleStandardsSyncPull(w http.ResponseWriter, r *http.Request) {
	c, err := loadStandardsSyncConfig()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "load config", err)
		return
	}
	if c.RepoURL == "" {
		writeError(w, http.StatusBadRequest, "尚未配置规范仓库", nil)
		return
	}
	sha, err := fetchLatestCommitSHA(c.RepoURL, c.Branch)
	if err != nil {
		writeError(w, http.StatusBadGateway, "拉远端 SHA 失败", err)
		return
	}
	zipData, err := downloadArchiveZip(c.RepoURL, c.Branch)
	if err != nil {
		writeError(w, http.StatusBadGateway, "下载 zip 失败", err)
		return
	}
	files, err := extractStandardsFromZip(zipData)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "解 zip 失败", err)
		return
	}
	if len(files) == 0 {
		writeError(w, http.StatusInternalServerError, "zip 内未找到任何 .md/.png 文件", nil)
		return
	}
	changed, err := writeStandardsToReferences(files)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "写盘失败", err)
		return
	}
	c.LastSyncedSHA = sha
	c.LastSyncedAt = time.Now()
	c.LastCheckError = ""
	_ = saveStandardsSyncConfig(c)

	writeJSON(w, http.StatusOK, map[string]any{
		"ok":            true,
		"latest_sha":    sha,
		"files_total":   len(files),
		"files_changed": len(changed),
		"changed":       changed,
		"message":       fmt.Sprintf("已同步 %d 个文件，%d 个变更", len(files), len(changed)),
	})
}

// 默认团队规范仓（公开 readonly mirror）
// 用户首次启动 App 时会自动写入此 URL，省去手动配置。
// 想改成自己的 mirror，去 App 里「⚡ 安装 → 更新与同步 → ☁️ 规范云端同步 → ⚙️ 配置」改即可。
const DefaultStandardsRepoURL = "https://github.com/kary2999/standards.git"
const DefaultStandardsBranch = "main"

// ---------- 后台定时检查（每小时一次） ----------
//
// v1.7.46: 启动时跑一次，然后 1h ticker；发现远端 SHA != LastSyncedSHA 时
// **静默自动覆盖** references/ 下的规范文档（用户已确认：references/ 改动无需重启
// Claude Code / Cursor，下次触发 skill 时读到最新内容）。
// 设计：失败不阻塞、不弹错；错误记到 LastCheckError，前端可见即可。

// StandardsSyncTickInterval 定时间隔；测试时可临时调小。
const StandardsSyncTickInterval = 1 * time.Hour

func standardsSyncBackgroundLoop() {
	runStandardsSyncTick()
	t := time.NewTicker(StandardsSyncTickInterval)
	defer t.Stop()
	for range t.C {
		runStandardsSyncTick()
	}
}

// runStandardsSyncTick 单次检查 + 静默自动覆盖。返回值仅用于 log，外层不关心。
func runStandardsSyncTick() {
	c, err := loadStandardsSyncConfig()
	if err != nil {
		return
	}

	// 首次启动：自动注入默认配置（团队公开 mirror）
	if c.RepoURL == "" {
		c.RepoURL = DefaultStandardsRepoURL
		c.Branch = DefaultStandardsBranch
		_ = saveStandardsSyncConfig(c)
	}

	sha, err := fetchLatestCommitSHA(c.RepoURL, c.Branch)
	c.LastCheckedAt = time.Now()
	if err != nil {
		c.LastCheckError = err.Error()
		_ = saveStandardsSyncConfig(c)
		return
	}
	c.LastCheckError = ""

	// 远端 SHA 与上次同步一致 → 啥也不干
	if sha == c.LastSyncedSHA {
		_ = saveStandardsSyncConfig(c)
		return
	}

	// SHA 不同 → 静默自动覆盖
	zipData, err := downloadArchiveZip(c.RepoURL, c.Branch)
	if err != nil {
		c.LastCheckError = "auto-pull download: " + err.Error()
		_ = saveStandardsSyncConfig(c)
		return
	}
	files, err := extractStandardsFromZip(zipData)
	if err != nil {
		c.LastCheckError = "auto-pull extract: " + err.Error()
		_ = saveStandardsSyncConfig(c)
		return
	}
	if len(files) == 0 {
		c.LastCheckError = "auto-pull: zip 内无 .md/.png 文件"
		_ = saveStandardsSyncConfig(c)
		return
	}
	changed, err := writeStandardsToReferences(files)
	if err != nil {
		c.LastCheckError = "auto-pull write: " + err.Error()
		_ = saveStandardsSyncConfig(c)
		return
	}
	c.LastSyncedSHA = sha
	c.LastSyncedAt = time.Now()
	_ = saveStandardsSyncConfig(c)

	shortSHA := sha
	if len(shortSHA) > 8 {
		shortSHA = shortSHA[:8]
	}
	log.Printf("[standards-sync] auto-pulled %s · %d files · %d changed",
		shortSHA, len(files), len(changed))
}

// standardsSyncBackgroundCheck 兼容旧调用点（启动时跑一次，不进 ticker）。
// 新代码应调用 standardsSyncBackgroundLoop。
func standardsSyncBackgroundCheck() {
	runStandardsSyncTick()
}

// ---------- 工具 ----------

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
