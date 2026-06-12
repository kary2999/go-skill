# Team Standards — Go 微服务团队编码规范

单一源（`standards/*.md`）→ 双平台分发（Claude Code Skill + Cursor `.mdc`）+ 轻量 demo 模板 + 术语表 + 图形化 Installer。

## 快速开始

```bash
# 双击或终端运行
./team-standards-installer
```
浏览器会自动弹出 → 选择"全局"或"项目" → 点"一键安装"。

无 GUI 环境？命令行 fallback：
```bash
./install.sh   # 零参数，默认全局覆盖安装 Claude + Cursor
```

## 安装位置

| 场景 | Claude Skill | Cursor Rules |
|---|---|---|
| 全局 | `~/.claude/skills/go-team-standards/` | `~/.cursor/rules/` |
| 项目 | `<项目>/.claude/skills/go-team-standards/` | `<项目>/.cursor/rules/` + `.cursorignore`（首次） |

## 目录结构

```
team-standards/
├── VERSION                     # 版本号
├── CHANGELOG.md                # 变更日志（App 的 History Tab 展示）
├── standards/                  # ← SSOT，规范原文
│   ├── go-style.md
│   ├── naming-logging.md
│   ├── error-codes.md
│   ├── api-design.md
│   ├── database.md
│   ├── testing.md
│   ├── commit.md
│   ├── ci-pipeline.md
│   ├── cursor-usage.md
│   ├── glossary.md             # ← 新增：团队 + Go + 云原生 术语表
│   └── naming-visual.png
├── demos/                      # ← 新增：10 个轻量代码模板（Cursor/Claude 模糊触发）
│   ├── kratos-service-min.go
│   ├── wire-providerset.go
│   ├── kafka-producer.go
│   ├── kafka-consumer.go
│   ├── pg-migration.sql
│   ├── pg-gorm-repo.go
│   ├── redis-idempotency.go
│   ├── errno-xerror.go
│   ├── slog-trace.go
│   └── table-driven-test.go
├── claude/go-team-standards/   # Claude Skill（含 references 软链）
├── cursor/
│   ├── rules/                  # 12 份 .mdc（含 common-lib 偏好 + demo 触发）
│   └── .cursorignore.template
├── assets/.golangci.yml
│
├── main.go / catalog.go / install.go    # Installer App 源码
├── web/                                  # App UI（内嵌）
├── go.mod
├── team-standards-installer              # ← 编译好的二进制（8.6MB）
└── install.sh                            # 命令行 fallback
```

## App 功能（4 个 Tab）

1. **Skills** — 11 个规范模块，展示触发条件 / 作用范围
2. **Demos** — 10 份最佳实践模板，点击左侧即时预览 + 一键复制
3. **Glossary** — 术语表，支持搜索（面向 PHP/Swoole 转 Go 的同学）
4. **History** — 版本变更日志

顶部是安装面板：Radio 选 global/project + 目标（Claude/Cursor/两个都装）+ 一键按钮。

## Demo 模糊触发

在 Claude Code / Cursor 里用**自然语言**描述需求即可，例如：
- "帮我写个 kafka 生产者" → 参考 `demos/kafka-producer.go`
- "给个 redis 幂等示例" → `demos/redis-idempotency.go`
- "pg 建表模板" → `demos/pg-migration.sql`
- "新建一个 kratos 服务" → `demos/kratos-service-min.go`

关键词不必精确匹配。详细触发映射见 `cursor/rules/11-demos.mdc` 和 SKILL.md 的 Demo 路由表。

## 技术栈前提

所有规范和 demo 假设团队栈 = **Kratos v2** + **Wire DI** + **Buf proto** + **mask-go-common-lib**。  
**禁止绕过 common-lib 直连底层库**（sarama / go-redis / otel 裸客户端等）—— 详见 SKILL.md 的替换表。

## 升级规范

1. 改 `standards/*.md`
2. Bump `VERSION` + 写 `CHANGELOG.md`
3. 如果影响要点：同步修改 `cursor/rules/*.mdc` 和 `demos/`
4. 重新编译：`go build -o team-standards-installer .`
5. 分发新二进制，用户双击重装（覆盖式，幂等）

## 依赖

- 开发：Go 1.22+（编译 installer）
- 运行：**无依赖**，single-binary 跨平台（macOS/Linux/Windows）
- Cursor / Claude Code 未安装也能装，目录不存在时自动创建（不需额外授权）
