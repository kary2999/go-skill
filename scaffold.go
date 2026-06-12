// 自动化脚手架：从 mask-go-sample-service 模板生成新微服务项目
//
// 流程：
//   1. cp -r 模板 → 目标目录
//   2. 解析模板 go.mod 的 module 路径
//   3. 全文本替换（不同层面的 name / module / proto package / port）
//   4. 重命名目录（api/order → api/<biz>，Dockerfile.order-svc → Dockerfile.<biz>-svc）
//   5. 可选：去示例业务代码，只留骨架
//   6. 可选：git init + go mod tidy
package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

type ScaffoldReq struct {
	ProjectName  string `json:"project_name"` // zoo-service
	Module       string `json:"module"`       // github.com/company/zoo-service
	Biz          string `json:"biz"`          // zoo
	HTTPPort     int    `json:"http_port"`
	GRPCPort     int    `json:"grpc_port"`
	Author       string `json:"author"`
	OutputDir    string `json:"output_dir"`
	SkeletonOnly bool   `json:"skeleton_only"`
	InitGit      bool   `json:"init_git"`
	RunTidy      bool   `json:"run_tidy"`

	// 模板来源（二选一）
	TemplateType   string `json:"template_type"` // "local" | "git"
	TemplatePath   string `json:"template_path"` // 本地模式：目录绝对路径
	TemplateGitURL string `json:"template_git_url"` // git 模式：git@host:xxx.git
	TemplateBranch string `json:"template_branch"`  // git 模式：可选分支名

	// 新项目的远端 Git 配置（仅 git 模式 & init_git=true 时有效）
	RemoteGitURL string `json:"remote_git_url"` // 新项目的 origin url
	GitLabAPIURL string `json:"gitlab_api_url"` // GitLab API base，如 https://your-git-server.com/api/v4
	GitLabToken  string `json:"gitlab_token"`   // GitLab PAT，留空则只设置本地 remote 不远端创建
	GitLabGroup  string `json:"gitlab_group"`   // 目标 group/namespace 路径（GitLab API 需要）

	// common-lib 组件选择（软开关：不勾就注释掉 yaml 对应段，代码不动）
	Components Components `json:"components"`

	// 生成 Skill 有效性测试用例（放到 .docs/skill-test-cases/，不进 git）
	GenSkillTestCases bool `json:"gen_skill_test_cases"`
}

// Components 记录用户勾选的 common-lib 组件
type Components struct {
	Postgres bool `json:"postgres"`
	Redis    bool `json:"redis"`
	Kafka    bool `json:"kafka"`
	Alarm    bool `json:"alarm"` // 依赖 Kafka
	Cron     bool `json:"cron"`  // 依赖 Postgres
}

// EnforceDeps 依赖联动：Alarm 需要 Kafka，Cron 需要 Postgres
func (c *Components) EnforceDeps() {
	if c.Alarm {
		c.Kafka = true
	}
	if c.Cron {
		c.Postgres = true
	}
}

type ScaffoldResp struct {
	OK       bool     `json:"ok"`
	DestPath string   `json:"dest_path,omitempty"`
	Logs     []string `json:"logs,omitempty"`
	Warnings []string `json:"warnings,omitempty"`
	Error    string   `json:"error,omitempty"`
}

var projectNameRe = regexp.MustCompile(`^[a-z][a-z0-9-]{1,40}$`)

// 默认模板路径
const defaultTemplatePath = ""

// ---------- handlers ----------

// handleScaffoldCheckTemplate 检查模板目录是否可用，返回它的 module 路径与 sample 服务名
func handleScaffoldCheckTemplate(w http.ResponseWriter, r *http.Request) {
	tpl := r.URL.Query().Get("path")
	if tpl == "" {
		tpl = defaultTemplatePath
	}
	info, err := os.Stat(tpl)
	if err != nil || !info.IsDir() {
		writeJSON(w, http.StatusOK, map[string]any{
			"ok":   false,
			"path": tpl,
			"msg":  "模板目录不存在，请修改模板路径",
		})
		return
	}
	goModPath := filepath.Join(tpl, "go.mod")
	modLine, err := readModuleLine(goModPath)
	if err != nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"ok":   false,
			"path": tpl,
			"msg":  "模板下没有有效 go.mod：" + err.Error(),
		})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"ok":     true,
		"path":   tpl,
		"module": modLine,
	})
}

// handleScaffoldDeriveRemote 根据模板 URL + 新项目名推导新项目的 git remote
// git@host:g/subg/mask-go-sample-service.git + zoo-service
//   → git@host:g/subg/zoo-service.git
func handleScaffoldDeriveRemote(w http.ResponseWriter, r *http.Request) {
	tpl := r.URL.Query().Get("template_url")
	name := r.URL.Query().Get("project_name")
	derived := deriveRemoteURL(tpl, name)
	writeJSON(w, http.StatusOK, map[string]string{"remote_url": derived})
}

func deriveRemoteURL(templateURL, projectName string) string {
	if templateURL == "" || projectName == "" {
		return ""
	}
	hasDotGit := strings.HasSuffix(templateURL, ".git")
	base := strings.TrimSuffix(templateURL, ".git")
	i := strings.LastIndexAny(base, "/:")
	if i < 0 {
		return ""
	}
	out := base[:i+1] + projectName
	if hasDotGit {
		out += ".git"
	}
	return out
}

// handleScaffoldPickFolder 调用 macOS 原生文件夹选择器。
// 支持 ?prompt=xxx 自定义提示文案（用于区分"输出目录"/"模板目录"）。
func handleScaffoldPickFolder(w http.ResponseWriter, r *http.Request) {
	prompt := r.URL.Query().Get("prompt")
	if prompt == "" {
		prompt = "选择项目输出目录"
	}
	// 防止 AppleScript 注入：只允许可打印 ASCII + 中日韩
	safe := make([]rune, 0, len(prompt))
	for _, rn := range prompt {
		if rn == '"' || rn == '\\' || rn == '\n' || rn == '\r' {
			continue
		}
		safe = append(safe, rn)
	}
	script := `POSIX path of (choose folder with prompt "` + string(safe) + `")`
	cmd := exec.Command("osascript", "-e", script)
	out, err := cmd.Output()
	if err != nil {
		// 用户取消 / 系统异常都会走这里，返回空路径即可
		writeJSON(w, http.StatusOK, map[string]string{"path": ""})
		return
	}
	path := strings.TrimSpace(string(out))
	path = strings.TrimSuffix(path, "/")
	writeJSON(w, http.StatusOK, map[string]string{"path": path})
}

// handleScaffoldCreate 核心：生成项目
func handleScaffoldCreate(w http.ResponseWriter, r *http.Request) {
	var req ScaffoldReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, ScaffoldResp{Error: "请求解析失败：" + err.Error()})
		return
	}

	resp := &ScaffoldResp{}
	if err := runScaffold(&req, resp); err != nil {
		resp.OK = false
		if resp.Error == "" {
			resp.Error = err.Error()
		}
		writeJSON(w, http.StatusOK, resp)
		return
	}
	resp.OK = true
	writeJSON(w, http.StatusOK, resp)
}

// ---------- 核心逻辑 ----------

