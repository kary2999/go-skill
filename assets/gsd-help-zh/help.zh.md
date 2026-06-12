<purpose>
展示完整的 GSD 命令参考（中文版）。仅输出参考内容。不要添加项目特定分析、git 状态、下一步建议或参考之外的评论。
</purpose>

<reference>
# GSD 命令参考

**GSD**（Get Shit Done）为 Claude Code 单人 agentic 开发场景生成分层项目计划。

## 快速开始

1. `/gsd-new-project` - 初始化项目（含调研、需求、roadmap）
2. `/gsd-plan-phase 1` - 为第一个 phase 创建详细 plan
3. `/gsd-execute-phase 1` - 执行 phase

## 保持升级

GSD 迭代很快，定期升级：

```bash
npx get-shit-done-cc@latest
```

## 主流程

```
/gsd-new-project → /gsd-plan-phase → /gsd-execute-phase → 循环
```

### 项目初始化

**`/gsd-new-project`**
通过统一流程初始化新项目。

一条命令从想法走到可以开始 plan：
- 深度提问理解你要构建什么
- 可选领域调研（spawn 4 个并行 researcher agent）
- 需求定义（带 v1 / v2 / 范围外 分级）
- Roadmap 创建（phase 拆分 + 成功标准）

生成所有 `.planning/` 产物：
- `PROJECT.md` — 愿景与需求
- `config.json` — workflow 模式（interactive/yolo）
- `research/` — 领域调研（如选了）
- `REQUIREMENTS.md` — 带 REQ-ID 的范围化需求
- `ROADMAP.md` — phase 与需求映射
- `STATE.md` — 项目记忆

用法：`/gsd-new-project`

**`/gsd-map-codebase [--fast] [--focus <area>] [--query <term>]`**
为存量 codebase（brownfield）做映射。

- `--fast` — 快速轻量评估（替代旧 `gsd-scan`）
- `--focus <area>` — 把 map 范围圈到某个领域
- `--query <term>` — 查 `.planning/intel/` 里的 codebase 情报索引（替代旧 `gsd-intel`）

- 用并行 Explore agent 分析代码库
- 在 `.planning/codebase/` 下生成 7 份聚焦文档
- 覆盖：技术栈、架构、结构、约定、测试、集成、关注点
- 在存量 codebase 上跑 `/gsd-new-project` 前用

用法：`/gsd-map-codebase`

### Phase 规划

**`/gsd-discuss-phase <number> [--chain | --analyze | --power | --assumptions] [--batch[=N]]`**
计划 phase 前帮你清晰表达想法。

- `--chain` — 链式 prompt discuss 流
- `--analyze` — 深度假设分析
- `--power` — power-user 模式（扩展提问集）
- `--assumptions` — 不开 interactive session，直接列出 Claude 对 phase 的实现假设

- 捕获你心目中 phase 应该如何运作
- 创建 CONTEXT.md（愿景 / 必要项 / 边界）
- 当你对实现有想法时用
- `--batch` 一次问 2-5 个相关问题（不是一个一个问）

用法：`/gsd-discuss-phase 2`
用法：`/gsd-discuss-phase 2 --batch`
用法：`/gsd-discuss-phase 2 --batch=3`

**`/gsd-mvp-phase <number> [--force]`**
按垂直 MVP 切片规划 phase —— 3 个结构化 user-story 提示（`作为 X / 我想 Y / 以便 Z`），故事过大就 SPIDR 切分，然后委托 `/gsd-plan-phase` 进 MVP 模式。

- 改 phase 的 ROADMAP 条目：写 `**Mode:** mvp` + 用拼好的 user story 替换 `**Goal:**`
- 通过 `gsd-sdk query user-story.validate` 校验故事（正则 `/^As a .+, I want to .+, so that .+\.$/`）
- `--force` 跳过状态守护（phase 已是 `in_progress` 或 `completed` 时需要）
- 配合 new-project 的模式提问（垂直 MVP vs 水平分层）

用法：`/gsd-mvp-phase 1`
用法：`/gsd-mvp-phase 2 --force`

**`/gsd-plan-phase <number> [--research] [--skip-research] [--research-phase <N>] [--view] [--gaps] [--skip-verify] [--tdd] [--mvp]`**
为指定 phase 创建详细执行 plan。

