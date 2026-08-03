// App 自动更新检查 —— L1（通知 + 引导）
//
// 设计：
//   - 拉 GitHub Releases API 拿最新 tag
//   - 与本地 VERSION 文件比对
//   - 启动时静默 check，发现新版顶部 banner
//   - 用户支持「跳过此版本」永久不再提示
//   - 用户点下载 → 走系统 open 命令在浏览器拉 DMG，自己手动装
//
// 配置：~/Library/Application Support/TeamStandards/app-update.json
// （含 last_checked_at / last_seen_version / skipped_versions[]）

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"
)

// ---------- 配置存储 ----------

type AppUpdateConfig struct {
	ReleaseProject   string    `json:"release_project"`    // "kary2999/go-skill"
	LastCheckedAt    time.Time `json:"last_checked_at"`
	LastSeenVersion  string    `json:"last_seen_version"`  // 上次拉到的 latest tag（带 v 前缀）
	SkippedVersions  []string  `json:"skipped_versions"`   // 用户点过"跳过此版本"的 tag
	LastCheckError   string    `json:"last_check_error"`
}

const (
	DefaultReleaseProject = "kary2999/go-skill"
)

func appUpdateConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(home, "Library", "Application Support", "TeamStandards")
	if err := os.MkdirAll(dir, 0700); err != nil {
		return "", err
	}
	return filepath.Join(dir, "app-update.json"), nil
}

func loadAppUpdateConfig() (*AppUpdateConfig, error) {
	p, err := appUpdateConfigPath()
	if err != nil {
		return nil, err
	}
	b, err := os.ReadFile(p)
	if err != nil {
		if os.IsNotExist(err) {
			return &AppUpdateConfig{
				ReleaseProject: DefaultReleaseProject,
			}, nil
		}
		return nil, err
	}
	var c AppUpdateConfig
	if err := json.Unmarshal(b, &c); err != nil {
		return nil, err
	}
	// 强制覆盖为正确值，防止旧版本写入了错误的 release_project
	// 导致其他用户的 App 始终连接错误仓库，看不到更新 banner。
	c.ReleaseProject = DefaultReleaseProject
	return &c, nil
}

func saveAppUpdateConfig(c *AppUpdateConfig) error {
	p, err := appUpdateConfigPath()
	if err != nil {
		return err
	}
	b, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(p, b, 0600)
}

// ---------- 读本地 VERSION（embedded） ----------

func currentAppVersion() string {
	if b, err := readEmbedFile("VERSION"); err == nil {
		return strings.TrimSpace(string(b))
	}
	return "dev"
}

// ---------- GitHub Releases API ----------

type githubReleaseAsset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

type githubRelease struct {
	TagName     string               `json:"tag_name"`
	Name        string               `json:"name"`
	Body        string               `json:"body"`
	PublishedAt time.Time            `json:"published_at"`
	Assets      []githubReleaseAsset `json:"assets"`
}

// 拉最新 release
func fetchLatestRelease(project string) (*githubRelease, error) {
	apiURL := fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", project)
	client := newProxyAwareHTTPClient(30 * time.Second)
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "TeamStandards/"+currentAppVersion())
	req.Header.Set("Accept", "application/vnd.github+json")
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("GitHub Releases API %d: %s", resp.StatusCode, truncate(string(body), 200))
	}
	var rel githubRelease
	if err := json.Unmarshal(body, &rel); err != nil {
		return nil, fmt.Errorf("parse release: %w", err)
	}
	if rel.TagName == "" {
		return nil, fmt.Errorf("仓库 %s 还没发布过 release", project)
	}
	return &rel, nil
}

// 比较两个 semver 形如 v1.7.34（左 > 右 返回 1，相等 0，左 < 右 -1）
// 容错：解析失败按字符串比较
func compareSemver(a, b string) int {
	a = strings.TrimPrefix(a, "v")
	b = strings.TrimPrefix(b, "v")
	pa := strings.Split(a, ".")
	pb := strings.Split(b, ".")
	maxLen := len(pa)
	if len(pb) > maxLen {
		maxLen = len(pb)
	}
	for i := 0; i < maxLen; i++ {
		var ai, bi int
		if i < len(pa) {
			fmt.Sscanf(pa[i], "%d", &ai)
		}
		if i < len(pb) {
			fmt.Sscanf(pb[i], "%d", &bi)
		}
		if ai > bi {
			return 1
		}
		if ai < bi {
			return -1
		}
	}
	return 0
}