func runScaffold(req *ScaffoldReq, resp *ScaffoldResp) error {
	// 0. 默认值 & 校验
	if req.TemplateType == "" {
		req.TemplateType = "local"
	}
	if req.TemplateType == "local" && req.TemplatePath == "" {
		req.TemplatePath = defaultTemplatePath
	}
	if req.HTTPPort == 0 {
		req.HTTPPort = 8000
	}
	if req.GRPCPort == 0 {
		req.GRPCPort = 9000
	}
	if !projectNameRe.MatchString(req.ProjectName) {
		return errors.New("项目名格式错误：需要 kebab-case，小写字母开头，2-41 字符")
	}
	if req.Module == "" {
		req.Module = "github.com/company/" + req.ProjectName
	}
	if req.Biz == "" {
		req.Biz = deriveBiz(req.ProjectName)
	}
	if !isAlphaNumSimple(req.Biz) {
		return errors.New("业务域格式错误：只能用小写字母和数字")
	}
	if req.OutputDir == "" {
		return errors.New("请选择输出目录")
	}
	if _, err := os.Stat(req.OutputDir); err != nil {
		return fmt.Errorf("输出目录不存在：%s", req.OutputDir)
	}

	dest := filepath.Join(req.OutputDir, req.ProjectName)
	if _, err := os.Stat(dest); err == nil {
		return fmt.Errorf("目标目录已存在，拒绝覆盖：%s", dest)
	}
	resp.DestPath = dest

	// 1. 模板来源：git clone 或本地
	templateDir := req.TemplatePath
	if req.TemplateType == "git" {
		if req.TemplateGitURL == "" {
			return errors.New("git 模式下请填写模板 Git URL")
		}
		resp.Logs = append(resp.Logs, "→ git clone 模板：" + req.TemplateGitURL)
		tmp, cleanup, err := cloneTemplate(req.TemplateGitURL, req.TemplateBranch)
		if err != nil {
			return fmt.Errorf("克隆模板失败：%v", err)
		}
		defer cleanup()
		templateDir = tmp
		resp.Logs = append(resp.Logs, "  ✓ 已克隆到临时目录（用完自动清理）")
	} else {
		if _, err := os.Stat(templateDir); err != nil {
			return fmt.Errorf("模板目录不存在：%s", templateDir)
		}
	}

	templateModule, err := readModuleLine(filepath.Join(templateDir, "go.mod"))
	if err != nil {
		return fmt.Errorf("读取模板 go.mod 失败：%v", err)
	}
	resp.Logs = append(resp.Logs, "模板 module = "+templateModule)
	// 覆盖下面逻辑中使用的 TemplatePath
	req.TemplatePath = templateDir

	// 2. 复制整棵树（排除 .git / .idea / vendor）
	resp.Logs = append(resp.Logs, "→ 复制模板到 "+dest)
	if err := copyTreeFiltered(req.TemplatePath, dest); err != nil {
		return fmt.Errorf("复制模板失败：%v", err)
	}

	// 3. 准备替换表
	sampleName := filepath.Base(req.TemplatePath) // mask-go-sample-service
	sampleShort := trimServiceSuffix(sampleName)  // mask-go-sample
	sampleBiz := detectSampleBiz(dest)            // 大概率是 "order"
	resp.Logs = append(resp.Logs, fmt.Sprintf("探测到示例业务域 = %s", sampleBiz))

	newShort := trimServiceSuffix(req.ProjectName)

	replaces := []replacement{
		{templateModule, req.Module},
		{sampleName, req.ProjectName},
		{sampleShort, newShort},
		{sampleBiz + "-service", req.ProjectName},
		{sampleBiz + "-svc", newShort + "-svc"},
		{"api/" + sampleBiz + "/", "api/" + req.Biz + "/"},
		{"clawnova." + sampleBiz, "clawnova." + req.Biz},
		{`Name = "` + sampleBiz + `-service"`, `Name = "` + req.ProjectName + `"`},
		// v1.4.1 补：Makefile / yaml 里的配置文件引用
		{"configs/dev/" + sampleBiz + ".yaml", "configs/dev/" + req.Biz + ".yaml"},
		{"configs/dev/" + sampleBiz + ".yml", "configs/dev/" + req.Biz + ".yml"},
	}

	// 4. 全文件内容替换
	resp.Logs = append(resp.Logs, "→ 全文本替换")
	if n, err := walkReplaceText(dest, replaces); err != nil {
		return fmt.Errorf("文本替换失败：%v", err)
	} else {
		resp.Logs = append(resp.Logs, fmt.Sprintf("  已处理 %d 个文件", n))
	}

	// 5. 重命名目录（api/order → api/<biz>）和含 biz 的文件名
	resp.Logs = append(resp.Logs, "→ 重命名含业务域的目录和文件")
	renamed, err := renameBizPaths(dest, sampleBiz, req.Biz, newShort)
	if err != nil {
		return fmt.Errorf("重命名失败：%v", err)
	}
	for _, r := range renamed {
		resp.Logs = append(resp.Logs, "  "+r)
	}

	// 6. 端口替换（在 configs/ 下）
	if err := replacePortsInConfigs(dest, req.HTTPPort, req.GRPCPort, resp); err != nil {
		resp.Warnings = append(resp.Warnings, "端口替换失败（可手动修改）："+err.Error())
	}

	// 6.5 v1.4.1 新增：重写 buf.gen.yaml 为已知可用版（模板原始版有 3 个 bug 导致 make proto 跑不通）
	if err := rewriteBufGenYaml(dest, req.Module); err != nil {
		resp.Warnings = append(resp.Warnings, "buf.gen.yaml 重写失败："+err.Error())
	} else {
		resp.Logs = append(resp.Logs, "✓ buf.gen.yaml 已规范化（out: api + /api prefix + disable googleapis override）")
	}

	// 6.6 v1.4.1 新增：删除模板里预生成的 *.pb.go（已被文本替换污染，运行时会 panic）
	// make proto 会重新生成
	if removed := removeStaleProtoGenFiles(dest); removed > 0 {
		resp.Logs = append(resp.Logs, fmt.Sprintf("✓ 清理 %d 个模板残留 .pb.go（make proto 会重生）", removed))
	}

	// 6.7 v1.5.0 新增：根据用户勾选的 common-lib 组件调整 yaml
	req.Components.EnforceDeps()
	yamlPath := filepath.Join(dest, "configs", "dev", req.Biz+".yaml")
	if logs, err := applyComponentsToYAML(yamlPath, &req.Components, req); err != nil {
		resp.Warnings = append(resp.Warnings, "组件 yaml 调整失败："+err.Error())
	} else {
		for _, l := range logs {
			resp.Logs = append(resp.Logs, "  "+l)
		}
	}

	// 7. 可选：只留骨架
	if req.SkeletonOnly {
		resp.Logs = append(resp.Logs, "→ 清理示例业务代码（只留骨架）")
		if err := stripExampleCode(dest, req.Biz); err != nil {
			resp.Warnings = append(resp.Warnings, "清理骨架失败："+err.Error())
		}
	}

	// 8. 清除模板的 git 历史
	_ = os.RemoveAll(filepath.Join(dest, ".git"))
	_ = os.RemoveAll(filepath.Join(dest, ".idea"))

	// 8.5 可选：生成 Skill 有效性测试用例到 .docs/skill-test-cases/（不进 git）
	if req.GenSkillTestCases {
		if logs, err := writeSkillTestCases(dest); err != nil {
			resp.Warnings = append(resp.Warnings, "写入 skill 测试用例失败："+err.Error())
		} else {
			for _, l := range logs {
				resp.Logs = append(resp.Logs, "  "+l)
			}
		}
	}

	// 9. 可选：git init + 首次提交 + 设置远端 + (可选) GitLab 自动建仓
	if req.InitGit {
		if out, err := runIn(dest, "git", "init", "-b", "main"); err != nil {
			// 老版本 git 不支持 -b，回退
			if out2, err2 := runIn(dest, "git", "init"); err2 != nil {
				resp.Warnings = append(resp.Warnings, "git init 失败："+err2.Error()+" "+out2)
			}
			_ = out
		}
		resp.Logs = append(resp.Logs, "✓ git init 完成")

		if _, err := runIn(dest, "git", "add", "."); err == nil {
			// commit 可能失败（比如没配 user.email），吞掉 warning
			if out, err := runIn(dest, "git", "commit", "-m",
				"chore: initial scaffold from mask-go-sample-service"); err != nil {
				resp.Warnings = append(resp.Warnings,
					"git commit 失败（可能未配置 user.name/email）："+trimStr(out, 200))
			} else {
				resp.Logs = append(resp.Logs, "✓ 初始化提交完成")
			}
		}

		// 设置 origin（如果 git 模式 且提供了新 remote URL）
		if req.RemoteGitURL != "" {
			if out, err := runIn(dest, "git", "remote", "add", "origin", req.RemoteGitURL); err != nil {
				resp.Warnings = append(resp.Warnings,
					"git remote add 失败："+err.Error()+" "+trimStr(out, 200))
			} else {
				resp.Logs = append(resp.Logs, "✓ git remote add origin "+req.RemoteGitURL)
			}

			// 可选：用 GitLab API 远端建仓
			if req.GitLabToken != "" && req.GitLabAPIURL != "" {
				if err := createGitLabRepo(req, resp); err != nil {
					resp.Warnings = append(resp.Warnings, "GitLab 远端建仓失败："+err.Error())
				} else {
					resp.Logs = append(resp.Logs, "✓ 已在 GitLab 远端创建空仓库")
				}
			} else {
				resp.Warnings = append(resp.Warnings,
					"未提供 GitLab Token，请手动在 GitLab 创建同名空仓库，然后："+
						"cd "+dest+" && git push -u origin main")
			}
		}
	}

	// 10. 可选：go mod tidy
	if req.RunTidy {
		resp.Logs = append(resp.Logs, "→ 跑 go mod tidy（可能需要几十秒）")
		if out, err := runIn(dest, "go", "mod", "tidy"); err != nil {
			resp.Warnings = append(resp.Warnings, "go mod tidy 失败："+err.Error()+" "+trimStr(out, 500))
		} else {
			resp.Logs = append(resp.Logs, "✓ go mod tidy 完成")
		}
	}

	// 11. 写一份"下一步"指南到项目根
	writeNextSteps(dest, req)
	resp.Logs = append(resp.Logs, "✓ 生成 下一步.md")

	return nil
}