- `--skip-research` — 跳过 researcher subagent
- `--research-phase <N>` — 仅调研模式。为 phase `<N>` 跑 researcher，写 `RESEARCH.md`，然后在 planner 跑之前退出。适合：跨 phase 调研、决定 plan 方法前先 doc review、纠正而不重新 plan 的循环。替代已删除的 `gsd-research-phase` 独立命令（#3042）。
  - 修饰符：`--research` 强制刷新（重 spawn researcher，不提问）。`--view` 把现有 `RESEARCH.md` 输出到 stdout 而不 spawn。都没给的话，`RESEARCH.md` 已存在时提示 `update / view / skip`。
- `--gaps` — 仅关注前次 plan-check 的缺口
- `--skip-verify` — 跳过 plan 之后的 verifier 循环
- `--tdd` — 按测试驱动顺序规划（先测试后代码）
- `--mvp` — 垂直切片 MVP 规划模式

- 生成 `.planning/phases/XX-phase-name/XX-YY-PLAN.md`
- 把 phase 拆成具体可执行任务
- 含验证准则和成功度量
- 一个 phase 支持多 plan（XX-01、XX-02 等）

用法：`/gsd-plan-phase 1`
用法：`/gsd-plan-phase --research-phase 2` — 仅 phase 2 调研（如已有 RESEARCH.md 会提示）
用法：`/gsd-plan-phase --research-phase 2 --view` — 打印现有 RESEARCH.md，不 spawn
用法：`/gsd-plan-phase --research-phase 2 --research` — 强制刷新，不提问
结果：创建 `.planning/phases/01-foundation/01-01-PLAN.md`

**PRD 快速路径**：传 `--prd path/to/requirements.md` 直接跳过 discuss-phase。你的 PRD 变成 CONTEXT.md 里锁定的决策。已有清晰验收标准时好用。

### 执行

**`/gsd-execute-phase <phase-number> [--wave N] [--gaps-only] [--tdd]`**
执行 phase 里所有 plan，或跑指定 wave。

- `--wave N` — 仅执行 wave N（见下面 *每个 wave 内的 plan*）
- `--gaps-only` — 仅重跑前次 verifier 标记为 gaps 的 plan
- `--tdd` — 强制按测试驱动顺序执行

- 按 frontmatter 里 wave 分组 plan，wave 顺序串行执行
- 每个 wave 内的 plan 通过 Task 工具并行跑
- 可选 `--wave N` 仅跑 wave N 然后停（除非 phase 已完全 complete）
- 所有 plan 完成后验证 phase 目标
- 更新 REQUIREMENTS.md / ROADMAP.md / STATE.md

用法：`/gsd-execute-phase 5`
用法：`/gsd-execute-phase 5 --wave 2`

### 智能路由

**`/gsd-progress --do "<description>"`**
把自由文本自动路由到对应 GSD 命令。

- 分析自然语言找最匹配的 GSD 命令
- 作为 dispatcher —— 自己不做活
- 歧义时让你在 top 候选间选
- 当你知道要做啥但不知道用哪个 `/gsd-*` 时用

用法：`/gsd-progress --do "修登录按钮"`
用法：`/gsd-progress --do "重构 auth 系统"`
用法：`/gsd-progress --do "我要开新 milestone"`

### 快速模式

**`/gsd-quick [--full] [--validate] [--discuss] [--research]`**
带 GSD 保证执行小型临时任务，但跳过可选 agent。

Quick 模式走同套系统但路径更短：
- spawn planner + executor（默认跳过 researcher / checker / verifier）
- Quick 任务存 `.planning/quick/`，与计划好的 phase 分开
- 更新 STATE.md 跟踪（不更新 ROADMAP.md）

flag 开启额外质量步骤：
- `--full` — 完整质量流水线：discussion + research + plan-checking + verification
- `--validate` — 仅 plan-checking（最多 2 次迭代）和执行后 verification
- `--discuss` — 轻量 discussion，规划前暴露灰色地带
- `--research` — 聚焦的 research agent 在规划前调研方法

flag 可组合：`--discuss --research --validate` 等同于 `--full`。

用法：`/gsd-quick`
用法：`/gsd-quick --full`
用法：`/gsd-quick --research --validate`
结果：创建 `.planning/quick/NNN-slug/PLAN.md` 和 `.planning/quick/NNN-slug/NNN-slug-SUMMARY.md`

