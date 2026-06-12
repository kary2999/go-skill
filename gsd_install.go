// gsd-build/get-shit-done 框架安装支持（v1.7.39 重写）
//
// 之前 v1.7.32 起我们 ship 了一份简版的 GTD 任务清单 skill。
// v1.7.39 起：换成调用 npx 安装真正的 GSD 框架（spec-driven development，含 ~50 个 gsd-* skill）。
//
// 上游：https://github.com/gsd-build/get-shit-done
// 安装命令：npx --yes get-shit-done-cc@latest --claude --global
//
// 行为：
//   - status: 扫 ~/.claude/skills/gsd-* 看装了几个；检测 node / npx 是否可用
//   - install: spawn `bash -lc 'npx --yes get-shit-done-cc@latest --claude --global'`，最多 8 分钟
//   - uninstall: 删 ~/.claude/skills/gsd-* 所有目录（best-effort）
//
// 注：旧版 ~/.claude/skills/gsd/（我们的 GTD 简版）不主动删，
// 用户如果要清理可手工 rm -rf ~/.claude/skills/gsd

package main

import (
	"bufio"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// ---------- 异步 install 进度状态 ----------

type gsdInstallProgress struct {
	Phase     string    `json:"phase"`      // idle / running / done / failed
	StartedAt time.Time `json:"started_at"`
	EndedAt   time.Time `json:"ended_at,omitempty"`
	ExitCode  int       `json:"exit_code"`
	ElapsedMS int64     `json:"elapsed_ms"`
	Log       string    `json:"log"` // 累积 stdout/stderr
	Error     string    `json:"error,omitempty"`
}

var (
	gsdInstallMu       sync.Mutex
	gsdInstallProgress_ gsdInstallProgress
	gsdInstallLogBuf   strings.Builder
)

func gsdInstallSet(phase string) {
	gsdInstallMu.Lock()
	defer gsdInstallMu.Unlock()
	gsdInstallProgress_.Phase = phase
}

func gsdInstallAppendLog(line string) {
	gsdInstallMu.Lock()
	defer gsdInstallMu.Unlock()
	gsdInstallLogBuf.WriteString(line)
	gsdInstallLogBuf.WriteString("\n")
	// 防止 log 无限增长：超过 200KB 截掉前半
	if gsdInstallLogBuf.Len() > 200_000 {
		s := gsdInstallLogBuf.String()
		gsdInstallLogBuf.Reset()
		gsdInstallLogBuf.WriteString("...（已截掉前面，仅保留后半）\n")
		gsdInstallLogBuf.WriteString(s[len(s)/2:])
	}
	gsdInstallProgress_.Log = gsdInstallLogBuf.String()
}

func gsdInstallSnapshot() gsdInstallProgress {
	gsdInstallMu.Lock()
	defer gsdInstallMu.Unlock()
	p := gsdInstallProgress_
	if !p.StartedAt.IsZero() && p.Phase == "running" {
		p.ElapsedMS = time.Since(p.StartedAt).Milliseconds()
	} else if !p.EndedAt.IsZero() {
		p.ElapsedMS = p.EndedAt.Sub(p.StartedAt).Milliseconds()
	}
	return p
}

// GSD 框架的核心 skill 列表（用来判断"基本装好了"）
var gsdCoreSkills = []string{
	"gsd-new-project",
	"gsd-discuss-phase",
	"gsd-plan-phase",
	"gsd-execute-phase",
	"gsd-help",
	"gsd-update",
}

func handleGSDStatus(w http.ResponseWriter, r *http.Request) {
	home, _ := os.UserHomeDir()
	skillsDir := filepath.Join(home, ".claude", "skills")

	// 扫所有 gsd-* skill 目录
	var installedSkills []string
	if entries, err := os.ReadDir(skillsDir); err == nil {
		for _, e := range entries {
			if !e.IsDir() {
				continue
			}
			n := e.Name()
			if strings.HasPrefix(n, "gsd-") || n == "gsd" {
				installedSkills = append(installedSkills, n)
			}
		}
	}

	// 检测核心 skill 都在不在
	installedSet := map[string]bool{}
	for _, n := range installedSkills {
		installedSet[n] = true
	}
	coreInstalled := 0
	for _, c := range gsdCoreSkills {
		if installedSet[c] {
			coreInstalled++
		}
	}

	// 检测 npx / node
	npxAvailable := false
	npxPath := ""
	nodeVersion := ""
	if p, err := lookupCommand("npx"); err == nil {
		npxAvailable = true
		npxPath = p
	}
	if out, err := exec.Command("bash", "-lc", "node --version 2>/dev/null").Output(); err == nil {
		nodeVersion = strings.TrimSpace(string(out))
	}

	// 检测老版残留（我们 v1.7.32~38 的简版）
	hasOldSimple := false
	if _, err := os.Stat(filepath.Join(skillsDir, "gsd", "SKILL.md")); err == nil {
		// 读 SKILL.md 看是不是我们的简版（标题含"任务清单"）
		if b, err := os.ReadFile(filepath.Join(skillsDir, "gsd", "SKILL.md")); err == nil {
			if strings.Contains(string(b), "Getting Shit Done · 任务清单") ||
				strings.Contains(string(b), "name: gsd\nversion:") {
				hasOldSimple = true
			}
		}
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"installed_skills":  installedSkills,
		"installed_count":   len(installedSkills),
		"core_total":        len(gsdCoreSkills),
		"core_installed":    coreInstalled,
		"framework_ready":   coreInstalled >= len(gsdCoreSkills)-1, // 容忍 1 个缺失
		"npx_available":     npxAvailable,
		"npx_path":          npxPath,
		"node_version":      nodeVersion,
		"install_command":   "npx --yes get-shit-done-cc@latest --claude --global",
		"upstream":          "https://github.com/gsd-build/get-shit-done",
		"has_old_simple":    hasOldSimple,
	})
}

