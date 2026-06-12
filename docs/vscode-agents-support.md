# 技术方案：Team Standards 支持 VSCode 全家桶 AI 扩展
<!-- [skill: go-team-standards · 技术方案]  需要支持vscode 中的扩展 方案 -->

> 版本：V1.0 | 日期：2026-06-12 | 状态：第一层已落地（v1.8.14），第二/三层待排期

---

## 1. 背景与目标

团队规范目前通过 Team Standards App 安装到 Claude Code（`~/.claude/skills/`）、
Cursor（`~/.cursor/rules/`）、Codex（`~/.codex/skills/`）。但团队成员在 VSCode
里实际使用的 AI 扩展五花八门：GitHub Copilot、Cline（常配 DeepSeek）、Roo Code、
Windsurf 插件、Continue、Gemini Code Assist 等。

**目标**：无论成员用哪个 AI 扩展，团队规范都能被自动注入，且规范本体只维护一份。

**非目标**：不为每家扩展维护一份完整规范副本；不做规范内容的多格式翻译。

## 2. 现状分析

| 痛点 | 说明 |
|---|---|
| 格式碎片化 | 每家扩展认自己的 rules 文件（`.clinerules`、`.windsurfrules`、`copilot-instructions.md`…），没有统一 skill 格式 |
| 内容漂移 | 若每家复制一份规范，更新时必然漏改 |
| 安装繁琐 | 成员手动建 N 个文件不现实，也无法保证一致 |

**可利用的事实标准**：`AGENTS.md`（agents.md 开放规范）已被 Codex、Copilot
agent 模式、Cline、Roo、Cursor 新版、Gemini CLI 等原生支持，是唯一的公共交集。

## 3. 总体方案 —— 三层递进

```
┌─────────────────────────────────────────────────────┐
│ 第三层：Team Standards 自有 VSCode 扩展（远期）        │
│   状态栏 + 命令面板 + 规范漂移检测 + 一键修复            │
├─────────────────────────────────────────────────────┤
│ 第二层：MCP Server（中期）                            │
│   规范查询作为 MCP 工具暴露，支持 MCP 的扩展全部受益      │
├─────────────────────────────────────────────────────┤
│ 第一层：薄适配文件（✅ 已落地 v1.8.14）                 │
│   AGENTS.md 单一本体 + 7 家入口指针文件                │
└─────────────────────────────────────────────────────┘
```

### 第一层：薄适配文件（已上线）

`POST /api/universal/install` 在项目根生成：

- `AGENTS.md` —— 规范本体（事实标准，多数扩展原生读取）
- 7 个指针文件，内容统一为"请先读 AGENTS.md"：
  `.github/copilot-instructions.md` / `.clinerules/team-standards.md` /
  `.roo/rules/team-standards.md` / `.windsurfrules` /
  `.continue/rules/team-standards.md` / `GEMINI.md` / `CLAUDE.md`

原则：**已存在的文件一律跳过不覆盖**（成员可能有自己的定制）。

### 第二层：MCP Server

Copilot、Cline、Roo、Continue、Claude 系均已支持 MCP。把 Team Standards
做成本地 MCP server（Go 实现，stdio 传输），暴露工具：

| MCP 工具 | 功能 |
|---|---|
| `standards_search(query)` | 按关键字检索规范条目 |
| `standards_get(reference)` | 取某条规范完整原文 |
| `standards_check(diff)` | 对 diff 做铁律快速检查（复用现有 eval 用例逻辑） |

优势：规范变成 AI **主动可查询**的数据源，而非一次性注入的静态文本，
长规范不再挤占上下文窗口。

### 第三层：自有 VSCode 扩展（可选远期）

发布 `team-standards` VSCode 扩展：

- 打开工作区时检测适配文件是否存在 / 过期 → 状态栏提示，一键调第一层接口修复
- 命令面板：`Team Standards: 适配全部 Agent`、`查看规范`、`检查更新`
- 与 App 共用 Go 后端逻辑（扩展内调用打包的 CLI 或本地 HTTP）

## 4. 接口设计

第一层已实现（见 `universal_install.go`）。第二层新增 MCP server 入口：

```
team-standards mcp        # stdio 模式启动，供扩展的 mcp.json 引用
```

各扩展配置示例（成员 mcp.json）：

```json
{ "mcpServers": { "team-standards": { "command": "team-standards", "args": ["mcp"] } } }
```

## 5. 异常处理 / 兜底

- 适配文件被成员手动改过 → 永不覆盖，UI 列为「跳过」，由成员自行合并
- 扩展不支持 AGENTS.md 也不支持 MCP → 第一层指针文件仍可被其 rules 机制读到
- MCP server 崩溃 → 扩展侧降级为纯静态 rules 文件，功能不中断

## 6. 上线计划与回滚

| 阶段 | 内容 | 状态 |
|---|---|---|
| P0 | 第一层薄适配 + UI 卡片 + 单测 | ✅ v1.8.14 已发布 |
| P1 | MCP server（search/get 两个工具） | 待排期 |
| P2 | `standards_check` 工具 + VSCode 扩展 | 待排期 |

回滚：第一层生成的均为新增文件，删除即回滚，无副作用；MCP server 独立进程，
停用只需从 mcp.json 移除。

## 7. 风险与待定事项

- **风险**：各扩展 rules 文件路径可能随版本变化（如 Cline 曾从单文件改为目录）——
  适配表集中在 `universalAdapters` 一处，跟进成本低
- **待定**：MCP server 是否随 App 分发还是独立二进制；指针文件是否需要内嵌
  铁律摘要（当前为纯指针，依赖 agent 主动读 AGENTS.md 的服从性）

🌟