---

**`/gsd-fast [description]`**
内联执行琐碎任务 —— 无 subagent / 无 planning 文件 / 无开销。

适合小到不值得 planning 的任务：拼写修复、配置改动、忘了的 commit、简单添加。在当前上下文跑，改完提交并记到 STATE.md。

- 不生 PLAN.md 或 SUMMARY.md
- 不 spawn subagent（内联跑）
- ≤ 3 处文件编辑 —— 任务不简单时重定向到 `/gsd-quick`
- atomic commit 带 conventional message

用法：`/gsd-fast "修 README 拼写"`
用法：`/gsd-fast "把 .env 加进 gitignore"`

### Roadmap 管理

**`/gsd-phase <description>`**
在当前 milestone 末尾加新 phase。

- 追加到 ROADMAP.md
- 用下一个顺序编号
- 更新 phase 目录结构

用法：`/gsd-phase "加 admin dashboard"`

**`/gsd-phase --insert <after> <description>`**
插入紧急工作为小数 phase 到已存在 phase 之间。

- 创建中间 phase（如 7.1 在 7 和 8 之间）
- 适合 milestone 中途发现必须做的活
- 维持 phase 顺序

用法：`/gsd-phase --insert 7 "修严重 auth bug"`
结果：创建 Phase 7.1

**`/gsd-phase --remove <number>`**
删除未来 phase 并对后续 phase 重新编号。

- 删除 phase 目录和所有引用
- 把所有后续 phase 重编号填补空缺
- 仅对未来（未开始）phase 有效
- Git commit 保留历史记录

用法：`/gsd-phase --remove 17`
结果：Phase 17 被删，phase 18-20 变成 17-19

**`/gsd-phase --edit <number> [--force]`**
原位编辑已存在 roadmap phase 的任意字段，保留编号和位置。

- 更新 `ROADMAP.md` 里的标题、描述、需求、依赖
- `--force` 允许编辑已开始的 phase（慎用）

### Milestone 管理

**`/gsd-new-milestone <name>`**
通过统一流程开启新 milestone。

- 深度提问理解你接下来要做什么
- 可选领域调研（spawn 4 个并行 researcher agent）
- 范围化需求定义
- Roadmap 创建（phase 拆分）
- 可选 `--reset-phase-numbers` 把编号重置为 Phase 1，并先归档旧 phase 目录确保安全

镜像 `/gsd-new-project` 流程，给存量项目用（已有 PROJECT.md）。

用法：`/gsd-new-milestone "v2.0 Features"`
用法：`/gsd-new-milestone --reset-phase-numbers "v2.0 Features"`

**`/gsd-complete-milestone <version>`**
归档已完成 milestone 并准备下个版本。

- 在 MILESTONES.md 创建条目附统计
- 把完整细节归档到 milestones/ 目录
- 为这个 release 创建 git tag
- 为下个版本准备 workspace

用法：`/gsd-complete-milestone 1.0.0`

### 进度跟踪

**`/gsd-progress [--next | --forensic | --do "<description>"]`**
查项目状态并智能路由到下一动作。

- 显示可视化进度条和完成率
- 从 SUMMARY 文件总结近期工作
- 显示当前位置和下一步
- 列出关键决策和未解 issue
- 提议执行下个 plan 或创建（如缺失）
- 检测 100% milestone 完成

模式：
- **default** — 进度报告 + 智能路由
- **`--next`** — 自动推进到下个逻辑步骤（`--next --force` 跳过安全门）
- **`--forensic`** — 在进度报告后附 6 项完整性审计
- **`--do "<text>"`** — 智能路由：把自由意图分派到匹配的 `/gsd-*` 命令（见上面 *智能路由*）

用法：`/gsd-progress`
用法：`/gsd-progress --next`
用法：`/gsd-progress --forensic`

### 会话管理

**`/gsd-resume-work`**
从上次会话恢复工作，完整还原上下文。

- 读 STATE.md 拿项目上下文
- 显示当前位置和近期进度
- 基于项目状态提议下一动作

用法：`/gsd-resume-work`

**`/gsd-pause-work [--report]`**
phase 中途暂停时创建上下文交接。