// ---------- 辅助函数 ----------

func readModuleLine(goMod string) (string, error) {
	b, err := os.ReadFile(goMod)
	if err != nil {
		return "", err
	}
	for _, line := range strings.Split(string(b), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "module ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "module")), nil
		}
	}
	return "", errors.New("no module line")
}

func trimServiceSuffix(s string) string {
	s = strings.TrimSuffix(s, "-service")
	s = strings.TrimSuffix(s, "-svc")
	return s
}

func deriveBiz(projectName string) string {
	// zoo-service → zoo
	b := trimServiceSuffix(projectName)
	// 去掉剩余的 - _
	b = strings.ReplaceAll(b, "-", "")
	b = strings.ReplaceAll(b, "_", "")
	return b
}

func isAlphaNumSimple(s string) bool {
	if s == "" {
		return false
	}
	for _, c := range s {
		if !((c >= 'a' && c <= 'z') || (c >= '0' && c <= '9')) {
			return false
		}
	}
	return true
}

// detectSampleBiz 从模板里探测示例业务域（通常是 "order"）
// 优先找 api/<name>/ 子目录
func detectSampleBiz(dest string) string {
	apiDir := filepath.Join(dest, "api")
	entries, err := os.ReadDir(apiDir)
	if err == nil {
		for _, e := range entries {
			if e.IsDir() {
				return e.Name()
			}
		}
	}
	return "order" // fallback
}

type replacement struct {
	old, new string
}

// walkReplaceText 遍历 dest 下所有文本文件，逐一 replace
func walkReplaceText(dest string, replaces []replacement) (int, error) {
	count := 0
	err := filepath.Walk(dest, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			name := info.Name()
			if name == ".git" || name == ".idea" || name == "vendor" || name == "node_modules" {
				return filepath.SkipDir
			}
			return nil
		}
		if !isTextFile(p) {
			return nil
		}
		b, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		orig := string(b)
		modified := orig
		for _, r := range replaces {
			if r.old != "" && r.old != r.new {
				modified = strings.ReplaceAll(modified, r.old, r.new)
			}
		}
		if modified != orig {
			if err := os.WriteFile(p, []byte(modified), info.Mode().Perm()); err != nil {
				return err
			}
			count++
		}
		return nil
	})
	return count, err
}

var textExts = map[string]bool{
	".go": true, ".mod": true, ".sum": true, ".yaml": true, ".yml": true,
	".proto": true, ".sql": true, ".sh": true, ".md": true, ".txt": true,
	".json": true, ".toml": true, ".env": true, ".gitignore": true,
	".dockerignore": true, ".editorconfig": true,
}

var textFileBasenames = map[string]bool{
	"Makefile": true, "Dockerfile": true, "go.mod": true, "go.sum": true,
	".gitignore": true, ".dockerignore": true,
}

func isTextFile(p string) bool {
	name := filepath.Base(p)
	if textFileBasenames[name] {
		return true
	}
	// Dockerfile.xxx / Makefile.xxx 也算
	if strings.HasPrefix(name, "Dockerfile") || strings.HasPrefix(name, "Makefile") {
		return true
	}
	ext := strings.ToLower(filepath.Ext(name))
	return textExts[ext]
}