func contains(list []string, s string) bool {
	for _, v := range list {
		if v == s {
			return true
		}
	}
	return false
}

// 找第一个名字带 .dmg 的 asset
func findDMGAsset(rel *githubRelease) *githubReleaseAsset {
	for i, a := range rel.Assets {
		if strings.HasSuffix(strings.ToLower(a.Name), ".dmg") {
			return &rel.Assets[i]
		}
	}
	return nil
}

// 找 upgrade-helper.sh asset（用于让老版本 binary 下载最新 helper，打破嵌入旧 helper 的鸡蛋困境）
func findHelperAsset(rel *githubRelease) *githubReleaseAsset {
	for i, a := range rel.Assets {
		if strings.EqualFold(a.Name, "upgrade-helper.sh") {
			return &rel.Assets[i]
		}
	}
	return nil
}

// ---------- HTTP API handlers ----------

// GET /api/app-update/check
//
// 返回：
//   {
//     "current_version":  "v1.7.33",
//     "latest_version":   "v1.7.34",
//     "has_update":       true,
//     "is_skipped":       false,
//     "release_url":      "https://...",
//     "dmg_url":          "https://.../TeamStandards-v1.7.34-...dmg",
//     "release_notes":    "<markdown>",
//     "released_at":      "2026-05-10T...",
//     "last_checked_at":  "2026-05-10T...",
//     "error":            "" (only when reachable=false)
//   }
func handleAppUpdateCheck(w http.ResponseWriter, r *http.Request) {
	c, err := loadAppUpdateConfig()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "load config", err)
		return
	}

	current := "v" + currentAppVersion()
	rel, err := fetchLatestRelease(c.ReleaseProject)
	c.LastCheckedAt = time.Now()
	if err != nil {
		c.LastCheckError = err.Error()
		_ = saveAppUpdateConfig(c)
		writeJSON(w, http.StatusOK, map[string]any{
			"current_version": current,
			"reachable":       false,
			"error":           err.Error(),
			"last_checked_at": c.LastCheckedAt,
		})
		return
	}
	c.LastCheckError = ""
	c.LastSeenVersion = rel.TagName
	_ = saveAppUpdateConfig(c)

	hasUpdate := compareSemver(rel.TagName, current) > 0
	isSkipped := contains(c.SkippedVersions, rel.TagName)

	dmg := findDMGAsset(rel)
	dmgURL := ""
	dmgName := ""
	if dmg != nil {
		dmgURL = dmg.BrowserDownloadURL
		dmgName = dmg.Name
	}
	releaseURL := fmt.Sprintf("https://github.com/%s/releases/tag/%s",
		c.ReleaseProject, rel.TagName)

	writeJSON(w, http.StatusOK, map[string]any{
		"current_version":  current,
		"latest_version":   rel.TagName,
		"has_update":       hasUpdate,
		"is_skipped":       isSkipped,
		"reachable":        true,
		"release_url":      releaseURL,
		"dmg_url":          dmgURL,
		"dmg_name":         dmgName,
		"release_notes":    rel.Body,
		"release_name":     rel.Name,
		"released_at":      rel.PublishedAt,
		"last_checked_at":  c.LastCheckedAt,
		"skipped_versions": c.SkippedVersions,
	})
}

// POST /api/app-update/skip
// body: {version: "v1.7.34"}
func handleAppUpdateSkip(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Version string `json:"version"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "bad json", err)
		return
	}
	if req.Version == "" {
		writeError(w, http.StatusBadRequest, "version 必填", nil)
		return
	}
	c, err := loadAppUpdateConfig()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "load", err)
		return
	}
	if !contains(c.SkippedVersions, req.Version) {
		c.SkippedVersions = append(c.SkippedVersions, req.Version)
	}
	if err := saveAppUpdateConfig(c); err != nil {
		writeError(w, http.StatusInternalServerError, "save", err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"ok":               true,
		"skipped_versions": c.SkippedVersions,
	})
}

// POST /api/app-update/unskip
// body: {version: "v1.7.34"}（清掉某个跳过项 / 全部清空）
func handleAppUpdateUnskip(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Version string `json:"version"`
		All     bool   `json:"all"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "bad json", err)
		return
	}
	c, err := loadAppUpdateConfig()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "load", err)
		return
	}
	if req.All {
		c.SkippedVersions = nil
	} else if req.Version != "" {
		var kept []string
		for _, v := range c.SkippedVersions {
			if v != req.Version {
				kept = append(kept, v)
			}
		}
		c.SkippedVersions = kept
	}
	if err := saveAppUpdateConfig(c); err != nil {
		writeError(w, http.StatusInternalServerError, "save", err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"ok":               true,
		"skipped_versions": c.SkippedVersions,
	})
}