- `--report` — 在 `.planning/reports/` 生成会话后摘要（commit / 文件改动 / phase 进度）
- 创建 .continue-here 文件附当前状态
- 更新 STATE.md 的会话连续性段
- 捕获进行中的工作上下文

用法：`/gsd-pause-work`

### 调试

**`/gsd-debug [issue description] [--diagnose]`**
跨上下文重置仍保留状态的系统化调试。

- `--diagnose` — 跑一次性诊断，不打开持久 debug session

- 通过自适应提问收集症状
- 创建 `.planning/debug/[slug].md` 跟踪调查
- 用科学方法调查（证据 → 假设 → 测试）
- 撑过 `/clear` —— 不带参跑 `/gsd-debug` 恢复
- 解决的 issue 归档到 `.planning/debug/resolved/`

用法：`/gsd-debug "登录按钮不工作"`
用法：`/gsd-debug`（恢复活跃 session）

### Spike 与 Sketch

**`/gsd-spike [idea] [--quick]`**
通过一次性实验快速 spike 验证想法可行性。

- 把想法分解为 2-5 个聚焦实验（按风险排序）
- 每个 spike 回答一个具体的 Given/When/Then 问题
- 写最小代码、跑、捕获判定（VALIDATED / INVALIDATED / PARTIAL）
- 存 `.planning/spikes/` 含 MANIFEST.md 跟踪
- 不依赖 `/gsd-new-project` —— 任何 repo 都能用
- `--quick` 跳过分解，立即开干

用法：`/gsd-spike "能不能用 WebSocket 流式输出 LLM？"`
用法：`/gsd-spike --quick "测下 pdfjs 能不能抽出表格"`

**`/gsd-sketch [idea] [--quick]`**
用一次性 HTML mockup 多变体探索快速草绘 UI / 设计想法。

- 构建前对话式拿氛围 / 方向
- 每个 sketch 产 2-3 个变体（tab HTML 页）
- 用户对比变体、挑元素、迭代
- 共享 CSS 主题系统跨 sketch 累积
- 存 `.planning/sketches/` 含 MANIFEST.md 跟踪
- 不依赖 `/gsd-new-project` —— 任何 repo 都能用
- `--quick` 跳过氛围环节，直接构建

用法：`/gsd-sketch "admin panel 的 dashboard 布局"`
用法：`/gsd-sketch --quick "表单卡片分组"`

**`/gsd-spike --wrap-up`**
把 spike 发现打包成持久化项目 skill。

- 逐个 curate spike（include/exclude/partial/UAT）
- 按特性领域分组发现
- 生成 `./.claude/skills/spike-findings-[project]/` 含 references 和源
- 写 summary 到 `.planning/spikes/WRAP-UP-SUMMARY.md`
- 给项目 CLAUDE.md 加自动加载路由行

用法：`/gsd-spike --wrap-up`

**`/gsd-sketch --wrap-up`**
把 sketch 设计发现打包成持久化项目 skill。

- 逐个 curate sketch（include/exclude/partial/revisit）
- 按设计领域分组发现
- 生成 `./.claude/skills/sketch-findings-[project]/` 含设计决策、CSS 模式、HTML 结构
- 写 summary 到 `.planning/sketches/WRAP-UP-SUMMARY.md`
- 给项目 CLAUDE.md 加自动加载路由行

用法：`/gsd-sketch --wrap-up`

### 捕获想法、笔记、待办

**`/gsd-capture [description]`**
从当前对话中把想法或任务捕获为结构化 todo。

- 从对话提上下文（或用提供的描述）
- 在 `.planning/todos/pending/` 创建结构化 todo 文件
- 从文件路径推断领域分组
- 创建前查重
- 更新 STATE.md 的 todo 计数

用法：`/gsd-capture`（从对话推断）
用法：`/gsd-capture 加 auth token 刷新`

**`/gsd-capture --note <text>`**
零摩擦笔记捕获 —— 一条命令、即时保存、不提问。

- 保存带时间戳的笔记到 `.planning/notes/`（或全局 `$HOME/.claude/notes/`）
- 3 个子命令：append（默认）、list、promote
- promote 把笔记变成结构化 todo
- 没项目也能跑（fallback 到全局范围）

用法：`/gsd-capture --note 重构 hook 系统`
用法：`/gsd-capture --note list`
用法：`/gsd-capture --note promote 3`
用法：`/gsd-capture --note --global 跨项目想法`