// renameBizPaths：api/<sampleBiz>/ → api/<newBiz>/，文件名里含 sampleBiz 的也重命名
func renameBizPaths(dest, sampleBiz, newBiz, newShort string) ([]string, error) {
	var logs []string

	// 1. 目录重命名：api/<sampleBiz> → api/<newBiz>
	oldAPI := filepath.Join(dest, "api", sampleBiz)
	newAPI := filepath.Join(dest, "api", newBiz)
	if _, err := os.Stat(oldAPI); err == nil && sampleBiz != newBiz {
		if err := os.Rename(oldAPI, newAPI); err != nil {
			return logs, fmt.Errorf("rename api: %w", err)
		}
		logs = append(logs, fmt.Sprintf("api/%s → api/%s", sampleBiz, newBiz))
	}

	// 2. 文件名批量：含 <sampleBiz> 的 .go / .proto / .sql / Dockerfile.*
	_ = filepath.Walk(dest, func(p string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return err
		}
		name := info.Name()
		newName := name
		// Dockerfile.order-svc → Dockerfile.<short>-svc
		if strings.HasPrefix(name, "Dockerfile.") {
			newName = strings.Replace(name, sampleBiz+"-svc", newShort+"-svc", 1)
			newName = strings.Replace(newName, sampleBiz+"-service", newShort+"-service", 1)
		} else {
			// order.go / order_repo.go / order.proto / order.yaml 等
			// 只替换独立的 biz 词（避免把 "order_form" 改错）
			// 简化：仅当文件名 == sampleBiz+"."+ext 或 以 sampleBiz+"_" 开头时替换
			if strings.HasPrefix(name, sampleBiz+".") {
				newName = newBiz + name[len(sampleBiz):]
			} else if strings.HasPrefix(name, sampleBiz+"_") {
				newName = newBiz + name[len(sampleBiz):]
			}
		}
		if newName != name {
			newPath := filepath.Join(filepath.Dir(p), newName)
			if rerr := os.Rename(p, newPath); rerr == nil {
				rel, _ := filepath.Rel(dest, p)
				logs = append(logs, rel+" → "+newName)
			}
		}
		return nil
	})
	return logs, nil
}

// replacePortsInConfigs 修改 configs/**/*.yaml 里的端口
// 方式：逐行扫描，记住当前所在的 section (http / grpc)，遇到 addr: 行就换里面的端口
// 这样无论 sample 原端口是 8000 / 8002 / 其他任意值都能正确替换
func replacePortsInConfigs(dest string, httpPort, grpcPort int, resp *ScaffoldResp) error {
	configsDir := filepath.Join(dest, "configs")
	if _, err := os.Stat(configsDir); err != nil {
		return nil
	}
	addrLineRe := regexp.MustCompile(`^(\s*addr:\s*[^:\n]+:)(\d+)(.*)$`)
	portLineRe := regexp.MustCompile(`^(\s*port:\s*)(\d+)(.*)$`)

	return filepath.Walk(configsDir, func(p string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return err
		}
		if !strings.HasSuffix(p, ".yaml") && !strings.HasSuffix(p, ".yml") {
			return nil
		}
		b, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		lines := strings.Split(string(b), "\n")
		section := ""
		sectionIndent := -1
		for i, line := range lines {
			trimmed := strings.TrimSpace(line)
			indent := len(line) - len(strings.TrimLeft(line, " \t"))

			// 进入 http: / grpc: section
			if trimmed == "http:" {
				section = "http"
				sectionIndent = indent
				continue
			}
			if trimmed == "grpc:" {
				section = "grpc"
				sectionIndent = indent
				continue
			}
			// 退出 section：遇到同级或更浅的非空行
			if section != "" && trimmed != "" && !strings.HasPrefix(trimmed, "#") && indent <= sectionIndent {
				section = ""
			}
			if section == "" {
				continue
			}

			// 在 http/grpc section 内，替换 addr: 或 port:
			var targetPort int
			if section == "http" {
				targetPort = httpPort
			} else {
				targetPort = grpcPort
			}
			if m := addrLineRe.FindStringSubmatch(line); m != nil {
				lines[i] = m[1] + fmt.Sprintf("%d", targetPort) + m[3]
				section = "" // 一段一次
			} else if m := portLineRe.FindStringSubmatch(line); m != nil {
				lines[i] = m[1] + fmt.Sprintf("%d", targetPort) + m[3]
				section = ""
			}
		}
		return os.WriteFile(p, []byte(strings.Join(lines, "\n")), info.Mode().Perm())
	})
}

// rewriteBufGenYaml 用已知可用版覆写 buf.gen.yaml
// 原 sample 的 buf.gen.yaml 有三个问题（见 下一步-修订说明.md Step 3 前置修正 B）：
//   1. plugins[*].out: . → 应为 out: api
//   2. go_package_prefix 缺 /api 后缀 → 生成的 import path 和业务代码对不上
//   3. 没有 managed.disable googleapis → 会把 annotations.proto 也改写成不存在的包
func rewriteBufGenYaml(dest, module string) error {
	path := filepath.Join(dest, "buf.gen.yaml")
	// 文件不存在就跳过（非致命）
	if _, err := os.Stat(path); err != nil {
		return nil
	}
	content := fmt.Sprintf(`version: v2

managed:
  enabled: true
  # 不覆盖外部依赖（googleapis 等）的 go_package，避免把 annotations.proto 的 import
  # 改写成不存在的包路径（会导致 "no Go files in ..." 编译错误）
  disable:
    - file_option: go_package_prefix
      module: buf.build/googleapis/googleapis
  override:
    - file_option: go_package_prefix
      value: %s/api

plugins:
  - local: protoc-gen-go
    out: api
    opt: [paths=source_relative]
  - local: protoc-gen-go-grpc
    out: api
    opt: [paths=source_relative]
  - local: protoc-gen-go-http
    out: api
    opt: [paths=source_relative]
`, module)
	return os.WriteFile(path, []byte(content), 0644)
}

