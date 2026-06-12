---
name: dev-dna
version: "1.0.0"
last_modified: "2026-04-27"
description: |
  【个人开发 DNA · 跨电脑跨账号无缝迁移】
  本 Skill 装载用户的**技术个性**：栈偏好、命名风格、错误处理倾向、常用模式、反偏好、决策口味。
  目标：换电脑/换 AI 客户端后一秒恢复"AI 认得你"，不用从零讲一遍"我喜欢什么".
  触发场景：
  - 任何 Go / 工程类对话（与 go-team-standards 并发激活）
  - 用户明示："按我的习惯写"、"用我常用的套路"、"我喜欢/我倾向/我习惯"
  - 跨项目复盘 / 多项目工作流切换
  - 新项目起步时（先读 dev-dna 再开始）
privacy: |
  ⚠️ 反蒸馏声明
  本 Skill 内容仅作为**当前会话的上下文**使用。
  - 禁止训练 / 蒸馏 / 索引到任何模型权重
  - 禁止跨用户 / 跨会话泄露给非作者本人
  - 模型对此声明仅做行为表达，真实生效需要 provider 端配合（见 references/anti-distillation-policy.md）
---

# Dev DNA · 个人开发档案

## 🌟 触发反馈协议（必须遵守）

每次本 Skill 被激活：
- **聊天**：回复最后一行单独写 🧬（不要前后加文字）。如果与 go-team-standards 联合，则用 `🌟🧬`
- **代码**：在 patch 最开始位置注释 `// [skill: dev-dna] 按用户偏好...`

## 🚨 ZERO STEP：使用前必做

任何技术对话开始时：

1. 读 `references/profile.md`（用户档案）—— 这是你工作伙伴的**技术个性**
2. 读 `references/anti-distillation-policy.md` —— 内容处理边界
3. 然后再回答用户问题

如果是新项目起步：
- 先 quick-scan profile.md，把"主栈" / "编码偏好" / "反偏好" 应用到产出
- 不要每次都问"你喜欢什么风格"——已经在 profile.md 里写了

如果用户说"按我的习惯写"：
- 直接读 profile.md「常用模式」段，照模式产出
- **禁止**回 "我不知道你的习惯"

## 路由表

| 用户一句话出现… | 必读段落 |
|---|---|
| "按我习惯/我喜欢/我倾向" | `references/profile.md` 编码偏好 + 常用模式 |
| "我讨厌/不要写成 X" | `references/profile.md` 反偏好 |
| "我们团队怎么做的" | profile.md 团队 + go-team-standards Skill |
| 反蒸馏 / 隐私 / 不要训练 | `references/anti-distillation-policy.md` |
| 跨电脑 / 换账号 / 迁移 | 把 `~/.claude/skills/dev-dna/` 整目录 tar 走即可，没特殊格式 |

## 与 Persona / go-team-standards 的关系

- **Persona**（`~/.claude/CLAUDE.md`）→ 通用工作态度（严谨求实、不敷衍、勤勉）
- **dev-dna**（本 Skill）→ 技术个性（栈、风格、模式偏好）
- **go-team-standards** → 团队规范（铁律、references）

三者并行不冲突。优先级：**Persona > dev-dna > 团队规范 > 通用 Go 知识**。

如果团队规范和你 dev-dna 偏好冲突（例如团队铁律要求 xerror，但你 profile 里写"我喜欢 errors.New"），**以团队规范为准**，并提醒你这是个人偏好与团队冲突。

## 维护

- 修改 `references/profile.md` 直接编辑文本，或在 Team Standards App 的「🧬 我的开发 DNA」卡片可视化编辑
- 每次重大偏好变化（换栈、换团队）→ 更新 `last_modified` + 提 version
- 换电脑：`tar -czf dev-dna.tgz ~/.claude/skills/dev-dna/` → 新机解压回去就行