**`/gsd-capture --list [area]`**
列出待办 todo 并挑一个开干。

- 列所有待办 todo（标题、领域、年龄）
- 可选 area 过滤（如 `/gsd-capture --list api`）
- 加载选中 todo 的完整上下文
- 路由到合适动作（立即做、加入 phase、头脑风暴）
- 开始工作时把 todo 移到 done/

用法：`/gsd-capture --list`
用法：`/gsd-capture --list api`

### 用户验收测试

**`/gsd-verify-work [phase]`**
通过对话式 UAT 验证已构建特性。

- 从 SUMMARY.md 文件抽出可测交付物
- 一次呈现一个测试（yes/no 响应）
- 失败自动诊断并创建修复 plan
- 发现问题准备重新执行

用法：`/gsd-verify-work 3`

### 发布

**`/gsd-ship [phase]`**
从已完成 phase 创建 PR，自动生成 body。

- 推 branch 到 remote
- 创建 PR（summary 取自 SUMMARY.md / VERIFICATION.md / REQUIREMENTS.md）
- 可选请求 code review
- 更新 STATE.md 发布状态

前置：phase 已验证、`gh` CLI 已装且已认证。

用法：`/gsd-ship 4` 或 `/gsd-ship 4 --draft`

---

**`/gsd-review --phase N [--gemini] [--claude] [--codex] [--coderabbit] [--opencode] [--qwen] [--cursor] [--all]`**
跨 AI peer review —— 调外部 AI CLI 独立 review phase plan。

- 检测可用 CLI（gemini / claude / codex / coderabbit）
- 每个 CLI 用相同结构化 prompt 独立 review plan
- CodeRabbit review 当前 git diff（不是 prompt）—— 最多需要 5 分钟
- 产 REVIEWS.md（含每个 reviewer 反馈 + consensus summary）
- 把 review 喂回 plan：`/gsd-plan-phase N --reviews`

用法：`/gsd-review --phase 3 --all`

---

**`/gsd-pr-branch [target]`**
通过过滤掉 .planning/ commit 创建干净的 PR 分支。

- 给 commit 分类：纯代码（含入）、纯 planning（排除）、混合（含但去 .planning/）
- 在干净分支上 cherry-pick 代码 commit
- Reviewer 只看到代码改动，不看 GSD 产物

用法：`/gsd-pr-branch` 或 `/gsd-pr-branch main`

---

**`/gsd-capture --seed [idea]`**
捕获带触发条件的前瞻性想法，自动浮现。

- Seed 保留 WHY、何时浮现、关联代码的面包屑
- `/gsd-new-milestone` 时触发条件匹配会自动浮现
- 比延后项好 —— 触发会被检查，不会被遗忘

用法：`/gsd-capture --seed "事件系统做完后加实时通知"`

**`/gsd-capture --backlog [description]`**
把想法加进 backlog 停车场，留给将来 milestone。

- 在 ROADMAP.md 用 999.x 编号创建 backlog 条目
- 保留想法但不承诺到当前 milestone
- 通过 `/gsd-review-backlog` 后续浮现 / 提升

用法：`/gsd-capture --backlog "事件上线后做实时通知"`

---

**`/gsd-audit-uat`**
跨 phase 审计所有未完成 UAT 和验证项。

- 扫每个 phase 找 pending / skipped / blocked / human_needed 项
- 与 codebase 交叉引用检测过期文档
- 产带优先级的人工测试 plan（按可测性分组）
- 开新 milestone 前清验证债

用法：`/gsd-audit-uat`

### Milestone 审计

**`/gsd-audit-milestone [version]`**
对照原始意图审计 milestone 完成度。

- 读所有 phase 的 VERIFICATION.md
- 检查需求覆盖
- spawn integration checker 检查跨 phase 接线
- 创建 MILESTONE-AUDIT.md 含缺口和技术债

用法：`/gsd-audit-milestone`

### 配置

**`/gsd-settings`**
交互式配置 workflow 开关和模型 profile。

- 切换 researcher / plan checker / verifier agent
- 选模型 profile（quality / balanced / budget / inherit）
- 更新 `.planning/config.json`

用法：`/gsd-settings`