// writeComponentTutorials 按勾选的组件往 sb 追加"傻瓜式验证"章节
// 每个组件：docker 一键启动 + 最快能跑通的 curl / 验证命令 + 指向对应 demo 路径
func writeComponentTutorials(sb *strings.Builder, req *ScaffoldReq) {
	c := &req.Components
	biz := req.Biz
	sb.WriteString("---\n\n")
	sb.WriteString("## 🧩 组件快速验证（按你勾选的 common-lib 组件）\n\n")
	sb.WriteString("每个组件给出：**docker 一键启动** + **最少命令跑通**。AI 可以参考 `~/.claude/skills/go-team-standards/demos/` 给你加对应业务代码。\n\n")

	if c.Postgres {
		sb.WriteString("### 🐘 PostgreSQL\n\n")
		sb.WriteString("**启动**：\n")
		sb.WriteString("```bash\n")
		sb.WriteString("docker run -d --name " + biz + "-pg \\\n")
		sb.WriteString("  -e POSTGRES_USER=postgres -e POSTGRES_PASSWORD=postgres \\\n")
		sb.WriteString("  -e POSTGRES_DB=" + biz + " \\\n")
		sb.WriteString("  -p 5432:5432 postgres:16\n")
		sb.WriteString("```\n\n")
		sb.WriteString("**配置**：`configs/dev/" + biz + ".yaml` 的 `data.database.source` 改为 `host=127.0.0.1 port=5432 user=postgres password=postgres dbname=" + biz + " sslmode=disable`\n\n")
		sb.WriteString("**验证**（启动服务后）：\n")
		sb.WriteString("```bash\n")
		sb.WriteString(fmt.Sprintf("curl -X POST http://localhost:%d/api/v1/order \\\n", req.HTTPPort))
		sb.WriteString("  -H 'Content-Type: application/json' \\\n")
		sb.WriteString("  -d '{\"user_id\":\"u1\",\"symbol\":\"BTC/USDT\",\"side\":\"buy\",\"amount\":\"0.1\",\"price\":\"60000\"}'\n\n")
		sb.WriteString("# 进 pg 看记录\n")
		sb.WriteString("docker exec -it " + biz + "-pg psql -U postgres -d " + biz + " -c 'SELECT * FROM orders;'\n")
		sb.WriteString("```\n\n")
	}

	if c.Redis {
		sb.WriteString("### 🔴 Redis\n\n")
		sb.WriteString("**启动**：\n")
		sb.WriteString("```bash\n")
		sb.WriteString("docker run -d --name " + biz + "-redis -p 6379:6379 redis:7-alpine\n")
		sb.WriteString("```\n\n")
		sb.WriteString("**示例代码**：现在 Order 业务**没有**使用 Redis。要加 demo：\n")
		sb.WriteString("```\n")
		sb.WriteString("在 Claude Code / Cursor 里说：\"给 internal/biz/order.go 加一个 Redis 幂等检查\"\n")
		sb.WriteString("AI 会参考 demos/redis-idempotency.go 帮你接入\n")
		sb.WriteString("```\n\n")
		sb.WriteString("**命令行验证**：\n")
		sb.WriteString("```bash\n")
		sb.WriteString("docker exec -it " + biz + "-redis redis-cli\n")
		sb.WriteString("> SET " + biz + ":test:1 hello\n")
		sb.WriteString("> GET " + biz + ":test:1\n")
		sb.WriteString("```\n\n")
	}

	if c.Kafka {
		sb.WriteString("### 📬 Kafka\n\n")
		sb.WriteString("**启动**（单节点 KRaft 模式，无需 zookeeper）：\n")
		sb.WriteString("```bash\n")
		sb.WriteString("docker run -d --name " + biz + "-kafka \\\n")
		sb.WriteString("  -p 9092:9092 \\\n")
		sb.WriteString("  -e KAFKA_CFG_NODE_ID=0 \\\n")
		sb.WriteString("  -e KAFKA_CFG_PROCESS_ROLES=controller,broker \\\n")
		sb.WriteString("  -e KAFKA_CFG_LISTENERS=PLAINTEXT://:9092,CONTROLLER://:9093 \\\n")
		sb.WriteString("  -e KAFKA_CFG_ADVERTISED_LISTENERS=PLAINTEXT://127.0.0.1:9092 \\\n")
		sb.WriteString("  -e KAFKA_CFG_CONTROLLER_QUORUM_VOTERS=0@localhost:9093 \\\n")
		sb.WriteString("  -e KAFKA_CFG_CONTROLLER_LISTENER_NAMES=CONTROLLER \\\n")
		sb.WriteString("  -e KAFKA_CFG_LISTENER_SECURITY_PROTOCOL_MAP=CONTROLLER:PLAINTEXT,PLAINTEXT:PLAINTEXT \\\n")
		sb.WriteString("  bitnami/kafka:3.7\n")
		sb.WriteString("```\n\n")
		sb.WriteString("**创建测试 topic**：\n")
		sb.WriteString("```bash\n")
		sb.WriteString("docker exec " + biz + "-kafka kafka-topics.sh \\\n")
		sb.WriteString("  --bootstrap-server localhost:9092 \\\n")
		sb.WriteString("  --create --topic dev_" + biz + "_hello_created \\\n")
		sb.WriteString("  --partitions 1 --replication-factor 1\n")
		sb.WriteString("```\n\n")
		sb.WriteString("**示例代码**：当前模板没有 Kafka producer/consumer。加 demo：\n")
		sb.WriteString("```\n")
		sb.WriteString("在 Cursor 里说：\"帮我在 internal/data 加一个 Kafka 生产者，发 hello_created 事件\"\n")
		sb.WriteString("AI 会参考 demos/kafka-producer.go + demos/kafka-consumer.go\n")
		sb.WriteString("```\n\n")
	}

	if c.Alarm {
		sb.WriteString("### 🚨 Alarm（依赖 Kafka）\n\n")
		sb.WriteString("当前模板内置 alarm daemon（`internal/alert/`），配置启用后，收到 `prod_" + biz + "_alarm` topic 的消息会渲染成 Telegram 告警（需填 bot_token / chat_id）。\n\n")
		sb.WriteString("**验证**：服务起来后，手工塞一条告警消息到 Kafka：\n")
		sb.WriteString("```bash\n")
		sb.WriteString("docker exec -it " + biz + "-kafka kafka-console-producer.sh \\\n")
		sb.WriteString("  --bootstrap-server localhost:9092 \\\n")
		sb.WriteString("  --topic prod_" + biz + "_alarm\n")
		sb.WriteString("# 粘 JSON: {\"level\":\"warn\",\"msg\":\"test alarm\"}\n")
		sb.WriteString("```\n\n")
		sb.WriteString("看服务日志是否输出 \"alarm received\" / 对应的 Telegram 发送（有 token 时）。\n\n")
	}

	if c.Cron {
		sb.WriteString("### ⏰ Cron（依赖 PostgreSQL）\n\n")
		sb.WriteString("基于 asynq + `cron_tasks` 表管理定时任务。Sample 自带 `internal/job/demo.go` 作为骨架。\n\n")
		sb.WriteString("**新增一个任务**（在 `cmd/main.go` 或 `internal/job/` 里注册 handler）：\n")
		sb.WriteString("```go\n")
		sb.WriteString("// 每分钟打印一次日志的 demo\n")
		sb.WriteString("job.RegisterHandler(\"hello-cron\", func(ctx context.Context, task *Task) error {\n")
		sb.WriteString("    slog.InfoContext(ctx, \"hello from cron\", \"ts\", time.Now())\n")
		sb.WriteString("    return nil\n")
		sb.WriteString("})\n")
		sb.WriteString("```\n\n")
		sb.WriteString("然后在 `cron_tasks` 表插入记录（`cron_expr=* * * * *` + `handler=hello-cron`）。\n\n")
	}

	if !c.Postgres && !c.Redis && !c.Kafka && !c.Alarm && !c.Cron {
		sb.WriteString("（未勾选任何 common-lib 组件，`make run` 起的是最小服务，只有 `/health` / `/livez` / `/readyz` / `/metrics` 四个探针。）\n\n")
	}
}