// POST /api/gsd/install
//
// 异步启动 npx 安装。立刻返回 {ok, already_running}，
// 前端通过 polling GET /api/gsd/install/progress 拿实时日志。
//
// 重要：用 userLoginShell()（zsh 优先 fallback bash）拿用户的 nvm/asdf 配置。
// 设 CI=1 等环境变量让 npx 完全非交互。
func handleGSDInstall(w http.ResponseWriter, r *http.Request) {
	// 不允许并发跑两次
	gsdInstallMu.Lock()
	if gsdInstallProgress_.Phase == "running" {
		gsdInstallMu.Unlock()
		writeJSON(w, http.StatusOK, map[string]any{
			"ok":              false,
			"already_running": true,
			"error":           "安装已在进行中，看实时日志面板",
		})
		return
	}
	// reset
	gsdInstallProgress_ = gsdInstallProgress{
		Phase:     "running",
		StartedAt: time.Now(),
	}
	gsdInstallLogBuf.Reset()
	gsdInstallMu.Unlock()

	// 检测 npx
	if _, err := lookupCommand("npx"); err != nil {
		gsdInstallMu.Lock()
		gsdInstallProgress_.Phase = "failed"
		gsdInstallProgress_.Error = "找不到 npx，请先 brew install node"
		gsdInstallProgress_.EndedAt = time.Now()
		gsdInstallMu.Unlock()
		gsdInstallAppendLog("✗ 找不到 npx，请先：brew install node")
		gsdInstallAppendLog("  或访问 https://nodejs.org")
		writeJSON(w, http.StatusOK, map[string]any{
			"ok":    false,
			"error": "找不到 npx",
		})
		return
	}

	go runGSDInstallAsync()

	writeJSON(w, http.StatusOK, map[string]any{
		"ok":    true,
		"phase": "running",
		"hint":  "polling /api/gsd/install/progress 拿实时日志",
	})
}

// runGSDInstallAsync 后台跑 npx，把 stdout/stderr 一行行实时写入 log buffer
func runGSDInstallAsync() {
	shell := userLoginShell()
	gsdInstallAppendLog("→ shell: " + shell)
	gsdInstallAppendLog("→ command: npx --yes get-shit-done-cc@latest --claude --global")
	gsdInstallAppendLog("")

	cmd := exec.Command(shell, "-lc",
		"npx --yes get-shit-done-cc@latest --claude --global 2>&1")

	// 设环境：禁交互、走系统代理（如有）
	env := os.Environ()
	env = append(env,
		"CI=1",                                     // 大多数 CLI 通过 $CI 知道是非交互模式
		"NPM_CONFIG_YES=true",                      // npm 强制 yes
		"NO_COLOR=1",                               // 关闭彩色输出避免 ANSI 干扰前端显示
		"FORCE_COLOR=0",
	)
	cmd.Env = env

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		gsdInstallAppendLog("✗ StdoutPipe 失败: " + err.Error())
		finishGSDInstall(-1, err.Error())
		return
	}
	cmd.Stderr = cmd.Stdout // stderr 走同一管道

	if err := cmd.Start(); err != nil {
		gsdInstallAppendLog("✗ 启动失败: " + err.Error())
		finishGSDInstall(-1, err.Error())
		return
	}

	// 实时读 stdout 一行行写 log
	scanner := bufio.NewScanner(stdout)
	scanner.Buffer(make([]byte, 64*1024), 1024*1024)
	go func() {
		for scanner.Scan() {
			gsdInstallAppendLog(scanner.Text())
		}
	}()

	// 自己控超时（8 分钟）
	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()
	select {
	case waitErr := <-done:
		exit := 0
		if cmd.ProcessState != nil {
			exit = cmd.ProcessState.ExitCode()
		}
		// 读完剩余 buffer
		io.Copy(io.Discard, stdout)
		errStr := ""
		if waitErr != nil {
			errStr = waitErr.Error()
		}
		finishGSDInstall(exit, errStr)
	case <-time.After(10 * time.Minute):
		_ = cmd.Process.Kill()
		gsdInstallAppendLog("")
		gsdInstallAppendLog("✗ 超时（10 分钟）— 已 kill 进程")
		gsdInstallAppendLog("  请在终端手工跑试试：")
		gsdInstallAppendLog("  " + shell + " -lc 'npx --yes get-shit-done-cc@latest --claude --global'")
		finishGSDInstall(-1, "timeout")
	}
}