**`/gsd-config [--profile <profile> | --advanced | --integrations]`**
配置 GSD 超出基本设置：模型 profile / 高级调优 / 第三方集成。

- `--profile <profile>` — 快切模型 profile（`quality | balanced | budget | inherit`）
- `--advanced` — power-user 调优：plan 反弹 / 超时 / branch 模板 / 跨 AI 执行（替代旧 `gsd-settings-advanced`）
- `--integrations` — 第三方 API key / code-review CLI 路由 / agent-skill 注入（替代旧 `gsd-settings-integrations`）

- `quality` — Opus 全场（除验证）
- `balanced` — Opus 做规划，Sonnet 做执行（默认）
- `budget` — Sonnet 写，Haiku 调研 / 验证
- `inherit` — 所有 agent 用当前 session 模型（OpenCode `/model`）

用法：`/gsd-config --profile budget`

### 工具命令

**`/gsd-cleanup`**
归档已完成 milestone 的累积 phase 目录。

- 找已完成 milestone 但仍在 `.planning/phases/` 的 phase
- 动手前显示 dry-run summary
- 把 phase 目录移到 `.planning/milestones/v{X.Y}-phases/`
- 多 milestone 后用，减少 `.planning/phases/` 杂乱

用法：`/gsd-cleanup`

**`/gsd-help`**
显示英文版命令参考。

**`/gsd-help-zh`**
显示**本文档（中文版命令参考）**。

**`/gsd-update [--sync] [--reapply]`**
升级 GSD 到最新版，预览 changelog。

- `--sync` — 跨 runtime 根目录同步托管的 GSD skill（替代旧 `gsd-sync-skills`）
- `--reapply` — 升级后重新应用本地修改（替代旧 `gsd-reapply-patches`)

- 展示已装 vs 最新版本对比
- 显示你错过的版本的 changelog 条目
- 高亮 breaking changes
- 装之前确认
- 比裸 `npx get-shit-done-cc` 好

用法：`/gsd-update`

## 额外命令

上面是日常最常用的。下面是按用途分组的所有 `/gsd-*` slash 命令。

### 发现与规范

- **`/gsd-explore`** — 苏格拉底式 ideation 和 idea 路由。先想清楚再做 plan。
- **`/gsd-spec-phase <phase> [--auto] [--text]`** — 带歧义评分澄清 phase 交付物，discuss-phase 前产出 SPEC.md。
- **`/gsd-ai-integration-phase [phase]`** — 为含 AI 系统的 phase 生成 AI-SPEC.md 设计契约。
- **`/gsd-ui-phase [phase]`** — 为前端 phase 生成 UI 设计契约（UI-SPEC.md）。
- **`/gsd-import --from <filepath> | --from-gsd2`** — 带项目决策冲突检测导入外部 plan，或把 GSD-2（`.gsd/`）项目反迁回 GSD v1（`.planning/`）格式。
- **`/gsd-ingest-docs [path] [--mode new|merge] [--manifest <file>] [--resolve auto|interactive]`** — 从现有 ADR / PRD / SPEC / docs 自举或合并 `.planning/` 配置。

### 规划与执行

- **`/gsd-ultraplan-phase [phase]`** — [BETA] 把 plan phase 卸载到 Claude Code 的 ultraplan 云；在浏览器 review 后导回。
- **`/gsd-plan-review-convergence <phase> [--codex] [--gemini] [--claude] [--opencode] [--ollama] [--lm-studio] [--llama-cpp] [--all] [--text] [--ws <name>] [--max-cycles N]`** — 跨 AI plan 收敛循环：按 review 反馈重 plan 直到无 HIGH 问题。支持云 reviewer（Codex/Gemini/Claude/OpenCode）和本地模型运行时（Ollama / LM Studio / llama.cpp）。
- **`/gsd-autonomous [--from N] [--to N] [--only N] [--interactive]`** — 自动跑剩余所有 phase：每个走 discuss → plan → execute。

### 质量、Review 与验证