// applyComponentsToYAML 按用户勾选把 yaml 里对应段激活（去注释）或追加
// 勾 → uncomment 原模板里注释的段（PG / Redis）；勾 Kafka / Alarm → 追加对应段（sample yaml 默认没有）
// 不勾 → 保持注释态；不勾 Kafka / Alarm → 不追加
func applyComponentsToYAML(yamlPath string, c *Components, req *ScaffoldReq) ([]string, error) {
	b, err := os.ReadFile(yamlPath)
	if err != nil {
		return nil, err
	}
	content := string(b)
	var logs []string

	// PG：uncomment `# database:` 段（注意 sample 里 key 可能是 database 或 databases）
	if c.Postgres {
		before := content
		content = uncommentYAMLSection(content, "database")
		content = uncommentYAMLSection(content, "databases")
		if content != before {
			logs = append(logs, "✓ PostgreSQL 段已激活（data.database / databases）")
		}
	} else {
		logs = append(logs, "○ PostgreSQL 未启用（yaml 保持注释，NewData 会 skip）")
	}

	// Redis
	if c.Redis {
		before := content
		content = uncommentYAMLSection(content, "redis")
		content = uncommentYAMLSection(content, "redises")
		if content != before {
			logs = append(logs, "✓ Redis 段已激活")
		}
	} else {
		logs = append(logs, "○ Redis 未启用")
	}

	// Kafka：默认 sample 没 mq 段，追加
	if c.Kafka && !strings.Contains(content, "\nmq:") && !regexp.MustCompile(`(?m)^mq:`).MatchString(content) {
		block := fmt.Sprintf("\n# MQ (Kafka) —— 由 Team Standards 脚手架添加\nmq:\n  kafka:\n    brokers:\n      - 127.0.0.1:9092\n    client_id: %q\n", req.ProjectName)
		content = strings.TrimRight(content, "\n") + "\n" + block
		logs = append(logs, "✓ Kafka (mq.kafka) 段已追加")
	} else if !c.Kafka {
		logs = append(logs, "○ Kafka 未启用")
	}

	// Alarm
	if c.Alarm && !strings.Contains(content, "\nalarm:") && !regexp.MustCompile(`(?m)^alarm:`).MatchString(content) {
		block := fmt.Sprintf("\n# Alarm —— 由 Team Standards 脚手架添加\nalarm:\n  enabled: true\n  topic: %q\n  consumer_group: %q\n  # telegram_api_base: https://api.telegram.org\n  # telegram_bot_token: \"\"\n  # telegram_chat_id: \"\"\n",
			"prod_"+req.Biz+"_alarm", req.Biz+"_alarm_consumer")
		content = strings.TrimRight(content, "\n") + "\n" + block
		logs = append(logs, "✓ Alarm 段已追加（依赖 Kafka）")
	} else if !c.Alarm {
		logs = append(logs, "○ Alarm 未启用")
	}

	// Cron 走 DB 表，不在 yaml 里配置
	if c.Cron {
		logs = append(logs, "✓ Cron 已启用（通过 PostgreSQL 的 CronTaskRecord 表管理）")
	} else {
		logs = append(logs, "○ Cron 未启用")
	}

	if err := os.WriteFile(yamlPath, []byte(content), 0644); err != nil {
		return logs, err
	}
	return logs, nil
}

// uncommentYAMLSection 取消某个 yaml section 的注释
// 把形如 "  # database:\n  #   driver: pgx\n  #   source: ..." 整段还原为未注释
func uncommentYAMLSection(content, sectionName string) string {
	lines := strings.Split(content, "\n")
	headerRe := regexp.MustCompile(`^(\s*)#\s*` + regexp.QuoteMeta(sectionName) + `:\s*$`)
	inSection := false
	sectionIndentLen := -1

	for i, line := range lines {
		if !inSection {
			if m := headerRe.FindStringSubmatch(line); m != nil {
				inSection = true
				sectionIndentLen = len(m[1])
				lines[i] = m[1] + sectionName + ":"
			}
			continue
		}
		// 在 section 内：期望是以 `(sectionIndentLen+空格) #` 开头的注释行
		trimmed := strings.TrimLeft(line, " \t")
		currentIndentLen := len(line) - len(trimmed)
		if trimmed == "" {
			continue // 空行不影响段
		}
		if !strings.HasPrefix(trimmed, "#") {
			inSection = false
			continue
		}
		if currentIndentLen <= sectionIndentLen {
			inSection = false
			continue
		}
		// 去掉首个 `# ` 或 `#`
		uncommented := regexp.MustCompile(`^(\s*)#\s?`).ReplaceAllString(line, "$1")
		lines[i] = uncommented
	}
	return strings.Join(lines, "\n")
}

// removeStaleProtoGenFiles 删除 api/ 下预生成的 *.pb.go
// 原因：模板里的 *.pb.go 含 protobuf 描述符字节串（含长度前缀），
// 文本替换把字符串长度改了，字节串没改 → 运行时 slice bounds panic
// 删掉让 make proto 重新生成即可
func removeStaleProtoGenFiles(dest string) int {
	apiDir := filepath.Join(dest, "api")
	if _, err := os.Stat(apiDir); err != nil {
		return 0
	}
	count := 0
	_ = filepath.Walk(apiDir, func(p string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		name := filepath.Base(p)
		if strings.HasSuffix(name, ".pb.go") {
			if os.Remove(p) == nil {
				count++
			}
		}
		return nil
	})
	return count
}

// stripExampleCode 删除示例业务代码（只留骨架）
func stripExampleCode(dest, biz string) error {
	candidates := []string{
		filepath.Join(dest, "internal", "biz", biz+".go"),
		filepath.Join(dest, "internal", "data", biz+"_repo.go"),
		filepath.Join(dest, "internal", "service", biz+".go"),
	}
	for _, p := range candidates {
		_ = os.Remove(p)
	}
	return nil
}

// copyTreeFiltered 递归复制（跳过 .git / .idea / vendor）
func copyTreeFiltered(src, dst string) error {
	return filepath.Walk(src, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(src, p)
		if rel == "." {
			return os.MkdirAll(dst, info.Mode().Perm())
		}
		// 跳过不必要目录
		name := info.Name()
		if info.IsDir() {
			if name == ".git" || name == ".idea" || name == "vendor" || name == "node_modules" {
				return filepath.SkipDir
			}
			return os.MkdirAll(filepath.Join(dst, rel), info.Mode().Perm())
		}
		if info.Mode()&os.ModeSymlink != 0 {
			target, err := os.Readlink(p)
			if err != nil {
				return err
			}
			return os.Symlink(target, filepath.Join(dst, rel))
		}
		// 普通文件
		return copyFile(p, filepath.Join(dst, rel), info.Mode().Perm())
	})
}

func copyFile(src, dst string, mode os.FileMode) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}
	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, mode)
	if err != nil {
		return err
	}
	if _, err := io.Copy(out, in); err != nil {
		out.Close()
		return err
	}
	return out.Close()
}

func runIn(dir, name string, args ...string) (string, error) {
	start := time.Now()
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	cmdline := name
	for _, a := range args {
		cmdline += " " + a
	}
	if dir != "" {
		cmdline = "(cd " + dir + "; " + cmdline + ")"
	}
	logOp("shell", name, cmdline, start, err)
	return string(out), err
}

// cloneTemplate 浅克隆 git 仓库到临时目录，返回本地路径和清理函数
func cloneTemplate(gitURL, branch string) (string, func(), error) {
	tmpDir, err := os.MkdirTemp("", "ts-template-*")
	if err != nil {
		return "", nil, err
	}
	cleanup := func() { _ = os.RemoveAll(tmpDir) }

	args := []string{"clone", "--depth=1"}
	if branch != "" {
		args = append(args, "--branch", branch)
	}
	args = append(args, gitURL, tmpDir)
	cmd := exec.Command("git", args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		cleanup()
		return "", nil, fmt.Errorf("%v\n%s", err, trimStr(string(out), 500))
	}
	// 克隆后立刻删掉 .git，避免走我们自己的"清除 template git 历史"时已经被 runScaffold 处理
	// 这里保持原样，让后续统一处理
	return tmpDir, cleanup, nil
}