func finishGSDInstall(exitCode int, errMsg string) {
	gsdInstallMu.Lock()
	gsdInstallProgress_.ExitCode = exitCode
	gsdInstallProgress_.EndedAt = time.Now()
	if exitCode == 0 {
		gsdInstallProgress_.Phase = "done"
	} else {
		gsdInstallProgress_.Phase = "failed"
		gsdInstallProgress_.Error = errMsg
	}
	gsdInstallMu.Unlock()
	gsdInstallAppendLog("")
	if exitCode == 0 {
		gsdInstallAppendLog("✓ 安装完成（exit 0）")
		gsdInstallAppendLog("  Cmd+Q 重启 Claude Code / Cursor 才能加载 gsd-* skill")
	} else {
		gsdInstallAppendLog("✗ 失败（exit " + strings.TrimSpace(errMsg) + "）")
	}
}

// GET /api/gsd/install/progress
//
// 返回实时进度 + log。前端 500ms polling 一次。
func handleGSDInstallProgress(w http.ResponseWriter, r *http.Request) {
	p := gsdInstallSnapshot()
	writeJSON(w, http.StatusOK, p)
}

// POST /api/gsd/uninstall
//
// 删除所有 ~/.claude/skills/gsd-* 目录 + 我们老版的 gsd/
//
// body: {include_old_simple?: bool} 默认 true
func handleGSDUninstall(w http.ResponseWriter, r *http.Request) {
	var req struct {
		IncludeOldSimple *bool `json:"include_old_simple,omitempty"`
	}
	_ = json.NewDecoder(r.Body).Decode(&req)
	rmOldSimple := true
	if req.IncludeOldSimple != nil {
		rmOldSimple = *req.IncludeOldSimple
	}

	home, _ := os.UserHomeDir()
	skillsDir := filepath.Join(home, ".claude", "skills")
	var removed []string
	var failed []string

	if entries, err := os.ReadDir(skillsDir); err == nil {
		for _, e := range entries {
			if !e.IsDir() {
				continue
			}
			n := e.Name()
			isFrameworkSkill := strings.HasPrefix(n, "gsd-")
			isOldSimple := (n == "gsd")
			if !isFrameworkSkill && !(isOldSimple && rmOldSimple) {
				continue
			}
			p := filepath.Join(skillsDir, n)
			if err := os.RemoveAll(p); err != nil {
				failed = append(failed, n+": "+err.Error())
			} else {
				removed = append(removed, n)
			}
		}
	}

	// Cursor 那边也清一下（如果存在）
	cursorDir := filepath.Join(home, ".cursor", "skills-cursor")
	if entries, err := os.ReadDir(cursorDir); err == nil {
		for _, e := range entries {
			if !e.IsDir() {
				continue
			}
			n := e.Name()
			if strings.HasPrefix(n, "gsd-") || (n == "gsd" && rmOldSimple) {
				p := filepath.Join(cursorDir, n)
				if err := os.RemoveAll(p); err == nil {
					removed = append(removed, "cursor/"+n)
				}
			}
		}
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"ok":       true,
		"removed":  removed,
		"failed":   failed,
		"note":     "个人数据（如 ~/Library/Application Support/TeamStandards/gsd/ 任务清单）未删，要彻底清手工 rm",
	})
}

// 兼容旧 API：返回内置 list 的某文件（v1.7.32~38 旧前端可能调）
// v1.7.39 简版 skill 已删，所有 list 返回 404
func handleGSDListGet(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"deprecated": true,
		"note":       "v1.7.39 起任务清单 GTD 简版已废弃；改用上游框架 https://github.com/gsd-build/get-shit-done",
	})
}

// ---------- helpers ----------

// userLoginShell 拿用户的 login shell（macOS 默认 zsh，老系统可能 bash）
// 重要：用户的 nvm/asdf/brew 之类配置在 ~/.zshrc，bash -lc 看不到
func userLoginShell() string {
	// 1. 试 $SHELL（最准——这是用户实际登录 shell）
	if s := os.Getenv("SHELL"); s != "" {
		if _, err := os.Stat(s); err == nil {
			return s
		}
	}
	// 2. 优先 /bin/zsh（macOS 10.15+ 默认）
	if _, err := os.Stat("/bin/zsh"); err == nil {
		return "/bin/zsh"
	}
	// 3. 兜底 bash
	return "/bin/bash"
}

// lookupCommand 通过 login shell 找命令绝对路径（兼容 nvm/asdf）
func lookupCommand(name string) (string, error) {
	out, err := exec.Command(userLoginShell(), "-lc", "command -v "+name+" 2>/dev/null").Output()
	if err != nil {
		return "", err
	}
	p := strings.TrimSpace(string(out))
	if p == "" {
		return "", os.ErrNotExist
	}
	return p, nil
}

func errString(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}