- **`/gsd-code-review <phase> [--depth=quick|standard|deep] [--files file1,file2,...] [--fix [--all] [--auto]]`** — review phase 内改动文件的 bug / 安全 / 代码质量问题。
- **`/gsd-secure-phase [phase]`** — 事后验证已完成 phase 的威胁缓解。
- **`/gsd-validate-phase [phase]`** — 已完成 phase 的 Nyquist 验证缺口回顾审计/填补。
- **`/gsd-ui-review [phase]`** — 已实现前端代码的 6 维度回顾式视觉审计。
- **`/gsd-eval-review [phase]`** — 审计已执行 AI phase 的 eval 覆盖率，产 EVAL-REVIEW.md 修复方案。
- **`/gsd-audit-fix --source <audit-uat> [--severity medium|high|all] [--max N] [--dry-run]`** — 自治 audit→fix 流水线：找问题 / 分类 / 修 / 测 / 提交。
- **`/gsd-add-tests <phase> [additional instructions]`** — 按 UAT 准则和实现给已完成 phase 生成测试。

### 诊断与维护

- **`/gsd-health [--repair] [--context]`** — 诊断 planning 目录健康度，可选修复。
- **`/gsd-forensics [problem description]`** — 失败 GSD 流程的法医调查：诊断哪里出问题。
- **`/gsd-undo --last N | --phase NN | --plan NN-MM`** — 安全 git revert。按 phase manifest 带依赖检查回滚 phase 或 plan commit。
- **`/gsd-docs-update [--force] [--verify-only]`** — 生成或更新项目文档，与代码库对照验证。
- **`/gsd-extract-learnings <phase>`** — 从已完成 phase 产物提取决策 / 教训 / 模式 / 意外。

### 知识与上下文

- **`/gsd-graphify [build|query <term>|status|diff]`** — 在 `.planning/graphs/` 构建 / 查询 / 审项目知识图谱。
- **`/gsd-thread [list [--open|--resolved] | close <slug> | status <slug> | name | description]`** — 管理跨会话的持久化上下文 thread。
- **`/gsd-profile-user [--questionnaire] [--refresh]`** — 生成开发者行为档案 + Claude 可发现的产物。
- **`/gsd-stats`** — 显示项目统计：phase / plan / requirements / git metrics / timeline。

### Workflow 与编排

- **`/gsd-manager [--analyze-deps]`** — 在一个终端管理多个 phase 的交互式指挥中心。`--analyze-deps` 在并行执行前扫 ROADMAP 找依赖关系。
- **`/gsd-workspace [--new | --list | --remove] [name]`** — 管理 GSD workspace：创建 / 列出 / 删除隔离工作区。
- **`/gsd-workstreams`** — 管理并行 workstream：列出 / 创建 / 切换 / 状态 / 进度 / 完成 / 恢复。
- **`/gsd-review-backlog`** — 评审 backlog 项并提升到当前 milestone。
- **`/gsd-milestone-summary [version]`** — 从 milestone 产物生成完整项目摘要（团队 onboarding / 评审用）。

### 仓库集成

- **`/gsd-inbox [--issues] [--prs] [--label] [--close-incomplete] [--repo owner/repo]`** — 按项目模板和贡献指南分诊 / 评审未关 GitHub issue & PR。

### 命名空间路由（面向模型的元 skill）

这 6 个 skill 主要让模型在 60+ skill 上做两段式分层路由。你也能直接调来交互式浏览类别。

- **`/gsd-context`** — 代码库情报路由（map / graphify / docs / learnings）。
- **`/gsd-ideate`** — 探索 / 捕获路由（explore / sketch / spike / spec / capture）。
- **`/gsd-manage`** — 配置和 workspace 路由（workstreams / thread / update / ship / inbox）。
- **`/gsd-project`** — 项目生命周期路由（milestones / audits / summary）。
- **`/gsd-review`** — 质量门路由（code review / debug / audit / security / eval / ui）。
- **`/gsd-workflow`** — Phase 管线路由（discuss / plan / execute / verify / phase / progress）。

## 文件与结构