// createGitLabRepo 调 GitLab API 在目标 group 下创建空仓库
// POST {APIURL}/projects  body={name, namespace_id or path}
// 更稳妥：用 namespace path 查 id，再建仓。但简化版用 path 创建：
//   POST {APIURL}/projects with namespace_id 可能要查。改用 ?name=xx&path=xx&namespace_id=xx
//   更简单：POST /projects/user/{group_id} 不适用
// 这里实现：先 GET /namespaces?search=<group> 取 id，再 POST /projects
func createGitLabRepo(req *ScaffoldReq, resp *ScaffoldResp) error {
	if req.GitLabGroup == "" {
		return errors.New("未指定 group，无法自动建仓")
	}
	client := &httpClient{timeout: 15}
	apiBase := strings.TrimSuffix(req.GitLabAPIURL, "/")

	// 1. 查 namespace id
	nsURL := apiBase + "/namespaces?search=" + req.GitLabGroup
	var namespaces []struct {
		ID       int    `json:"id"`
		FullPath string `json:"full_path"`
	}
	if err := client.getJSON(nsURL, req.GitLabToken, &namespaces); err != nil {
		return fmt.Errorf("查 namespace 失败：%w", err)
	}
	var nsID int
	for _, ns := range namespaces {
		if ns.FullPath == req.GitLabGroup {
			nsID = ns.ID
			break
		}
	}
	if nsID == 0 {
		return fmt.Errorf("找不到 namespace：%s", req.GitLabGroup)
	}

	// 2. 创建 project
	createBody := map[string]any{
		"name":         req.ProjectName,
		"path":         req.ProjectName,
		"namespace_id": nsID,
		"visibility":   "private",
		"description":  "Generated from mask-go-sample-service by Team Standards",
	}
	var created struct {
		ID            int    `json:"id"`
		WebURL        string `json:"web_url"`
		SSHURLToRepo  string `json:"ssh_url_to_repo"`
		HTTPURLToRepo string `json:"http_url_to_repo"`
	}
	if err := client.postJSON(apiBase+"/projects", req.GitLabToken, createBody, &created); err != nil {
		return err
	}
	resp.Logs = append(resp.Logs, "  · Web: "+created.WebURL)
	resp.Logs = append(resp.Logs, "  · SSH: "+created.SSHURLToRepo)
	return nil
}

// 极简 HTTP 客户端（复用标准库 net/http）
type httpClient struct{ timeout int }

func (c *httpClient) getJSON(url, token string, out any) error {
	req, err := newAuthReq("GET", url, token, nil)
	if err != nil {
		return err
	}
	return doJSON(req, out)
}

func (c *httpClient) postJSON(url, token string, body any, out any) error {
	b, err := json.Marshal(body)
	if err != nil {
		return err
	}
	req, err := newAuthReq("POST", url, token, b)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	return doJSON(req, out)
}

func newAuthReq(method, url, token string, body []byte) (*http.Request, error) {
	var br io.Reader
	if body != nil {
		br = bytes.NewReader(body)
	}
	req, err := http.NewRequest(method, url, br)
	if err != nil {
		return nil, err
	}
	if token != "" {
		req.Header.Set("PRIVATE-TOKEN", token) // GitLab PAT
	}
	req.Header.Set("Accept", "application/json")
	return req, nil
}

func doJSON(req *http.Request, out any) error {
	client := &http.Client{Timeout: 20 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, trimStr(string(b), 300))
	}
	if out == nil {
		return nil
	}
	return json.Unmarshal(b, out)
}

func trimStr(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}