// POST /api/app-update/open-download
// body: {url: "https://...dmg"}
// 用系统 open 命令打开 URL（浏览器接管下载）
func handleAppUpdateOpenDownload(w http.ResponseWriter, r *http.Request) {
	var req struct {
		URL string `json:"url"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "bad json", err)
		return
	}
	if req.URL == "" {
		writeError(w, http.StatusBadRequest, "url 必填", nil)
		return
	}
	if !strings.HasPrefix(req.URL, "https://") {
		writeError(w, http.StatusBadRequest, "URL 必须 https://", nil)
		return
	}
	cmd := exec.Command("open", req.URL)
	if err := cmd.Run(); err != nil {
		writeError(w, http.StatusInternalServerError, "open 失败", err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

// ============================================================
// L2 · 一键自升级（v1.7.38）
// ============================================================

type AutoUpgradeProgress struct {
	Phase      string `json:"phase"` // idle / downloading / ready / failed
	Percent    int    `json:"percent"`
	BytesDone  int64  `json:"bytes_done"`
	BytesTotal int64  `json:"bytes_total"`
	Message    string `json:"message"`
	Error      string `json:"error,omitempty"`
	LogPath    string `json:"log_path,omitempty"`
	TargetApp  string `json:"target_app,omitempty"`
}

var (
	upgradeMu       sync.Mutex
	upgradeProgress AutoUpgradeProgress
)

func setUpgradePhase(phase, msg string) {
	upgradeMu.Lock()
	upgradeProgress.Phase = phase
	upgradeProgress.Message = msg
	upgradeMu.Unlock()
}

func setUpgradeProgress(done, total int64) {
	upgradeMu.Lock()
	upgradeProgress.BytesDone = done
	upgradeProgress.BytesTotal = total
	if total > 0 {
		upgradeProgress.Percent = int(done * 100 / total)
	}
	upgradeMu.Unlock()
}

func setUpgradeError(err string) {
	upgradeMu.Lock()
	upgradeProgress.Phase = "failed"
	upgradeProgress.Error = err
	upgradeMu.Unlock()
}

func currentAppBundlePath() (string, error) {
	exe, err := os.Executable()
	if err != nil {
		return "", err
	}
	p := exe
	for i := 0; i < 6; i++ {
		if strings.HasSuffix(p, ".app") {
			return p, nil
		}
		parent := filepath.Dir(p)
		if parent == p {
			break
		}
		p = parent
	}
	return "", fmt.Errorf("not running as .app bundle (exe=%s)", exe)
}

type progressReader struct {
	r       io.Reader
	done    *int64
	total   int64
	lastUpd time.Time
}

func (pr *progressReader) Read(p []byte) (int, error) {
	n, err := pr.r.Read(p)
	if n > 0 {
		*pr.done += int64(n)
		if time.Since(pr.lastUpd) > 100*time.Millisecond {
			setUpgradeProgress(*pr.done, pr.total)
			pr.lastUpd = time.Now()
		}
	}
	return n, err
}

func handleAppUpdateAutoInstall(w http.ResponseWriter, r *http.Request) {
	upgradeMu.Lock()
	if upgradeProgress.Phase == "downloading" || upgradeProgress.Phase == "ready" {
		curr := upgradeProgress.Phase
		msg := upgradeProgress.Message
		upgradeMu.Unlock()
		// 后台自动升级已在进行中，不是失败，告知前端正常等待即可
		writeJSON(w, http.StatusOK, map[string]any{
			"ok":      true,
			"phase":   curr,
			"message": msg,
			"already_running": true,
		})
		return
	}
	upgradeProgress = AutoUpgradeProgress{Phase: "downloading", Message: "正在准备…"}
	upgradeMu.Unlock()

	go performAutoUpgrade()

	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "phase": "downloading"})
}

func handleAppUpdateProgress(w http.ResponseWriter, r *http.Request) {
	upgradeMu.Lock()
	p := upgradeProgress
	upgradeMu.Unlock()
	writeJSON(w, http.StatusOK, p)
}

func performAutoUpgrade() {
	currentApp, err := currentAppBundlePath()
	if err != nil {
		setUpgradeError("无法定位当前 App 路径（可能不是从 .app 启动）：" + err.Error())
		return
	}
	upgradeMu.Lock()
	upgradeProgress.TargetApp = currentApp
	upgradeMu.Unlock()

	setUpgradePhase("downloading", "拉取最新 release…")
	c, err := loadAppUpdateConfig()
	if err != nil {
		setUpgradeError("加载配置失败：" + err.Error())
		return
	}
	rel, err := fetchLatestRelease(c.ReleaseProject)
	if err != nil {
		setUpgradeError("拉 release 失败：" + err.Error())
		return
	}
	dmg := findDMGAsset(rel)
	if dmg == nil {
		setUpgradeError("最新 release 里没有 .dmg 文件")
		return
	}
	dmgURL := dmg.BrowserDownloadURL

	tmpDMG := filepath.Join(os.TempDir(), "team-standards-update.dmg")
	_ = os.Remove(tmpDMG)

	setUpgradePhase("downloading", "下载 "+dmg.Name+"…")
	if err := downloadDMGWithProgress(dmgURL, tmpDMG); err != nil {
		setUpgradeError("下载失败：" + err.Error())
		return
	}

	// v1.7.54: 在 Go 里预先 mount DMG，把挂载路径作为第 5 个参数传给 helper。
	// 避免 helper 脚本用 awk 解析含空格的挂载路径时截断失败。
	setUpgradePhase("downloading", "挂载安装包…")
	mountPath, err := mountDMGPath(tmpDMG)
	if err != nil {
		setUpgradeError("挂载 DMG 失败：" + err.Error())
		return
	}

	setUpgradePhase("downloading", "准备 helper…")
	// 优先从 release assets 下载最新 upgrade-helper.sh：
	//   老版本 binary 嵌入的是旧 helper（awk 截断含空格路径），
	//   下载最新 helper 可让任何老版本都能正确自升级，打破鸡蛋困境。
	helperPath := filepath.Join(os.TempDir(), "team-standards-upgrade-helper.sh")
	helperBytes := downloadHelperFromRelease(rel)
	if helperBytes == nil {
		// 下载失败，回退到嵌入版
		helperBytes, err = readEmbedFile("scripts/upgrade-helper.sh")
		if err != nil {
			_ = exec.Command("hdiutil", "detach", "-force", mountPath).Run()
			setUpgradeError("读 helper 脚本失败：" + err.Error())
			return
		}
	}
	if err := os.WriteFile(helperPath, helperBytes, 0755); err != nil {
		_ = exec.Command("hdiutil", "detach", "-force", mountPath).Run()
		setUpgradeError("写 helper 失败：" + err.Error())
		return
	}

	logPath := filepath.Join(os.TempDir(), "team-standards-upgrade.log")
	upgradeMu.Lock()
	upgradeProgress.LogPath = logPath
	upgradeMu.Unlock()

	// 版本升级时自动拉取规范同步（异步，不阻塞升级流程）
	go func() {
		_ = doStandardsSyncPull()
	}()

	pidStr := fmt.Sprintf("%d", os.Getpid())
	// 第 5 个参数传预挂载路径；helper 若收到则跳过自己的 mount 步骤
	cmd := exec.Command("/bin/bash", helperPath, pidStr, tmpDMG, currentApp, logPath, mountPath)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.Stdin = nil
	cmd.Stdout = nil
	cmd.Stderr = nil
	if err := cmd.Start(); err != nil {
		_ = exec.Command("hdiutil", "detach", "-force", mountPath).Run()
		setUpgradeError("启动 helper 失败：" + err.Error())
		return
	}
	_ = cmd.Process.Release()

	setUpgradePhase("ready", "准备就绪，App 3 秒后退出 → helper 接手安装")

	// helper 在等父进程退出，App 必须主动 quit，否则 helper 60s 超时 abort。
	// 给 webview 3 秒时间把 phase=ready 状态推给前端，然后退出。
	go func() {
		time.Sleep(3 * time.Second)
		os.Exit(0)
	}()
}

// mountDMGPath 用 hdiutil attach -plist 挂载 DMG，返回挂载点路径。
// 使用 plist 输出避免含空格的卷名被 awk/fields 截断。
func mountDMGPath(dmgPath string) (string, error) {
	out, err := exec.Command("hdiutil", "attach", "-nobrowse", "-readonly", "-plist", dmgPath).Output()
	if err != nil {
		return "", fmt.Errorf("hdiutil attach: %w", err)
	}
	// 用 python3 解析 plist（Go 标准库无 plist 支持）
	py := exec.Command("python3", "-c", `
import plistlib, sys
d = plistlib.loads(sys.stdin.buffer.read())
mps = [e.get('mount-point','') for e in d.get('system-entities',[]) if e.get('mount-point','')]
print(mps[-1] if mps else '', end='')
`)
	py.Stdin = bytes.NewReader(out)
	mp, err := py.Output()
	if err != nil {
		return "", fmt.Errorf("parse plist: %w", err)
	}
	mount := strings.TrimSpace(string(mp))
	if mount == "" {
		return "", fmt.Errorf("hdiutil 未返回挂载点")
	}
	return mount, nil
}

// downloadHelperFromRelease 尝试从 release assets 下载最新 upgrade-helper.sh。
// 任何错误（网络、找不到 asset）直接返回 nil，调用方回退到嵌入版。
func downloadHelperFromRelease(rel *githubRelease) []byte {
	asset := findHelperAsset(rel)
	if asset == nil {
		return nil
	}
	u := asset.BrowserDownloadURL
	if u == "" {
		return nil
	}
	client := newProxyAwareHTTPClient(30 * time.Second)
	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		return nil
	}
	req.Header.Set("User-Agent", "TeamStandards-App/auto-upgrade")
	resp, err := client.Do(req)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil
	}
	b, err := io.ReadAll(resp.Body)
	if err != nil || len(b) == 0 {
		return nil
	}
	return b
}

func downloadDMGWithProgress(rawURL, dst string) error {
	req, err := http.NewRequest("GET", rawURL, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", "TeamStandards-App/auto-upgrade")
	// 走代理（如配置了）+ 5min timeout
	client := newProxyAwareHTTPClient(5 * time.Minute)
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return fmt.Errorf("HTTP %d", resp.StatusCode)
	}
	total := resp.ContentLength
	if total > 0 {
		setUpgradeProgress(0, total)
	}
	f, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer f.Close()
	var done int64
	pr := &progressReader{r: resp.Body, done: &done, total: total}
	if _, err := io.Copy(f, pr); err != nil {
		return err
	}
	setUpgradeProgress(done, done)
	return nil
}

// ---------- 启动后台静默检查 ----------
//
// v1.7.51: 每分钟检测一次，发现新版本立即静默自动升级（下载 DMG → helper 安装）。
// 同一版本只触发一次，避免重复下载。

const AppUpdateTickInterval = 1 * time.Minute

func appUpdateBackgroundLoop() {
	appUpdateBackgroundCheck()
	t := time.NewTicker(AppUpdateTickInterval)
	defer t.Stop()
	for range t.C {
		appUpdateBackgroundCheck()
	}
}

func appUpdateBackgroundCheck() {
	c, err := loadAppUpdateConfig()
	if err != nil {
		return
	}
	rel, err := fetchLatestRelease(c.ReleaseProject)
	c.LastCheckedAt = time.Now()
	if err != nil {
		c.LastCheckError = err.Error()
		_ = saveAppUpdateConfig(c)
		return
	}
	c.LastCheckError = ""
	c.LastSeenVersion = rel.TagName
	_ = saveAppUpdateConfig(c)

	current := "v" + currentAppVersion()
	if compareSemver(rel.TagName, current) > 0 {
		// 有新版本——触发自动升级（幂等：upgradeProgress.Phase 保护重复触发）
		upgradeMu.Lock()
		alreadyRunning := upgradeProgress.Phase == "downloading" || upgradeProgress.Phase == "ready"
		upgradeMu.Unlock()
		if !alreadyRunning {
			upgradeMu.Lock()
			upgradeProgress = AutoUpgradeProgress{Phase: "downloading", Message: "后台自动升级：" + rel.TagName}
			upgradeMu.Unlock()
			go performAutoUpgrade()
		}
	}
}