```
.planning/
├── PROJECT.md            # 项目愿景
├── ROADMAP.md            # 当前 phase 拆分
├── STATE.md              # 项目记忆与上下文
├── RETROSPECTIVE.md      # 活复盘（每 milestone 更新）
├── config.json           # workflow 模式与门
├── todos/                # 捕获的想法和任务
│   ├── pending/          # 待办 todo
│   └── done/             # 已完成 todo
├── spikes/               # spike 实验（/gsd-spike）
│   ├── MANIFEST.md       # spike 清单和判定
│   └── NNN-name/         # 单个 spike 目录
├── sketches/             # 设计 sketch（/gsd-sketch）
│   ├── MANIFEST.md       # sketch 清单和优胜
│   ├── themes/           # 共享 CSS 主题文件
│   └── NNN-name/         # 单个 sketch 目录（HTML + README）
├── debug/                # 活动 debug session
│   └── resolved/         # 归档已解决 issue
├── milestones/
│   ├── v1.0-ROADMAP.md       # 归档 roadmap 快照
│   ├── v1.0-REQUIREMENTS.md  # 归档需求
│   └── v1.0-phases/          # 归档 phase 目录（通过 /gsd-cleanup 或 --archive-phases）
│       ├── 01-foundation/
│       └── 02-core-features/
├── codebase/             # 代码库地图（存量项目）
│   ├── STACK.md          # 语言、框架、依赖
│   ├── ARCHITECTURE.md   # 模式、分层、数据流
│   ├── STRUCTURE.md      # 目录布局、关键文件
│   ├── CONVENTIONS.md    # 编码标准、命名
│   ├── TESTING.md        # 测试设置、模式
│   ├── INTEGRATIONS.md   # 外部服务、API
│   └── CONCERNS.md       # 技术债、已知问题
└── phases/
    ├── 01-foundation/
    │   ├── 01-01-PLAN.md
    │   └── 01-01-SUMMARY.md
    └── 02-core-features/
        ├── 02-01-PLAN.md
        └── 02-01-SUMMARY.md
```

## Workflow 模式

在 `/gsd-new-project` 时设置：

**Interactive 模式**

- 每个重要决策都确认
- 在检查点暂停等批准
- 全程更多指引

**YOLO 模式**

- 自动批准大多数决策
- 执行 plan 不确认
- 仅在关键检查点停下

随时改：编辑 `.planning/config.json`。

## 规划配置

在 `.planning/config.json` 配置规划产物如何管理：

**`planning.commit_docs`**（默认 `true`）
- `true`：规划产物提交到 git（标准 workflow）
- `false`：规划产物仅本地，不提交

`commit_docs: false` 时：
- 把 `.planning/` 加进 `.gitignore`
- 适合 OSS 贡献、客户项目、或想私有保存 planning
- 所有 planning 文件照常工作，只是不进 git

**`planning.search_gitignored`**（默认 `false`）
- `true`：给大范围 ripgrep 搜索加 `--no-ignore`
- 仅当 `.planning/` 被 gitignore 且你想项目级搜索包含它时需要

示例配置：
```json
{
  "planning": {
    "commit_docs": false,
    "search_gitignored": true
  }
}
```

## 常见 workflow

**开新项目：**

```
/gsd-new-project        # 统一流程：提问 → 调研 → 需求 → roadmap
/clear
/gsd-plan-phase 1       # 为第一个 phase 创建 plan
/clear
/gsd-execute-phase 1    # 执行 phase 所有 plan
```

**休息后恢复工作：**

```
/gsd-progress  # 看上次到哪、继续
```

**milestone 中途加紧急工作：**

```
/gsd-phase --insert 5 "严重安全修复"
/gsd-plan-phase 5.1
/gsd-execute-phase 5.1
```

**完成 milestone：**

```
/gsd-complete-milestone 1.0.0
/clear
/gsd-new-milestone  # 开下个 milestone（提问 → 调研 → 需求 → roadmap）
```

**工作中捕获想法：**

```
/gsd-capture                                  # 从对话上下文捕获
/gsd-capture 修 modal z-index                # 带显式描述捕获
/gsd-capture --note 重构 auth 系统           # 快速无摩擦笔记
/gsd-capture --seed "实时通知"               # 带触发的前瞻性想法
/gsd-capture --list                          # review 并开干 todo
/gsd-capture --list api                      # 按 area 过滤
```

**调试 issue：**

```
/gsd-debug "表单悄悄提交失败"          # 开 debug session
# ... 调查中，上下文填满 ...
/clear
/gsd-debug                              # 从上次的位置恢复
```

## 求助

- 读 `.planning/PROJECT.md` 看项目愿景
- 读 `.planning/STATE.md` 看当前上下文
- 看 `.planning/ROADMAP.md` 看 phase 状态
- 跑 `/gsd-progress` 看做到哪
</reference>