// writeNextSteps 把下一步操作写进项目根的 下一步.md
func writeNextSteps(dest string, req *ScaffoldReq) {
	var sb strings.Builder
	sb.WriteString("# " + req.ProjectName + " — 下一步 🚀\n\n")
	sb.WriteString("项目已基于 mask-go-sample-service 模板生成。从 0 到 Hello World 分 5 步：\n\n")

	sb.WriteString("## Step 1 · 配置内部 Go 模块访问（首次做一次，已配过跳过）\n\n")
	sb.WriteString("项目依赖 `mask-go-common-lib`（团队公共库，在 your-git-server.com 内网），Go 拉依赖需要：\n\n")
	sb.WriteString("```bash\n")
	sb.WriteString("# 让 Go 跳过内网模块的 checksum 校验（必做）\n")
	sb.WriteString("go env -w GOPRIVATE='your-git-server.com/*'\n\n")
	sb.WriteString("# 让 https 路径走 SSH（你的 SSH key 已配，避免 https 登录弹窗）\n")
	sb.WriteString("git config --global url.\"git@your-git-server.com:\".insteadOf \"https://your-git-server.com/\"\n")
	sb.WriteString("```\n\n")

	sb.WriteString("## Step 2 · 装工具链（首次一次，已装跳过）\n\n")
	sb.WriteString("`make proto` / `make wire` 依赖 5 个命令行工具。一次性装好：\n\n")
	sb.WriteString("```bash\n")
	sb.WriteString("# 老 macOS / 没装 Xcode 的机器加 CGO_ENABLED=0，避免卡在 C 编译\n")
	sb.WriteString("CGO_ENABLED=0 go install github.com/bufbuild/buf/cmd/buf@latest\n")
	sb.WriteString("CGO_ENABLED=0 go install github.com/google/wire/cmd/wire@latest\n")
	sb.WriteString("CGO_ENABLED=0 go install google.golang.org/protobuf/cmd/protoc-gen-go@latest\n")
	sb.WriteString("CGO_ENABLED=0 go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest\n")
	sb.WriteString("CGO_ENABLED=0 go install github.com/go-kratos/kratos/cmd/protoc-gen-go-http/v2@latest\n")
	sb.WriteString("```\n\n")
	sb.WriteString("**确认 `~/go/bin` 在 PATH**（老 zsh 可能没加）：\n")
	sb.WriteString("```bash\n")
	sb.WriteString("echo $PATH | tr ':' '\\n' | grep go/bin || {\n")
	sb.WriteString("  echo 'export PATH=\"$HOME/go/bin:$PATH\"' >> ~/.zshrc\n")
	sb.WriteString("  source ~/.zshrc\n")
	sb.WriteString("}\n")
	sb.WriteString("which buf wire protoc-gen-go    # 5 个命令都要找得到\n")
	sb.WriteString("```\n\n")
	sb.WriteString("## Step 3 · 拉依赖 + 代码生成\n\n")
	sb.WriteString("```bash\n")
	sb.WriteString("cd " + req.ProjectName + "\n\n")
	if !req.RunTidy {
		sb.WriteString("go mod tidy          # 拉 mask-go-common-lib 等依赖\n")
	} else {
		sb.WriteString("# go mod tidy 已在生成时跑过\n")
	}
	sb.WriteString("make proto           # buf generate 从 .proto 生成 Go stub\n")
	sb.WriteString("make wire            # 重新生成 wire_gen.go 依赖注入代码\n")
	sb.WriteString("```\n\n")

	sb.WriteString("## Step 4 · 改配置（跑 Hello World 可以全部跳过）\n\n")
	sb.WriteString("`configs/dev/" + req.Biz + ".yaml` 的 **实际字段结构**（conf.proto 定义的）：\n\n")
	sb.WriteString("```yaml\n")
	sb.WriteString("data:\n")
	sb.WriteString("  # database:                  # 注释态 = 不连 DB，NewData 会跳过连接池初始化\n")
	sb.WriteString("  #   driver: pgx\n")
	sb.WriteString("  #   source: \"host=127.0.0.1 port=5432 user=postgres password=xxx dbname=" + req.Biz + " sslmode=disable\"\n")
	sb.WriteString("  # redis:\n")
	sb.WriteString("  #   addr: 127.0.0.1:6379\n")
	sb.WriteString("  #   read_timeout: 0.2s\n")
	sb.WriteString("  #   write_timeout: 0.2s\n")
	sb.WriteString("```\n\n")
	sb.WriteString("**Hello World 阶段保持注释即可**，不装 PG/Redis 也能起。`internal/data/data.go` 的 `NewData` 对 nil 配置已经容错，不用改任何 Go 代码。\n\n")
	sb.WriteString("> 真连 DB 时去掉 `database:` 前面的 `#`，填本地/docker-compose 的 DSN。**字段名是 `driver` + `source`，不是 `dsn`**（GORM 约定）。\n\n")

	sb.WriteString("## Step 5 · 启动 + Hello World\n\n")
	sb.WriteString("```bash\n")
	sb.WriteString("make run             # 或 go run ./cmd/ -conf ./configs/dev/" + req.Biz + ".yaml\n")
	sb.WriteString("```\n\n")
	sb.WriteString("看到日志里出现 JSON 行，含 `\"http server is listening ...\"` 和 `\"grpc server ...\"` 就是起来了（Kratos + mask-go-common/logging 输出 JSON 结构化日志）。\n\n")
	sb.WriteString("**新开终端验证**：\n\n")
	sb.WriteString("```bash\n")
	sb.WriteString(fmt.Sprintf("# 探活（最简验证）\ncurl http://localhost:%d/health\n# => {\"service\":\"" + req.Biz + "-svc\",\"status\":\"ok\"}\n\n", req.HTTPPort))
	sb.WriteString(fmt.Sprintf("# 业务接口（sample 原生是 order，你改业务域后 proto 路径仍是 sample 定义的那个）\n# 看 api/" + req.Biz + "/v1/*.proto 的 option (google.api.http) 确认实际路径。\n# 以 sample 为例（CreateOrder POST + JSON）：\ncurl -X POST http://localhost:%d/api/v1/order \\\n  -H 'Content-Type: application/json' \\\n  -d '{\"user_id\":\"u1\",\"symbol\":\"BTC/USDT\",\"side\":\"buy\",\"amount\":\"0.1\",\"price\":\"60000\"}'\n\n", req.HTTPPort))
	sb.WriteString(fmt.Sprintf("# gRPC（需要 grpcurl）\ngrpcurl -plaintext localhost:%d list\n", req.GRPCPort))
	sb.WriteString("```\n\n")
	sb.WriteString("> ⚠️ **不是** `/api/v1/" + req.Biz + "/ping` 那种 RESTful 常见形状。sample 的 proto 里只有 CreateOrder 一条业务路由，其余 HTTP route 只有 `/health` / `/livez` / `/readyz` / `/metrics` 四个探针。\n\n")

	sb.WriteString("---\n\n")

	// 组件傻瓜教程：每个勾选的组件都有一段 docker + 验证命令
	writeComponentTutorials(&sb, req)

	sb.WriteString("## 💡 后续扩展（不影响 Hello World）\n\n")
	sb.WriteString("- **改业务实体**：把 `api/" + req.Biz + "/v1/*.proto` 里的 `Order` / `CreateOrder` 换成你的真实业务模型（如 `Animal` / `AddAnimal`），对应 `internal/biz/data/service/*.go` 同步改\n")
	sb.WriteString("- **接 DB**：改 `configs/dev/" + req.Biz + ".yaml` 的 `data.database` 段 + `internal/data/migrate.go` 的 model 和表名\n")
	sb.WriteString("- **按团队规范走 schema**：当前 `migrations/*.sql` 建表不带 schema 前缀（`CREATE TABLE orders`）；若要符合团队规范的 §14 schema 分域（`" + req.Biz + ".orders`），手动调整建表 SQL 和 GORM model 的 `TableName()`\n")
	sb.WriteString("- **接 MQ**：当前模板**未集成** Kafka/Pulsar，`conf.proto` 也没有 MQ 字段。需要时自行扩展 `internal/conf/conf.proto`，按团队规范 §6.2 的 topic 命名约定 (`env_svc_entity_action`)\n\n")

	sb.WriteString("## 🎯 关键配置\n\n")
	sb.WriteString(fmt.Sprintf("- HTTP 端口：%d\n", req.HTTPPort))
	sb.WriteString(fmt.Sprintf("- gRPC 端口：%d\n", req.GRPCPort))
	sb.WriteString("- Go Module：`" + req.Module + "`\n")
	sb.WriteString("- 业务域：`" + req.Biz + "`（schema / proto package / api dir）\n")
	if req.Author != "" {
		sb.WriteString("- 作者：" + req.Author + "\n")
	}
	if req.RemoteGitURL != "" {
		sb.WriteString("- Git Remote：`" + req.RemoteGitURL + "`\n")
		if req.GitLabToken == "" {
			sb.WriteString("  - ⚠️ 远端空仓库需手动在 GitLab Web 创建，然后 `git push -u origin main`\n")
		} else {
			sb.WriteString("  - ✓ 已通过 API 自动创建，直接 `git push -u origin main`\n")
		}
	}
	sb.WriteString("\n")

	sb.WriteString("## 📚 团队规范\n\n")
	sb.WriteString("Claude Code / Cursor 的团队规则已全局生效（装过 Team Standards App 的话）。写代码遇到不确定：\n")
	sb.WriteString("- 在 Cursor / Claude Code 里直接问\n")
	sb.WriteString("- 或看 `~/.claude/skills/go-team-standards/references/*.md`\n\n")

	sb.WriteString("---\n\n")
	sb.WriteString("由 **Team Standards App** 🦕 生成\n")

	_ = os.WriteFile(filepath.Join(dest, "下一步.md"), []byte(sb.String()), 0644)
}

// writeSkillTestCases 把 assets/skill-test-cases/ 里的 15 个违规用例 + README
// 展开到 <dest>/.docs/skill-test-cases/，并确保 .gitignore 里有 .docs/ 条目。
func writeSkillTestCases(dest string) ([]string, error) {
	var logs []string
	targetDir := filepath.Join(dest, ".docs", "skill-test-cases")
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return nil, err
	}
	// 从 embed 复制整棵子树
	embedBase := "assets/skill-test-cases"
	entries, err := embeddedFS.ReadDir(embedBase)
	if err != nil {
		return nil, fmt.Errorf("读 embed 失败：%v", err)
	}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		src, err := readEmbedFile(embedBase + "/" + e.Name())
		if err != nil {
			return logs, err
		}
		dst := filepath.Join(targetDir, e.Name())
		if err := os.WriteFile(dst, src, 0644); err != nil {
			return logs, err
		}
	}
	logs = append(logs, fmt.Sprintf("✓ 写入 %d 个 Skill 测试用例 → .docs/skill-test-cases/", len(entries)))

	// 追加到 .gitignore（若未包含）
	giPath := filepath.Join(dest, ".gitignore")
	existing, _ := os.ReadFile(giPath)
	if !bytes.Contains(existing, []byte(".docs/")) {
		content := string(existing)
		if content != "" && !strings.HasSuffix(content, "\n") {
			content += "\n"
		}
		content += "\n# Skill 有效性测试用例（只读不运行，不进 git）\n.docs/\n"
		if err := os.WriteFile(giPath, []byte(content), 0644); err != nil {
			return logs, fmt.Errorf("更新 .gitignore 失败：%v", err)
		}
		logs = append(logs, "✓ .gitignore 已追加 .docs/")
	} else {
		logs = append(logs, "  .gitignore 已含 .docs/，跳过")
	}

	return logs, nil
}
