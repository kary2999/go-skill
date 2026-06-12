// gsd-build/get-shit-done 框架的中文翻译覆盖层（v1.7.41）
//
// 不改磁盘上的 SKILL.md（AI 读原版英文）；仅在 App 的「规范模块」UI 上展示中文。
// 用户点开 modal 看完整 SKILL.md 时还是英文（保持上游一致，避免翻译版漂移）。
//
// 66 个翻译来自 v1.7.40 smoke test 实测扫到的 ID 清单（GSD 框架 v1.39 版本）。
// 上游加新 skill 时，没翻译的会 fallback 到 frontmatter 英文。

package main

type GSDTranslation struct {
	Title string // 中文 title，覆盖默认的 "💪 Plan Phase" 格式
	Desc  string // 中文描述，覆盖 frontmatter description
}

// 66 个 skill 的中文翻译
//
// Title 前缀 💪 由 scanGSDFrameworkSkills 加，这里只写中文名。
var gsdTranslations = map[string]GSDTranslation{
	"gsd-add-tests": {
		Title: "补测试",
		Desc:  "按 UAT 准则给已完成 phase 生成测试",
	},
	"gsd-ai-integration-phase": {
		Title: "AI 集成 Phase",
		Desc:  "为含 AI 系统的 phase 生成 AI-SPEC.md 设计契约",
	},
	"gsd-audit-fix": {
		Title: "审计→修复 流水线",
		Desc:  "自治流程：找问题 → 分类 → 修 → 测 → 提交",
	},
	"gsd-audit-milestone": {
		Title: "审计 Milestone",
		Desc:  "归档前对照原始意图审计 milestone 完成度",
	},
	"gsd-audit-uat": {
		Title: "审计未完成 UAT",
		Desc:  "跨 phase 审计所有未完成 UAT 与验证项",
	},
	"gsd-autonomous": {
		Title: "自治跑所有 Phase",
		Desc:  "自动跑剩余所有 phase：每个走 discuss → plan → execute",
	},
	"gsd-capture": {
		Title: "捕获想法/任务",
		Desc:  "把 ideas / tasks / notes / seeds 落到对应目录",
	},
	"gsd-cleanup": {
		Title: "清理已完成 Phase 目录",
		Desc:  "归档已完成 milestone 的累积 phase 目录",
	},
	"gsd-code-review": {
		Title: "代码评审",
		Desc:  "评审 phase 内改动文件的 bug / 安全 / 代码质量问题",
	},
	"gsd-complete-milestone": {
		Title: "完成 Milestone",
		Desc:  "归档已完成 milestone，准备下一版本",
	},
	"gsd-config": {
		Title: "配置 GSD",
		Desc:  "workflow 开关 / 高级旋钮 / 集成 / 模型 profile",
	},
	"gsd-debug": {
		Title: "系统化调试",
		Desc:  "跨上下文重置仍保留状态的系统化调试",
	},
	"gsd-discuss-phase": {
		Title: "讨论 Phase（澄清需求）",
		Desc:  "通过自适应提问收集 phase 上下文，进 plan 前必走",
	},
	"gsd-docs-update": {
		Title: "更新文档",
		Desc:  "生成/更新项目文档，与代码库实情对照验证",
	},
	"gsd-eval-review": {
		Title: "Eval 审计",
		Desc:  "审计已执行 AI phase 的 eval 覆盖率，产出 EVAL-REVIEW.md 修复方案",
	},
	"gsd-execute-phase": {
		Title: "执行 Phase",
		Desc:  "按波次并行执行 phase 所有 plan",
	},
	"gsd-explore": {
		Title: "苏格拉底式 Ideation",
		Desc:  "先想清楚再做 plan：idea 分流 / 路由",
	},
	"gsd-extract-learnings": {
		Title: "提取经验",
		Desc:  "从已完成 phase 产物提取决策 / 教训 / 模式 / 意外",
	},
	"gsd-fast": {
		Title: "极速执行（无 overhead）",
		Desc:  "直接执行琐碎任务：无 subagent / 无 planning 开销",
	},
	"gsd-forensics": {
		Title: "失败流程取证",
		Desc:  "失败 GSD 流程的法医调查：诊断哪里出了问题",
	},
	"gsd-graphify": {
		Title: "项目知识图谱",
		Desc:  "在 .planning/graphs/ 建/查/审项目知识图谱",
	},
	"gsd-health": {
		Title: "健康检查",
		Desc:  "诊断 .planning/ 目录健康度，可选自动修复",
	},
	"gsd-help": {
		Title: "命令帮助",
		Desc:  "展示可用 GSD 命令 + 使用指南",
	},
	"gsd-import": {
		Title: "导入外部 Plan",
		Desc:  "导入外部 plan，写入前对照项目决策检测冲突",
	},
	"gsd-inbox": {
		Title: "Inbox（issue/PR 分诊）",
		Desc:  "按项目模板和贡献指南分诊/评审未关 GitHub issue & PR",
	},
	"gsd-ingest-docs": {
		Title: "导入现有文档",
		Desc:  "从现有 ADR / PRD / SPEC / docs 自举或合并 .planning/ 配置",
	},
	"gsd-manager": {
		Title: "Phase 指挥中心",
		Desc:  "一个终端管理多个 phase 的交互式控制台",
	},
	"gsd-map-codebase": {
		Title: "代码库地图",
		Desc:  "并行 mapper agent 分析代码库，产出 .planning/codebase/ 文档",
	},
	"gsd-milestone-summary": {
		Title: "Milestone 摘要",
		Desc:  "从 milestone 产物生成完整项目摘要（团队 onboarding / 评审用）",
	},
	"gsd-mvp-phase": {
		Title: "MVP Phase 规划",
		Desc:  "按垂直 MVP 切片：user story → SPIDR 切分 → plan-phase",
	},
	"gsd-new-milestone": {
		Title: "新 Milestone",
		Desc:  "开启新 milestone：更新 PROJECT.md，导到 requirements",
	},
	"gsd-new-project": {
		Title: "新建项目",
		Desc:  "深度收集上下文初始化新项目 + 写 PROJECT.md",
	},
	"gsd-ns-context": {
		Title: "代码库情报命名空间",
		Desc:  "map / graphify / docs / learnings",
	},
	"gsd-ns-ideate": {
		Title: "Ideation 命名空间",
		Desc:  "explore / sketch / spike / spec / capture",
	},
	"gsd-ns-manage": {
		Title: "管理命名空间",
		Desc:  "config / workspace / workstreams / thread / update / ship / inbox",
	},
	"gsd-ns-project": {
		Title: "项目生命周期命名空间",
		Desc:  "milestone / audit / summary",
	},
	"gsd-ns-review": {
		Title: "质量门命名空间",
		Desc:  "code review / debug / audit / security / eval / ui",
	},
	"gsd-ns-workflow": {
		Title: "主 Workflow 命名空间",
		Desc:  "discuss / plan / execute / verify / phase / progress",
	},
	"gsd-pause-work": {
		Title: "暂停工作",
		Desc:  "phase 中途暂停时生成上下文交接",
	},
	"gsd-phase": {
		Title: "Phase CRUD",
		Desc:  "ROADMAP.md 里 phase 的增/插/删/改",
	},
	"gsd-plan-phase": {
		Title: "计划 Phase（产出 PLAN.md）",
		Desc:  "带验证循环的 phase 详细 plan",
	},
	"gsd-plan-review-convergence": {
		Title: "Plan 跨 AI 收敛",
		Desc:  "按 review 反馈重 plan，直到无 HIGH 级问题",
	},
	"gsd-pr-branch": {
		Title: "干净 PR 分支",
		Desc:  "过滤 .planning/ commit 创建干净分支，准备 code review",
	},
	"gsd-profile-user": {
		Title: "开发者 Profile",
		Desc:  "生成开发者行为档案 + Claude 可发现的产物",
	},
	"gsd-progress": {
		Title: "推进 Workflow",
		Desc:  "查进度 / 推进 / 自由意图分派",
	},
	"gsd-quick": {
		Title: "快速任务（带 GSD 保证）",
		Desc:  "atomic commit + 状态跟踪，但跳过可选 agent",
	},
	"gsd-resume-work": {
		Title: "恢复工作",
		Desc:  "从上一会话完整还原上下文",
	},
	"gsd-review": {
		Title: "跨 AI Plan 评审",
		Desc:  "请求外部 AI CLI 对 phase plan 做 peer review",
	},
	"gsd-review-backlog": {
		Title: "Backlog 评审",
		Desc:  "评审 backlog 项 + 提升到当前 milestone",
	},
	"gsd-secure-phase": {
		Title: "安全审计 Phase",
		Desc:  "事后验证已完成 phase 的威胁缓解",
	},
	"gsd-settings": {
		Title: "Workflow 设置",
		Desc:  "GSD workflow 开关 + 模型 profile",
	},
	"gsd-ship": {
		Title: "Ship（建 PR + 合并）",
		Desc:  "验证通过后创建 PR / 跑 review / 准备 merge",
	},
	"gsd-sketch": {
		Title: "UI 草图",
		Desc:  "用一次性 HTML mockup 草绘 UI / 设计想法",
	},
	"gsd-spec-phase": {
		Title: "Spec Phase（澄清 What）",
		Desc:  "带歧义评分澄清 phase 交付物，discuss-phase 前产出 SPEC.md",
	},
	"gsd-spike": {
		Title: "Spike 探索",
		Desc:  "通过体验式探索 spike idea",
	},
	"gsd-stats": {
		Title: "项目统计",
		Desc:  "phase / plan / requirements / git metrics / timeline",
	},
	"gsd-thread": {
		Title: "上下文 Thread",
		Desc:  "管理跨会话的持久化上下文 thread",
	},
	"gsd-ui-phase": {
		Title: "UI Phase（产出 UI-SPEC.md）",
		Desc:  "为前端 phase 生成 UI 设计契约",
	},
	"gsd-ui-review": {
		Title: "UI 视觉审计",
		Desc:  "已实现前端代码的 6 维度回顾式视觉审计",
	},
	"gsd-ultraplan-phase": {
		Title: "Ultraplan Phase [BETA]",
		Desc:  "把 plan phase 卸载到 Claude Code ultraplan 云",
	},
	"gsd-undo": {
		Title: "安全回滚",
		Desc:  "按 phase manifest 带依赖检查回滚 phase 或 plan commit",
	},
	"gsd-update": {
		Title: "升级 GSD",
		Desc:  "升级 GSD 到最新版 + 显示 changelog",
	},
	"gsd-validate-phase": {
		Title: "验证 Phase（Nyquist）",
		Desc:  "已完成 phase 的 Nyquist 验证缺口回顾审计/填补",
	},
	"gsd-verify-work": {
		Title: "对话式 UAT 验证",
		Desc:  "通过对话式 UAT 验证已建特性",
	},
	"gsd-workspace": {
		Title: "Workspace 管理",
		Desc:  "创建 / 列出 / 删除隔离工作区",
	},
	"gsd-workstreams": {
		Title: "并行 Workstream",
		Desc:  "管理并行 workstream：列出/创建/切换/状态/进度/完成/恢复",
	},
}
