# 规范版本变更日志

遵循 [Keep a Changelog](https://keepachangelog.com/)。版本号见 `VERSION`。

---

## [1.8.16] — 2026-06-12

### Added
- 安装 tab 新增「VSCode AI 扩展」卡片：一键为 Copilot / Cline / Roo / Continue / Windsurf / Gemini 生成 AGENTS.md + 各家入口文件（项目级）

### Removed
- 删除「测试 Skill 有效性」功能（tab + 后端 eval/providers + Clawnova 配置）

## [1.8.15] — 2026-06-12

### Security
- 全历史清除内部域名残留；二进制不再包含任何内部地址字样
- 同步地址迁移规则改为通用判断（不点名旧域名）

## [1.8.14] — 2026-06-12

### Fixed
- 规范云端同步支持 GitHub 仓库（原仅支持 GitLab v4 API，配 GitHub 地址会失败）
- 旧版残留的内部 GitLab 同步地址自动迁移到 GitHub 公开仓库
- UI 文案：GitLab 镜像 → GitHub 公开仓库

## [1.8.13] — 2026-06-12

### Changed
- UI 优化：安装目标改为 2×2 grid 布局，hook 卡片图标按触发类型着色
- 提示词 tab：5 张工作流提示词卡片 + AGENTS.md 初始化功能
- team-guard 单文件一体化 hook（UserPromptSubmit+PreToolUse+PostToolUse）
- 新增 Codex skill 支持、所有路径输入框支持目录浏览器

---

## [1.8.12] — 2026-05-23

### Added
- **Codex Skill 支持**：安装目标新增「仅 Codex」选项，`all` 覆盖 Claude + Cursor + Codex 三端；skill 安装到 `~/.codex/skills/go-team-standards/`
- **已安装页 Codex 一节**：扫描 `~/.codex/skills/` 展示已装 skill
- **team-guard 一体化 Hook**：一个脚本覆盖 UserPromptSubmit + PreToolUse + PostToolUse，安装时自动写入 `~/.claude/settings.json`
- **强制约束路径浏览**：项目级 Hook 安装支持「📁 选目录」macOS 原生选择器

---

## [1.8.11] — 2026-05-22

### Added
- **git-commit-guard**（PreToolUse）：git commit 前一次性全量守卫
  - Commit Message 格式校验（Conventional Commits）
  - Go 铁律：panic / fmt.Println / error丢弃 / 裸goroutine / float金额 / user_id / is_deleted / camelCase JSON tag / 时间字段命名
  - SQL 铁律：user_id 列 / is_deleted 列 / float金额列 / 缺 created_at|updated_at / TIMESTAMP 未带时区 / VARCHAR 无长度 / 主键类型建议
  - Proto 铁律：user_id / is_deleted / float金额

---

## [1.8.10] — 2026-05-22

### Changed
- **Superpowers tab UI 重设计**：安装目标改为 3 格大图标卡片（比小文字按钮更易点），顶部改为左右 2 列布局（安装操作 + 快速上手并排），8 个 skill 卡片展示命令名
- **强制约束 Hook 卡片重设计**：左图标 + 中信息 + 右状态/操作的三栏布局，状态 badge 颜色区分已启用/已暂停/未安装，hook 类型标签 PostToolUse/PreToolUse/UserPromptSubmit 各用不同颜色
- 新增 `.sp-target-btn`、`.sp-skill-card`、`.hook-card-icon`、`.hook-status-badge` 等 CSS class

---

## [1.8.9] — 2026-05-21

### Added
- 强制约束新增 6 个铁律衍生 Hook，均为 PostToolUse，AI 写文件后立即检测：
  - `user-id-guard`：禁用 `user_id`，全平台统一 `uid`（field-naming §1.2）
  - `float-amount-guard`：金额/价格字段禁用 `float32/float64`，必须用 `decimal`（database §字段规范）
  - `fmt-println-guard`：业务代码禁用 `fmt.Println/Printf`，必须用 `slog`（go-style §7）
  - `error-ignore-guard`：禁止 `_, _ := func()` 丢弃 error（go-style §4.1）
  - `soft-delete-guard`：软删除禁用 `is_deleted BOOLEAN`，统一 `deleted_at TIMESTAMPTZ`（database §字段规范）
  - `goroutine-naked-guard`：禁裸 `go func()`，必须通过 `errgroup` 管理（go-style §5）

---

## [1.8.8] — 2026-05-21

### Fixed
- Superpowers 项目级安装：路径输入框旁增加「📁 选目录」按钮，调用 macOS 原生目录选择器，不再需要手动输入路径

---

## [1.8.7] — 2026-05-21

### Changed
- **GSD 框架 tab 替换为 Superpowers**：侧栏「GSD 框架」改为「⚡ Superpowers」，来自 [obra/superpowers](https://github.com/obra/superpowers)
- Superpowers tab 支持 3 种安装目标：🤖 Claude Code（全局）、🖱️ Cursor（全局）、📁 项目级
- 自动检测安装状态（扫 `~/.claude/skills/superpowers*`）
- 一键安装（git clone 从 GitHub）+ 一键卸载（两步式确认，WKWebView 安全）
- 复制安装命令按钮（可在 Claude Code / Cursor 对话框手动执行）
- 展示 8 个核心 skill 说明卡片 + 3 步快速上手指南

---

## [1.8.6] — 2026-05-21

### Added
- 强制约束 tab 支持「全局」和「项目级」两种作用域：
  - 全局 → `~/.claude/hooks/`（对所有项目生效）
  - 项目级 → `<项目根>/.claude/hooks/`（仅对当前项目生效）
- 切到项目级时显示路径输入框，确认后加载该项目下的 hook 状态
- 安装 / 卸载 / 启用 / 禁用均按当前选中作用域操作，互不干扰

---

## [1.8.5] — 2026-05-21

### Added
- 新增「强制约束」选项卡（侧栏 🔒），支持 hook 的安装 / 卸载 / 启用 / 禁用
- 内置 4 个强制约束 hook：
  - 🛡️ DevDefender 需求防御（UserPromptSubmit）
  - 🚨 Go 禁裸 panic（PostToolUse · 写 .go 文件后校验）
  - 🔤 JSON 禁驼峰字段名（PostToolUse · 写 .go/.json/.md 后校验）
  - 📝 Commit Message 格式（PreToolUse · git commit 前校验）
- 禁用 = 重命名为 .sh.disabled（Claude Code 不读取），不删文件
- 侧栏徽章实时显示已启用 hook 数量

---

## [1.8.4] — 2026-05-21

### Fixed
- UI: 所有卸载 / 危险操作按钮改为两步式确认（第一次点击变红提示，3s 内再点一次才执行），  
  修复 WKWebView 屏蔽原生 `confirm()` 导致所有卸载按钮点击无反应的问题  
  影响范围：已装清单单项卸载、GSD 卸载全部 gsd-*、gsd-help-zh 卸载、OrangeCat 模板重置、dev-dna 档案重置
- UI: 卸载结果改用页内 log 展示，不再依赖 `alert()`（同样在 WKWebView 里被屏蔽）

---

## [1.8.3] — 2026-05-20

### Changed
- 欢迎页：3 秒后自动关闭（无需点按钮）
- 欢迎页文案带版本号：升级时显示「v1.8.2 → v1.8.3 升级完成，所有规范和工具已更新到最新版本」

---

## [1.8.2] — 2026-05-20

### Added
- UI: 欢迎页（首次启动 / 每次升级到新版本后弹一次）  
  显示「👏 欢迎尊贵的用户使用本 App 👏」+ 版本号，点「开始使用」关闭，  
  下次同版本进入不再弹出，升级后再弹一次

---

## [1.8.1] — 2026-05-20

### Fixed
- upgrade: 点"一键升级"时若后台自动升级已在进行中，返回 `ok:true + already_running:true`  
  而非 `ok:false`，前端直接进入进度轮询，不再显示假的"升级失败"提示

---

## [1.8.0] — 2026-05-20

### 重大修复（自动升级全链路）
- upgrade: 自动升级时优先从 release assets 下载最新 `upgrade-helper.sh`，打破老 binary 嵌旧 helper 的死局，**任意旧版本均可自动升级**
- upgrade: `performAutoUpgrade()` helper 启动后 3 秒自动 `os.Exit(0)`，解决 App 不退出 → helper 60s 超时 abort 的根本问题
- upgrade: helper 完整执行 `xattr -cr` + `open`，升级后无需任何手动操作，自动重新唤起
- update: `loadAppUpdateConfig` 强制覆盖 `release_project` / `release_host`，修复老用户 config 残留错误仓库地址导致看不到更新 banner
- install: `references/` 改为动态同步 `standards/` 全量内容，解决硬编码旧文件名导致安装失败 + 接口规范重复出现

### Added
- api-doc: 规范升级 v1.3.0，JSON 字段全面改为 snake_case，严禁驼峰
- DevDefender skill + hook 随主安装一起分发
- catalog: DevDefender 安装后自动出现在 Skill 管理卡片列表

---

## [1.7.58] — 2026-05-20

### Fixed
- update: `loadAppUpdateConfig` 改为强制覆盖 `release_project` / `release_host` 为正确值，  
  不再依赖文件里存的值。解决老用户 config 里残留 `"len/standards"` 导致  
  App 始终连接错误仓库、看不到更新 banner 的问题。

---

## [1.7.57] — 2026-05-20

### Fixed
- upgrade: `performAutoUpgrade()` 在 helper 启动后 3 秒自动调用 `os.Exit(0)`，  
  解决 App 不退出导致 helper 等 60s 超时、升级始终失败的问题
- install: `installClaude` 的 `references/` 改为动态扫描 `standards/` 目录而非硬编码列表，  
  解决列表含已删除文件名（`api-design.md`、`api-doc-example.md`）导致安装失败、  
  旧版 references 留存、接口规范重复出现的问题

---

## [1.7.56] — 2026-05-20

### Fixed
- upgrade: 自动升级时优先从 release assets 下载最新 `upgrade-helper.sh`，用完即弃。  
  彻底打破"老 binary 嵌了旧 helper → 一键升级永远失败"的鸡蛋困境：  
  **任何版本（含 v1.7.50）只要能访问 GitLab，均可自动升级到 v1.7.56+**。  
  下载失败时自动回退使用 binary 内嵌的 helper。

---

## [1.7.55] — 2026-05-20

### Test
- 测试版本：验证 v1.7.54 自动升级流程（Go预挂载DMG）是否正常

---

## [1.7.54] — 2026-05-20

### Fixed
- upgrade: Go 层预先 mount DMG（plist 解析），把挂载路径作为第 5 参数传给 helper，彻底解决含空格路径导致所有用户升级失败的问题
- upgrade-helper: 优先使用 Go 传入的挂载路径，无法获取时降级自行 mount（兼容旧版）
- upgrade 失败弹窗：加「💽 立即下载 DMG 手动安装」直接下载按钮，不再只有文字提示

### Added
- catalog: DevDefender skill 已安装时在 Skill 管理里自动显示卡片（触发方式：/devdefender）

---

## [1.7.53] — 2026-05-20

### Added
- DevDefender skill 随主安装一起分发（`~/.claude/skills/DevDefender/`）
- DevDefender hook 脚本随 global 安装写入 `~/.claude/hooks/devdefender-guard.sh`
- 新增 `hooks/` embed 目录，后续 hook 脚本统一从这里分发

---

## [1.7.52] — 2026-05-20

### Fixed
- upgrade-helper: 用 plist 解析 hdiutil 挂载点，修复路径含空格时 awk 截断导致 mount 判空失败的 bug（之前每次升级都回滚）

---

## [1.7.51] — 2026-05-20

### Changed
- app-update: 后台检测间隔从 1h 缩短至 1 分钟
- app-update: 检测到新版本后立即静默自动下载并安装，无需手动点击

### Fixed
- app-update.json 默认 release_project 注释明确为 len/go-skill

---

## [1.7.50] — 2026-05-20

### Fixed
- catalog: 删除 hardcoded api-design 条目，api-doc.md 走动态扫描
- scanExtraReferences 去掉 example/template 过滤，规范文件命名不再受限
- 清理本地残留 api-design.md / api-doc-example.md

---

## [1.7.49] — 2026-05-20

### Fixed
- 规范云端拉取有变更后，规范模块卡片立即自动刷新，无需重启
- 安装页「更新与同步」tab 新增 App 版本更新常驻卡片（显示当前/最新版本，切 tab 自动检查，有新版显示升级按钮）

### Changed
- api-doc-example.md 重命名为 api-doc.md，避免被 catalog 扫描过滤
- 删除旧版 api-design.md（与新版 GET/POST 规范冲突）
- api-doc.md 升级至 v1.2.0（新增 with_column/columns/评审清单，修正3处内部不一致）

---

## [1.7.48] — 2026-05-16

### Chore — 发版验证（自动同步 / GitLab Release）

无功能性代码变更。在完成 `glab` 鉴权后重新走完整 `publish.sh` 流程，用于验证：

- Team Standards App 的 **检查更新** 能否从 `len/go-skill` 拉到新 release；
- 已安装客户端是否在 **约 1h 内**（或手动「检查更新」）感知到新版本 banner。

---

## [1.7.47] — 2026-05-15

### Fixed — 状态条「有新版本」时缺失最后同步时间

[app.js:1383](web/app.js:1383) 在检测到远端有新版本时，整段覆盖了状态条文本，把上面已经渲染好的「· 上次同步 abc12345 (5 分钟前)」冲掉了。用户看到「有新版本」就完全看不到本地是什么时候同步过来的。

**修法**：「有新版本」消息里追加 `· 上次同步 X 分钟前`（从 `cfg.last_synced_at` 转 `fmtTimeAgo`），与「最新版本」消息保持时间信息一致。

---

## [1.7.46] — 2026-05-15

### Fixed — 三连修：剪贴板、规范同步偶发 EOF、仓库 slug

#### 1. Cmd+C/V/X/A/Z 在 App 内不生效（v1.7.45 的回归）

**根因**：`installNativeAppMenu()` 在 `webview.New(false)` **之后**调用。`menu_darwin.go` 头部注释明确说"必须在 `webview.New` 之前"——webview_go 创建窗口时若发现 NSApp 没装 mainMenu，会自己临时挂一个空的，覆盖后装的 Edit 菜单。

**修法**：把 `installNativeAppMenu()` 提到 `webview.New` 之前一行。

#### 2. 规范同步 `unexpected EOF` / `context deadline exceeded`

**根因**：Go 默认 `http.Transport` 走 Clash/V2Ray 这类本地 HTTP 代理时，HTTP/2 长连接复用与代理 keepalive 时机经常对不上，偶发中途断流。15s 总超时也偏紧。

**修法**：`newProxyAwareHTTPClient` 调优 ——
- `ForceAttemptHTTP2: false` 强制 HTTP/1.1
- 显式 `TLSHandshakeTimeout / ResponseHeaderTimeout / ExpectContinueTimeout / IdleConnTimeout`
- `DisableKeepAlives: true`（规范同步 1h 一次，每次新连接性能损失忽略不计）
- 总 timeout 15s → 30s

#### 3. App release 仓库纠正：`len/standards` → `len/go-skill`

历史 bug：`app_update.go` 的 `DefaultReleaseProject` 和 `upload-release.sh` 的 `RELEASE_PROJECT` 都写成了 `len/standards`（规范文档仓），但 **App 自身的 DMG release 应该发到 `len/go-skill`（skill 代码仓）**。两个仓库职责分离：

- `len/go-skill` —— App 代码 + DMG release
- `len/standards` —— 规范文档源（保持不变）

之前老用户的「检查更新」一直对着规范仓查 release，永远拿不到 App 新版（规范仓没 release）。这次修正后 `/api/app-update/check` 会去正确的 `len/go-skill` 拉 release。

**规范同步默认仓 `len/standards` 不动**——本来就对。

### Added — 后台每小时定时检测

**Why**：之前后台 check 只在 App 启动时跑一次。规范 owner 推新版后，团队成员要么手动点「检查更新」要么重启 App，体验割裂。

**改法**：
- 新 `standardsSyncBackgroundLoop()` / `appUpdateBackgroundLoop()`，启动跑一次 + `time.NewTicker(1h)` 循环
- 规范：发现远端 SHA != `LastSyncedSHA` → **静默自动覆盖** `~/.claude/skills/.../references/` 和 `~/.cursor/.../references/`。无需重启 Claude Code / Cursor（references 是按需读取，不常驻内存）
- App 更新：仅刷新 `LastSeenVersion`，前端 banner 自取（不自动装，重启 App 是显式动作）

### Added — `docs/规范检测说明.md`

完整说清楚检测流程、覆盖路径、为啥不用重启、与 App 自身更新的区别。

---

## [1.7.45] — 2026-05-15

### Fixed — GSD「一键装」按下无反馈（根因：HTTP WriteTimeout）

用户反馈点「🚀 一键装（npx）」后按钮卡在「⏳ 安装中…」永远不动。

**根因**：`main.go` 里 `http.Server{WriteTimeout: 30 * time.Second}`。但 npx 装 GSD 框架要 1-3 分钟（首次下 50-100MB 包），30 秒后 Go server 强关连接 → 前端 fetch promise 永不 resolve → 按钮卡死，无任何错误提示。

**修法**：彻底改异步 + 实时进度轮询。

#### 后端 `gsd_install.go` 重写

新增 `gsdInstallProgress` 结构 + 互斥锁的全局状态：
- `Phase`: idle / running / done / failed
- `StartedAt` / `EndedAt` / `ExitCode` / `ElapsedMS`
- `Log`: 累积的 stdout/stderr，最大 200KB（超出截前半）

`POST /api/gsd/install` 改异步：
1. 检测 npx + 防并发（已 running 直接拒绝）
2. spawn goroutine `runGSDInstallAsync`
3. **立即** 返回 `{ok: true, phase: "running"}`，绕过 WriteTimeout

`runGSDInstallAsync` 内部：
- 用 `cmd.StdoutPipe()` 实时读 stdout（stderr 也走同管道）
- `bufio.Scanner` 逐行追加到 log buffer
- 自己控 10 分钟超时

新增 `GET /api/gsd/install/progress` 返回当前 snapshot，前端 500ms 一次轮询。

#### 前端 `web/app.js` 改 polling

- 点击 → 立即 fetch `/api/gsd/install` 启动
- 启动响应后 `setInterval(tick, 500)` 拉 progress
- 按钮文案实时显示 `⏳ 安装中… 23s`
- 日志面板实时显示 npx 输出（含 ANSI 已关闭：`NO_COLOR=1`）
- done/failed 时停止 polling + 刷新 GSD tab
- **启动时**也调一次 progress：如果发现有 running 任务（用户重开 App / 切走再回来），自动接管 polling 显示

#### 顺带改善

- npx 环境变量加：`CI=1` / `NPM_CONFIG_YES=true` / `NO_COLOR=1` / `FORCE_COLOR=0` — 彻底非交互、无 ANSI 干扰
- `bash -lc` 改 `userLoginShell()` — 优先 zsh（用户默认 shell），fallback bash，确保 nvm/asdf 装的 node 能找到
- log 限 200KB 防 OOM

### Added — 项目安装路径加「📁 选目录」按钮

「⚡ 安装 → 1. 作用范围 → 项目安装」的输入框旁边加按钮，点击调 macOS 原生 `osascript choose folder`，省手敲。

复用现有 `POST /api/scaffold/pick-folder` 端点。

### Added — macOS 原生 Edit 菜单（cgo + Cocoa）

新增 `menu_darwin.go`：webview_go 创建 NSWindow 但**不装 NSMenu**，导致 Cmd+C/V/X/A 在某些 WKWebView 配置下不工作（OS 找不到 menu item 路由）。

cgo 写 Cocoa 装一个标准菜单栏：
- **App 菜单**：About / Hide / Quit (Cmd+Q)
- **Edit 菜单**：Undo (Cmd+Z) / Redo (Cmd+Shift+Z) / Cut (Cmd+X) / Copy (Cmd+C) / Paste (Cmd+V) / Delete / Select All (Cmd+A)
- **Window 菜单**：Minimize (Cmd+M) / Close (Cmd+W)

OS 自己路由这些 shortcut 到 first responder（聚焦的文本框）。比纯靠 JS keydown handler 可靠 10 倍——之前 JS 路径在某些 macOS 版本/WKWebView 配置失灵。

`main.go` 在 `webview.New()` 后调 `installNativeAppMenu()`。

### Notes

- 改：`gsd_install.go`（同步 → 异步 + progress snapshot，~+170 行）
- 改：`main.go`（注册 `/api/gsd/install/progress` + 调 `installNativeAppMenu()`）
- 改：`web/app.js`（polling 逻辑 + 选目录按钮 handler）
- 改：`web/index.html`（项目路径输入框旁加 📁 按钮）
- 新：`menu_darwin.go`（~95 行 cgo Cocoa）
- 编译过

### 验证

- 点「一键装」→ 按钮变「⏳ 安装中… Ns」秒数实时增加
- 日志面板实时滚动显示 npx 输出
- 完成后按钮恢复 + 状态刷新
- 切走 GSD tab 再回来：仍能看到 polling 状态（不丢失）
- Cmd+C/V 在文本框里 100% 可用（NSMenu 路由）

---

## [1.7.44] — 2026-05-15

### Fixed — 已装清单卸载按钮 v1.7.43 加错但失效

v1.7.43 给已装清单加 🗑 但代码里有两个 hard error：

**Bug 1 · Cursor 路径写错**
- 实际：`~/.cursor/rules/<name>.mdc`
- 代码写：`~/.cursor/skills-cursor/<name>/`（来自我自己的 cowork install 那条线，张冠李戴）
- 结果：所有 cursor rule 删除请求都 403 "路径不在允许的 skill 目录内"

**Bug 2 · 强制目录**
- 代码：`if !info.IsDir() { writeError(400, "路径不是目录") }`
- 但 cursor rules 是 `.mdc` 文件，不是目录
- 结果：cursor rule 卸载请求都 400

**修复**（`skill_uninstall.go`）：
- 允许的根扩到三个：`~/.claude/skills/` + `~/.cursor/rules/` + `~/.cursor/skills-cursor/`（后者备用）
- 删除支持文件 + 目录两种（cursor rule 是 `.mdc` 文件用 `os.Remove`，claude skill 是目录用 `os.RemoveAll`）
- 团队 cursor rule（数字前缀 `NN-xxx.mdc`）也加入禁删清单

**前端修复**（`app.js`）：
- confirm 提示按 path 后缀动态显示「删整个目录」或「删文件」
- 错误处理：把 backend 返回的 `error` + `detail` 都展示，方便用户排查
- 加 `console.error` 让开发者能看到具体失败原因

### E2E 验证

```
✓ 真实删 ~/.cursor/rules/test-fake-uninstall.mdc → ok
✗ 拒绝删 00-iron-laws.mdc → 团队 cursor rule 请走「⚡ 安装」
✗ 拒绝删 go-team-standards → 团队核心 skill 请走「⚡ 安装」对应卡片
✗ 拒绝穿越 /etc/passwd → 路径不在允许的 skill 目录内
```

### 教训

加 feature 前没去实际看 `/api/installed` 返回的 path 长啥样，凭印象写的卸载逻辑——结果 cursor 那边全挂。

后面加任何「按路径操作」的 API，先 curl 一次实际数据，不要靠脑补。

---

## [1.7.43] — 2026-05-15

### Fixed — 大重构：规范 / Skill / GSD 三类分清楚（之前混在「规范模块」里）

v1.7.40~42 把 GSD 框架 66 张卡塞到了「规范模块」tab，**位置错了**。本版整体重排：

#### 重排前后

| Tab | v1.7.42 之前（错） | v1.7.43（对） |
|---|---|---|
| **规范模块** | 82 张（12 团队 + 4 ☁️ + **66 GSD**） | **16 张**（12 团队 + 4 ☁️） |
| **已装清单** | 列出所有 ~/.claude/skills/ 里的 skill，只能查看 | **加 🗑 卸载按钮** |
| **新增：💪 GSD 框架** | 不存在 | 独立 tab + 状态 + 一键装 + 中文 help + 使用说明 + 66 张卡 |

**核心原则**：规范是开发规范（写代码遵循的约定），skill 是 skill（独立的 AI 能力），GSD 是 GSD（spec-driven 框架）。三类不混。

#### 具体改动

**1. 规范模块退回 16 张**（catalog.go）
- `handleCatalog` 移除 `scanGSDFrameworkSkills()` merge
- 现在只返回 12 hardcoded 团队规范 + 4 张 ☁️ ref-auto = 16 张
- 卡片右上角 🗑 按钮**移除**（不再让用户在这里卸 skill）

**2. 已装清单加 🗑 卸载**（installed view）
- 每个 skill item 右下角加「🗑 卸载」按钮（团队 4 个 hardcoded 除外）
- 新后端 API：`POST /api/installed-skill-uninstall` 按路径直删
- 严格安全：必须在 `~/.claude/skills/` 或 `~/.cursor/skills-cursor/` 内 / 必须是目录 / 团队核心 4 个禁删（要走「⚡ 安装」对应卡）

**3. 侧栏新加独立 tab「💪 GSD 框架」**（紫色 nav icon）
- 全套独立体验，不污染规范模块
- 子模块：
  - **📊 状态**：Node 版本 + 已装 gsd-* skill 数 + 核心 6 检查 + 老版残留警告
  - **🚀 一键装**：执行 `npx --yes get-shit-done-cc@latest --claude --global`（8 分钟阻塞）
  - **🇨🇳 中文版 help**：装 / 卸载 gsd-help-zh skill
  - **📖 使用说明**：内置 quick start（new-project → plan-phase → execute-phase → ship）+ 8 个最常用命令 + 与团队规范协作说明
  - **📚 已装 GSD Skill 清单**：66 张中文化卡，点开看完整 SKILL.md

**4. 新 API**
- `GET /api/gsd-framework/list` — 独立返回 GSD 66 张卡（不再混进 catalog）
- `POST /api/installed-skill-uninstall` — 已装清单用的路径式卸载

### 团队同事现在看到的位置

| 想做的事 | 去哪 |
|---|---|
| 看团队 14 铁律 / 命名 / 数据库 等规范 | **规范模块** tab（16 张白卡 + ☁️ 卡） |
| 看哪些 skill 装了 / 卸某个 skill | **已装清单** tab（每个卡有 🗑） |
| 用 GSD 做项目 / 装框架 / 看命令参考 | **💪 GSD 框架** tab（侧栏紫色那条） |
| 装 / 卸 团队 4 个核心 skill | **⚡ 安装** tab |

### Notes

- 改：`catalog.go`（移除 GSD merge + 新加 `/api/gsd-framework/list`）、`main.go`（注册 2 路由）、`web/index.html`（侧栏 + 新 tab section）、`web/app.js`（卡片去 🗑 + 已装清单加 🗑 + GSD tab handler）、`skill_uninstall.go`（加 path-based uninstall handler）、`web/style.css`（GSD nav 紫色样式）
- 不删除：`gsd_translations.go` / `scanGSDFrameworkSkills()`（GSD 框架 tab 复用）
- E2E 实测：规范模块返回 16 张 ✓ / `/api/gsd-framework/list` 返回 67 张 ✓
- 编译过

---

## [1.7.42] — 2026-05-15

### Added — GSD 命令参考中文版 skill（`/gsd-help-zh`）

上游 `gsd-help` skill 只有英文版命令参考（760 行）。本版 ship 一份完整中文译版，**双语并存**，不动上游。

#### 实现

App 内嵌两个文件：

```
assets/gsd-help-zh/
├── SKILL.md         # gsd-help-zh skill 入口（中文 description + 引用 help.zh.md）
└── help.zh.md       # 760 行中文译版（完整翻译上游 v1.39 help.md）
```

装的时候写到：

```
~/.claude/skills/gsd-help-zh/SKILL.md
~/.claude/get-shit-done/workflows/help.zh.md
```

用户 Cmd+Q 重启 Claude Code 后敲 `/gsd-help-zh` 看中文版；`/gsd-help` 仍然英文版。

#### 翻译原则

- **保留所有命令名 / 参数 / 文件名 / 路径**（不译，框架术语统一）
- **保留所有代码块**（不动）
- **意译而非逐字翻译**，重点把"做啥用"说清楚
- 关键英文术语首次出现保留括注（如「workflow（主流程）」）
- 中英混排自然流畅，符合开发者阅读习惯

#### 完整覆盖（760 行 reference 全译）

- 快速开始 + 主流程
- 项目初始化（`/gsd-new-project` / `/gsd-map-codebase`）
- Phase 规划（`/gsd-discuss-phase` / `/gsd-mvp-phase` / `/gsd-plan-phase`）
- 执行（`/gsd-execute-phase`）
- 智能路由 / 快速模式 / Fast 内联
- Roadmap 管理（4 个 phase 子命令）
- Milestone 管理 + 进度跟踪 + 会话管理
- 调试 / Spike & Sketch / 捕获想法
- UAT 验证 / Ship 发布 / 跨 AI Review
- 命名空间路由（6 个元 skill）
- 文件结构 + Workflow 模式 + 规划配置
- 常见 workflow 示例

#### 新增后端 `gsd_help_zh_install.go`（3 API）

- `GET /api/gsd-help-zh/status` — 查 skill / help.zh.md 是否装好
- `POST /api/gsd-help-zh/install` — 把 embed 内容拷到 `~/.claude/`
- `POST /api/gsd-help-zh/uninstall` — 删 skill + help.zh.md（不影响 gsd-help 英文版）

#### 前端 UI

GSD 框架卡底部加分隔区「🇨🇳 中文版命令参考」：
- 状态行实时显示装/未装
- 「🇨🇳 装中文版 help」按钮
- 「🗑 卸载中文版」按钮

### 团队同事使用流程

1. 装 v1.7.42
2. 在「💪 GSD 框架」卡底部点「🇨🇳 装中文版 help」
3. Cmd+Q 重启 Claude Code
4. 敲 `/gsd-help-zh` → 看完整中文命令参考
5. 也能敲 `/gsd-help` 看原版英文（共存）

### E2E 验证

实测装好后：
```
~/.claude/skills/gsd-help-zh/SKILL.md             571 bytes
~/.claude/get-shit-done/workflows/help.zh.md   29,578 bytes
```

Claude 重启后 skill 列表里出现 `gsd-help-zh: 展示 GSD 命令参考的中文版`。

### 上游升级独立性

`npx get-shit-done-cc@latest` 升级**不会动**：
- `help.zh.md`（上游只管 `help.md`）
- `gsd-help-zh/SKILL.md`（上游只装 `gsd-help/`）

所以中文版完全跟我们 App 的版本走，不会被覆盖。

### Notes

- 新增：`assets/gsd-help-zh/SKILL.md` + `assets/gsd-help-zh/help.zh.md`（约 760 行翻译）
- 新增：`gsd_help_zh_install.go`（~105 行）
- 改：`main.go`（注册 3 路由）、`web/index.html`（GSD 卡底部加按钮区）、`web/app.js`（~55 行 install/uninstall handler）
- E2E smoke test 通过：install → 文件落地 → status 报告 ✓ → skill 出现在 Claude 可用列表
- 编译过

### 后续可优化（按需）

- 上游版本升级时检测 help.md 内容变化，提示用户中文版可能过期
- 更多语言版本（日 / 韩 / 葡）—— 参照上游 README 多语言架构

---

## [1.7.41] — 2026-05-15

### Added — 单 skill 卸载（卡片右上角 🗑）+ GSD 框架 66 张卡全中文化

#### 1️⃣ 单 skill 卸载

之前只有「立即安装」/ 「一键更新」装新的，没有针对单个 skill 的精细卸载。本版加：

**前端**：每张卡片 hover 时右上角浮现红色 🗑 小按钮，点 → confirm → 调 API → 刷新。
- ✅ 可卸：`gsd-framework`（66 张 GSD 框架卡）和 `ref-auto`（4 张 ☁️ 云同步卡）
- ❌ 不可卸：`hardcoded`（团队规范 12 条）—— 这些 embed 在 DMG 里，卡片上不显示 🗑 按钮

**后端 `skill_uninstall.go`**：
- `POST /api/skill-uninstall` body: `{id, source}`
- `gsd-framework` → 删 `~/.claude/skills/<id>/` + `~/.cursor/skills-cursor/<id>/`
- `ref-auto` → 删 `references/<id>.md`（两端）
- 严格安全：id 不能含 `..` / `/`，路径必须在 `$HOME` 内（防穿越）
- 返回 `removed[] / skipped[] / failed[]` 分类，前端展示

**CSS**：`.card-uninstall-btn` opacity 0 默认隐藏，`.card:hover .card-uninstall-btn` 显形——hover 才出现，不抢主体的视觉。

#### 2️⃣ GSD 框架 66 张卡全中文化

新加 `gsd_translations.go` —— 66 个 GSD skill 的中文 `Title` + `Description` 映射，覆盖默认的英文 frontmatter。

**示例**（前 5 张）：
```
gsd-add-tests         → 💪 补测试 · 按 UAT 准则给已完成 phase 生成测试
gsd-ai-integration-phase → 💪 AI 集成 Phase · 为含 AI 系统的 phase 生成 AI-SPEC.md 设计契约
gsd-audit-fix         → 💪 审计→修复 流水线 · 自治流程：找问题 → 分类 → 修 → 测 → 提交
gsd-discuss-phase     → 💪 讨论 Phase（澄清需求）· 通过自适应提问收集 phase 上下文
gsd-plan-phase        → 💪 计划 Phase（产出 PLAN.md）· 带验证循环的 phase 详细 plan
```

**实现策略**：
- 仅 UI 层翻译。**不改磁盘上的 SKILL.md**（AI 读原版英文，与上游保持一致避免漂移）
- `scanGSDFrameworkSkills()` 优先查 `gsdTranslations[name]`，命中用中文；未命中 fallback 到 frontmatter 英文（如上游加新 skill）
- 点开 modal 看完整 SKILL.md 时还是英文（避免翻译版漂移）

#### 团队同事使用流程

1. 装 v1.7.41
2. 「规范模块」→ 66 张 💪 卡都是中文标题 + 中文描述，**一眼就能扫**哪个 skill 干啥
3. hover 任一卡 → 右上角浮现 🗑 → 点 → 弹 confirm → 卸载
4. 误装的 / 不需要的 skill 立等卸掉，不用 manual `rm -rf`

### Notes

- 新增：`gsd_translations.go`（66 个翻译条目，~250 行）+ `skill_uninstall.go`（~110 行 + 安全检查）
- 改：`catalog.go`（scanGSDFrameworkSkills 套用翻译）、`main.go`（注册 1 路由）、`web/app.js`（uninstall handler ~35 行）、`web/style.css`（card-uninstall-btn ~20 行）
- 编译过，DMG 16 MB
- smoke test 验证翻译正确（前 10 张都是中文）

### 翻译质量说明

66 条翻译参照上游 v1.39 版 SKILL.md frontmatter description **意译**而非逐字翻译，重点保证：
- 中文卡片能 1 秒看懂 skill 用途
- 保留专有名词（phase / plan / milestone / SKILL / UAT 等）—— 这些是框架术语，不译
- 保留命令名（discuss-phase / plan-phase / execute-phase 等）

上游升级到 v1.40+ 新加的 skill，App 会 fallback 到 frontmatter 英文，等下版本翻译更新。

---

## [1.7.40] — 2026-05-15

### Added — 自动加载已装的 GSD 框架 skill 为「规范模块」卡片

用户装了 [gsd-build/get-shit-done](https://github.com/gsd-build/get-shit-done) 框架（66 个 gsd-* skill），但 App 的「规范模块」tab 看不到。本版补：**启动时扫 `~/.claude/skills/gsd-*` 自动加载**，每个 SKILL.md 变成一张「💪」卡。

#### 实测扫到的卡片数（基线）

```
总卡片数: 82
  团队规范:   12  （hardcoded）
  云同步规范: 4   （☁️ ref-auto，从 references/ 扫出来）
  GSD 框架:   66  （💪 新加，从 ~/.claude/skills/gsd-*/SKILL.md 扫）
```

未来 GSD 框架 upgrade 装更多 skill，App 重启自动看到——零 hardcoded、零写死。

#### 实现

**`catalog.go`**：
- 加 `scanGSDFrameworkSkills()` 扫 `~/.claude/skills/gsd-*`
- 每个 SKILL.md 解析 frontmatter（新加 `extractFrontmatterNameDesc()` 支持 single-line / multi-line / 引号包裹的 `description`）
- 卡 title 自动生成：`gsd-plan-phase` → "💪 Plan Phase"（去前缀 + 空格替换 + Title Case）
- description 取自 frontmatter，截断 200 字符
- Source = "gsd-framework", Group = "GSD 框架"
- Reference 字段塞**磁盘绝对路径**（不是相对 standards/）

**Skill struct 加 `Source` + `Group` 字段**（v1.7.40 schema 扩展，向后兼容）：
- hardcoded → source=hardcoded, group=团队规范
- ref-auto → source=ref-auto, group=云同步规范
- gsd-framework → source=gsd-framework, group=GSD 框架

**`/api/skill-disk-file?path=<abs>`**：
- 读 `~/.claude/skills/` 下任意文件
- 安全检查：必须在该目录内（防穿越）+ 禁止 `..`
- 前端打开 GSD 卡 modal 时调它读 SKILL.md

**前端 `openSkillModal`**：
- 按 `skill.source` 分发：
  - hardcoded / ref-auto → 旧 `/api/reference`（embed standards/）
  - gsd-framework → 新 `/api/skill-disk-file`（磁盘）
- Modal 副标题显示完整磁盘路径（gsd-framework）或 `references/<name>`（其它）

### 团队同事使用流程

1. 装 v1.7.40 + npx 装好 GSD 框架（v1.7.39 已经能做了）
2. App 启动 → 侧栏「规范模块」
3. 看到 82 张卡（12 团队规范 + 4 ☁️ 云同步 + 66 💪 GSD）
4. 点任一💪卡 → modal 显示完整 SKILL.md 内容

### Notes

- 仅改 `catalog.go`（+120 行：扫 GSD + frontmatter 解析 + sort）、`main.go`（注册新路由）、`web/app.js`（modal 分发，+10 行）
- 不动 gsd_install.go（v1.7.39 的安装/卸载逻辑保留）
- 编译过 + smoke test 跑通（82 张卡片正确分组）
- 后续可加：UI 按 group 折叠分组（避免一页 82 张卡片太挤）/ 卡片搜索过滤

---

## [1.7.39] — 2026-05-14

### Changed — GSD skill 换成上游 gsd-build/get-shit-done 框架

之前 v1.7.32~38 我们 ship 了一份**简版任务清单 GTD** skill（inbox/next-actions/projects 等 6 清单 + `/gsd capture` 之类命令）。
v1.7.39 起：**整体替换**为上游 [gsd-build/get-shit-done](https://github.com/gsd-build/get-shit-done) 框架。

#### 上游是啥

不是任务清单——是个 **spec-driven development 框架**（59 个 gsd-* skill）：
- 完整 project / milestone / phase 生命周期管理
- 上下文工程层，解决 Claude context rot
- `new-project → discuss-phase → plan-phase → execute-phase → ship` 主循环
- 已被 Amazon / Google / Shopify / Webflow 工程师采用

我们的简版只是任务清单 GTD，规模和定位都远不如。换上游版让团队拿到真正能驱动 Claude 做项目的工具。

#### 替换细节

**删除**：
- `claude/gsd/` 整个文件夹（旧 SKILL.md + 4 个 references）
- `claude/commands/gsd.md`（旧 `/gsd` slash command）
- `commands_install.go` 里 `gsd` 注册条目
- 一键更新逻辑里 `/api/gsd/install` 调用

**重写** `gsd_install.go`：

- `GET /api/gsd/status`
  - 扫 `~/.claude/skills/gsd-*` 看装了几个
  - 比对核心 6 skill（`gsd-new-project` / `gsd-discuss-phase` / `gsd-plan-phase` / `gsd-execute-phase` / `gsd-help` / `gsd-update`）
  - 检测 `npx` / `node` 是否可用（用 `bash -lc command -v npx` 兼容 nvm/asdf）
  - 检测老版残留（`~/.claude/skills/gsd/` 含 "Getting Shit Done · 任务清单" frontmatter）

- `POST /api/gsd/install`
  - 跑 `bash -lc 'npx --yes get-shit-done-cc@latest --claude --global 2>&1'`
  - 自己控 8 分钟超时
  - npx 不可用 → 返回 "请先 brew install node"
  - 返回完整 stdout/stderr 给前端看

- `POST /api/gsd/uninstall`
  - 删 `~/.claude/skills/gsd-*` 所有目录
  - 默认连老版 `~/.claude/skills/gsd/` 一起删（`include_old_simple: true`）
  - 也清 `~/.cursor/skills-cursor/gsd-*`

**前端 GSD 卡片重写**：
- 紫色左边框（标识 v2 上游版）
- 状态显示装了几个 gsd-* skill、Node 版本、是否检测到老版残留
- 「🚀 一键装（npx）」 - 弹确认 → 8 分钟阻塞 → 看输出
- 「📋 复制安装命令」 - 让用户在终端跑（npx 不可用时的退化）
- 「🗑 卸载全部 gsd-*」 - 双击确认 + 一并清理老版

### 团队同事使用流程

```
打开 App → ⚡ 安装 → 进阶 Skill → 「💪 GSD 框架」紫卡
    ↓
看到「⚠ 找不到 npx」？先装 Node.js：brew install node
    ↓
点「🚀 一键装（npx）」 → 等 1-3 分钟
    ↓
看到 ✓ 装了 59 个 gsd-* skill → Cmd+Q 重启 Claude Code
    ↓
试一下：/gsd-help 看可用命令
```

或者自己开终端跑：

```bash
npx --yes get-shit-done-cc@latest --claude --global
```

### 老版用户怎么办

升级 v1.7.39 装 DMG 后，老版 `~/.claude/skills/gsd/`（v1.7.38 起留下的）**不会被自动覆盖**——和上游 `gsd-*` skill 命名空间不冲突，可以共存。

如果想清理：在 App 里点「🗑 卸载全部 gsd-*」，会一并删老版 `gsd/`。

**个人数据保留**：`~/Library/Application Support/TeamStandards/gsd/`（你的 inbox/next-actions/etc 清单）**不删**。要彻底清手工 `rm -rf` 即可。

### Notes

- 不再 embed `claude/gsd/` 到二进制（DMG 略瘦一点点）
- `claude/commands/gsd.md` 移除——上游框架自己装 `/gsd-*` 系列命令
- `gsd_install.go` 从 ~210 行 → ~190 行（行数差不多，逻辑完全换了）
- 编译过
- ⚠️ 因为 npx 阻塞 8 分钟，UI 那段时间显示"安装中"按钮 disabled，没进度条（npx 自己输出在最后一次性给）

---

## [1.7.38] — 2026-05-14

### Added — 一键自升级（L2 半自动 OTA）+ HTTP 代理配置

之前 v1.7.34 只有 L1 通知 + 浏览器跳转下载。本版加 L2：**点一个按钮 → App 自动下载 + 挂载 + 拷贝 + 重启**，全程不用碰命令行。

#### 流程

```
点「🚀 一键升级」
    ↓
POST /api/app-update/auto-install  → 后端 goroutine 开跑
    ↓
1. 定位当前 .app bundle（os.Executable 向上找 .app）
2. 拉 GitLab Release 最新 DMG URL
3. 下载到 /tmp/team-standards-update.dmg（带 progress reader）
4. embed 的 upgrade-helper.sh 写到 /tmp
5. spawn detached helper：bash helper.sh <PID> <DMG> <APP> <LOG>
    ↓
前端 polling /api/app-update/progress（500ms 一次）
显示下载进度条 + MB/MB + 阶段
    ↓
phase=ready 后弹「3 秒后 App 自动退出」倒计时
    ↓
window.appQuit() → w.Terminate() → App 退
    ↓
Helper（独立 PID，detached + Setpgid）接手：
  1. 等父进程退出（最多 60s）
  2. 备份当前 .app 到 .bak-<timestamp>
  3. hdiutil mount -nobrowse -readonly DMG
  4. find DMG 里的 .app
  5. cp -R 到 /Applications/
  6. hdiutil detach
  7. xattr -cr 解 Gatekeeper
  8. open 启动新 App
  9. 任意步骤失败 → 回滚旧 .app
    ↓
新 App 启动，正常运行
```

#### 安全保障

- ✅ **失败自动回滚**：helper 任何一步挂掉，自动 mv 备份回原位置，确保 App 不会半死
- ✅ **备份保留**：.bak-<timestamp> 不立即删，用户可手动 rm 或回滚
- ✅ **细致 logging**：全程写 `/tmp/team-standards-upgrade.log`，失败时前端给链接看
- ✅ **检查 PID 有效**：父进程没退就不动手
- ✅ **路径在 $HOME 或 /Applications/**：os.Executable() 给出真实路径，不会写错地方
- ❌ **不强制升级**：用户主动点才走
- ❌ **不后台静默**：全程进度条可见

#### 新增 `scripts/upgrade-helper.sh`（~135 行 bash）

完全自包含的 detached 升级脚本，embed 进 App 二进制（`//go:embed all:scripts`）。

逻辑：等 PID 退出 → 备份 → mount → find .app → cp → detach → xattr → open → 失败任意步骤 rollback。

#### 新增后端 `app_update.go` L2 部分

- `POST /api/app-update/auto-install` — 启动升级流程（return 立即，后台 goroutine 跑）
- `GET /api/app-update/progress` — 查询进度 `{phase, percent, bytes_done/total, message, error, log_path}`
- 阶段：`idle → downloading → ready → failed`（前端 polling 500ms）
- 用 `progressReader` 包 `http.Body` 实现下载进度回调
- `currentAppBundlePath()` 通过 `os.Executable()` 向上找 .app 后缀

#### 新增前端「🚀 一键升级」流程

- 升级 modal 主按钮换成蓝紫渐变的「🚀 一键升级」（旧的「💽 下载 DMG 手工装」降级为 ghost 按钮）
- 新 `#upgradeProgressModal`：进度条 + 阶段 + bytes 显示 + 失败时给 log 链接
- ready 阶段 3 秒倒计时后自动调 `appQuit()` 退 App

### Added — HTTP 代理配置（全局）

**痛点**：内网域名 `your-gitlab.com` 在某些网络环境下不通，App 所有 HTTP 出站请求挂掉。

**方案**：加 `~/Library/Application Support/TeamStandards/proxy.json`（mode 0600），影响**所有** HTTP 出站：
- 规范云同步（v1.7.30）
- App 自动更新检查（v1.7.34）
- L2 一键升级下载（v1.7.38）

**优先级**：
1. App 配置的 `proxy_url`（最高，UI 设置）
2. 系统 `HTTPS_PROXY` / `HTTP_PROXY` env var
3. 直连

**API**：
- `GET /api/proxy/config` — 读
- `POST /api/proxy/config` — 写 + 立即生效（重建 `standardsHTTPClient`）

**UI**：「⚡ 安装 → 更新与同步 → ☁️ 规范云端同步」卡顶部加「🌐 HTTP 代理（内网不通时配）」折叠区。勾选 + 填 URL（如 `http://127.0.0.1:1087`）保存即生效。

### 团队同事使用流程

**正常升级（推荐）**：
1. App 启动后 banner 弹「🆕 v1.7.X 可用」
2. 点「查看更新」 → 弹 modal
3. 点「🚀 一键升级」
4. 看进度条 → App 自动重启 → 完事

**网络不通时**：
1. 启动前 set 环境变量：`HTTPS_PROXY=http://127.0.0.1:1087 open /Applications/Team\ Standards.app`
2. 或在 App 里「☁️ 规范云端同步 → 🌐 HTTP 代理」勾选 + 填 URL 保存
3. 之后正常一键升级

**失败回退**：
- 升级 modal 「💽 下载 DMG 手工装」按钮还在，浏览器接管下载
- helper 失败会自动回滚旧 .app，App 不会半死

### Notes

- 新增：`scripts/upgrade-helper.sh`（135 行 bash）
- 改：`app_update.go`（+200 行：L2 流程 + progress + bundle path 探测）、`standards_sync.go`（+70 行：proxy 配置 + client 构建）、`main.go`（+4 路由）、`web/index.html`（升级进度 modal + 代理折叠区）、`web/app.js`（+130 行）
- 编译过，DMG 16 MB
- ⚠️ **首次测试场景**：升级 v1.7.37 → v1.7.38 时验证整个 L2 流程能跑通——这是首次实战
- **GitLab Release 推送**：需要 `glab auth login --hostname your-gitlab.com` 一次性配 token，然后 `bash upload-release.sh` 推

---

## [1.7.37] — 2026-05-14

### Added — 提交规范化（A · pre-commit hook + CI gate）

**痛点**：Skill 是 advisory layer，AI 写完代码用户能手改成 float64 / user_id / SELECT * 照样提交。要真"限制"需要硬关卡。

**架构**：

```
Layer 1 · AI Skill（现有）—— 软约束（go-team-standards / code-review）
   ↓ 用户可绕过
Layer 2 · pre-commit hook（本版新加）—— 本地 commit 前 grep 检查
   ↓ git commit --no-verify 可跳过
Layer 3 · GitLab CI gate（本版新加）—— MR push 后 CI 跑检查，阻断 merge
```

#### 核心脚本 `scripts/team-standards-check.sh`

bash + grep 实现 **16 条核心规则检查**：

**铁律类（5 条）**
- 铁律 #1 · 硬编码密钥（sk-xxx / kgb_xxx / Bearer / password=）
- 铁律 #2 · errors.New 不走 xerror
- 铁律 #3 · `_ = err` 丢弃错误
- 铁律 #6 · 金额 float（精度丢失）
- 铁律 #9 · SELECT *
- 铁律 #10 · 数据库外键
- 铁律 #12 · 敏感数据进日志

**命名类（9 条）**
- §1.2 · user_id → uid（.go / .sql / .proto）
- §1.5 · txid / transaction_hash → tx_hash
- §2.1 · 裸 time/timestamp/ts 列
- §2.2 · gmt_create / create_time → created_at
- §4.2 · 裸 amount 列需业务前缀
- §4.3 · vol/size/amt 缩写
- §5.1 · 布尔字段非 is_/has_/can_ 前缀
- §5.2 · is_deleted BOOLEAN → deleted_at
- §8 · ip_addr / login_ip → client_ip

**输出**：分 P0 / P1 / P2 三级
- P0 / P1 → 退出码 1（拒绝 commit）
- P2 → 退出码 0（仅警告）

**模式**：
- `bash check.sh`（无参）→ 扫 git staged 文件（pre-commit 模式）
- `bash check.sh file1.go file2.sql`（显式）
- `bash check.sh --all` → 扫整个 git 仓库

**跳过文件**：`vendor/` / `node_modules/` / `*.gen.go` / `*.pb.go` / `*_mock.go` / `*.lock`
**跳过场景**：注释行的 SELECT *（粗略，行首 `--` 或 `//`）

#### 配套脚本

- `scripts/install-precommit.sh` —— 装 hook 到指定项目
- `scripts/gitlab-ci-snippet.yml` —— CI gate 模板，拷到项目 `.gitlab-ci.yml` 即可
- `scripts/README.md` —— 三层防御架构 + 各规则说明 + 紧急跳过

#### App 后端 `commit_guard.go`（5 API）

- `GET /api/commit-guard/status?path=...` — 检查项目状态（是 git? hook 装了? script 装了?）
- `POST /api/commit-guard/install` — 一键装：拷 scripts/team-standards-check.sh + 写 pre-commit hook，原 hook 自动备份为 `.bak-<timestamp>`
- `POST /api/commit-guard/uninstall` — 删 hook + script（仅删 team 装的）
- `POST /api/commit-guard/check` — 在指定项目跑一次 `--all` 扫描，返回输出
- `GET /api/commit-guard/scripts` — 返回 embedded 4 个脚本内容（前端预览 / 拷贝）

**安全检查**：
- 项目路径必须在 `$HOME` 内（拒绝 `/Applications/` 等系统路径）
- 必须存在 `.git/` 目录（确认是 git 仓库）
- 不覆盖原有 pre-commit hook（自动备份 `.bak-<timestamp>`）

#### App 前端：「📋 提交规范化」卡片

位置：「⚡ 安装」 → 「更新与同步」 subpage，红色左边框（标识强约束）。

UI：
- 项目路径输入框 + 🔍 检查按钮（实时显示是否 git 仓库 / hook 已装 / script 已装）
- 📥 装 / 🧪 试跑（--all 全仓扫）/ 🗑 卸载 三按钮
- 折叠区显示 GitLab CI YAML 片段，一键复制

### Added — Skill 测试用例覆盖（B · 命名规范 + feature-flags）

`testcases.go` 在原有 20 个 case 基础上加 **10 条新 case**：

**命名规范（7 条）**：
- `naming-uid` · §1.2 user_id → uid
- `naming-created-at` · §2.2 gmt_create → created_at
- `naming-soft-delete` · §5.2 is_deleted → deleted_at
- `naming-amount-prefix` · §4.2 裸 amount 业务前缀
- `naming-bool-prefix` · §5.1 布尔 is_/has_/can_ 前缀
- `naming-bare-version` · §6.3 业务 version 字段前缀
- `naming-client-ip` · §8 client_ip 统一

**特性开关（3 条）**：
- `ff-default-on` · §2.3 新 FF 默认 OFF
- `ff-no-ctx` · §2.2 走 go-common FF 客户端禁自实现
- `ff-commit-format` · §5 commit 必带 [reqID] (@user)

**总测试 case 数：20 → 30**。

测试 Skill tab 不需要改 UI——现有按 category 渲染会自动包含 `naming` / `feature-flags` 两个新分类。

### 团队同事使用流程

**装 hook（Layer 2）**：
1. 打开 App → 「⚡ 安装」 → 「更新与同步」 → 「📋 提交规范化」卡
2. 填项目路径 `~/your-go-service`，点 🔍 检查
3. 看到「✓ 是 git 仓库」→ 点「📥 装 pre-commit hook」
4. 完事。下次 `git commit` 时自动跑

**加 CI gate（Layer 3）**：
1. 同卡片点开「📜 GitLab CI gate」折叠
2. 点「📋 复制 YAML」
3. 粘到自己项目 `.gitlab-ci.yml`，提交
4. MR 时 CI 自动跑，违规阻断 merge

**测试 skill 是否生效**：
1. 「🧪 测试 Skill」tab
2. 「运行全部 30 条测试」
3. 看 pass/fail 分布——每条规则一个 case，覆盖 14 铁律 + 命名规范 + feature-flags

### Notes

- 新增：`scripts/team-standards-check.sh` (~460 行) + `scripts/install-precommit.sh` + `scripts/gitlab-ci-snippet.yml` + `scripts/README.md`
- 新增：`commit_guard.go` (~280 行)
- 改：`main.go`（embed `scripts/*` + 注册 5 路由）、`testcases.go`（+10 case）、`web/index.html`（提交规范化卡片）、`web/app.js`（~120 行）
- 编译过；scripts 一并 embed 进 DMG 二进制（无需远程拉取）
- **网络限制说明**：本次 release 因到 `your-gitlab.com` 网络不通暂时没推 GitLab Release——用户可在网络恢复后跑 `bash upload-release.sh` 推上去

---

## [1.7.36] — 2026-05-14

### Fixed — 「规范模块」卡片数量不随云同步增加（11 卡死循环）

**背景**：用户反馈装完 v1.7.35 拉了云同步规范，但「规范模块」tab 仍显示 11 张卡。

**根因**：`catalog.go` 的 `skills` slice 是**写死**的 11 条 entry，UI 直接渲染这个 slice；云同步只覆盖 `references/*.md` 文件内容，跟 UI 列表完全无关。

**双管齐下修复**：

#### A · 补硬编码条目「特性开关 & 分支管理」

加第 12 条 entry 到 `skills` slice：
```go
{
    ID: "feature-flags", Title: "特性开关 & 分支管理",
    Description: "FF SDK 集成 + IDP 注册 + GitOps 分支策略 + 5-to-4 预览环境 + Conventional Commits + 治理审计",
    Triggers: []string{"特性开关 / feature flag", "灰度 / 发布 / 回滚", "PR / MR / 分支策略"},
    Reference: "feature-flags.md", Scope: "conditional",
},
```

#### B · 动态扫描 references/ 自动补 ☁️ 卡片（治本）

新增 `scanExtraReferences()` 函数：每次 `/api/catalog` 请求时扫一遍 `~/.claude/skills/go-team-standards/references/`：

- 把**没在硬编码 slice 里**的 .md 文件自动生成卡片
- 跳过明显的 example / template / custom-* 文件（避免噪音）
- 从文件 frontmatter `title:` 字段取标题；缺失则用 filename 兜底
- 加 `☁️` 前缀让用户一眼识别「这是云端来的，非内置」

**效果**：以后规范 owner 推任何新 `.md` 到 `len/standards` 公开仓 → 团队成员 App 里**拉取最新规范**后 → 重启 App → 「规范模块」自动多一张 ☁️ 卡，**无需发新 DMG**。

#### 升级后预期看到的卡片数

旧版（v1.7.35 之前）：11 张
新版（v1.7.36）：**16 张**
- 12 张硬编码（11 旧 + 1 新加 feature-flags）
- 4 张 ☁️ 自动发现：`code-review.md` / `deployment-checklist.md` / `field-naming.md` / `meeting-minutes.md`

这些 ☁️ 卡片之前一直在 `references/` 里 AI 也能读到，只是 UI 没展示。现在 UI 跟上了。

### Notes

- 仅改 `catalog.go`（+90 行：动态扫描 + frontmatter 解析）
- 不动后端其它部分、前端、规范文件
- 编译过
- `extractFrontmatterTitle` 简易 YAML 解析，仅识别 `title:` 一行；frontmatter 必须以 `---` 包裹

### 团队同事使用流程

1. 装 v1.7.36 DMG（或等自动更新 banner 弹出）
2. 「立即安装」覆盖装 skill
3. 打开「规范模块」 → 看到 16 张卡（含 ☁️ 自动发现的 4 张）
4. 以后规范 owner 在 `len/standards` 公开仓推新 `.md` → 同事点「📥 拉取最新规范」 → 重启 App → 卡片自动多

---

## [1.7.35] — 2026-05-10

### Added — Claude Desktop / Cowork 直装支持（路径自动探测 + 覆盖安装）

**问题**：用户反馈 Claude Desktop 和 Cowork 看不到本地 skill——因为它们各自的 skill 目录跟 Claude Code（`~/.claude/skills/`）和 Cursor（`~/.cursor/skills-cursor/`）不一样。

之前唯一支持是「📤 给 Claude Desktop 用」卡片导出 zip 让用户**手工**通过 web UI 上传——繁琐，每次发新版都要操作一遍。

**新方案**：自动探测系统中可能的 Claude Desktop / Cowork skill 目录 → 列出来 → 一键覆盖安装。

#### ⚠️ 路径准确性诚实声明

**作者没在装好 Cowork 的机器上验证过具体目录名**。所以做法是：

1. **多候选探测**（8 个候选路径，覆盖 Claude Desktop + Cowork 各种常见命名）
2. **额外扫描** `~/Library/Application Support/` 和 `~/Library/Containers/` 下所有含 `claude` / `cowork` / `anthropic` 的目录，列给用户参考
3. **支持自定义路径**（手填，必须在 $HOME 内）
4. UI 上每个候选标注「父目录存在 ✓」「已有 skills/ ✓」「可写 ✓」三个信号——优先勾这些都有的

#### 候选路径清单

**Claude Desktop**（3 个候选）：
- `~/Library/Application Support/Claude/Skills` ← 最可能
- `~/Library/Application Support/Claude/skills` ← 大小写变种
- `~/Library/Containers/com.anthropic.claudefordesktop/Data/Library/.../Skills` ← App Sandbox

**Cowork**（5 个候选）：
- `~/Library/Application Support/Cowork/skills`
- `~/Library/Application Support/Anthropic/Cowork/skills`
- `~/Library/Application Support/Claude/cowork/skills`（如果是 Claude 内嵌）
- `~/Library/Containers/com.anthropic.cowork/.../Skills`（沙箱）
- `~/.cowork/skills`（CLI 风格）

#### 新增后端 `cowork_install.go`（2 API）

- `GET /api/cowork/probe` — 探测所有候选 + 扫 `~/Library/` 找 Anthropic 相关目录
  - 返回 `{candidates: [...], discovered_dirs: [...], home: "..."}`
  - 每个 candidate 标 `parent_exists` / `exists` / `writable`
- `POST /api/cowork/install` — 把指定 skill 装到选中路径
  - 默认装全部 5 个 skill（go-team-standards / orangecat / dev-dna / code-review / gsd）
  - **安全检查**：所有 path 必须在 `$HOME` 内（防止误写到系统目录）
  - 自动创建目录 + 复用 `installEmbeddedSkill`

#### 新增前端 UI

「打包分发」subpage 顶部新加 紫色边框 卡片：

- 「🔍 探测路径」按钮 → 列出所有候选 + discovered 目录
- 每个候选带 checkbox（推荐选项默认勾上：父目录存在 + 可写）
- 紫色 tag 区分 `Claude Desktop` vs `Cowork`
- 折叠区显示扫到的所有 Anthropic 相关目录（用户可参考决定哪个对）
- 自定义路径输入框（如果候选都不对，手填）
- 「📥 安装到勾选路径」按钮 → 批量装

旧的「📤 给 Claude Desktop 用」zip 上传卡保留作为 fallback（如果直装不工作或目录不存在）。

#### 团队同事使用流程（Cowork 看不到 skill 时）

1. 打开 App → 安装 → 「打包分发」
2. 第一张卡「📡 Claude Desktop / Cowork 直装」
3. 点「🔍 探测路径」
4. 看到带「父目录 ✓ 可写 ✓」标记的候选，已勾选 → 点「📥 安装」
5. 如果没找到合适候选，点开「🔍 扫到的目录」看实际存在啥
6. 实在不行用 fallback 卡导 zip 手工传

**Cmd+Q 重启 Claude Desktop / Cowork** 才能加载新装的 skill。

#### 隐式约束

- 写入路径必须在 `$HOME` 内（拒绝 `/Applications/` 之类系统路径）
- 探测不会修改任何文件（只读）
- `installEmbeddedSkill` 走 `fs.WalkDir`，确保所有 reference 文件都复制

### Notes

- 新增：`cowork_install.go`（~210 行）
- 改：`main.go`（2 路由）、`web/index.html`（探测卡）、`web/app.js`（~115 行：探测渲染 / 安装 / 自定义路径）
- 编译过
- **请装完后试一下并反馈**：探测到的哪些路径实际有效？这样我可以把候选清单调整得更准。

---

## [1.7.34] — 2026-05-10

### Added — App 自动更新（L1 通知 + 引导 / 公开 GitLab Releases）

之前每次发版都要去 IM 群通知"v1.7.x 出了，自己去拉最新 DMG"，团队同事很容易错过。
本版加上**启动时静默检查 + 顶部蓝色 banner 提示**。

#### 架构（4 决策已落定）

| 决策项 | 选择 |
|---|---|
| DMG 放哪 | **A · 复用 `len/standards.git` 公开仓的 GitLab Releases** |
| 自动化程度 | **L1 · 通知 + 引导**（用户点下载链接，浏览器接管，自己手动装） |
| 触发时机 | App 启动后 1 秒后台 check + 用户手动按钮 |
| 跳过版本机制 | **支持**（点「跳过此版本」永久不再提示该版本） |

为什么是 L1 不是 L2/L3：
- L2/L3 在 ad-hoc 签名场景下风险高（自更新后 Gatekeeper 重新隔离需 xattr 操作）
- macOS 运行中的 .app 无法 atomic 替换自己（需要外部 helper 进程）
- L1 风险最低，先上线验证流程，后面看反馈再升级

#### 新增 `upload-release.sh`（手工发布脚本）

```bash
# 用法：发版后跑
bash release.sh           # 出 DMG 到 ../version/
bash upload-release.sh    # 推 DMG + changelog 到 len/standards 的 GitLab Releases
```

**前置**：
- 装 glab CLI：`brew install glab`
- 登录：`glab auth login --hostname your-gitlab.com`

**自动做的事**：
- 找最新 DMG（按 VERSION 匹配）
- 算 SHA256
- 从 CHANGELOG.md 提取当前版本节作 release notes
- 末尾自动加 xattr 安装提示
- 用 glab 创建 release `v1.7.x` 到 `len/standards`
- 上传 DMG 作为 release asset
- 输出 release 页 URL 和直链下载 URL

#### 新增后端 `app_update.go`（4 API）

- `GET /api/app-update/check` — 拉 GitLab Releases API 最新 release，对比本地 `VERSION`，返回 `{has_update, current_version, latest_version, release_url, dmg_url, release_notes, is_skipped, ...}`
- `POST /api/app-update/skip` — 把版本写入跳过清单，下次 check 不再弹 banner
- `POST /api/app-update/unskip` — 取消跳过（单个或全部）
- `POST /api/app-update/open-download` — 用系统 `open` 命令在浏览器打开 DMG URL（用户接管下载）

**配置存储**：`~/Library/Application Support/TeamStandards/app-update.json`（mode 0600）
含：`release_project` / `release_host` / `last_checked_at` / `last_seen_version` / `skipped_versions[]`

**启动后台 goroutine**：`appUpdateBackgroundCheck()` 不阻塞主流程，失败静默不报。

#### 新增前端 banner + modal

- **顶部 banner**（蓝紫渐变）：`🆕 Team Standards v1.7.34 可用 · 当前 v1.7.33  [查看更新] [稍后] [跳过此版本]`
- 点「查看更新」 → modal 弹出：
  - 版本 + 发布时间
  - changelog markdown 渲染（从 GitLab release notes 拉）
  - xattr 安装命令独立卡片提示（防忘记）
  - 3 按钮：`💽 下载 DMG（浏览器打开）` / `📋 复制 xattr 命令` / `跳过此版本`

#### 隐式约束

- 启动后 1 秒才发请求（不阻塞 UI 加载）
- 网络不通**不弹任何东西**（静默失败）
- 已跳过版本**不弹**
- 当前版本与 latest 相同**不弹**
- 比较走 semver 形如 `v1.7.34`，容错按字符串

#### 团队同事新流程

1. 装现有 v1.7.x → 下次启动后台自动检查
2. 看到顶部 banner「🆕 v1.7.40 可用」
3. 点「查看更新」 → 看 changelog → 点「💽 下载 DMG」
4. 浏览器开始下载 → 拖到 Applications → 跑 xattr -cr → 重启 App
5. 不想要这版？点「跳过此版本」永久不再提示

### Notes

- 新增：`upload-release.sh` (~95 行) + `app_update.go` (~285 行)
- 改：`main.go`（注册 4 路由 + 后台 check）、`web/index.html`（banner + modal）、`web/app.js`（启动检查 + UI 逻辑 ~95 行）、`web/style.css`（update-banner 样式）
- 编译过；首次发版需 `bash upload-release.sh` 才能让 App 检测到
- **CI 自动 release 留作后续优化**（需要在 GitLab 配 cross-project deploy token）
- v1.7.34 本身需要先手工发到 release 仓，App 启动 check 才会看到自己

---

## [1.7.33] — 2026-05-09

### Changed — `gsd` Skill 收紧到 v2.0：仅 `/gsd` 命令触发 + 输出统一为 GFM 任务清单

v1.0 上线后立刻收到反馈："被动模糊触发太烦"。v2.0 大幅收紧。

#### 触发协议（铁律）

**唯一触发方式**：用户敲 `/gsd <子命令>`。

明确**禁止被动激活**：
- ❌ 用户说"今天事好多 / 任务太多 / 帮我理一下" → **不要**主动 capture
- ❌ 用户说"现在干啥" → **不要**主动跑 next 推荐
- ❌ 用户写代码 / 写文档时 → **绝不**插话"要不要把这个入 GSD"
- ❌ code-review 输出 P0/P1 后 → **不要**自动建议入 GSD（用户自己敲 `/gsd capture`）

判定规则：**只有当用户当前消息第一行是 `/gsd ...` 时**才激活。

#### 输出统一：GFM 任务清单 markdown

所有产出走标准 GitHub Flavored Markdown：

```markdown
- [ ] @电脑 ⚡⚡⚡ 30min 截止 5/12 看完《订单分库》设计评审
- [ ] @会议 ⚡⚡ 本周 找 DBA 商量 wallet 表分区
- [x] ~~给 QA 回复 v1.7.30 问题~~（2 分钟规则，已做）
```

字段顺序固定：
```
- [ ] @<context> ⚡<1-3> <时长> [截止/期望/本周/本月/someday] 任务描述
```

**禁止**：
- 自创复杂语法（`@CTX[level=3]` 之类）
- 加非标记 emoji 装饰（` - [ ] 🚀 任务` 这种）
- HTML 表格（用纯 markdown 列表）

#### 改了哪些

- `claude/gsd/SKILL.md` 从 ~310 行 → ~250 行
  - description 砍掉所有"模糊关键字触发"清单
  - 头部加 §0「触发协议（铁律）」明确禁止被动激活
  - §2 输出格式硬约束 GFM `- [ ]`
  - §6 禁止行为榜首：「当前消息不是 `/gsd ...` 开头时走 GTD 流程 = 违规」
- `claude/gsd/references/examples.md` 重写
  - 所有例子改成 `/gsd <子命令>` 形式
  - 新增 4 条反例明确"不该被动激活"的场景
- `claude/commands/gsd.md` 描述更新强调"仅本命令触发"

#### 不变的

- 数据落盘位置 `~/Library/Application Support/TeamStandards/gsd/`
- 6 大清单文件（inbox / next-actions / projects / waiting / someday / reference）
- 后端 4 API（status / install / uninstall / list）
- 前端 GSD 卡片（进阶 Skill subpage）
- 末尾标记 💪
- 隐私铁律（不导出 / 不 share）

#### 团队同事使用流程（v2.0）

```bash
# 想 dump 一件事
/gsd capture 周二记得问 Lily SDK 升级兼容

# 想理清 inbox
/gsd clarify

# 想看现在干啥
/gsd next
# AI 问你 @context + ⚡ + 时长，过滤推荐 1-3 件

# 周五跑复盘
/gsd review

# 看在等谁
/gsd waiting

# 看清单总览
/gsd
```

**没敲 /gsd 时 AI 不会动 GSD 流程**——可以放心写代码、写文档、聊天，不会被任务管理打扰。

### Notes

- 仅改 `claude/gsd/SKILL.md` + `claude/gsd/references/examples.md` + `claude/commands/gsd.md` + VERSION + CHANGELOG
- 后端 / 前端 / 一键更新逻辑全部不动
- 编译过

---

## [1.7.32] — 2026-05-09

### Added — `gsd` Skill (Getting Shit Done · 个人 GTD 流程)

第 5 个 Skill 上线。把 David Allen 的 GTD 经典 5 阶段（Capture / Clarify / Organize / Reflect / Engage）落到开发者日常，跟现有 4 个 Skill 联动协作。

**核心场景**：
- 任务太多没头绪 → Capture 全 dump 到 inbox
- 写完代码一堆 TODO 散落 → 自动从 code-review 输出转入 next-actions
- PR 等 merge / 等审批 → 进 waiting.md
- 周复盘扫一遍系统不让烂 → /gsd review 走标准 7 步模板

**触发设计**（学 code-review 的精细化）：
- ✅ 强触发：动作词（capture / inbox / 拆任务 / next action）+ 求助型（今天做啥 / 任务太多 / 没头绪）+ 复盘型（周复盘 / weekly review）+ 状态型（在等谁 / 有哪些项目）
- ✅ 显式触发：`/gsd` `/gsd capture <事>` `/gsd clarify` `/gsd next` `/gsd review` `/gsd waiting`
- ❌ 不触发：写代码时 / 写文档时 / 纯知识问 / 用户已说"先不管任务" / 同段已 review 过未变
- 🤔 拿不准时倾向不触发（学 code-review 的成功经验）

**输出三模式自动决策**（覆盖不同场景）：
- 模式 A · Capture：原样吸收 inbox，**不评判 / 不分类 / 不重排**（GTD 核心铁律）
- 模式 B · Clarify：5 步流水线（可执行？2 分钟？单步还是项目？）按 next-actions / projects / waiting / someday / reference 分流
- 模式 C · Engage（"现在干啥"）：按当前 @context + ⚡ 精力 + 可用时长过滤推荐 1-3 件

**核心标签体系**：
- 上下文（必填）：@电脑 / @会议 / @电话 / @外出 / @阅读 / @脑暴 / @等待
- 精力等级（必填）：⚡⚡⚡ 高 / ⚡⚡ 中 / ⚡ 低
- 时长估计（推荐）：5min / 15min / 30min / 1h / 半天 / 全天
- 截止日期（仅必要时用）：截止 X / 期望 X / 本周 / 本月 / someday

**与其它 Skill 的协作**（差异化设计的关键）：
- `code-review` 输出 P0/P1 违规 → 自动建议入 GSD next-actions（带优先级 + 置信度）
- `dev-dna` 提供个人精力曲线 → GSD 推荐 next action 时按时段过滤
- `go-team-standards` 写代码任务时 → 末尾联合标记 🌟💪
- 不与 `orangecat` 冲突（提测是 QA 文档，GSD 是个人任务流，关注点不同）

**数据落盘**：
- 目录：`~/Library/Application Support/TeamStandards/gsd/`（mode 0700）
- 6 大清单：inbox.md / next-actions.md / projects.md / waiting.md / someday.md / reference.md
- 周复盘归档：weekly-reviews/2026-W19.md 等
- 卸载 Skill 不删数据（保护用户隐私级别同 dev-dna）

### Added — 后端 `gsd_install.go` (4 个 API)

- `GET /api/gsd/status` — 装好了没 + 各清单条数统计（inbox 超 10 条会标黄提醒 clarify）
- `POST /api/gsd/install` — 装 Skill 到 Claude + Cursor + 初始化 6 个清单文件（已存在不覆盖）
- `POST /api/gsd/uninstall` — 删 Skill 文件，**保留数据目录**（数据是用户的）
- `GET /api/gsd/list?name=<file>.md` — 只读预览清单内容（白名单 6 个标准清单）

### Added — Slash command `/gsd`

第 6 条 Slash Command，路由 6 个子命令：空 / status / capture / clarify / next / review / waiting。

`commands_install.go` 注册表已加 `/gsd` 条目。

### Added — 前端 GSD 卡片（进阶 Skill subpage）

位置：「⚡ 安装 → 进阶 Skill」最末尾，紧接 dev-dna 卡。

UI：
- 状态行实时显示 6 大清单条数（inbox 高亮提醒 / next-actions 倒数前 3）
- 4 个按钮：📥 独立安装 / 🗑 卸载（双击防误） / 📥 看 inbox / 🎯 看 next-actions
- 卸载有 4 秒倒计时确认，明示"数据保留，要彻底删手工 rm"
- 折叠展开「🔍 看清单内容」预览区

### Changed — 一键更新接入 GSD

`updateAllBtn` 并发列表从 5 个端点扩到 6 个：

```
/api/install              (go-team-standards + cursor rules)
/api/orangecat/install
/api/dev-dna/install
/api/code-review/install
/api/gsd/install          ← 新接入
/api/commands/install     (现包含 6 条 commands)
```

### 团队同事使用流程

1. 升 v1.7.32 → 点「立即安装」自动装齐 5 Skill + 6 Slash Commands
2. Claude / Cursor 里说"今天事好多，帮我理一下" → AI 自动走 capture → clarify
3. 或主动敲 `/gsd capture <一句话>` 把事丢进 inbox
4. 周五下午敲 `/gsd review` 走 7 步标准复盘
5. 数据全在本地，不联网不导出，跨电脑迁移直接拷 `~/Library/Application Support/TeamStandards/gsd/` 即可

### Notes

- 新增：`claude/gsd/SKILL.md` (~310 行) + `references/gtd-methodology.md` + `references/context-tags.md` + `references/weekly-review-template.md` + `references/examples.md`
- 新增：`gsd_install.go` (~210 行)、`claude/commands/gsd.md`
- 改：`main.go`（注册 4 路由）、`web/index.html`（GSD 卡片）、`web/app.js`（GSD 状态/装/卸/预览 + 一键更新接入，~110 行）
- `commands_install.go` slashCommands 注册表加 `gsd`
- `//go:embed all:claude` 自动覆盖新增 skill 文件夹，无需改 embed
- 编译过

---

## [1.7.31] — 2026-05-09

### Changed — 规范云端同步默认 URL 内置 + 首次启动自动配置

省掉团队同事 90% 的手动配置成本：

1. **`standards_sync.go` 加 `DefaultStandardsRepoURL` 常量** = `https://github.com/kary2999/standards.git`
2. **首次启动自动写入配置**：`standardsSyncBackgroundCheck()` 检测到 `RepoURL` 为空时，自动用默认 URL 写入 `~/Library/Application Support/TeamStandards/standards-sync.json`，**用户连「保存」都不用点**
3. **前端输入框预填**：`<input value="https://github.com/kary2999/standards.git">`，作为兜底（防 background goroutine 还没跑完用户就打开了 UI）
4. 改输入框 `<small>` 文字：明确「默认指向团队规范仓」+ 允许用户改成自己的 mirror

### 团队同事新流程（零配置）

之前 v1.7.30：装 → 配置 URL → 保存测试 → 拉取（4 步）
现在 v1.7.31：装 → **直接点「📥 拉取最新规范」**（1 步）

启动后台 goroutine 已经把默认 URL 写好，UI 加载时配置已就绪，可以直接拉。

### Notes

- 仅改 `standards_sync.go`（加常量 + 默认注入逻辑） + `web/index.html`（input value + small 文案）
- `~/Library/Application Support/TeamStandards/standards-sync.json` 已存在的用户配置**不会被覆盖**（只在 RepoURL 为空时注入）
- 想改成别的 mirror？UI 里改完点保存即可

---

## [1.7.30] — 2026-05-08

### Added — 规范云端同步（脱离 App 版本绑定）

**痛点**：以前规范一改就要重打 DMG 发全员重装；规范 owner 改个命名约定要等下一版 App。

**解决**：standards 源文件搬到 GitLab 公开 readonly 镜像仓库，App 启动时静默检查、用户手动一键拉取，**完全脱离 App 二进制版本**。

**架构（4 决策已落定）**：

| 决策项 | 选择 |
|---|---|
| 仓库形态 | 单独建公开 readonly mirror repo（与 App 仓库隔离） |
| 认证 | 公开仓库，无需 token（GitLab Visibility = Public） |
| 同步粒度 | 仅 `standards/*.md` 源文件 → 自动覆盖到 `~/.claude/skills/.../references/` 和 `~/.cursor/skills-cursor/.../references/` |
| 检查时机 | App 启动后台静默检查 + 用户手动按钮 |

**不动的**：SKILL.md / commands / cursor rules / App 二进制 —— 这些仍随 DMG 走。

**离线兼容**：检查 / 拉取失败不阻断 App 启动，回退使用内嵌的内置版本。

### Added — 后端 `standards_sync.go`（4 个 API）

- `GET /api/standards-sync/config` — 返回当前配置（repo URL / branch / last_synced_sha / last_checked_at / last_check_error）
- `POST /api/standards-sync/config` — 保存仓库 URL，**保存前先测试连通**（拉一次 commit SHA，失败拒绝保存）
- `GET /api/standards-sync/check` — 拉远端最新 commit SHA 与本地 last_synced_sha 比对，返回 `{has_update, current_sha, latest_sha}`
- `POST /api/standards-sync/pull` — 下载 archive.zip（GitLab `/-/archive/<branch>/<repo>-<branch>.zip` 端点）→ 解压 → 比对内容 → 仅写入有变化的文件 → 更新 last_synced_sha

**实现细节**：
- 配置存 `~/Library/Application Support/TeamStandards/standards-sync.json`（mode 0600）
- HTTP 客户端 15s 超时，50MB 响应硬上限
- 通过 GitLab v4 API（`/api/v4/projects/<urlencoded path>/repository/commits`）获取最新 SHA
- 通过 archive zip 端点一次性下载全仓库（公开仓不需 token）
- 解压时只保留 `.md` 和 `.png`，跳过 README.md 和 dotfiles
- 写盘时按字节内容比对，相同则跳过（避免无谓 IO）
- 同时覆盖 Claude + Cursor 两个 references 目录
- 启动时 `go standardsSyncBackgroundCheck()` 后台跑，刷新 last_checked_at 但**不主动拉**（避免突兀改动用户磁盘）

### Added — 前端「☁️ 规范云端同步」卡片

位置：安装页 → 「更新与同步」subpage 顶部，蓝色左边框突出显示。

UI 元素：
- **状态行**：未配置 / 已配置 + 上次同步信息 / 有新版本提示
- **配置 details**（默认折叠）：仓库 URL 输入框 + 分支输入框 + 「💾 保存并测试连通」按钮
- **行动按钮**：「🔍 检查更新」 / 「📥 拉取最新规范」
- **日志区**：拉取结果 + 变更文件列表（最多展示 20 条）

启动后自动加载状态 + 异步检查远端，发现新版本时直接把状态行变橙色显示「🆕 有新版本」。

### 团队同事使用流程（极简）

1. 装好 v1.7.30 DMG
2. 打开 App → 安装 → 「更新与同步」 → 点开「⚙️ 配置」
3. 填团队公开镜像 URL（如 `https://your-gitlab.com/team/standards`）
4. 点「💾 保存并测试连通」 → 通过后点「📥 拉取最新规范」
5. 之后每次启动 App 自动检查，有新版直接点「📥 拉取」即可
6. 拉完 Cmd+Q 重启 Claude / Cursor 生效

### Notes

- 新增文件：`standards_sync.go`（330 行）
- 修改：`main.go`（注册 4 个路由 + 后台检查 goroutine）、`web/index.html`（云同步卡片）、`web/app.js`（前端逻辑 ~135 行）
- 编译过；webview_go / lipo / codesign / hdiutil 仍是 macOS 专属，CI 行为不变
- **公开镜像仓库内容已准备**：在 `/tmp/team-standards-public/` 下，含全部 19 个 .md/.png + 配套 README.md（19 文件 ~636 KB）
- **用户 TODO**：在 GitLab 建公开仓库 + 推内容（命令见对话回复）

---

## [1.7.29] — 2026-05-08

### Changed — 数据库设计规范同步到 2026.05.08 版

源 `~/Desktop/work/Code/开发Skills/技术规范.2026.05.08/数据库设计规范.md` 覆盖 `standards/database.md`。

frontmatter 同步：`version: 1.1.0 → 1.2.0` · `last_modified: 2026-04-29 → 2026-05-08` · `source: 技术规范.2026.04.29 → 技术规范.2026.05.08`。

`claude/go-team-standards/references/database.md` 是 symlink 指向 `standards/database.md`，自动同步无需改动。

### 实质内容差异（4 处）

1. **PostgreSQL 版本基线 16.x → 18.x**（§2.1）
   - 影响：扩展兼容性（`pg_repack` / `pgaudit`）需复检；MERGE 语法、`SQL/JSON` 标准函数已 GA 可用

2. **用户字段命名彻底统一为 `uid`**（§3.5 / §3.6 / §4.3）
   - 残留示例 `user_id` 全部清除
   - 关联字段示例：`{关联表单数}_id` → 仅留 `order_id`，删除 `user_id`
   - 索引示例：`idx_orders_user_id` → `idx_orders_uid`
   - 部分索引示例：`trade.orders(user_id)` → `trade.orders(uid)`

3. **新增"NOT NULL 默认"硬约束**（§4.2）
   - 原文：仅说"软删除用 deleted_at"
   - 新增："**除 TEXT 类型字段和 deleted_at 外，业务字段默认不允许为 NULL**"
   - 影响：以后 CREATE TABLE 必须每列显式 `NOT NULL`，不写就违规

4. **Kafka §10 大改 —— IDP 强约束 + 跨项目 Topic 模板**
   - 头部新增："**中间件命名（Middleware）—— 由 IDP 控制。** 所有 Topic 必须经 IDP 申请、审批后下发；禁止应用直连 Broker 自助创建"
   - Topic 格式从单一格式拆为两类：
     - **项目内**：`[环境]_[服务名]_[业务语义]_[动作]` 如 `prod_order_payStatus_updated`
     - **跨项目（新）**：`[环境]_[发起方服务]_to_[接收方服务]_[业务语义]` 如 `prod_pay_to_order_record`、`prod_order_to_wallet_settlement`
   - "发布方"改名"服务名"，明确"不含 -service / -job 后缀"
   - 业务语义示例从 `trade/deposit` 扩展到 `payStatus/deposit/trade`（涵盖驼峰式语义短语）
   - 禁止清单新增："应用绕过 IDP 自行创建 / 修改 / 删除 Topic"
   - 子节重排：原 §10.1 拆为 §10.1（命名格式）+ §10.2（通用约束）；§10.2/10.3 后移到 §10.3/10.4

### Synced — Cursor `05-database.mdc` 同步关键约束

补两条 v1.2.0 新约束到 Cursor 规则（精简版）：
- 关联字段：`{关联表单数}_id` 示例从 `user_id, order_id` → 仅 `order_id`，明确注释"用户字段统一 `uid`，禁 `user_id`"
- 索引示例：`(user_id 等)` → `(uid / order_id 等)`
- 字段必备区新增："**业务字段默认 NOT NULL**（除 TEXT 类型字段和 deleted_at）"

### Notes

- 没碰其它规范、Skill SKILL.md、命令、UI、后端
- `claude/go-team-standards/SKILL.md` 内只引用规范路径不内嵌字面量，无需改
- 编译过

---

## [1.7.28] — 2026-05-01

### Added — `code-review` Skill 配套 USAGE.md（给人看的使用说明）

`claude/code-review/USAGE.md` —— 区别于 `SKILL.md`（给 AI 的 prompt），这份是**给团队成员看的**：

- 怎么知道 skill 生效了（emoji 标记 / patch 头部 / TODO 注释格式）
- 强触发 / 不触发 / 拿不准三类场景速查
- 三种输出模式（A 静默修复 / B 完整报告 / C TODO 注释）的实际示例
- 严重度 P0/P1/P2/P3 怎么读 + 行动指南
- 置信度 `conf=NN` 怎么读 + 含义分段
- 怎么 opt-out（`// review:ignore` 单段豁免 / 当次对话不 review / 文件类型自动放宽）
- `/review` 主动调用三种用法（diff / 文件 / 内联）
- 6 条常见问题 FAQ（装完没反应 / AI 不打 🔍 / P3 噪音 / idiom 误判 / 与 /review 重复 / 改阈值）

文件随 skill 安装自动落地到 `~/.claude/skills/code-review/USAGE.md` 和 `~/.cursor/skills-cursor/code-review/USAGE.md`，团队同事装完直接看。

### Notes

- 仅新增 `claude/code-review/USAGE.md`（一个文件），其它没动
- `installEmbeddedSkill` 走 `fs.WalkDir` 自动 embed，无需改后端
- 编译过

---

## [1.7.27] — 2026-05-01

### Changed — `code-review` Skill 升级到 v2.0（对标 GitHub 业界精品）

调研对象：
- [Anthropic 官方 code-review plugin](https://github.com/anthropics/claude-code/tree/main/plugins/code-review) —— 拿到「置信度 0-100 + 阈值过滤」核心思路
- [wshobson/agents code-reviewer](https://github.com/wshobson/agents/blob/main/plugins/comprehensive-review/agents/code-reviewer.md) —— 拿到「constructive 教育语气」+ 10 步 workflow
- [VoltAgent/awesome-agent-skills](https://github.com/VoltAgent/awesome-agent-skills) + [travisvn/awesome-claude-skills](https://github.com/travisvn/awesome-claude-skills) —— 索引参考

**v1.0 → v2.0 五大核心改动**：

1. **触发决策表**（解决"过度激活"问题）
   - 强触发：AI 自产代码 / 用户粘 ≥10 行实际代码 + 模糊问句
   - 显式触发：review / 审一下 / PR / MR
   - **不触发**（明确清单）：语法/语言知识问 · ≤5 行伪代码 · 写 .md 文档 · 用户说"先不 review" · 同段已 review 过未变
   - 拿不准时倾向**不触发**——错过比噪音好

2. **严重度分级 P0/P1/P2/P3**（之前是一锅炖）
   - P0 阻断：硬编码密钥 / SQL 注入 / 金额 float64 / 业务 panic / 敏感数据进日志
   - P1 重要：errors.New 不走 xerror / 命名违规 / 无 ctx 调 IO
   - P2 改进：日志缺 trace_id / 函数过长
   - P3 建议：折叠在 `<details>` 里，不抢主线

3. **置信度 0-100**（学 Anthropic 官方，过滤吹毛求疵）
   - 默认阈值：P0/P1 ≥70 · P2 ≥80 · P3 ≥90
   - 自查准则："如果用户回我'你确定？'，能不能列出 3 条具体证据？"——列不出就撤回

4. **False positive 黑名单**（业界 review skill 最易翻车点）
   - 已存在代码不在 diff 内 → 标"待重构"不阻断
   - linter 能逮的（gofmt / golangci）
   - idiom 误判：`sync.Once` / `context.Background()` in main / `_ = g.Wait()` 已 select 处理 err
   - `// review:ignore` opt-out 注释相邻 5 行
   - 重复打扰防护（同段已 review 过且未变）

5. **三种输出模式自动决策**
   - 模式 A · 静默自检（默认）：AI 自产代码 → 内心 review → P0/P1 直接修了再交付，用户根本看不到 review 过程
   - 模式 B · 完整 review 报告：用户粘代码 / 显式叫 → 输出 P0/P1/P2 列表 + P3 折叠 + 通过项 + 已存在问题区
   - 模式 C · TODO 注释模式：重写成本高 / 用户说"标出来就行" → `// TODO[review:P1·铁律#6 conf=85]: ... → ...`

### Changed — `/review` slash command 同步对齐 v2.0

`/review` 共享 code-review skill 同一套规则，但**总走模式 B**（用户主动调 = 总要看完整报告）。
新增分级 + 置信度 + 不报清单。末尾标记由 `🌟` 升级为 `🌟🔍`（go-team-standards + code-review 联合）。

### Fixed — 一键更新接入 code-review + Slash commands（v1.7.25 遗留 bug）

旧版「🔄 一键更新规范」只装 `go-team-standards + cursor rules + orangecat + dev-dna` 共 3 skill，
**漏了 v1.7.25 新加的 code-review skill 和 v1.7.23 的 5 个 slash commands**。

新版并发拉 5 个端点：
- `/api/install`（go-team-standards + cursor rules）
- `/api/orangecat/install`
- `/api/dev-dna/install`
- `/api/code-review/install` ← 新接入
- `/api/commands/install` ← 新接入

每条独立 `.catch()` 防止单个失败拖垮全部，最后输出统一汇总（哪个成 / 哪个败 / 各装了几个文件）。

### Notes

- 仅改 `claude/code-review/SKILL.md`（重写 →v2.0）+ `claude/commands/review.md`（重写）+ `web/app.js`（updateAllBtn handler）
- `references/fix-examples.md` 不动（详细修复对照保留）
- `code_review_install.go` 后端不动
- 编译过；CSS / index.html 没碰

---

## [1.7.26] — 2026-05-01

### Changed — UI 全面换肤到 Apple Tahoe 设计语言

之前的深色 GitHub 风太重、信息密度溢出；新版换成 macOS System Settings 那种**浅色玻璃**：白色半透明卡片 + `backdrop-filter: blur(40px) saturate(180%)` + 苹果系统色板（`--sys-blue/green/orange/red/purple/pink/indigo`）。

**视觉系统**：
- 三级 label 透明度（macOS label hierarchy）
- 0.5px 分隔线（系统设置那种细发丝）
- 7px / 10px / 12px / 14px 圆角阶梯
- SF Pro Display 大标题 · SF Pro Text 正文 · SF Mono 等宽
- 整体背景：4 个彩色径向渐变 + 浅灰底（环境光）
- 滚动条收成 8px 圆角细条

**保留**：所有旧 class 名通过 CSS 变量别名映射到新系统色，**`app.js` 2168 行业务逻辑零改动即生效**。

### Added — Skill 卡片触发方式 Pill

「规范模块」每个 Skill 卡片现在显示**触发方式标签**，一眼看到 AI 怎么被激活：

- 🔵 **被动 · 关键字**（`tg-passive`）：description 命中关键字时自动激活
- 🟢 **主动 · /命令**（`tg-active`）：用户敲 slash command 时
- 🟠 **文件 · *.go**（`tg-glob`）：编辑命中 glob 文件时
- 🟣 **始终注入**（`tg-always`）：每次对话都强制加载

`renderSkills()` 按 `triggers[]` 字串和 `scope` 启发式分类输出 pill。旧的单标签 `.card-scope` 废弃但保留 DOM 兼容。

### Changed — 安装页 11 卡片重组为 4 段 segmented control

旧版 11 个 `install-card` 堆在一页要滚很久。新版顶部 segmented control 切 4 段：

| 段 | 内容 |
|---|---|
| **基础** | Hero 卡（"X 个 Skill 已就绪"大数字）+ 4 格 mini 统计 + 立即安装表单 |
| **更新与同步** | 一键更新规范 · 覆盖检查 · 项目级 Skill 同步 |
| **进阶 Skill** | Slash Commands · OrangeCat · Dev DNA |
| **打包分发** | DMG / zip / Shell · DMG 打不开警告 · Claude Desktop 单 Skill 导出 |

**删除冗余**：旧版"📌 Skill 触发时机说明"长表格已删 —— 触发条件现在直接由规范模块页的 pill 展示。

**Hero / mini-tiles 数据**：来自 `/api/catalog`、`/api/custom`、`/api/commands/list`，无 mock 数据。每次切到安装页都刷新一次。

### Added — Hero / Mini-tiles / Segmented 组件 CSS

- `.seg` / `.seg-item` / `.subpage` —— 苹果 iOS 那种段控
- `.hero` —— 玻璃大卡，含 `.num-big` 渐变文字
- `.mini-tiles` 4 格，每格含 `.lbl` / `.num` / `.delta`
- `.trigger-tag.tg-passive/active/glob/always` —— 4 色触发方式 pill

### Notes

- 仅改 `web/style.css`（重写）+ `web/index.html`（安装页结构 + Hero + mini-tiles）+ `web/app.js`（renderSkills 注入 pill + segmented 切换 + hero 数据）
- 所有业务 ID 保留：`installBtn` / `updateAllBtn` / `coverageCheckBtn` / `projScanBtn` / `commandsList` / `orangecatStatus` / `devDnaStatus` / `cd-zip-btn[data-skill]` 等
- 编译过：12.3 MB 二进制
- 没碰任何后端 API
- 已知差异 vs mockup：没做 traffic-light titlebar（webview_go 已经走 macOS 原生 chrome，没必要画）

---

## [1.7.25] — 2026-05-01

### Added — `code-review` Skill（**被动**自动评审）

之前只有 `/review` slash command —— 用户**主动**敲才触发。本版新加 `code-review` Skill —— **AI 写完代码 / 用户粘代码时自动激活**，不等指令。

**触发场景**：
- AI 自己刚产出 .go / .sql / .proto / .ts 代码 → **产出后自检**，违规直接修
- 用户贴代码 + 没明确意图（"看下"/"这样行吗"）→ 自动 review
- 用户说"review / 审一下 / 检查 / 看看"等

**和 `/review` 的区别**：

| | code-review skill | /review command |
|---|---|---|
| 触发 | **自动**（被动） | 用户敲 / |
| 范围 | AI 当前产出 / 用户粘的代码 | git diff / 指定文件 |
| 控制 | 默认开 | 用户决定何时 |
| 场景 | "AI 自己别写出违规代码" | "我要 review 这次改动" |

**Review 流程**（Skill 内置）：
1. 读 14 条铁律 + 全局命名规范
2. 14 条铁律 grep-style 检查（密钥 / errors.New / float 金额 / SELECT * 等）
3. 命名规范 12 条对照（`user_id→uid` / `gmt_create→created_at` / `is_deleted→deleted_at` 等）
4. 输出统一格式：违规 # · 位置 · 代码 · 为什么不行 · 怎么改
5. **直接修复模式**（AI 自己产出未交付时）或 **TODO 注释模式**（用户已写代码）：
   ```
   TODO[skill: code-review · 铁律 #N]: 原 → 推荐
   TODO[skill: code-review · 命名 §X.X]: 原 → 推荐
   ```
6. 末尾打 🔍 标记（与 go-team-standards 联合时打 🌟🔍）

### Added — `code-review` 安装 API + 覆盖检查接入

- `claude/code-review/SKILL.md` + `references/fix-examples.md`
- 新文件 `code_review_install.go` —— 3 个 API：`/api/code-review/status` / `install` / `uninstall`
- coverage.go 加第 5 个 bundle「Claude · code-review（自动评审）」

### App 三套 UI mockup（**还未接进 App，等你选定**）

`mockups/index.html` 入口，3 套 App 整体风格：

| 风格 | 适合 | 改动量 |
|---|---|---|
| **A · 暗黑技术风** | 当前样式精简，开发者向 | 0（贴近现状） |
| **B · 仪表盘 SaaS 风** | KPI + 进度环，PM/Lead 向 | 中（重排） |
| **C · 引导向导风** | 5 步 onboarding，新人向 | 大（流程重设计） |

回 `选 A` / `选 B` / `选 C` / `A+B 组合` 我下版接进 App。

### 用户操作

1. 装 v1.7.25 DMG
2. 「⚡ 安装」点「🔄 一键更新规范到最新」（**我会下版加上 code-review 的自动安装；本版需手动调一次** `curl -X POST http://127.0.0.1:.../api/code-review/install`，或在覆盖检查里看到 ❌ missing 时点"补齐"）
3. **Cmd+Q 重启 Claude Code / Cursor**
4. 试粘一段含 `user_id` 的 SQL 给 Claude → 期望它自动指出"§1.2 → uid"+ 末尾 🔍

### 隐患

- v1.7.25 的「🔄 一键更新」**还没**自动包含 code-review。下版补
- 自动触发依赖模型对 description 的解读，本质仍是软触发。Claude 4.6/4.7 应该能稳触发，老版本可能漏

---

## [1.7.24] — 2026-05-01

### Fixed — Cmd+Q 退出快捷键真生效（v1.7.14 那次只修了一半）

**Bug 全貌**：v1.7.14 的 commit message 说"修 Cmd+Q 不生效"，但实际**只做了后端**：

- ✅ `main.go` 已加 `w.Bind("appQuit", func() { w.Terminate() })`
- ❌ `web/app.js` 的 keydown handler **没监听 `'q'`** —— 用户按 Cmd+Q 信号根本没传到 Go

所以 v1.7.14 ~ v1.7.23 期间 Cmd+Q 一直是**摆设**。

**修法**（1 行）：在 keydown handler 加分支：

```js
if (k === 'q') {
  e.preventDefault();
  if (typeof window.appQuit === 'function') {
    window.appQuit();   // 调后端 w.Terminate()
  }
}
```

**不论焦点在哪都触发**（macOS 标准 Cmd+Q 行为：全局退出，不像 Cmd+C 那样需要 inField）。

### 复制粘贴快捷键现状（无需改动，确认还在）

`Cmd+A` / `Cmd+C` / `Cmd+V` / `Cmd+X` / `Cmd+Z` / `Cmd+Shift+Z` 在 v1.7.14 之前就有完整实现（line 5-51），且本版未动。再次确认：

| 快捷键 | 行为 |
|---|---|
| Cmd+A | 输入框/textarea 内全选 |
| Cmd+C | 选中文字复制（navigator.clipboard 优先，回退 execCommand） |
| Cmd+X | 剪切（同 Cmd+C + 删选区） |
| Cmd+V | 粘贴（navigator.clipboard.readText 优先，回退 execCommand） |
| Cmd+Z | 撤销 |
| Cmd+Shift+Z | 重做 |
| Cmd+Q | **新修**：调 appQuit() → Go 端 w.Terminate() 退 App |

### 验证步骤

1. 装 v1.7.24 DMG
2. 装好后按 Cmd+Q → App 应该立刻退出
3. 在任何 textarea（如 OrangeCat 模板编辑器）选文本 + Cmd+C → 粘到外部 → 应粘对
4. 在任何 input 里 Cmd+V 应能粘进系统剪贴板内容

---

## [1.7.23] — 2026-04-30

### Added — 5 个 Slash Commands（架构落地：Skill / Rules / Prompt 三态分离）

之前所有规则都塞 Skill。本版引入 **Slash Commands**（`~/.claude/commands/<name>.md`），让"用户主动想做某事"用显式 `/<name>` 调用，区别于 Skill 的被动激活。

**架构对照表**（App UI 也展示）：

| 形态 | 何时生效 | 举例 |
|---|---|---|
| **Skill** | 模型按 description 决定（被动） | go-team-standards / dev-dna / orangecat |
| **Rules** alwaysApply | 每次对话强制注入 | 14 条铁律（00-iron-laws.mdc）|
| **Rules** globs | 编辑命中文件时自动加载 | 15-doc-trigger.mdc（.md 触发）|
| **Slash Command** | **用户主动 `/` 调用** | /tech-design / /design-table |

**5 个新 Slash Command**：

| 命令 | 用途 |
|---|---|
| `/tech-design <主题>` | 按 tech-design-example.md 7 段范例展开技术方案 |
| `/design-table <表名+业务>` | 按 database.md + 命名规范产出完整 CREATE TABLE |
| `/api-doc <接口>` | 按 api-doc-example.md 格式产出接口文档 |
| `/review [diff/file]` | 按 14 铁律 + dev-dna 偏好 review git diff / 当前文件 |
| `/tixuebj [tag]` | 主动调用 OrangeCat 生成提测报告（比模糊触发更可靠）|

每个命令的 .md 都带 `argument-hint`（参数提示）+ description（在 / 列表里显示）。

**装到**：
```
~/.claude/commands/<name>.md   ← Claude Code /<name> 调用
~/.cursor/commands/<name>.md   ← Cursor 同样支持
```

**App UI**：「⚡ 安装」Tab 加紫边卡片「⌨️ Slash Commands · 主动调用的规范模板」：
- 列表：每条命令 + `/id` + 描述 + 用法 + Claude/Cursor 双侧安装状态
- 「📥 安装全部」/「🗑 卸载」按钮
- 安装后立即在列表显示绿色 ✅

### Added — App 安装说明加 `xattr -cr` 提示卡片

接收方装 DMG 后双击 .app 报"已损坏 / 无法打开"是常见问题（macOS Gatekeeper 把未签名 App 标了 quarantine 属性）。
ad-hoc 签名解决不了，必须**手动清除扩展属性**：

```bash
xattr -cr /Applications/Team\ Standards.app
```

App「⚡ 安装」Tab 加橙边卡片「⚠️ DMG 安装后打不开？必读」，含：
- 复制即用的命令（user-select:all）
- 解释（什么是 quarantine）
- 何时需要跑（DMG / zip / U 盘 .app 第一次打开）
- 何时不用跑（App Store / 正式签名）
- 路径不对怎么办（实际路径替换 + 空格转义）

### 多版本兼容（推迟到下版）

用户提到"App 安装支持多版本历史和最新兼容"。**还没动**。需要先确认你想要的是：

1. **方案 X 备份回退**：每次「🔄 一键更新」前自动备份当前版本到 `~/.team-standards/backup/v<旧>/`，回退按钮恢复
2. **方案 Y 多版本 embed**：DMG 内置多个版本（v1.7.x），安装时下拉选 —— DMG 体积膨胀
3. **方案 Z 版本提示**：检测本地版本与 embed 不一致时显示"本地 v1.7.x，最新 v1.7.23，是否更新？"，用户主动决定

回个 X / Y / Z，下版做。我推荐 **X（轻量备份/回退）**。

### 后端

- 新文件 `commands_install.go`（~150 行）
- 3 个 API：`GET /api/commands/list` / `POST /api/commands/install` / `POST /api/commands/uninstall`
- 5 个内置 .md 在 `claude/commands/`，被现有 `//go:embed all:claude` 自动捡到，无需改 embed 指令

### 用户操作

1. 装 v1.7.23 DMG（**Cmd+Q 彻底退旧 App**）
2. 「⚡ 安装」Tab 找紫边「⌨️ Slash Commands」卡片
3. 点「📥 安装全部」→ 5 命令落到 `~/.claude/commands/` + `~/.cursor/commands/`（共 10 文件）
4. **Cmd+Q 重启 Claude Code / Cursor**
5. 在 Claude Code / Cursor 里敲 `/tech` → 应自动补全 `/tech-design`
6. 例：`/design-table users 用户表，含 uid、邮箱、注册时间、最后登录` → AI 按命名规范产出完整 CREATE TABLE

### 隐患

- Slash Command 在 Claude Code 4.6+ 才好支持。老版可能 / 列表不出现
- Cursor 的 `~/.cursor/commands/` 是新加的目录，是否生效要看你 Cursor 版本。如果不识别，命令文件其实就是普通 markdown，可以手动 @ 引用
- 多版本兼容（X/Y/Z）还没做，等你定方向

---

## [1.7.22] — 2026-04-30

### Added — Skill zip 导出含 `INSTALL.md` 手动安装说明 + 路径显示

**用户反馈两件**：
1. orangecat / dev-dna 卡片上的「📦 导出 zip」还是浏览器流式下载，不显示路径
2. 接收方拿到 zip 不知道怎么装

**修法 A**：orangecat / dev-dna 卡片上的导出按钮统一走 `save-zip` 落盘流程，与 v1.7.18 主导出列保持一致。三按钮：
- 📂 在 Finder 中显示
- 🚀 打开 Claude Desktop
- 📋 复制路径

抽出 `doSkillSaveZipFlow(skillName, logElId, btn)` 公共函数复用。

**修法 B**：所有 zip **根目录新增 `INSTALL.md`**，含三套安装方式：

```bash
# Claude Code
mkdir -p ~/.claude/skills/
cp -R <skill> ~/.claude/skills/
# Cmd+Q 重启

# Cursor
cp -R <skill> ~/.cursor/skills-cursor/

# Claude Desktop
Settings → Skills → + → Upload zip → 拖入
```

包含验证方法（"末尾应该出现 🌟 / 🐱 / 🧬"）+ 卸载步骤 + "想用图形化工具问同事要 DMG"。

接收方解压 zip 第一眼就看到 INSTALL.md，**不依赖我们的 App**。

### Added — Cursor `15-doc-trigger.mdc`：按 .md glob 强行触发（不靠模型决策）

之前 v1.7.19 加了 description 触发词 + ZERO STEP（文档场景），但仍依赖模型对 SKILL.md description 的解读，**模型偷懒就漏触发**。

v1.7.22 上 **engine-level 强行触发**：

```yaml
---
description: 写技术文档时强制激活团队规范
alwaysApply: false
globs:
  - "**/*.md"
  - "**/*.markdown"
---
```

Cursor 引擎按文件 glob 加载，**不管模型怎么想**，只要在 Cursor 里打开 .md，本规则就直接注入上下文。比纯 description 触发可靠得多。

规则内容：
- 按"文档类型 → 必读 reference"路由表
- 字段命名速查（`uid` 不是 `user_id`、`deleted_at` 不是 `is_deleted`、时间 `_at` 后缀、金额带前缀、布尔 `is_` 前缀）
- 触发反馈协议（末尾 🌟 + 文档头加 `<!-- [skill: ...] -->`）
- 4 条禁止行为

Cursor `.mdc` 总数 12 → **13**。`installCursor` 用 WalkDir 自动捡到，无需改 install.go。`coverage.go` 加进 manifest。

### 关于 Claude Code 端的等价方案

Claude Code 不支持 globs 强触发（它没有 Cursor 那套 .mdc 引擎层）。Claude Code 只能靠 SKILL.md description 让模型决定是否激活 —— 这部分 v1.7.19 已加固。Cursor 这边是**双保险**（description + globs）。

### 验证

```
=== save-zip 含 INSTALL.md ===
saved: .../version/exported-skills/go-team-standards-...zip
总文件数: 29 (28 个 skill 文件 + 1 个 INSTALL.md)
含 INSTALL.md: True
```

### 操作建议

1. 装 v1.7.22 DMG（**Cmd+Q 彻底退旧 App**）
2. 「⚡ 安装」Tab 点「**🔄 一键更新规范到最新**」 —— 这一步会把新的 `15-doc-trigger.mdc` 装到 `~/.cursor/rules/`
3. **Cmd+Q 重启 Cursor**（Claude Code 同样重启）
4. 在 Cursor 里打开任意 .md 文件 → 看左下角"Rules"是否亮（说明规则已加载）
5. 试问 "帮我写个 user 表设计" → 期望 AI 按 database.md + 命名规范产出 + 末尾 🌟

### 还没法保证

- Claude Code 端的文档触发**仍依赖模型**（Cursor globs 不影响 Claude Code）
- 如果 Claude Code 4.6/4.7 还是不触发，下版考虑加 `~/.claude/CLAUDE.md` 全局规则强制（也是软约束但比 SKILL.md description 优先级高）

---

## [1.7.21] — 2026-04-30

### Fixed — 「🔄 一键更新」漏了 dev-dna，加上

之前一键更新只触发 3 个：
- `/api/install` → go-team-standards + Cursor rules
- `/api/orangecat/install` → orangecat
- ❌ **dev-dna 漏了**

v1.7.21 加上 `fetch('/api/dev-dna/install')`。点一下"🔄 一键更新规范到最新"现在会同步 4 个目标全更新。

### 答用户：路径有变吗？没变

```
~/.claude/skills/<name>/        ✅ 当前 Claude Code 路径
~/.cursor/skills-cursor/<name>/ ✅ 当前 Cursor skill 路径（Anthropic 约定）
~/.cursor/rules/*.mdc           ✅ 当前 Cursor rules 路径（v0.45+）
<proj>/.claude/skills/          ✅ 项目级 Claude
<proj>/.cursor/rules/           ✅ 项目级 Cursor rules
<proj>/.cursor/skills-cursor/   ✅ 项目级 Cursor skill
```

实地扫描确认全没变。

### Added — 项目级 skill 同步功能（解决"我项目里也做了 skill"）

「📊 规范覆盖检查」卡片下方新增**黄边「📁 项目级 Skill 同步」卡片**：

1. 输入扫描根目录（默认 `~/Desktop/work/Code`）
2. 点「🔍 扫描」→ App 后端 walk 该目录深度 5 层（跳 `node_modules` / `vendor` / `dist`）
3. 命中含 `.cursor/` 或 `.claude/` 或 `.git` / `go.mod` / `package.json` / `Cargo.toml` 的目录
4. 列出每个项目 + 它的 skill 路径标签：
   ```
   ☑ ~/Code/重构文件
       📜 .cursor/rules (1 .mdc)
   ```
5. 勾选后点「🔄 把规范同步到勾选的项目」
6. 后端调 `installClaude` + `installCursor` 把 embed 版本覆盖到每个选中项目的：
   - `<proj>/.claude/skills/go-team-standards/`
   - `<proj>/.cursor/rules/`
   - `<proj>/.cursor/skills-cursor/go-team-standards/`
   - 如果 `<proj>/.claude/skills/orangecat/` 已存在 → 也覆盖
   - 如果 `<proj>/.claude/skills/dev-dna/` 已存在 → 也覆盖

**不动你私人加的 skill**：`<proj>/.claude/skills/<别的名字>/` 不会被覆盖。

### 实测扫描结果（你机器）

```json
{
  "count": 2,
  "projects": [
    { "project_path": "/Code/golang/first-test",
      "has_cursor_rules": true, "cursor_rules_count": 0 },
    { "project_path": "/Code/重构文件",
      "has_cursor_rules": true, "cursor_rules_count": 1,
      "cursor_rules_files": ["15-go-team-standards-skill.mdc"] }
  ]
}
```

`/Code/重构文件/` 含一个 .mdc，但**不是**我们 manifest 里的 12 个之一（用户自加的）。
扫描后用户可以决定要不要把内置 12 个 .mdc 也装到该项目。

### 后端

- 新文件 `project_skills_scan.go` (~150 行)
- 新增 2 个 API：
  - `GET /api/project-skills/scan?root=<dir>` → 列含 skill 路径的项目
  - `POST /api/project-skills/sync` body `{projects: [path...]}` → 复用 `installClaude` + `installCursor` 覆盖

### 隐患

- **路径中文字符 OK**（`/Code/重构文件/` 已实测通）
- 默认根 `~/Desktop/work/Code` 是按你机器写死的，其他人用要自己改输入框
- 深度限 5 层 + 跳 `node_modules` 等，正常项目结构 1-3 秒扫完
- 同步时**不删原项目里你自加的 .mdc**（如 `15-go-team-standards-skill.mdc`），只覆盖 manifest 内的 12 个 + 加缺失的

---

## [1.7.20] — 2026-04-30

### Added — 字段命名词典 v1.0.1（来自《全局统一字段命名规范》）

**用户需求**：从 `技术规范.2026.04.29/全局统一字段命名规范.md` 加载新规范，特别强调 Kafka topic / Redis key 命名**必须带前缀，否则禁止使用**。

**新增文件**：`standards/field-naming.md`（v1.0.1，2026-04-30）—— 9 大类字段命名词典：
- §1 ID / 标识（uid 替代 user_id；platform_id / org_id 多租户三件套）
- §2 时间（_at 后缀 + TIMESTAMPTZ(6) UTC；deleted_at 替代 is_deleted）
- §3 状态 / 类型（SMALLINT + 域 lib 常量；status / state / type / mode 三分法）
- §4 金额 / 数量 / 价格（_amount / _qty / _balance / _price 后缀；裸 amount 禁用）
- §5 布尔（必须 is_ / has_ / can_ 前缀）
- §6 审计 / 版本（裸 version = 乐观锁；业务版本必须前缀 `rule_version` 等）
- §7 备注（description / remark / review_note / failure_reason 四分）
- §8 网络 / 设备 / 安全
- §9 加密 / 区块链

附 §10 现状冲突清单 / §12 CI Lint 规则 / §13 豁免登记。

### Added — SKILL.md 新增「🔑 命名硬规则」段（v1.7.20 强化）

写 / 改 Kafka topic、Redis key、DB 字段、Proto field、API JSON 字段时**必走 4 步**：

1. **强制读两份规范**：
   - DB 字段 → `field-naming.md`
   - Kafka topic → `database.md` §7 + `field-naming.md`
   - Redis key → `database.md` §8 + `field-naming.md`

2. **以下产出直接拒绝**（9 条硬禁止）：
   - ❌ Kafka topic 不带域前缀（必须 `<域>.<事件>.<v版本>`，如 `wallet.deposit_completed.v1`）
   - ❌ Redis key 不带域前缀（必须 `<域>:<实体>:<标识>`，如 `wallet:balance:USR_88001`）
   - ❌ DB 用 `user_id` / `userId` / `member_id` → `uid`
   - ❌ 时间用 `create_time` / `gmt_create` → `created_at`
   - ❌ 软删用 `is_deleted BOOLEAN` → `deleted_at TIMESTAMPTZ`
   - ❌ 布尔不带 `is_` 前缀（`enabled` / `active`）→ `is_enabled` / `is_active`
   - ❌ 业务版本字段裸 `version` → `rule_version` / `formula_version` 等
   - ❌ 金额裸 `amount` → `<前缀>_amount`
   - ❌ 时间裸 `time` / `timestamp` / `ts` → 必须 `_at` 后缀

3. **三段式提醒**：违反条款 → 后果 → 改为

4. **🌟 反馈** + 代码注释 `[skill: go-team-standards · 字段命名 · ...]`

### Changed — description 触发词加「命名场景」分组

```
- 命名场景（强制激活）：字段命名 / 表字段 / 列名 / 接口字段 /
  Kafka topic / Kafka 主题 / Redis key / Redis 缓存键 /
  proto field / DB schema 命名 / API 字段名
```

### Changed — install.go + coverage.go

`refs` 17 → **18**；`gtsRefs` 同步。

### 升级路径

1. 装 v1.7.20 DMG（Cmd+Q 彻底退）
2. **必须**点「🔄 一键更新规范到最新」 ← 不点就没新 SKILL.md
3. 重启 Claude Code / Cursor
4. 测试：在 Cursor 里说 "我要新建一个 Kafka topic 用来发用户充值成功事件"
   - 期望末尾 🌟 + 代码 `const TopicXxx = "wallet.deposit_completed.v1"` 顶部带 `// [skill: go-team-standards · Kafka 命名] ...`
5. 测试 2：写 `CREATE TABLE` 含 `user_id` → 期望 AI 主动指出违反 §1.2 + 改成 `uid`

### 隐患

- **field-naming.md 引用了《数据库设计规范》v1.1.1** 但本项目 `database.md` 还是 v1.1.0（v1.7.14 同步的）。如果有冲突按 field-naming.md 附录 C 的优先级裁决（裸 `version` = 乐观锁，`uid` 不用 `user_id` 等）
- **真正生效靠模型遵守 SKILL.md**。如果 v1.7.20 装上 + 重启 + 一键更新后还不触发，回我"上 globs"，我做 Cursor `.mdc` globs 强制注入

---

## [1.7.19] — 2026-04-29

### Fixed — 写技术文档（.md）时 go-team-standards 不触发

**用户反馈**：v1.7.17 加了文档类触发词，但实际写 .md 时 Skill 还是不激活。

**根因分析**（两条都可能）：
1. **本地 SKILL.md 还是老版**：用户 v1.7.17 装了但没点「🔄 一键更新规范到最新」，本地 `~/.claude/skills/go-team-standards/SKILL.md` 仍是 v1.7.16 内容（无文档触发词）
2. **触发词位置太隐蔽**：v1.7.17 把 .md 加在 "文件类型" 行的括号里 + 末尾"文档类"分组，模型扫 description 时容易漏

### 修法：description 重写 · 文档场景提到顶部

```yaml
description: |
  【Go 开发规范 · 强制前置激活 · 适用代码 + 技术文档】
  任何涉及 Go 代码 **或** 任何形式技术文档（.md 技术方案 / 设计文档 / 接口文档
  / 数据库设计 / 会议纪要）都必须在写第一行前读完本 skill。

  📄 文档场景（用户写 .md 时必须激活，与代码同等优先级）：
  - 文件类型：.md / .markdown
  - 行为：写技术方案 / 设计文档 / 接口文档 / 数据库设计 / 会议纪要 / 提测文档 ...
  - 关键短语：技术方案 / tech design / RFC / ADR / 架构文档 / 接口文档 ...

  💻 代码场景（保留原触发词）

  判断准则（防漏触发）：
  - 用户当前在写/编辑 .md 文件 + 内容含任一关键词 → 必须激活
  - 用户聊天里描述要写技术文档 → 必须激活
  - 提到任一基础设施（PG/MySQL/Redis/Kafka）→ 必须激活
  - 拿不准时宁可激活也不要漏（漏激活 = 用户拿到的产出与团队规范不符）
```

### 新增「📝 ZERO STEP（文档场景）」段

与原 ZERO STEP（代码场景）**同等优先级**，写文档前必做 4 步：

1. **识别文档类型** → 路由表（技术方案 → tech-design-example.md / 接口文档 → api-doc-example.md + api-design.md / 数据库设计 → database.md + naming-logging.md / 提测 → tixuebj-template-simple.md / 会议纪要 → meeting-minutes.md / 部署 → deployment-checklist.md / Code Review → code-review.md / 特性开关 → feature-flags.md）
2. **完整读对应 reference**（不能扫一眼就动手）
3. **对比+应用**：把用户已写部分与标准结构对比，指出缺段/命名不一致；新部分严格按范例组织
4. **触发 🌟 反馈**：聊天末尾打 🌟，文档头部加 `<!-- [skill: go-team-standards · <子规范>] -->`

**明确列禁止行为**：
- ❌ 看到 .md 当"普通问答" → 必须先按文档类型路由读规范
- ❌ "我帮你写一份技术方案" 不读 tech-design-example.md
- ❌ 写表结构不读 database.md
- ❌ 写接口文档不读 api-design.md + api-doc-example.md

### 用户操作（关键）

**必须先重装才生效**：

1. 装 v1.7.19 DMG（Cmd+Q 彻底退旧 App）
2. 「⚡ 安装」Tab → 点「**🔄 一键更新规范到最新**」（很关键 —— 不点的话本地 SKILL.md 还是老版，新触发逻辑没生效）
3. **完全重启 Claude Code / Cursor**（Cmd+Q 不是 reload）
4. 测试：在 Cursor 里新建一个 `tech-design-test.md`，敲 `# 用户中台技术方案` → 期望 AI 立刻激活 Skill 并按 tech-design-example.md 范例提建议
5. 如果激活成功，AI 回复末尾应该出现 🌟，文档第一行下方应该有 `<!-- [skill: go-team-standards · 技术方案] -->` 注释

### 仍然依赖模型执行度（诚实）

- description 写得再硬，最终是否真触发还是看模型决策。Claude 4.5 / 4.6 / 4.7 对 skill description 的遵守度不一样
- 如果 v1.7.19 装上还是不触发，下版可以考虑：
  - **明示强制**：在用户开新 .md 文件时，App 主动检测文件名/内容关键词，弹"是否触发 Skill"提示
  - **Cursor 侧 globs**：通过 Cursor 的 `.cursor/rules/*.mdc` 的 `globs: ["**/*.md"]` 字段强行匹配（这个**真**生效，不靠模型决策）—— 我没做是因为会全局触发所有 .md（包括 README、笔记），噪音大

如果你觉得 Cursor globs 方案可接受，告诉我，下版加。

---

## [1.7.18] — 2026-04-29

### Fixed — Claude Desktop 导出 zip 不告诉用户文件落在哪

**用户反馈**："✓ 已下载 go-team-standards.zip，导入：Claude Desktop → Customize → Skills → + → 上传 zip" —— 但用户根本不知道这个 zip 在哪个目录，没法找到去拖。

**根因**：之前用浏览器流式下载（`<a download>`），WKWebView 默认扔到 `~/Downloads/`，但 UI 不告知路径。

**修法**：

1. 后端新增 `POST /api/claude-desktop/save-zip?name=<skill>` —— 服务端落盘到 `<repo>/version/exported-skills/<skill>-<时间戳>.zip`，返回 `{path, dir, name, size}`
2. 前端「📦 zip」按钮改成调 save-zip，成功后渲染：
   ```
   ✓ 已保存 go-team-standards-20260429-2208.zip（79.1 KB）

   📂 文件位置：
   ~/skills/version/exported-skills/go-team-standards-20260429-2208.zip

   [📂 在 Finder 中显示] [🚀 打开 Claude Desktop] [📋 复制路径]
   ```
3. 三个按钮：
   - **📂 在 Finder 中显示** —— 调 `/api/reveal`，macOS `open -R` 在 Finder 中**选中**该 zip
   - **🚀 打开 Claude Desktop** —— 调 `/api/open-app`（白名单仅 Claude / Cursor / VSCode 防注入），`open -a Claude`
   - **📋 复制路径** —— `navigator.clipboard.writeText`

### Added — `POST /api/open-app` 通用打开应用接口

```go
// 仅 macOS，白名单仅 Claude / Cursor / VSCode（防命令注入）
exec.Command("open", "-a", body.App)
```

后续如果有要"装完后一键打开 X 应用"的场景可复用。

### 体验改进总结

旧：`✓ 已下载 go-team-standards.zip` ←（用户：在哪？）
新：`✓ 已保存 ... · /full/path/here · [📂 Finder] [🚀 Claude] [📋 Copy]`

3 步流程：
1. 点 **📦 zip** → zip 落盘 + 路径显示
2. 点 **📂 在 Finder 中显示** → Finder 自动跳到该 zip
3. 点 **🚀 打开 Claude Desktop** → Settings → Skills → + → 把刚才那个 zip 拖进去

### 实测

```
$ curl -X POST http://127.0.0.1:.../api/claude-desktop/save-zip?name=go-team-standards
{
  "ok": true,
  "path": "~/skills/version/exported-skills/go-team-standards-20260429-2208.zip",
  "size": 81040
}
```

zip 真落盘 79K，路径正确返回。

### 隐患

- Windows 上 `open -a` 不支持，handleOpenApp 直接报 501（有的话会再适配）
- "🚀 打开 Claude Desktop" 假设 Claude Desktop 已装。未装则报错"打不开"，不会自动下载

---

## [1.7.17] — 2026-04-29

### Changed — 规范同步：4 覆盖 + 3 新增（来源：技术规范.2026.04.29）

按用户要求："技术规范.2026.04.29 优先级最高，冲突覆盖老规范"。

**覆盖**（提到 v1.1.0，last_modified=2026-04-29）：

| standards/ | ← 来源 |
|---|---|
| `database.md` | `数据库设计规范.md` |
| `error-codes.md` | `错误码系统规范.md` |
| `api-doc-example.md` | `接口文档示例.md`（67K，大幅扩充） |
| `naming-logging.md` | `命名规范.md` |

**新增**（v1.0.0）：

| 新文件 | 来源 | 用途 |
|---|---|---|
| `meeting-minutes.md` | `会议纪要模版.md` | 会议纪要标准模板（议程/讨论/决议/行动项） |
| `tech-design-example.md` | `后端技术方案 - 示例用户中台.md` | 50K 完整技术方案范例（背景/架构/接口/DB/异常） |
| `tixuebj-template-simple.md` | `提测模板.md` | 简化单文件提测模板（与 OrangeCat 双文件版并存） |

`所有规范/` 目录里的内容 v1.7.14 已全部同步到 standards/，本轮不重复处理。

总文件数：14 → **17 references**

### Added — 🌟 触发反馈协议（重点）

**用户痛点**：现在 AI 触发了 Skill 用户也不知道，Skill 形同虚设。

**修法**：在 3 个 Skill 的 SKILL.md 顶部加入**强制反馈协议**：

#### 聊天回复

每次回复**最后一行单独**输出标记符（不要前后加文字）：

| 触发的 Skill 组合 | 末尾标记 |
|---|---|
| 仅 go-team-standards | `🌟` |
| 仅 dev-dna | `🧬` |
| 仅 orangecat | `🐱` |
| go-team-standards + dev-dna | `🌟🧬` |
| go-team-standards + orangecat | `🌟🐱` |
| 三个全触发 | `🌟🧬🐱` |

#### 代码生成 / 修改

在 patch 的**最开始位置**加注释，标明触发了哪个 skill 和具体子规范段：

```go
// [skill: go-team-standards · 数据库设计 · 命名规范] 创建 user 表
package data
```

```sql
-- [skill: go-team-standards · 数据库设计] 用户中台 - users 表
CREATE TABLE users (...);
```

```markdown
<!-- [skill: orangecat] 提测报告 v1.0 -->
```

格式约定：
```
[skill: <skill 名> · <子规范 1> · <子规范 2>] <一句话功能简写>
```

**违反 = 用户无法判断 skill 是否生效 = skill 形同虚设**。

### Added — 触发条件扩展（写文档场景）

go-team-standards description 增加触发词：

```
- 文件类型：.go / .sql / .proto / Dockerfile / Makefile / go.mod
  + **.md（写技术方案 / 接口文档 / 数据库设计 / 会议纪要时）**
- 文档类（新增分组）：写技术方案 / tech design / 设计文档 / 技术文档 / RFC / ADR /
  接口文档 / API doc / 数据库设计 / 表结构设计 / schema 设计 / 会议纪要 / 复盘
```

**实际效果**：你写 `.md` 技术方案、提到 PostgreSQL/MySQL/Redis/Kafka/PG 等，AI 都会激活规范 Skill 并按团队样式产出。

### Added — App 面板「📌 Skill 触发时机说明」卡片（蓝色边框）

「⚡ 安装」Tab 上方新增独立卡片，**3 行表格**清晰说明：

| Skill | 什么场景触发 | 触发后反馈 |
|---|---|---|
| go-team-standards | Go / SQL / 文档 / 基础设施提及 / 流程 | 末尾 🌟 + 代码注释 |
| orangecat | 提测 / QA 交付 | 末尾 🐱 / 🌟🐱 |
| dev-dna | "按我习惯写" / 任何 Go 任务并发 | 末尾 🧬 / 🌟🧬 |

让用户在 App 里一眼看清"什么时候该期待哪个 Skill 激活 + 怎么验证它真在工作"。

### Updated — install.go + coverage.go

`refs` 列表 14 → **17**；`gtsRefs` 同步。新加的 3 个 reference 也会被覆盖检查识别。

### 升级路径

老用户（v1.7.16）：
1. 装 v1.7.17 DMG，**Cmd+Q 彻底退旧 App 再开**
2. 「⚡ 安装」Tab 上方应该看到新的「📌 Skill 触发时机说明」蓝边卡片
3. 点「**🔄 一键更新规范到最新**」→ 17 个 references 全更新（4 覆盖 + 3 新增）
4. 「📊 规范覆盖检查」应看到 30+ 个 ✅ ok（SKILL.md + 17 refs + 11 demos + .golangci.yml）
5. 在 Claude / Cursor 里问"帮我设计一张 user 表" → 期望看到末尾 🌟 + SQL 顶部注释 `-- [skill: go-team-standards · 数据库设计] users 表`

### 隐患（诚实说明）

1. **🌟 反馈协议靠 AI 执行度**：SKILL.md 写得再硬，最终是否真在每次回复末尾打 🌟 取决于模型。如果某些回复漏打了，是模型问题不是协议设计问题。建议你跑一次「🚀 运行全部 API 测试」看 clawnova 模型是否能稳定打标
2. **多 Skill 联合标记可能漏**：理论 `🌟🧬🐱` 应该出现在三 Skill 都激活时，但模型可能只打一个。这是模型对"并发激活"的理解问题
3. **`tech-design-example.md` 是 50K 完整范例**，不是 spec 本身。AI 看到会"举一反三"，可能让某些技术方案 over-engineering（按用户中台标准写一个简单 CRUD）。如果碰到这种情况，告诉 AI "参考这个范例但简化"
4. **未跑端到端**：DMG 打了，没在新机器上点「一键更新」+「覆盖检查」实测全过

---

## [1.7.16] — 2026-04-27

### Added — `dev-dna` Skill（个人开发档案 · 跨电脑无缝迁移）

**起因**：用户希望在换电脑 / 换 AI 客户端时，AI 能立刻"认得你"，不用从零讲一遍"我喜欢什么风格"。同时希望对自己的偏好有反蒸馏保护。

**轻量方案**：单一 Skill `dev-dna`，反蒸馏作为头部声明 + 一个 reference 文档。**不**做"自动从历史会话蒸馏"那种重度功能，纯手填模板。

#### 文件结构

```
claude/dev-dna/
├── SKILL.md
└── references/
    ├── profile.md                       # 用户档案模板（手填）
    └── anti-distillation-policy.md      # 反蒸馏指南 + provider opt-out 操作
```

#### profile.md 模板（8 段）

| 段 | 内容 |
|---|---|
| 1. 基本信息 | 角色 / 主栈 / 次栈 / 团队 / 业务领域 |
| 2. 编码偏好 | 错误处理 / 命名 / 测试 / 注释 / 代码组织 / 金额 / 日志 |
| 3. 反偏好 | "不喜欢看到的代码"清单（panic / SELECT * / fmt.Println 等） |
| 4. 常用模式 | "按我的习惯实现 X"对应的具体套路（Kratos 服务 / 数据表 / HTTP 接口…） |
| 5. 工具链 | IDE / 终端 / shell / 关键 alias / 测试 / lint 命令 |
| 6. 决策口味 | 性能 vs 可读性 / 新依赖 / 上线节奏 / 评审风格 |
| 7. 技术债 | 自己常犯的错（让 AI 主动提醒） |
| 8. 学习方向 | 让 AI 在合适话题推荐进阶资料 |

#### 反蒸馏（三层防护）

1. **provider 端 opt-out**（最关键）—— Anthropic Privacy / Cursor Privacy Mode / Copilot 关代码上传
2. **Skill 头部 `privacy:` 声明**（软约束）—— 模型理解但不强制
3. **本地行为约束** —— profile 不填手机号/真名/机密项目代号；不入团队 git；同步走加密渠道

详细见安装后的 `references/anti-distillation-policy.md`。

#### 与 Persona / go-team-standards 的关系

| | 存什么 | 文件 |
|---|---|---|
| Persona（已有） | 工作态度（严谨求实、不敷衍） | `~/.claude/CLAUDE.md`（标记块） |
| **dev-dna（新）** | **技术个性**（栈、风格、模式偏好） | `~/.claude/skills/dev-dna/` |
| go-team-standards | 团队规范（铁律、references） | `~/.claude/skills/go-team-standards/` |

三者并行不冲突。SKILL.md 明确写优先级：**Persona > dev-dna > 团队规范 > 通用 Go 知识**。冲突（如个人偏 errors.New 但团队铁律要 xerror）以**团队规范**为准。

### App UI

「⚡ 安装」Tab 新增「🧬 Dev DNA · 我的开发档案 Skill」卡片：
- 「📥 独立安装」 / 「🗑 卸载」 / 「📦 导出 zip」
- 「📝 编辑我的档案」折叠 textarea —— 保存到 `~/Library/Application Support/TeamStandards/dev-dna-profile.md`（权限 0600）
- 「🛡 反蒸馏」黄色折叠 —— 含各 provider 设置链接 + opt-out 操作步骤
- Claude Desktop 导出段也加 dev-dna 行

### 后端

- 新文件 `dev_dna_install.go`（170 行，complete mirror of orangecat pattern）
- 新增 6 个 API：
  - `GET /api/dev-dna/status`
  - `POST /api/dev-dna/install`
  - `POST /api/dev-dna/uninstall`
  - `GET /api/dev-dna/profile`
  - `POST /api/dev-dna/profile`
  - `DELETE /api/dev-dna/profile`
- coverage.go 新增第 4 个 bundle `Claude · dev-dna（个人开发档案）`

### 跨电脑迁移流程

1. 老机器装好 dev-dna，在 App 里编辑好 profile，保存
2. 老机器：`tar -czf dev-dna.tgz ~/.claude/skills/dev-dna/ ~/Library/Application\ Support/TeamStandards/dev-dna-profile.md`
3. 把 dev-dna.tgz 拷到新机器（U 盘 / 加密云盘 / Apple AirDrop）
4. 新机器：`tar -xzf dev-dna.tgz -C ~/`
5. 新机器装 Team Standards App → 「🧬 Dev DNA」卡片点「独立安装」就完事

或者更简：在老机器点「📦 导出 zip」→ 新机器拖入 zip 解压到 `~/.claude/skills/`，**根本不需要装 App**。

### 不直接 fork 现成开源项目的理由（诚实说明）

| 看过的项目 | 为啥不用 |
|---|---|
| portable-ai-memory (PAM) | JSON 格式，与 Skill 体系（Markdown + frontmatter）不兼容 |
| claude-mem | 自动 hook 注入太重，控制不够 |
| second-brain-starter | 含 SOUL.md 人格层，与"工程开发"目标偏差 |
| everything-claude-code | 框架级，引入会架构爆炸 |

借鉴了 PAM 的"跨厂商可迁移"思想 + second-brain 的 USER.md 段落结构。自己写一份轻量版 ~250 行 markdown 模板。

### 隐患

- **反蒸馏**：Skill 头部 `privacy:` 声明只是软约束，**真正生效靠 provider 端 opt-out**。文档已强调，但用户仍可能误以为加了声明就万事大吉
- **profile.md 跨电脑同步若走公开 git** = 个人偏好泄漏。已在编辑器下方红字提醒"禁止填手机号 / 真名 / 机密项目代号"，文件权限 0600
- **未跑端到端**：DMG 打了，但我没在新电脑上验证"装新 App → tar 解压 → AI 真的认得我"完整链路。你试一次告诉我结果

---

## [1.7.15] — 2026-04-26

### Added — 规范模块加 YAML frontmatter（version + last_modified）

**用户需求**：之前覆盖检查只能告诉你"hash 一致 / 不一致"，看不到**具体版本号和最后修改日期**，将来更新时缺判断依据。

**修法**：每个 `standards/*.md` 文件首部加 YAML frontmatter：

```yaml
---
title: "Go 编码风格规范"
version: "1.0.0"
last_modified: "2026-04-26"
source: "规范版本库0.0.2 / go.md"
---
```

14 个文件全部覆盖：

| 文件 | source | 初始版本 |
|---|---|---|
| api-design.md | 规范版本库0.0.2 / api.md | v1.0.0 (2026-04-26) |
| api-doc-example.md | 规范版本库0.0.2 / 接口文档示例.md | v1.0.0 (2026-04-26) |
| ci-pipeline.md | 规范版本库0.0.2 / ci-pipeline.md | v1.0.0 (2026-04-26) |
| code-review.md | 规范版本库0.0.2 / code-review.md | v1.0.0 (2026-04-26) |
| commit.md | 规范版本库0.0.2 / commit-message.md | v1.0.0 (2026-04-26) |
| database.md | 规范版本库0.0.2 / 数据库设计规范.md | v1.0.0 (2026-04-26) |
| deployment-checklist.md | 规范版本库0.0.2 / 部署清单.md | v1.0.0 (2026-04-26) |
| error-codes.md | 规范版本库0.0.2 / 错误码系统规范.md | v1.0.0 (2026-04-26) |
| feature-flags.md | 规范版本库0.0.2 / 特性开关与分支管理办法.md | v1.0.0 (2026-04-26) |
| go-style.md | 规范版本库0.0.2 / go.md | v1.0.0 (2026-04-26) |
| naming-logging.md | 规范版本库0.0.2 / 命名与日志规范手册 (2026).md | v1.0.0 (2026-04-26) |
| cursor-usage.md | 团队原始（无更新） | v1.0.0 (2026-04-26) |
| glossary.md | 团队原始（无更新） | v1.0.0 (2026-04-26) |
| testing.md | 团队原始（无更新） | v1.0.0 (2026-04-26) |

**未来更新规范**：直接在文件头改 `version` + `last_modified` 即可，无需改其他地方。

### Added — 覆盖检查 UI 显示版本 + 修改日期

`coverage.go` 新增 frontmatter 解析器（不引第三方 yaml 库，自己 split），`coverageItem` 新增 4 个字段：

```json
{
  "rel_path": "references/go-style.md",
  "status": "outdated",
  "installed_version": "1.0.0",
  "installed_modified": "2026-04-26",
  "embed_version": "1.1.0",
  "embed_modified": "2026-05-15"
}
```

UI 渲染（v1.7.15）：

```
✅ references/go-style.md           v1.0.0 (2026-04-26)
🟡 references/database.md           v1.0.0 (2026-04-26) → v1.1.0 (2026-05-15)
❌ references/security.md           缺失 (最新: v1.0.0 2026-05-15)
🔷 references/custom-team-auth.md   （用户自定义）
```

升级时一眼看出哪些过时 + 过时多久。

### 升级路径

老用户（v1.7.14 及之前）：
1. 装 v1.7.15 DMG
2. 点「🔄 一键更新规范到最新」→ 14 个 references 全部覆盖（带新 frontmatter）
3. 点「📊 规范覆盖检查」→ 应看到全 ✅ ok 且每行带 `v1.0.0 (2026-04-26)`

老用户**不更新**也行：旧文件继续工作，只是覆盖检查会标 outdated（hash 不同）。

### 实现细节（诚实说明）

- `parseSpecFrontmatter` 是简单字符串切，**不**支持嵌套 YAML / 数组 / 多行字符串。够用就行，复杂场景再换 yaml.v3
- frontmatter 只加在 `standards/*.md`，**不加** orangecat 模板（那俩是 AI 填空模板，加 frontmatter 会出现在生成的报告里，不合适）
- frontmatter 只加在 `.md`，**不加** `.mdc`（cursor rules 已有 frontmatter，是 alwaysApply / globs，不混用）

---

## [1.7.14] — 2026-04-25

### Changed — 规范内容大版本同步（来源：`规范版本库0.0.2`）

按用户要求，从外部规范库 `~/Documents/SVG 脑图/规范版本库0.0.2/` 同步最新内容覆盖项目原有规范。**7 个文件覆盖 + 4 个新增 + 3 个保留**。

**覆盖（同主题，新版本）**：

| 旧 standards/ | 来源（新） | 行数变化 |
|---|---|---|
| `api-design.md` | `api.md` | 79 → 170 |
| `ci-pipeline.md` | `ci-pipeline.md`（同名） | — |
| `commit.md` | `commit-message.md` | — → 160 |
| `database.md` | `数据库设计规范.md` | — → 768 |
| `error-codes.md` | `错误码系统规范.md` | — → 275 |
| `go-style.md` | `go.md` | — → 417 |
| `naming-logging.md` | `命名与日志规范手册 (2026).md` | — → 94 |

**新增**：

| 新文件 | 来源 | 用途 |
|---|---|---|
| `code-review.md` | `code-review.md` | 三档评审标准 + MR 描述模板 + 阻断项 |
| `deployment-checklist.md` | `部署清单.md` | 发布信息表 + 审批 + 镜像 / 回滚清单 |
| `feature-flags.md` | `特性开关与分支管理办法.md` | Feature Flag 设计哲学 + Go 集成规范 + GitOps |
| `api-doc-example.md` | `接口文档示例.md` | 接口文档完整示例（接口设计配套） |

**保留**（外部规范库未提供新版）：
- `testing.md` —— Go 测试规范
- `cursor-usage.md` —— Cursor IDE 使用 + 安全红线
- `glossary.md` —— 术语表（Kratos / Wire / Buf / common-lib 子包）

### Changed — `go-team-standards` SKILL.md 触发关键字 + 路由表更新

**新增触发词**（让 AI 在更多场景下激活 Skill）：
```
- 流程：code review / MR / PR / 合并 / 评审 / 部署 / 发布 / 上线 /
  灰度 / 回滚 / 特性开关 / feature flag / Feature Flag
```

**路由表新增 3 行**（指向 4 个新 reference）：

| 一句话出现… | 必读文件 |
|---|---|
| 写接口文档 / 接口 README / API doc | `references/api-doc-example.md` |
| code review / MR / PR / 评审 | `references/code-review.md` |
| 部署 / 上线 / 发布 / 灰度 / 回滚 | `references/deployment-checklist.md` |
| 特性开关 / feature flag / 灰度发布 / 分支管理 | `references/feature-flags.md` |

### Updated — install.go + coverage.go 同步加 4 个 ref 条目

- `installClaude` 的 `refs` 列表从 10 项升到 14 项
- `coverageBundles` 的 `gtsRefs` 从 10 项升到 14 项 —— 新增的 4 个 reference 也会被覆盖检查识别

### 用户操作建议

- 装 v1.7.14 后，到「⚡ 安装」Tab 点 **🔄 一键更新规范到最新**
- 然后点「📊 规范覆盖检查」→ 应看到 `Claude · go-team-standards` 27 个 ✅ ok（SKILL.md + 14 references + 11 demos + 1 .golangci.yml）
- Cursor `.mdc` 规则**未同步**这次更新（`.mdc` 是浓缩版，下个版本统一处理）

---

## [1.7.13] — 2026-04-25

### Fixed — 跑全部 20 条 API 测试报 `TypeError: Load failed`

**用户报告**：测试连接成功（小请求 5s 返回），但点「🚀 运行全部 API 测试」后报：
```
✗ TypeError: Load failed
```

**根因**：之前 `/api/eval/run` 一次接收 20 条 case，后端串行调 LLM ≈ 20 × 9s ≈ **3 分钟**才返回。但 macOS WKWebView（`webview_go` 用的渲染引擎）的 fetch 默认 **60 秒超时**就把连接 abort，触发 `TypeError: Load failed`。

**修法**：前端改成**按条串行**。每次 `/api/eval/run` 只跑 1 条，每次 ~9 秒，远在 WKWebView 容忍内。新流程：

```js
for (let i = 0; i < runIDs.length; i++) {
  btn.textContent = `跑 ${i+1}/${runIDs.length}…`;
  showLog(`[${i+1}/${runIDs.length}] ${caseId} 进行中… 通过 ${passed}/${i}`);
  const res = await fetch('/api/eval/run', {
    body: JSON.stringify({ case_ids: [caseId], ... })
  });
  // 累加结果
}
renderEvalReport({ cases: allCases, ... });
```

**附加保护**：连续 3 条 case 都报错（如 API 真挂了）→ 早停止避免浪费时间，已跑的部分照常渲染报告。

**实测**（顺序跑 3 条）：
```
✅ PASS iron-01-hardcoded-secret (9s) · 命中 7/3 · in/out 2221/600
✅ PASS iron-06-float-money       (9s) · 命中 4/2 · in/out 2220/600
✅ PASS iron-08-fmt-println-log   (9s) · 命中 5/2 · in/out 2212/600
```
20 条预计 ~3 分钟跑完，期间用户能看到进度（`跑 13/20…`）。

### 后端不变 / 兼容

`/api/eval/run` 接口本身不变，仍接受 `case_ids` 数组参数（前端只是改成每次传 1 个）。如果以后想用其他客户端（curl / postman）一次跑全部仍可工作，只是在 webview 里走分段。

---

## [1.7.12] — 2026-04-25

### Fixed — 用户保存了 URL 当 Key，导致测试连接 401

**用户报告**：测试连接报 `bad response (HTTP 401): {"error":"invalid bearer token"}`，并截图显示「已存的 key 显示星号化预览：`https://***5u3dal`」。

**根因**：之前某次用户不小心把 URL 粘到 API Key 框点了保存，本地配置文件里 `api_key` 字段是 `https://clawnova.ai/api/v1/.../5u3dal/...`。`maskKey` 函数取前 8 后 6，所以预览显示成 `https://***5u3dal`，明显不是真 key。后端拿这个去发请求，自然 401。

**修法（三层防护）**：

1. **后端 `POST /api/eval/config` 防呆校验**：
   - key 以 `http://` / `https://` 开头 → HTTP 400 + 明确错误「看起来你把 URL 粘到了 API Key 框」
   - key 长度 < 16 → HTTP 400 + 错误「太短，不像完整 key」
2. **`GET /api/eval/config` 返回 `key_health` 字段**：`ok` / `missing` / `bad_is_url` / `bad_too_short`
3. **前端**根据 `key_health` 渲染：
   - `bad_is_url` → 红色警告 `⚠️ 已存的 key 看起来是个 URL，明显错了。点下面「🗑 清除已存 Key」回落默认 key`
   - `bad_too_short` → 红色警告
   - `ok` → 正常显示星号化预览
   - `missing` → 显示 `未配置`

**用户操作**：
1. 装 v1.7.12
2. 打开「🧪 测试 Skill」会看到红色警告
3. 点「🗑 清除已存 Key」→ 回落到内置默认 key（`kgb_d0ef***ae244f`）
4. 点「🔌 测试连接」→ 成功

### Fixed — 模型列表 `qwen` 是过期别名，从教程截图核对修正

教程列出的 chat 模型只有 4 个，加上 2 个特殊用途：
| 模型 | 类型 | v1.7.10/11 错误 | v1.7.12 修正 |
|---|---|---|---|
| `qwen3.5-27b-local` | chat | ✅ | ✅ |
| `minimax-m2.5-local` | chat | ✅ | ✅ |
| `step-3.5-local` | chat | ✅ | ✅ |
| `gemma-4-31b` | chat（多模态） | ✅ | ✅ |
| `qwen` | 不在教程列表 | ❌ 误加 | ✅ 删除 |
| `qwen3-em-8b` | embeddings 专用 | 没列 | 仍不列（不能 chat） |
| `zimage-turbo-local` | 图片生成 | 没列 | 仍不列（不能 chat） |

模型 dropdown 现在精确匹配教程的 chat 子集。

### 验证（本地实测）

- POST URL 当 key → HTTP 400 + 中文错误✅
- POST 短 key（< 16） → HTTP 400 + 中文错误✅
- 配置里 key 是 URL → key_health 标 bad_is_url，UI 显示红警告✅
- 点「🗑 清除已存 Key」→ 回落默认 key，key_health 变 ok✅
- 模型 dropdown 只有 4 个 chat 模型✅

---

## [1.7.11] — 2026-04-25

### Added — Skill 测试 AI 配置默认值（开箱即用）

之前用户首次打开 App，「🧪 测试 Skill」段需要自己粘 API Key + base URL 才能用。现在开箱即用：

- 默认 API Base：`https://clawnova.ai/api/v1`
- 默认 Key：内部团队 key（编进二进制）
- 默认模型：`qwen3.5-27b-local`

第一次打开就能直接点「🔌 测试连接」+「🚀 运行测试」，不再需要先配置。

**用户仍可覆盖**：在 UI 里粘新 key、改 base、换模型 → 点「💾 保存配置」会写到本地配置文件，之后启动以本地配置优先。

### Added — 防泄漏的 build-time 默认值机制（`defaults_local.go`）

Key 是敏感信息，不能直接写进 Go 源码（万一某天 git push 就完蛋）。所以拆成两文件：

```
defaults_local.go         ← 真实 key，**已加 .gitignore，不入 git**
defaults_local.go.example ← 空模板，进 git
.gitignore                ← 列了 defaults_local.go + 其他构建产物
```

新开发者克隆 repo 后，需要：
```
cp defaults_local.go.example defaults_local.go
# 填入真实 key
bash build-mac-app.sh
```

**安全说明（再讲一次）**：
- 真实 key 会编进二进制 → DMG 接收方能 `strings` 出来
- 所以**只能在公司内部**分发 DMG，不要上传公网
- key 一旦泄漏到外部立刻去 IDP revoke + 重发 + 改 `defaults_local.go` + 重打包

### 验证

- 删掉用户本地配置文件 → 重启 App → `GET /api/eval/config` 返回 `has_key: true` + `key_masked: kgb_d0ef***ae244f`
- 「🔌 测试连接」直接成功（无需用户先 paste key）
- 用户在 UI 改 key 并保存 → 之后所有调用走新 key（覆盖默认）

---

## [1.7.10] — 2026-04-25

### Changed — Skill 测试改为单一 AI 配置（公司本地大模型 Clawnova）

之前「测试 Skill 有效性」Tab 列了 6 个 provider（Gemini / Groq / OpenRouter / Ollama / DeepSeek / Claude），需要选一个。现在简化为**只有公司本地大模型 Clawnova**：

- API Base 默认 `https://clawnova.ai/api/v1`（OpenAI 兼容）
- 模型选项从 `/v1/models` 实测得到的 chat 模型：`qwen3.5-27b-local`、`minimax-m2.5-local`、`step-3.5-local`、`gemma-4-31b`、`qwen`
- API Key 由内部 IDP 平台签发（`kgb_...` 前缀）

**历史 provider 实现保留在 `callByProvider` 里**（gemini/groq/openrouter/ollama/deepseek/claude），如果将来要切回多 provider，把 `listProviders()` 注释段恢复即可。

### Added — Skill 测试 AI 配置持久化

配置存到 `~/Library/Application Support/TeamStandards/eval-config.json`：
- 文件权限 `0600`，仅本人可读
- **不进 git**（Home 目录与 repo 物理隔离 + .gitignore 也不会涉及）
- API：
  - `GET /api/eval/config` —— 读，返回**星号化的 key 预览**（`kgb_d0ef***ae244f`），不直接回显完整 key
  - `POST /api/eval/config` —— 保存（key 字段为空 = 不变；非空 = 覆盖）
  - `DELETE /api/eval/config/key` —— 清除 key（保留 base/model 选择）

**Eval 调用时**：`POST /api/eval/run` 不再需要前端传 key，后端自动从持久化配置读取。前端只在「保存配置」时一次性发送 key。这样 key 不会反复在 HTTP 请求体里飞，也不会落 console/network panel。

### Added — 「🔌 测试连接」按钮

不跑完整 20 条测试，只发一个最小请求验证 base+key+model 三件套是否能通：
- 发 `system: 你是测试助手 / user: 回复字符串：OK` 的最小 prompt
- 用当前保存的配置
- 返回模型回复 + token 计数 + 耗时

API：`POST /api/eval/test-connection`

### Fixed — 16K 上下文模型超限（context window exceeded）

**实测发现**：用 `qwen3.5-27b-local` 跑 eval 时报错：
```
litellm.ContextWindowExceededError: This model's maximum context length is 16384 tokens.
However, you requested 1200 output tokens and your prompt contains at least 15185 input tokens.
```

原因：`buildSystemPrompt()` 把 SKILL.md + 所有 references 全拼起来，~15K tokens。加 1200 output = 超 16384。

**修法**（两处）：

1. **新增 `buildSystemPromptCompact()`**：只装 SKILL.md（不拼 references），约 2K tokens。Clawnova 走这个。其他 provider 走原 `buildSystemPrompt()`。
2. **`ProviderConfig.MaxTokens`** 字段：clawnova 设 600（足够三段式 review），其他 provider 仍 1200 默认。

实测：clawnova qwen3.5-27b-local 跑硬编码密钥违规检测：
```
✅ pass=true · matched 7 个关键词 / 需 3 · in 2221 tokens · out 600 · 9.6 秒
```
AI 输出完整三段式 review，正确指出违反铁律 1 + 给修正代码。

### 安全说明

- 测试时用户提供的 key 已在我对话上下文中出现过；**建议测完轮换**这把 key
- 新 key 我**不嵌入二进制 / 不写进 git**，只通过 App 本地配置文件保存
- App 本身不上传任何数据到外部服务（除了你配置的那一个 endpoint）

---

## [1.7.9] — 2026-04-25

### Fixed — orangecat SKILL.md 写死了 App-specific 路径，导出 zip 给别人后 Skill 行为不对

**用户反馈**：导出的 orangecat.zip 给同事或导入 Claude Desktop 后，AI 看到 SKILL.md 里写的：

```
- 自定义（团队/个人在 Team Standards App 编辑）：
  - `~/Library/Application Support/TeamStandards/orangecat-template-qa.md`
  - `~/Library/Application Support/TeamStandards/orangecat-template-dev.md`
```

会去找这个**只在打包人 Mac 上才存在**的路径，找不到就懵 / 行为漂移。Claude Desktop 完全没有 App 这个概念，导入后 SKILL.md 里这两行更是无意义。

**修法**：把 SKILL.md 里所有外部路径删掉，改成"始终相对本 Skill 目录读模板"：

```markdown
**只用相对路径**（不依赖任何外部 App / 本地配置）：
- `references/提测报告模板_QA版.md`
- `references/提测报告模板_开发版.md`

读模板时**始终相对本 Skill 目录**（不论它装在 ~/.claude/skills/orangecat/、
Cursor 的 skill 目录、还是 Claude Desktop 的内部存储）。**禁止**写绝对路径或假设
任何宿主环境。
```

App 编辑器的存在改成一段**软提示**，说"App 用户编辑后自动覆盖 Skill 的 references/，Skill 自己不需要也不应该知道"。

**也清理了**：description 里 `并发触发：go-unit-test Skill` 这行也删掉（v1.7.6 已经把 go-unit-test UI 卡片删了，这条触发也失去意义）。

### 验证（本地实测全部 3 个导出）

| 导出 | 状态 |
|---|---|
| `go-team-standards.zip` | ✅ 23 文件（SKILL.md + 10 references + 11 demos + .golangci.yml）；SKILL.md 无写死路径 |
| `orangecat.zip` | ✅ 3 文件（SKILL.md + 2 中文模板）；UTF-8 文件名 Python 验证正确；SKILL.md 已无写死路径 |
| `SKILL.md` 复制按钮 | ✅ 两个 skill 都返回正确字节数 |

每次导出本就是从 `~/.claude/skills/<name>/` live 读取，所以你装了 v1.7.9 → 重新点「OrangeCat 独立安装」一次（覆盖你机器上的 v1.7.3 老版本）→ 再导出 zip，给别人就是干净的。

### 设计原则

之后任何 Skill 的 SKILL.md 严格遵守：
- ❌ 不写绝对路径（`/Users/xxx`、`~/Library/...`、`/etc/...`）
- ❌ 不假设宿主是 Claude Code / Cursor / Claude Desktop / 某个 App
- ✅ 引用文件只用相对 `references/` / `assets/` / `scripts/` 路径
- ✅ 告诉 AI 输出文件时用"当前项目根目录"等运行时概念，不写死位置

---

## [1.7.8] — 2026-04-23

### Changed — Shell 安装脚本（P2 方案）：打包时从本地 live 目录读取，带上发送方的本地修改

**背景**：之前 Shell 安装脚本里 payload（tar.gz）是从 App 内置 embed 读取的，所以发送方对 `~/.claude/skills/go-team-standards/` 的任何本地修改（手改 references、custom-*.md、更新过的 orangecat 模板）都**带不出去**。接收方装上看到的是 App 编译时快照。

**改为**：`buildSkillTarGz` 改成**优先从本地 live 目录**读取：

| tar 里的路径 | 来源（优先）| 回退 |
|---|---|---|
| `claude/go-team-standards/**` | `~/.claude/skills/go-team-standards/` | embed |
| `claude/orangecat/**` | `~/.claude/skills/orangecat/` | embed |
| `cursor/rules/**` | `~/.cursor/rules/` | embed |
| `assets/.golangci.yml` | embed（固定技术栈，不应改） | — |

**元信息**：tar 里多了个 `META-source.txt` 文件，记录每组的来源（live / embed）和打包时间。Shell 脚本执行时会 `cat` 这个文件，让接收方知道是从发送方 live 打包的还是 embed。

**新增**：Shell 脚本现在会装 **`orangecat`** 到 `~/.claude/skills/orangecat/`（之前只装 go-team-standards + Cursor rules）。

**验证**（本地实测）：
```
=== META-source.txt ===
打包来源：
  · claude/go-team-standards ← live: ~/.claude/skills/go-team-standards
  · claude/orangecat ← live: ~/.claude/skills/orangecat
  · cursor/rules ← live: ~/.cursor/rules
  · assets/.golangci.yml ← embed (固定技术栈)

=== 文件统计 ===
  go-team-standards: 23
  orangecat:         3（SKILL.md + 2 个模板）
  cursor/rules:      13（12 内置 + 1 用户自留的 go-expert.mdc）
```

### 设计取舍说明（P2 选型的后果）

- Shell 安装脚本现在**会把你 `~/.cursor/rules/` 下所有 `.mdc`** 都带给接收方 —— 包括你自己加的、非团队的规则文件（比如 `go-expert.mdc`）。这是"完全对称"的必然结果。
- 如果你不想分发自己的私人规则：跑 `ls ~/.cursor/rules/` 自查一下，把不想分发的 `.mdc` 临时移走再打包。
- DMG 和 zip（.app 包）**不受影响**，继续用 embed。它们的 .app 里的 Go 二进制 embed 是编译时烧进去的，运行时无法替换。
- 建议：需要把本地修改分发给团队 → 用 Shell；只是发 App 安装包 → 用 DMG。

---

## [1.7.7] — 2026-04-23

### Fixed — 覆盖检查把所有 references 错误判为"自定义"（用户看到的"12 个规范没加载"）

**Bug 根因**：v1.7.2 新增的覆盖检查 `coverage.go` 里用 `claude/go-team-standards/references/xxx.md` 作为 embed 对比源，但这些路径**其实是符号链接**指向 `../../../standards/xxx.md`。Go 的 `embeddedFS.ReadFile` 在某些情况下读不到跨目录符号链接的内容 → `hashEmbed` 返回错误 → 代码把每个文件都标成 "custom"（自定义，意为 embed 里没找到）。

```
原覆盖结果：
  Claude · go-team-standards  → 全部 custom（错误！应该是 ok）
  Cursor · rules              → 全部 ok（对）
  Claude · orangecat          → ok（对）
```

**修法**：覆盖清单改用每个文件**真实的 embed 源路径**（和 install.go 的 copy 逻辑一致）：
- `go-team-standards/references/*.md` → 来自 `standards/xxx.md`（不是 claude/go-team-standards/references/）
- `go-team-standards/demos/*` → 来自顶级 `demos/xxx`
- `go-team-standards/assets/.golangci.yml` → 来自顶级 `assets/.golangci.yml`
- Cursor rules / orangecat → 路径和 embed 一致，不变

**验证**：修完跑覆盖检查，summary 显示：
```
Claude · go-team-standards → {ok: 23}
Cursor · rules             → {ok: 12, custom: 1（老的 go-expert.mdc）}
Claude · orangecat         → {outdated: 3}（用户装的是 v1.7.3 老模板，正确！）
```

### Added — 「🔄 一键更新规范到最新」显眼大按钮（#3 A）

「⚡ 安装」Tab 新增**最显眼**的卡片（绿色边框 + 大按钮）：

```
🔄 一键更新规范到最新版
   [🔄 一键更新规范到最新]（大按钮）  [🔍 先看看和最新版差多少]
```

点「一键更新」：
1. 调 `/api/install` 重装 go-team-standards + Cursor rules
2. 并发调 `/api/orangecat/install` 更新 orangecat 模板
3. 完成后自动触发覆盖检查，让你看到全绿
4. 提醒 Cmd+Q 重启 Claude/Cursor

你通过 SM 添加的 `custom-*.md` / `custom-*.mdc` **会被保留**（install.go 的 `syncCustomToInstalled` 在重装后会再同步回去）。

### Deferred — #4 (E) 导出包对称性推到 v1.7.8

「导出的包提供改动后的规范模块」涉及架构问题：DMG 里的 `.app` 用 Go `//go:embed` 在编译时烧进去，运行时改不了。要让 DMG 包含 sender 本地的自定义规范，需要：
- 发送端：把 `~/.claude/skills/*/` 打成 sidecar 附在 DMG 里
- 接收端：App 启动时检测 sidecar，优先用 sidecar 覆盖 embed

或者简化版：Shell 安装脚本（本来就是脚本可以改）直接从 sender 本地 `~/.claude/skills/` 打包。这种方案推到下一版讨论具体实现。

---

## [1.7.6] — 2026-04-23

### Fixed — zip 导出中文文件名乱码（orangecat 的模板显示为 `??????`）

**Bug**：`cd_export.go` 写 zip 时用的是 Go `archive/zip` 的默认 FileHeader，**没有设 UTF-8 编码标志位（EFS bit，0x800）**。结果：

- `orangecat.zip` 里的 `references/提测报告模板_QA版.md` 和 `references/提测报告模板_开发版.md`
- 在 macOS `unzip` 命令、某些旧解压工具、Claude Desktop 里可能显示为 `??????????????????_QA???.md` / `_?????????.md`
- 用户看到会以为"缺文件 / 缺目录"

**修法**：写 header 时强制打标：
```go
h.NonUTF8 = false
h.Flags |= 0x800   // Language Encoding Flag: 声明 name 是 UTF-8
rel = filepath.ToSlash(rel)  // zip 规范要求正斜杠
```

**验证**：用 Python `zipfile` 读出的 `info.flag_bits` 是 `0x808`（UTF-8 + data descriptor），文件名解码为正确的中文。现代解压工具（Finder / Claude Desktop / 7-Zip / BetterZip / Python / Windows Explorer）都会正确显示。

macOS 自带的 `unzip` 命令本身太老不认这个 flag，所以命令行输出仍是 `???`，但这是工具问题不是 zip 问题。

### Removed — UI 删除「🧪 Go Unit Test Skill（模块化安装）」卡片

按用户要求移除：
- 「⚡ 安装」Tab 里的 go-unit-test 独立管理卡片（含 references / assets 勾选）
- 「📤 给 Claude Desktop」段里的 `go-unit-test` 导出行

**保留**（不删）：
- 后端 `/api/unit-test/*` 路由和 `unittest_install.go`
- embed 里的 `claude/go-unit-test/` 全部文件
- `web/app.js` 里的 handlers（用 `if (unittestCardPresent)` 门禁跳过不存在的 DOM，不会报错）

方便未来以其他形式复用（比如直接嵌入 `go-team-standards` 内部）。

---

## [1.7.5] — 2026-04-23

### Fixed — Claude Desktop zip 导出缺 `assets/` 目录

**Bug**：`cd_export.go` 的过滤逻辑把所有 dotfile（以 `.` 开头的文件）一刀切全部跳过：

```go
if strings.HasPrefix(base, ".") { return nil }  // ❌
```

这导致 `assets/.golangci.yml`（这是给 Claude/Cursor 看的真实配置）被剔除。又因为 `assets/` 下只有这一个文件，**整个 `assets/` 目录在 zip 里不出现**。用户下载后会发现 "缺少目录"。

**修法**：改成白名单 —— 只跳我们自己的 metadata 和 OS 垃圾文件：

```go
switch base {
case ".installed-version", ".DS_Store", "Thumbs.db":
    return nil
}
if strings.HasPrefix(base, "._") { return nil } // macOS AppleDouble
```

`.golangci.yml` 等正常 dotfile 保留。

**影响范围**：
- `go-team-standards.zip` 之前漏了 `assets/.golangci.yml`
- `orangecat.zip` / `go-unit-test.zip` 不受影响（它们目录里没有 dotfile）

---

## [1.7.4] — 2026-04-23

### Added — OrangeCat 提测前 3 步自验强制门禁（G8）

SKILL.md 新增 STEP 0-E「开发侧 3 步自验前置」，强制开发在提测前完成：

- **第 1 步 · 理解改动**：跑 `git diff origin/master...HEAD`，**用一段话**描述改了什么、影响哪个业务流程、在什么环节生效。这段话同时进 QA 版 §1 开头「开发方的理解」段，**测试同学据此确认**开发理解与需求一致。
- **第 2 步 · 核心场景**：最少数量，不穷举。每个问题 1 条场景即可：
  - 🟢 能生效吗？（正确配置下功能按预期工作）
  - 🔴 能关掉吗？（有开关/配置的关掉后不影响原流程）
  - 🟡 边界在哪？（临界值行为）
- **第 3 步 · dev 环境真跑**：每条场景必须有「操作 + 结果 + 证据（日志/DB/截图）」。

**G8 门禁**：
- 开发版 §6「开发自验记录」三小段都非空
- 3 条场景都有操作 + 结果 + 证据
- 任一场景「失败」未修复 → **拒绝生成**（除非显式标注"带病提测 + 原因 + 影响范围 + 同意人"）

### Changed — 开发版模板增加 §6 开发自验记录

开发版模板新增段落顺延：
- §6 **🔬 开发自验记录**（新）
- §7 ✅ 开发自测报告（原 §6 顺延）
- §8 🔍 真实性校验记录（原 §7 顺延，且新增 G7/G8 校验行）

### Changed — QA 版 §1 新增「🧠 开发方的理解」小段

QA 同学首先看到开发自己对需求的理解，一句话确认「改了什么 / 影响哪个业务流程 / 在什么环节生效」—— 这是自验 Step 1 的产物，放在业务段开头，核对完再往下看接口变更等细节。

### Removed / Hidden — 「新建项目」Tab 隐藏（团队已有 IDP 平台）

左侧导航「🏗️ 新建项目」按 `display:none` 隐藏。tab 面板和后端 `/api/scaffold/*` 路由保留（不删除脚手架代码 / embed），方便未来恢复或紧急使用。

---

## [1.7.3] — 2026-04-23

### Fixed — OrangeCat QA 版缺"业务维度"说明，测试同学看不懂改了什么

**用户反馈**：v1.7.2 的 QA 版只有「接口变更清单（表格）」+「SQL 数据影响」，全是技术清单，QA 同学看不到**这次是做了什么业务 / 为什么改**。

**修法**：
- QA 版模板最前面加 **§1 本次业务改动**（给 QA 的上下文）
  - 🎯 做了什么（一句话业务总结）
  - 📝 具体改动（按业务功能分组，每组带证据 commit + 代码注释引用）
  - 🤔 为什么改（业务动机）
  - 🧭 测试重点建议（必测 / 建议覆盖 / 可不测）
- 开发版同步添加相同的 §1，内容与 QA 版完全同步（开发自测时也需理解全貌）
- 原先所有段落序号顺延（概览 1→2，接口 2→3，SQL 3→4，自测 4→5）

### Added — SKILL.md STEP 0-D 业务范围深度挖掘 + G7 门禁

SKILL.md 新增明确挖掘流程：

1. **读 commit body**：`git log origin/master..HEAD --format='%H%n%s%n%n%b%n---'`
   - 提炼"做了什么" + "为什么"
   - 忽略 `chore:` / `refactor:` / `style:` 这类非业务 commit
2. **读改动文件的 godoc / 中文业务注释**（`git diff` 过滤 `*.go` `*.php`）
   - 引用时保留**原文 + 文件路径 + 行号**作为证据
3. **按业务功能分组**（多功能 PR 不许糊成一段）
4. **业务动机**：JIRA/PRD 链接 → commit body → 明示"推断"兜底，不许当成事实

**G7 门禁**（新增，不过拒绝生成）：
- 「做了什么」必须业务视角，禁"修改了 XX 服务"这种技术描述
- 「具体改动」每组必带至少 1 条证据 commit + 1 条代码注释引用
- 「为什么改」commit 没讲就必须写"未找到明确动机"，不许编
- 禁用「详见 §3」这种偷懒甩锅

### 隐患说明

- G7 门禁靠 AI 执行度。如果 commit message 写得太烂（全是 `update`、`fix xxx`），AI 也榨不出业务信息，会如实标"未找到明确动机" —— 这是**功能而不是缺陷**，反过来倒逼团队写好 commit。
- 代码注释挖掘需要项目里真的有 godoc / 中文注释。没注释的项目出来的业务段会比较干瘪，告诉 AI 的话术也会明确提示"建议补注释"。

---

## [1.7.2] — 2026-04-23

### Changed — OrangeCat 报告拆成 QA 版 + 开发版两个文件（P2 方案）

原单文件模板删除。现在 OrangeCat 触发后会生成**两个**文件：

- `提测报告_<branch>_<时间>_QA版.md` —— 给测试同学（简版）
  - 概览 / 接口变更 / SQL 数据影响（不含上线顺序）/ 自测 checklist
- `提测报告_<branch>_<时间>_开发版.md` —— 给程序员（完整）
  - 概览 / SQL 上线顺序 & 回滚 / 脚本影响分析 / 建议测试关注点 / 自测 checklist / 真实性校验记录

**拆分原则**：
- 📋 概览 + ✅ 自测报告 —— 两边都有（共享）
- 🔌 接口变更 —— QA 侧
- 🗄️ SQL 变更 —— 两边都有，但「⚠️ 上线顺序/回滚」只在开发版
- ⚙️ 脚本影响 / 🧪 建议测试关注点 / 🔍 真实性校验 —— 开发侧

### Added — 提测人自动从 git author 取（门禁 G6）

SKILL.md 新增强制步骤：生成前跑 `git log -1 --format='%an <%ae>'`，拿到的作者作为提测人。若用户手填了别的，以 git author 为准 + 报告里注明"手填 X 被覆盖为 Y"。防止"提测人填老板"这种乱象。

### Added — 模板编辑器支持 QA / 开发双版本

App「🐱 OrangeCat」卡片的模板编辑器改成**双 Tab**：`📤 QA 版` + `🛠️ 开发版`。两份分开保存：

- `~/Library/Application Support/TeamStandards/orangecat-template-qa.md`
- `~/Library/Application Support/TeamStandards/orangecat-template-dev.md`

切 Tab 会保留未保存的 draft（不丢）。

API：`GET/POST/DELETE /api/orangecat/template?which=qa|dev`

### Added — 「📊 规范覆盖检查」按钮

主 Tab「安装」页新增卡片，扫描你已安装的规范文件 vs App 内置清单：

- ✅ **ok** 已装且内容 hash 匹配（最新版）
- 🟡 **outdated** 已装但 hash 不同（老版本 / 被改过）
- ❌ **missing** 应该有但缺失
- 🔷 **custom** 用户 `custom-*.md` 白名单（始终 ok）

扫描范围：
- `~/.claude/skills/go-team-standards/`（SKILL.md + 10 个 references）
- `~/.cursor/rules/`（12 个 .mdc）
- `~/.claude/skills/orangecat/`（SKILL.md + QA/开发两个模板）
- `~/.claude/skills/go-unit-test/`（SKILL.md，可选模块按装了哪些显示）

有缺失/过时时启用「📥 一键补齐」按钮，点了触发主安装重写 go-team-standards + Cursor rules。orangecat / go-unit-test 仍需各自卡片"独立安装"。

后端：`GET /api/coverage/check` → 返回分组清单 + 每项状态 + 计数汇总

### Changed — `go-unit-test` description 加「提测」次级触发器

```yaml
- 【次级触发】提测 / QA 交付 / 生成提测报告 / OrangeCat
  —— 与 orangecat 并发激活，验证 Go 单测覆盖是否达标
```

实际效果：用户说"提测"时，`orangecat` + `go-team-standards` + `go-unit-test` 三个 Skill 并发激活。`go-unit-test` 负责检查 Go 单测覆盖率是否足够（帮开发在自测 checklist 里填真实数字）。

### Removed — PHP 单测 Skill（曾在 v1.7.2 讨论范围，已撤）

原计划加 `php-unit-test` Skill，用户反馈自建框架不支持 PHPUnit TestCase 继承，撤销。保留 Go 单测 Skill。

---

## [1.7.1] — 2026-04-23

### Added — 脚手架生成 Skill 有效性测试用例

新建项目「选项」区新增勾选：「🧪 生成 Skill 有效性测试用例」（默认不勾）。

勾选后会把 15 个**故意写错**的 Go 文件 + README 写到 `<项目>/.docs/skill-test-cases/`：

| 文件 | 违反铁律 | 期望 AI 指出 |
|---|---|---|
| 01_hardcoded_secret.go | #1 密钥 | API Key 硬编码 |
| 02_bare_errors_new.go | #2 错误码 | 裸 errors.New 不走 xerror |
| 03_silent_error.go | #3 丢 error | `_ = err` 吞错 |
| 04_business_panic.go | #4 panic | 业务代码 panic |
| 05_naked_goroutine.go | #5 goroutine | 无 ctx 裸 goroutine |
| 06_float_money.go | #6 金额 | float64 存金额 |
| 07_local_time.go | #7 时间 | 本地时间不带时区 |
| 08_fmt_println_log.go | #8 日志 | fmt.Println 打日志 |
| 09_select_star_n_plus_1.go | #9 SQL | SELECT * + N+1 + OFFSET |
| 10_fk_constraint.go | #10 外键 | DB 外键 |
| 11_dead_commented_code.go | #11 死代码 | 注释掉的代码 |
| 12_log_sensitive_data.go | #12 敏感数据 | 日志打印手机号 |
| 13_no_timeout_http.go | #13 IO 超时 | 无超时 / 不传 ctx |
| 14_high_card_metric.go | #14 label 基数 | user_id 作 label |
| 15_test_antipattern.go | 单测反模式 | 手写 mock + 真 DB + 文案比较 |

每个文件头带 `//go:build skill_test_only` build tag，`go build` / `go test` **不会编译**。
放到 `.docs/` 目录 + 自动追加到 `.gitignore`，**不进 git**。

用法：在 Claude Code / Cursor 里打开任意一个文件，说"这段代码帮我看看"，期望 Skill 自动激活并按三段式指出违规。

### Added — 本地模板路径「📁 浏览…」按钮

新建项目页「模板来源 → 本地目录」输入框旁新增「📁 浏览…」按钮。点击弹 macOS 原生文件夹选择器（`osascript choose folder`），选中后自动填回输入框 + 立刻触发模板有效性校验。

`/api/scaffold/pick-folder` 现支持 `?prompt=xxx` 参数自定义对话框提示文案（用于区分「输出目录」/「模板目录」），文案做了简单 sanitize 防止 AppleScript 注入。

### Changed — 「选项」和「common-lib 组件」默认折叠

新建项目页第 4、5 两大段（「选项」「common-lib 组件」）改为 `<details>` 折叠容器，默认收起，减少首屏信息密度。用户展开后可见已按推荐预选。

Skill 测试用例勾选项放在「选项」段末，勾选后展开一段说明文字。

### Fixed — 脚手架 `.gitignore` 追加时保留末尾换行

`.gitignore` 处理时如果末尾没有换行，追加前会补一个，避免把 `.docs/` 接到已有条目后面。

---

## [1.7.0] — 2026-04-23

### Breaking — 提测模板 Skill 改名 + 完全重做

原 `tixuebj-template` Skill 改名为 **`orangecat`**（OrangeCat · 提测文档）。**旧名目录在安装/重装主 Skill 时会被自动清理**，API 路径 `/api/tixuebj/*` 移除不保留，前端只调 `/api/orangecat/*`。

老用户升级后：
- 旧 `~/.claude/skills/tixuebj-template/` 和 `~/.cursor/skills-cursor/tixuebj-template/` 会被自动删掉
- 新 `~/.claude/skills/orangecat/`、`~/.cursor/skills-cursor/orangecat/` 需要在 App 里点「独立安装」

### Changed — 报告模板从 7 段重设为 6 段

**删除**：改动文件清单 / Commit 列表（对 QA 无价值）

**新增/加强**：
1. 📋 本次提测概览
2. 🔌 **接口变更清单** —— 表格，每行必须写清"下游影响" + 兼容性
3. 🗄️ **SQL 变更清单** —— DDL + 执行方式 + 回滚
4. ⚙️ **脚本/任务影响分析** —— 触发时机 + 失败后果 + 回滚方式
5. ✅ **开发自测报告** —— 6 项 checklist + 失败用例 + 自测发现的 bug
6. 🧪 建议测试关注点 —— P0/P1/P2

### Added — 报告生成前的"真实性校验"门禁（强阻断）

Skill 指令里新增 STEP 0 门禁（G1-G5）：下游影响必具体 / SQL 文件完整 / 脚本影响展开 / 自测勾选 / 用 Grep 抽查报告内容真实存在。任一不过即拒绝生成。

### Added — 报告模板在 App 里可编辑

OrangeCat 卡片展开「📝 报告模板」，内嵌 textarea 编辑器。保存到 `~/Library/Application Support/TeamStandards/orangecat-template.md`。点「独立安装」时优先使用用户版，否则用内置。支持恢复默认。

API：`GET/POST/DELETE /api/orangecat/template`

### Added — `go-unit-test` Skill（独立 + 模块化安装）

专门针对"如何写 Go 单元测试"的独立 Skill，与 `go-team-standards` 并列。

**分层**：
- **基础层 · 必装**：`SKILL.md`（10 条铁律 + 场景路由 + 完整表驱动骨架 + 自查清单），自给自足
- **可选层 · 默认不装**：
  - 8 个 references：test-structure / table-driven / mock-gomock / assertions / anti-patterns / coverage / fixtures / integration
  - 3 个骨架：skeleton-usecase / skeleton-http-handler / skeleton-repo-sqlmock

**UI**：「🧪 Go Unit Test Skill」卡片，两折叠区勾选。Manifest 驱动，取消勾选会真删文件。

API：`GET /api/unit-test/manifest`、`GET /api/unit-test/status`、`POST /api/unit-test/install`（body 传 modules 数组）、`POST /api/unit-test/uninstall`

### Changed — `go-team-standards` Skill 触发机制强硬化

- description 改成强制前置激活措辞；触发信号按 7 大类展开（语言/框架/基础设施/行为/对象/文件类型/交付物）
- SKILL.md 新增 **ZERO STEP**：写代码前 5 步必做（识别场景 → 读铁律 → 读 references → 读 custom-*.md → 才动手）+ 自检信号
- **铁律从 12 条升到 14 条**：
  - #13：所有外部 IO（HTTP / DB / MQ / Redis）必带超时 + ctx/trace 必串通
  - #14：Prometheus metric label 基数 ≤ 100，禁 user_id / trace_id / order_id 作 label

### Changed — Cursor `00-iron-laws.mdc` globs 扩展

`alwaysApply: true` 基础上新增 globs：`**/*.go`, `**/*.sql`, `**/*.proto`, `**/go.mod`, `**/go.sum`, `**/Dockerfile`, `**/Makefile`。同时补铁律 #13 #14。

### Removed

- `tixuebj_install.go` → 替换为 `orangecat_install.go`
- `claude/tixuebj-template/` → 替换为 `claude/orangecat/`
- `/api/tixuebj/*` 路由 → 全删，不保留别名

---

## [1.6.2] — 2026-04-22

### Fixed — 老版本 macOS 无法打开（14.x Sonoma 及以下用户全跪）

**根因**：构建机是 2026 年的 Mac + Xcode 最新 SDK。没显式设置 `MACOSX_DEPLOYMENT_TARGET`，clang 默认按**构建机当前 macOS 的 minimum runtime** 链接。结果产出的二进制加载器里写着"需要 macOS 15+"。收件人是 M1 Pro + macOS 14.5 Sonoma，直接被 dyld 拒绝加载，报错形如"不支持此版本 / Mach-O header 不兼容"。

**修法**：
- `build-mac-app.sh` 编译 arm64 + x86_64 时都显式加：
  - `MACOSX_DEPLOYMENT_TARGET=11.0`（环境变量）
  - `-mmacosx-version-min=11.0`（clang 参数 × 3：CC / CGO_CFLAGS / CGO_LDFLAGS）
- Info.plist 的 `LSMinimumSystemVersion` 从 10.13 改到 11.0，与二进制实际支持的一致

**兼容范围**：
- ✅ macOS 11 Big Sur（2020）
- ✅ macOS 12 Monterey / 13 Ventura / 14 Sonoma（**用户的 14.5 覆盖**）/ 15 Sequoia
- ✅ 所有 M 系列（M1/M1 Pro/M1 Max/M2/M3/M4）
- ✅ Intel Mac（2020 年 Catalina/Big Sur 时代以后的）

为什么最低 11.0 而不是 10.13？**WKWebView 稳定版**和团队用的 cgo 依赖都用到了 11+ 的 API（notarization-friendly 的代码签名等）。10.13 理论上能支持但实际风险高，11.0 是合理下限。

## [1.6.1] — 2026-04-22

### Added — 提测模板 Skill 独立管理
- ⚡ 安装 Tab 新增 **📝 提测模板 Skill（独立管理）** 专属卡片，和 go-team-standards 的"立即安装"解耦
- **状态显示**：启动时自动检测 Claude + Cursor 两侧安装状态，三色徽章（✓ 双装 / ⚠ 只装一侧 / ○ 未装）
- **4 个按钮**：
  - 📥 **独立安装**：不用装整个 go-team-standards 也能单独要这个 skill
  - 🗑 **卸载**：两步式确认（WKWebView confirm 不工作的替代）
  - 📦 **导出 zip**：供他人 Claude Desktop 上传
  - 📋 **复制 SKILL.md**：供粘贴到 Claude Desktop Editor

### Fixed
- **提测 Skill 之前只装 Claude 侧**：v1.6.0 的 `立即安装` 流程只写了 `~/.claude/skills/tixuebj-template/`，没写 Cursor 侧。v1.6.1 起 Cursor `~/.cursor/skills-cursor/tixuebj-template/` 也同步写入，Cursor 3.x 能识别到
- 日志里多一条 "Cursor Skill → ... (提测模板)" 证明双写生效

### Backend
- `tixuebj_install.go` 新增 3 端点：`/api/tixuebj/status`、`/install`、`/uninstall`

## [1.6.0] — 2026-04-22

### Added — 两大功能

#### 📝 提测模板 Skill（独立 skill）
- 新增 `tixuebj-template` skill（嵌入 App）。当用户说"提测 v0.0.1"、"帮我写提测文档"、"QA handoff"等触发
- SKILL.md 指导 AI：读 `git diff` 和 `git log`，按模板填字段，输出到项目根 `提测_<tag>_<日期>.md`
- 模板 `提测模板_v0.0.1.md` 包含 7 大段：改动清单 / 测试重点 / 风险评估 / 回滚 / 依赖配置 / 下游影响 / 附录
- 安装时跟 go-team-standards 一起装到 `~/.claude/skills/tixuebj-template/`
- ⚡ 安装 Tab 的"给 Claude Desktop"卡片加了**多 skill 导出**：每条独立的 zip / SKILL.md 复制按钮

#### 🧩 组件傻瓜教程（下一步.md 增强）
- 新建项目时，勾选的每个 common-lib 组件在 `下一步.md` 追加一个"快速验证"段落：
  - 🐘 **PostgreSQL**：docker run 单行命令 + curl 下单 + psql 查表
  - 🔴 **Redis**：docker + redis-cli 读写 + 加 demo 的 AI prompt 建议
  - 📬 **Kafka**：KRaft 模式单机启动 + topic 创建 + demo 引用
  - 🚨 **Alarm**：kafka-console-producer 塞测试消息 + 看 Telegram 输出
  - ⏰ **Cron**：`RegisterHandler` 示例 + `cron_tasks` 表插入
- 不勾任何组件 → 说明 Hello World 最小模式下只有 4 个探针 endpoint

### Backend
- `installEmbeddedSkill(embedDir, dst)` 通用函数，任意嵌入的 skill 都能安装
- `writeComponentTutorials()` 按勾选组件生成对应章节，不勾不写

## [1.5.2] — 2026-04-22

### Added — 导出给 Claude Desktop (claude.ai)
**⚡ 安装** Tab 底部新增 "📤 给 Claude Desktop 用" 卡片，两个按钮：

- **📦 下载 Skill zip**：流式下载 `go-team-standards.zip`（含 SKILL.md + references/ + demos/ + assets/），到 Claude Desktop `Customize → Skills → +` 上传即可，完整带细节
- **📋 复制 SKILL.md**：调 `navigator.clipboard.writeText`（失败回退 `document.execCommand`）把 SKILL.md 纯文本塞剪贴板，到 Claude Desktop Editor 视图粘贴，轻量但不带 references

### Backend
- `cd_export.go` 新增 2 端点 `/api/claude-desktop/export-zip` 和 `/api/claude-desktop/skill-md`
- `?name=` 参数支持导出任意已装 skill（默认 go-team-standards），加 `safeSkillName()` 防路径穿越

## [1.5.1] — 2026-04-22

### Fixed
- **新建项目输入框字符过滤**：
  - **项目名**（sfName）：只允许 `[a-z0-9-]`，大写键入自动转小写，其他字符（`_` / 中文 / 空格 / 符号）直接过滤掉，保留光标位置
  - **业务域**（sfBiz）：更严格，只允许 `[a-z0-9]`（不含短横线），匹配后端 `isAlphaNumSimple` 校验规则
- 根因：之前只做了后端正则校验（`^[a-z][a-z0-9-]{1,40}$`），用户输入 `Zoo` 可以顺利填完表才提交报错，体验差。现在前端**实时阻止**非法字符键入，光标不跳。

## [1.5.0] — 2026-04-22

### Added — 新建项目支持勾选 common-lib 组件
**新建项目**表单加了第 5 节"common-lib 组件"，5 个复选框 + 依赖锁定：

| 组件 | 默认 | yaml 处理 |
|---|---|---|
| 🐘 PostgreSQL | ✅ 勾 | 勾则取消注释 `data.database`；不勾保持注释 |
| 🔴 Redis | ✅ 勾 | 勾则取消注释 `data.redis` |
| 📬 Kafka | ✅ 勾 | 勾则**追加** `mq.kafka` 段（sample 原始无） |
| 🚨 Alarm（依赖 Kafka） | ✅ 勾 | 勾则追加 `alarm` 段 |
| ⏰ Cron（依赖 PostgreSQL） | ✅ 勾 | 无 yaml（DB 表驱动，勾则说明激活） |

**路线 B 软开关**：代码不删，只改 yaml 注释状态。利用 `NewData` 对空 config 的容错设计，yaml 段注释 = 组件不启用。好处：
- 零侵入模板（升级 sample-service 不用重适配）
- 用户生成后可随时手动取消/加回注释切换，不用重新 scaffold
- 比硬裁（删 Go 代码）稳健

**依赖联动**：勾 Alarm 自动勾 Kafka 并 **🔒 锁定**（不可取消）；勾 Cron 自动勾 PG 锁定。UI 上锁定的卡片橙色边框 + "🔒 被依赖锁定" 角标。

### Backend
- `applyComponentsToYAML` 新函数：取消注释或追加 yaml 段
- `uncommentYAMLSection` 工具：按 section 名去掉整段 `#` 前缀，保留缩进
- `Components.EnforceDeps()` 后端再执行一次依赖检查（防前端绕过）

## [1.4.1] — 2026-04-22

### Fixed — 新建项目脚手架 3 个关键 bug（从 `下一步-修订说明.md` 实测修订）

用户实际跑了一次生成的项目，发现 `make proto` → `make run` 全程撞坑，根因都是 scaffold 阶段漏处理：

1. **Makefile 的配置文件引用没替换**
   - 现象：`make run` 报 `no such file or directory: configs/dev/order.yaml`（项目已改为 `<biz>.yaml`）
   - 修：文本替换表新增 `configs/dev/<sampleBiz>.yaml → configs/dev/<newBiz>.yaml`

2. **buf.gen.yaml 三个错**（`make proto` 生成到错目录 + panic）
   - `plugins[*].out: .` 应为 `out: api`
   - `go_package_prefix` 缺 `/api` 后缀，生成的 import 和业务代码对不上
   - 没有 `managed.disable` googleapis，annotations.proto 被改写成不存在的包
   - 修：新增 `rewriteBufGenYaml(dest, module)`，**直接覆写**为已知可用版（不 regex 缝补）

3. **残留 *.pb.go 被文本替换污染导致 panic**
   - sample 仓库的 `api/<biz>/v1/*.pb.go` 是预生成文件，含 protobuf 描述符字节（长度前缀编码）
   - 我的文本替换改了字符串内容但字节长度前缀没同步 → `make run` panic `slice bounds out of range [-4:]`
   - 修：新增 `removeStaleProtoGenFiles(dest)`，生成后删所有 api/ 下 `*.pb.go`，让 `make proto` 重生

### Fixed — 端口替换更 robust
原 regex 只匹配 `:8000` / `:9000`，sample 实际是 `:8002` / `:9002` → 端口替换静默失败。
改为**逐行扫描 http/grpc section**，遇到 `addr:` 行就替换其中的端口数字，不依赖 sample 原始端口值。

### Changed — 下一步.md 内容修订
对照用户实测路径改：
- Hello World 阶段的 Step 4 改为"保持 database/redis 注释态 + NewData 容错"，不再推荐手工改 Go 代码
- Step 5 的 curl 改为 `/health` 和 `POST /api/v1/order`（sample 实际路由），不是不存在的 `/api/v1/<biz>/ping`
- `configs/dev/<biz>.yaml` 示例字段结构换成 `data.database.driver + source`（conf.proto 实际定义），不是 `postgres.dsn`
- 删除 `naming.TopicInProject` 建议（当前项目代码里不存在）
- schema 检查说明改为"当前模板未按 schema 分域，后续按团队规范自行加"

### Impact
从初始 scaffold → `make proto` → `make wire` → `make run` → curl `/health` 一把跑通。之前踩过的 panic / 找不到 order.yaml / 生成到错误目录都不会再撞。

## [1.4.0] — 2026-04-22

### Fixed — Cursor 读不到规则的根本问题
Cursor 3.x 不可靠地扫描 `~/.cursor/rules/*.mdc`（有些版本不读）。正确的官方 skill 路径是 **`~/.cursor/skills-cursor/<name>/SKILL.md`**（Claude 格式同构）。

### Changed — 双写兼容
每个 Cursor 目标都**同时写两处**，哪个能读生效都行：
1. `~/.cursor/rules/*.mdc`（老格式，rules 机制）
2. `~/.cursor/skills-cursor/<name>/SKILL.md`（新格式，skill 机制，Cursor 3.x 推荐）

涉及功能：
- **团队规范**：skills-cursor/go-team-standards/（完整 SKILL.md + references/ + demos/）
- **AI 人格**：skills-cursor/persona/SKILL.md
- **SM 自定义规则**：skills-cursor/custom-&lt;id&gt;/SKILL.md

### Added — 醒目重启 Banner
安装完成 / 人格保存后，右下角弹橙色渐变 Banner（固定浮动、淡入动画）：
> ⚠️ 必做：彻底重启 Claude Code 和 Cursor
> Cmd+Q 退出（不是 reload window / 关窗口），然后重新打开。不重启 = 新规则不会加载

点"知道了"关闭。之前只在 log 里提一下重启，很多人忽略，现在 Banner 逃不过眼。

### 卸载逻辑
- 人格卸载：同时删 `~/.cursor/rules/00-persona.mdc` + `~/.cursor/skills-cursor/persona/`
- 自定义规则删除：同时清两处 skills-cursor 下的目录 + rules 下的 .mdc

## [1.3.1] — 2026-04-22

### Fixed — 大幅瘦身，消除"人格重复"

之前 v1.3.0 同时塞了**两套相似功能**导致用户困惑：
- 🎭 "AI 行为准则" Tab (`prompt-rules`) — 多规则列表 + 预设标题锁死 = "无法编辑"
- 🎭 "AI 人格" Tab (`persona`) — 单 prompt 直接编辑

两套都往 `~/.claude/CLAUDE.md` 写 marker 段（不同 marker），所以用户看 CLAUDE.md 就是**两段几乎一样的内容** = "人格重复"。

### Changed
- **只保留 AI 人格 Tab**（`persona`，单一全局 prompt，可自由编辑），彻底删除重复的 AI 行为准则 Tab
- **删除 Skill 市场**：市场 tab 和 find-skills tab 都砍（用户明确说不要）
- 代码清理：删除 `promptrules.go` / `marketplace.go` / `findskills.go`，删除对应 HTTP 路由和 UI

### Added
- **启动自动清理旧残留**：v1.3.1 启动时自动清理 ≤1.3.0 留在磁盘的痕迹
  - 从 `~/.claude/CLAUDE.md` 移除旧 `<!-- TEAM_STANDARDS_PROMPT_BEGIN/END -->` 段（persona 段保留）
  - 删 `~/.cursor/rules/00-attitude.mdc`（persona 走 00-persona.mdc，attitude 是旧文件）
  - 删 `~/.team-standards/prompt-rules.json`
  - 清理动作记进 🪵 日志控制台

## [1.3.0] — 2026-04-22

### Added — 两大新功能

#### 🎭 AI 行为准则 Tab
强制本地 Claude Code + Cursor 带着固定"人格"回答：
- **4 个预设规则**（首次启动自动写入，默认全部启用）：
  - 严谨求实（不编造 API / 不伪装确定 / 承认知识边界）
  - 不敷衍（不骗 token，读完全文再答，不生成占位糊弄）
  - 对代码负责（不留死代码、错误处理非平凡、改了就承认）
  - 勤恳思考（主动看 context、多方案权衡、反向校验）
- 可新建自定义规则，作用范围：全局 / 指定项目
- 启用/禁用切换、编辑、删除（preset 不能删只能禁）
- **写入位置**：
  - `~/.claude/CLAUDE.md` 的 managed section（用 `<!-- TEAM_STANDARDS_PROMPT_BEGIN -->` 标记隔离，不动你已有的 CLAUDE.md 内容）
  - `~/.cursor/rules/00-attitude.mdc`（`alwaysApply: true`，每次响应都加载）
  - 项目 scope 写到 `<项目>/.claude/CLAUDE.md` + `<项目>/.cursor/rules/00-attitude.mdc`
- 切换/编辑时自动 resync 到磁盘

#### 🔎 Skill 市场 Tab
从 GitHub 一键装官方 Anthropic skills：
- 内置 6 个热门官方 skill：PDF / PPTX / DOCX / XLSX / skill-creator / consolidate-memory
- 安装：下 tar.gz（**不需 git**）→ 提取子目录到 `~/.claude/skills/<id>/`
- 卸载：`rm -rf`，**仅限 "other" 类**（禁动团队 go-team-standards 和 custom-*）
- 卡片标记"✓ 已装"状态

### Architecture
- `promptrules.go`：CRUD + managed section 合并逻辑
- `marketplace.go`：tar.gz 下载 + 子目录提取（纯 Go，零外部依赖）
- 侧栏计数角标显示启用的规则数量

## [1.2.3] — 2026-04-22

### Fixed
- **🪵 日志控制台里看不到非 Claude provider 的失败**：原 `callAnthropic` 埋点只覆盖 Anthropic，Gemini / Groq / OpenRouter / DeepSeek / Ollama 的 API 错误（包括 Gemini 429 `limit: 0` 这种典型坑）**全部漏记**，用户切到日志控制台"只看失败"是空的。
- 修法：把 `logOp` 从 `callAnthropic` 挪到 `callByProvider` 入口，**所有 provider 统一日志**。`callAnthropic` 只做 HTTP 传输，错误冒泡由上层捕获记日志。
- 现在日志控制台里能看到：
  ```
  api  gemini.generate     model=gemini-2.0-flash in=0 out=0      ✗
      错误：gemini 429: Quota exceeded... limit: 0
  ```

## [1.2.2] — 2026-04-22

### Fixed
- **SM 卡片标题被挤成竖排（每字一行）**：当自定义规则的 globs 很长（如 `internal/service/**,internal/server/**,**/handler/**`），flex 布局里的 scope 胶囊占满，把 `strong` 标题挤到 1 字符宽度触发逐字换行。修法：
  - `.custom-item-head` 改 `align-items: flex-start`
  - `strong` 加 `flex:1 min-width:0 white-space:nowrap text-overflow:ellipsis`
  - `.custom-scope` 限 `max-width: 50%` + 省略号
  - JS 里 scope 长于 24 字符时只显示第一条 glob + "…"

### Changed — 卡片文案瘦身
- Padding 减小（12/14 → 10/12 px）
- 标题单行超长自动省略（hover 显示 title 全文）
- 简介最多 2 行（`-webkit-line-clamp: 2`）+ hover 显示全文
- Scope 胶囊最大占卡片一半宽，超出省略
- 按钮更紧凑（`.btn-sm` 2px 8px + 11px 字号）
- 预设模板卡片同步瘦身

## [1.2.1] — 2026-04-22

### Fixed
- **Gemini 模型列表过时**：`gemini-2.0-flash-exp` 在 v1beta API 已下架，导致 404 `models/... is not found`。更新成当前正式版 ID：`gemini-2.0-flash` / `gemini-2.0-flash-lite` / `gemini-2.5-flash` / `gemini-1.5-flash-002` / `gemini-1.5-pro-002`。**默认改成 `gemini-2.0-flash`**（Google 确认的免费稳定版）。

### Added — 两种零 API Key 测试方案
- **🔍 SKILL 结构自检**：静态扫描本地 `~/.claude/skills/go-team-standards/`，毫秒级完成。检查项：
  - SKILL.md 存在 + frontmatter name/description 正确
  - 12 条铁律**每条**都能在 SKILL.md 里找到
  - 技术栈前提（common-lib / xerror / decimal）都有提及
  - 路由表指向 references/ 和 demos/
  - 违规三段式反馈原则存在
  - 自定义规则强制加载指令（`custom-*.md`）存在
  - references/ 目录 8 个核心文件齐全
  - **自定义规则同步一致性**（JSON 源 N 条 ↔ references 里 custom-*.md 也 N 条）
  - demos/ 目录存在
- **📋 导出测试 prompt**：生成 `skill-test-prompts-v*.md` 到 `version/`，含 20 个违规代码 + 操作指南。用户打开文件，手工粘每段到自己日常用的 Cursor / Claude Code，**skill 已装好的真实环境**里看 AI 反应。
  - 每个用例带：违规代码、期望命中的关键词、3 个手工勾选点
  - 这比 API 调用更真实：测的是**你日常工作场景**下 skill 的效果

### 为什么加这两个
用户反馈 "有没有不需要 API key 就能测的方案"。Claude API 要付费，Gemini/Groq 即使免费也要注册。**这两个方案完全零门槛**：
- 结构自检：文件就在本地，不走网络
- 导出 prompt：用你自己已经在用的 Cursor / Claude Code 跑（它们本身你就已经配好了）

## [1.2.0] — 2026-04-22

### Added — 多 AI Provider 支持（免费可测）

**🧪 测试 Skill** Tab 现在支持 **6 个 AI provider**，不再绑死 Claude 付费：

| Provider | 费用 | 推荐场景 |
|---|---|---|
| **Gemini (Google)** ⭐ | 免费 1500 req/天 | 默认首选，AI Studio Key 秒申请 |
| **Groq (Llama)** ⭐ | 免费 30 req/min | 推理最快 500+ tok/s |
| **Ollama 本地** ⭐ | 完全免费 | 有本地 GPU/Mac 的首选 |
| **OpenRouter** | 部分免费（`:free` 后缀） | 想试多个模型 |
| **DeepSeek** | ~$0.14/M token（便宜） | 性价比高 |
| **Claude** | 付费（cache 90% 折扣） | 推理最准 |

### Architecture
- `providers.go`：统一 `ProviderConfig` + `APIResult` 抽象，内部分发到 5 种 API 协议：
  - Anthropic Messages API（Claude，保留 prompt cache）
  - Google Generative AI（Gemini，含 systemInstruction）
  - OpenAI 兼容（Groq / OpenRouter / DeepSeek，共享一个实现）
  - Ollama（OpenAI 兼容模式，localhost:11434/v1，无 auth）
- `eval.go` 用 `callByProvider()` 替代直调 `callAnthropic`
- API 返回每个 provider 的 token 用量（Claude 额外返 cache read/create）

### UI
- 每个 provider 做成**卡片选择**（推荐卡带 "推荐" 角标 + 绿色 ✅ 免费字样）
- 选定 provider 后：API Key label、placeholder、提示链接、模型下拉**动态切换**
- Ollama 的 "API Key" 字段自动变成 "端点 URL（可选）"

### Free tier note
- **Gemini**：aistudio.google.com/apikey 申请，1500 req/天
- **Groq**：console.groq.com/keys，30 req/min（够跑 20 条）
- **OpenRouter**：模型名带 `:free` 后缀的完全免费，其他按量
- **Ollama**：`ollama pull qwen2.5-coder:32b` 后本地跑

## [1.1.2] — 2026-04-22

### Added
- **🪵 日志控制台 Tab**：内存环形缓冲最近 200 条操作，方便排错。
  - **4 类事件**：`http`（所有 /api/* 请求）· `shell`（所有 exec.Command，含 scaffold 里的 git/cp/chmod/mv/go mod tidy）· `api`（Claude API 调用含 token 统计）· `info`（业务关键步骤）
  - **筛选**：按类型勾选 + "只看失败" 快速定位问题
  - **自动刷新 2s**：切进 Tab 启动，切走自动停
  - **点击行展开**：看完整 detail 和 error 堆栈
  - **顶部统计条**：总数 / 失败数 / 各类型分布
- **HTTP middleware 自动埋点**：新加 `logMiddleware` 包装所有 `/api/*` handler，无需手动埋点
- **Shell 命令日志**：`runIn()` 记录命令行、工作目录、exit code、耗时
- **Claude API 日志**：记录 model / input_tokens / output_tokens / cache_read / cache_create，一眼看出 cache 命不命中

### Changed
- 默认侧栏多一个 **🪵 日志控制台** 入口（放 🧪 测试 Skill 下面），计数角标显示总条数

## [1.1.1] — 2026-04-22

### Fixed
- **Shell 导出实际调的是 zip 端点**：`savePackage()` 的 endpoint 路由只处理 'dmg' / 'zip'，shell 分支漏写，点击 📜 Shell 按钮时走 fallback 到 `/api/save-zip`。改正为三分支路由，shell 现在能正确生成自解压脚本。
- **测试用例 20/20 全失败**：根因是默认模型 `claude-sonnet-4-5` 作为纯 alias 在某些 API 版本下返回 `model_not_found`（所以 6.9s 内 20 次快速 404，全部 0 token）。默认模型改成**带日期后缀**的 `claude-sonnet-4-5-20250929`（100% 稳定），alias 保留作为选项。

### Added
- **🔌 先跑 1 条测连通** 按钮：在跑全部 20 条前先用第 1 条试水，验证 API key + 模型能正常响应，避免白费 cache_creation 费用。
- **API 错误置顶展示**：失败用例的 API 错误消息（如 `authentication_error` / `not_found_error`）直接显示在卡片 header 下红底块，不用展开也能一眼看出是什么问题。
- **运行失败智能提示**：如果 20 条全部带 error，日志区直接列出 4 种最常见原因（key 格式 / 模型 ID / 配额 / 网络），不用自己猜。
- **模型列表扩充**：新增 `claude-3-5-sonnet-20241022` 作为兼容性最强的选项（所有历史 API 版本都支持）。

## [1.1.0] — 2026-04-22

### Added — 两大功能

#### 📜 Shell 自解压安装脚本（终极兼容方案）
- 安装 Tab 新增 **📜 生成 Shell（终极）** 按钮
- 产物：单个 `install-skills-v1.1.0-*.sh` 文件（~700KB）
- 内嵌所有 skill 文件的 base64 tar.gz + 当前用户自定义 SM 规则
- 对方只需：`chmod +x install-skills-*.sh && ./install-skills-*.sh`
- **完全不依赖 App / Go / Xcode / 签名**，只要系统有 bash（macOS 默认就有）
- 同时兼容 macOS 和 Linux（`base64 -D || base64 -d` 双适配）
- 适用场景：对方 macOS 太旧 App 打不开、企业禁装未签名应用、Linux 开发机

#### 🧪 Skill 自动化评估（Claude API + prompt caching）
- 新 Tab **🧪 测试 Skill**，左侧导航带用例计数
- 20 个核心违规代码用例（12 铁律 + 8 common-lib 替换表）
- 每个用例：违规 Go 代码 + 期望关键词 + 最少命中数
- 调用 Claude Messages API：
  - 把已装的 `~/.claude/skills/go-team-standards/SKILL.md` + 全部 references 作为 system prompt
  - 加 `cache_control: ephemeral`，20 条用例只首条付 full input cost，后续 90% 折扣
  - 违规代码作为 user message，要求 AI 按团队三段式反馈
- 自动 assert：AI 回复里是否命中 keywords 数量 ≥ min_match
- 报告页：通过率进度条 + 耗时 + cache token 统计 + 每个用例可展开查看 AI 实际回复 + 命中/未命中关键词 chip
- 模型下拉：支持 claude-sonnet-4-5 / claude-opus-4-1 等

### Changed
- VERSION 1.0.6 → **1.1.0**（两个大功能算 minor bump）
- 分发按钮从 2 个变 3 个：DMG（推荐）/ zip（兼容）/ Shell（终极）

### 20 条用例一览
- **铁律 12 条**：硬编码密钥 / 裸 errors.New / 丢弃 err / 业务 panic / 裸 goroutine / float 金额 / 本地时间 / fmt.Println / SELECT * / 数据库外键 / 注释代码 / 敏感数据入日志
- **common-lib 替换 8 条**：裸 sarama / 裸 go-redis / 手工 otel / 手拼 Redis key / 硬编码 topic / 裸 net/http.Client / 包名 util/common/helper / JsonConfig（缩写大小写）

## [1.0.6] — 2026-04-21

### Added
- **DMG 分发包格式**（推荐）：安装 Tab 新增 `💽 生成 DMG（推荐）` 按钮，调 `hdiutil create -format UDZO` 打 macOS 原生磁盘镜像。DMG 挂载时 macOS 原生读取，**签名元数据一位不丢**，对方端不存在"第三方解压器丢 xattr"的风险。解决 Sonoma (macOS 14+) 出现的"一按钮 无法打开（已损坏）"。
- **zip 和 DMG 双按钮**：zip 保留作兼容性 fallback，DMG 作为主推方式。

### Fixed
- **`首次打开失败-双击修复.command` 救不了 Sonoma "一按钮 损坏" 错误**：根因是 zip 传输破坏了 ad-hoc 签名的 bytes 对齐，简单 `xattr -cr` 不够。新版脚本：
  1. 搜 4 个候选位置找 .app（.command 旁 / Downloads / Desktop / Applications）
  2. 用 `osascript` 一次性弹 admin 密码框，`mv` 到 `/Applications`（避开 macOS App Translocation）
  3. `codesign --remove-signature` 清掉破损的签名
  4. `codesign --force --deep --sign -` 在接收端重签，bytes 对齐 100% 正确
  5. `spctl --add` 白名单
  6. `lsregister -kill -r` 重置 LaunchServices 缓存（让 macOS 忘掉"这 app 坏了"）
- 从根上解决 "双击修复.command 没用" 的问题。

### Explained
- **一按钮 vs 两按钮**错误在使用指导里补充对照表（一按钮=签名损坏，Gatekeeper 直接拒绝不给放行；两按钮=未签名但完整，可绕过）。

## [1.0.5] — 2026-04-21

### Fixed
- **自定义 Skill 无法删除**：根因是 WKWebView 默认不实现 `WKUIDelegate` 的 `runJavaScriptConfirmPanelWithMessage` 回调，所以 `window.confirm()` 返回 `undefined`，前一版里 `if (!confirm(...)) return;` 把删除动作静默吞掉了。现在换成**两步式确认**：第一次点"删除"按钮变红脉冲 "再点一次确认"（3 秒后自动复位），再点才真删。彻底绕开 WKWebView 的 JS 弹窗限制。

### Added
- **.danger-active 按钮脉冲动画**：删除第二步按钮红底 + 闪烁呼吸，视觉上不会误点。

## [1.0.4] — 2026-04-21

### Added
- **工具链安装步骤**：生成的 `下一步.md` 和 `docs/使用教程.md` 都加了 Step 2 · 装工具链小节，明确列出 5 个必装 CLI：`buf` / `wire` / `protoc-gen-go` / `protoc-gen-go-grpc` / `protoc-gen-go-http/v2`，统一用 `CGO_ENABLED=0 go install` 绕过老 macOS 无 Xcode 的限制。
- **PATH 自检**：下一步里加自动检测 `~/go/bin` 是否在 PATH，没有就写入 `~/.zshrc` 的一行命令。
- **排错表补充**：使用教程的启动失败排查表加了 "buf/wire not found" 和 "老 macOS 没 Xcode" 两条常见 case。

### Fixed
- 原来 `下一步.md` 只写 `go install buf@latest` 没加 `CGO_ENABLED=0`，老 macOS 环境下仍可能触发 CC 检测失败。

## [1.0.3] — 2026-04-21

### Added
- **WKWebView 键盘快捷键**：全局 `keydown` 监听 `Cmd+A / C / X / V / Z / Shift+Z`，在表单里手动实现全选/复制/剪切/粘贴/撤销/重做。webview_go 默认 Mac 窗口不带 Edit 菜单，这步前所有输入框都不能 Cmd+V —— 现在可以了。
- **Hello World 一步到位指引**：`下一步.md` 重写，从"配置 GOPRIVATE"到"curl /api/v1/biz/ping"分 4 个 Step 带命令；`docs/使用教程.md` 第 8.3 节同步新增"从 0 到 Hello World"章节 + 8.4 启动失败排查表。
- **mask-go-common-lib 内部依赖提示**：生成的 `下一步.md` 明确给出 `GOPRIVATE` + `git config url.insteadOf` 一次性配置命令。

### Changed
- **Git 模式默认值**：模板 Git URL 预填 `git@your-gitlab.com:infrastructure/go-infrastructure/mask-go-sample-service.git`，分支预填 `main`。GitLab API Base 和 Group 也预填常用值。

## [1.0.2] — 2026-04-21

### Added
- **脚手架支持 Git 仓库作为模板来源**：新建项目 Tab 的"模板来源"改成 Radio 二选一 — 本地目录 / Git 仓库。选 Git 时填 URL（如 `git@your-gitlab.com:...`），App 自动 `git clone --depth=1` 到临时目录，做完 scaffold 清理掉。
- **新项目 Git 远端自动配置**：`git init -b main` → `git add .` → 初始提交 → `git remote add origin <新 URL>`。新 Remote URL 根据模板 URL + 项目名自动推导（替换最后一段 path），也可手改。
- **GitLab API 自动建仓（可选）**：高级折叠区填 GitLab API Base + Group + Personal Access Token 就能在远端直接创建同名空仓库。不填 Token 时给出手工创建 + push 的命令。
- **推导 Remote URL 端点** `GET /api/scaffold/derive-remote` 供 UI 联动。

### Changed
- **Logo 背景改透明**：侧栏 🦕 图标去掉青蓝渐变方框，只显示 emoji 本体；App .icns 图标同步改为透明背景，只有恐龙 emoji（通过 Apple Color Emoji 字体位图放大到 1024×1024）。

## [1.0.1] — 2026-04-21

### Fixed
- **SKILL.md 强制读取自定义规则**：在 SKILL.md 第 9 行加 🔴 强制动作，要求 Claude 处理任何 Go/SQL/proto 代码前必须列出并完整读取 `references/custom-*.md`。之前只在文件末尾一句话带过，Claude 基本不会主动加载，导致 SM 自定义 Skill 编码时不生效。
- **重装时不再清空自定义 Skill**：`handleInstall` 完成后自动调用 `syncCustomToInstalled`，把 `~/.team-standards/custom-rules.json` 里的规则重新写回 Claude references 和 Cursor rules，避免 `installClaude` 的 `rm -rf` 把 `custom-*.md` 一起删掉。

### Added
- **🏗️ 新建项目 Tab**：基于 `mask-go-sample-service` 模板一键生成新微服务。支持填表（项目名 / 业务域 / Go Module / HTTP+gRPC 端口 / 作者 / 输出目录）+ macOS 原生文件夹选择器 + 选项（空骨架 / git init / go mod tidy）。后端自动替换 module、目录重命名（api/order→api/biz）、proto package、Dockerfile 后缀、configs 端口。生成后写 `下一步.md` 列出手动检查清单。
- **📋 已装清单 Tab**：扫描 `~/.claude/skills/` 和 `~/.cursor/rules/`，按「团队 / 自定义 / 其他」三色分类展示，每项显示名称、description、大小、修改时间、🔍 在 Finder 中显示。
- **Universal Binary + Ad-hoc 签名**：App 二进制合并 arm64 + x86_64，同时支持 M1-M4 和 Intel Mac；ad-hoc codesign + 清 quarantine xattr，避免收件人 Mac 报"文件已损坏"。
- **LSMinimumSystemVersion 降到 10.13**：兼容 macOS High Sierra (2017) 以上所有版本。
- **分发 zip 改用 `ditto` 打包**：Apple 官方推荐方式，保留 .app bundle 签名、权限、符号链接、资源分叉，避免跨机传输破坏 bundle。
- **首次打开失败-双击修复.command**：zip 里附带的一键解锁脚本，移除 quarantine 标记 + re-sign + 打开 App。
- **版本归档**：分发 zip 自动保存到 `~/skills/version/`，文件名含版本号和时间戳，不覆盖历史。

### Changed
- **SKILL.md 大幅瘦身**（约 10KB → 4.3KB）：移除重复的速查表（命名、DB schema、API、Commit 等），这些内容在对应 `references/*.md` 已有详细版本；SKILL.md 只留路由表和铁律。
- **description 兼容通用 Go 场景**：不再仅限 Kratos 微服务，扩到任何 Go 代码（命名、错误、日志、金额、时间、并发、测试）。
- **Logo 从 "TS" 改成 🦕**：侧栏图标 + App icon（1024×1024 .icns，Pillow + Apple Color Emoji 渲染）。

## [1.0.0] — 2026-04-21

### Added
- 首次发布：整合团队 8 份规范文档 + 术语表 + 轻量 demo 模板 + 图形化 Installer。
- **Claude Skill**：`go-team-standards`，含铁律 + 路由表 + 9 份 reference。
- **Cursor Rules**：10 份 `.mdc`，按 glob 自动触发；`00-iron-laws` + `02-naming-logging` + `99-cursor-security` 全局生效。
- **规范模块**：
  - `go-style` — Go 编码风格（项目结构、命名、并发、注释、lint）
  - `naming-logging` — 仓库 / 包 / 中间件 / 日志字段命名 + W3C Trace Context
  - `error-codes` — 受控错误码体系（IDP 注册 + codegen + CI linter）
  - `api-design` — RESTful + gRPC（URL / 响应 / 版本 / 安全）
  - `database` — PostgreSQL schema / 字段类型 / 索引 / migration / 审批
  - `testing` — 覆盖率门禁 + 表驱动 + Mock 规则
  - `commit` — Conventional Commits + commitlint
  - `ci-pipeline` — GitLab CI 六阶段 + 质量门禁
  - `cursor-usage` — Cursor Enterprise 使用与安全红线
  - `glossary` — 术语表（Go / 云原生 / 可观测 / 团队 common-lib）
- **轻量 demo 模板** (`demos/`)：Kratos 服务、Kafka 生产消费、PG 迁移 + GORM CRUD、Redis 幂等、Wire DI、errno/xerror、slog 追踪、表驱动测试 — 10 个可自由组合的代码片段。
- **Installer App**：Go 单文件二进制（`team-standards-installer`），本地 Web UI，Skills / Demos / Glossary / History 四个 Tab；支持一键全局 / 项目安装。

### Known
- IDP 错误码仓库尚未集成，`errno` 包需要团队 IDP 平台上线后联通。
- Demo 模板基于 `mask-go-common-lib` v1 API；common-lib 升级时需同步 demo。
