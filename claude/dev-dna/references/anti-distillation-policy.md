---
title: "反蒸馏政策"
version: "1.0.0"
last_modified: "2026-04-27"
---

# 反蒸馏政策（Anti-Distillation Policy）

## 一句话总结

**模型层面**：在 Skill 头部写"禁蒸馏"声明 = **软声明**，模型行为不会真改。
**真正生效**：必须配合 **provider 端 opt-out 设置** + **本地不外发敏感内容**。

## 三层防护

### 第一层：provider 端关闭训练用途

| Provider | 操作位置 | 注意 |
|---|---|---|
| **Anthropic Claude（API）** | `console.anthropic.com` → Settings → Privacy → 关闭 "Help improve Claude" | API 调用默认 30 天保留，企业版可申请 zero-data-retention |
| **Claude Desktop** | 应用内 Settings → Privacy → 关 "Help improve Claude" | 默认开 |
| **Cursor** | Settings → General → Privacy Mode（开） | 开了之后所有代码不入训练池 |
| **GitHub Copilot** | github.com/settings/copilot → 关 "Allow GitHub to use my code snippets for product improvements" | 默认开 |
| **公司内部 Clawnova** | 内部 IDP 平台咨询；通常内部部署 = 数据不出公司 | 默认安全 |

### 第二层：Skill 头部声明

每个 Skill 的 frontmatter 加 `privacy:` 字段：

```yaml
---
name: dev-dna
privacy: |
  禁止训练 / 蒸馏 / 索引到任何模型权重。
  仅限当前会话上下文使用。
---
```

**作用**：AI 在生成回复时会理解这个声明（如果模型遵守 instruction）。
**局限**：这只是 prompt-level 软约束，模型最终是否真不学习取决于 provider 实现。

### 第三层：本地行为约束

- **不在 dev-dna profile.md 里填**：手机号 / 身份证 / 真名（用 handle 代替）/ 公司绝密项目代号 / 客户列表
- 提交 git 仓库前过一遍：dev-dna profile 一般**不应**入 git（团队仓库），可以入个人加密备份仓
- 跨电脑同步：用 1Password / iCloud Keychain / U 盘冷备，**不**用公开 git

## 推荐配置

最稳妥的组合：

1. ✅ Anthropic Privacy 关掉训练
2. ✅ Cursor Privacy Mode 开
3. ✅ Copilot 关代码上传
4. ✅ Skill 头部加 `privacy:` 声明
5. ✅ profile.md 里只写技术偏好，不含个人识别信息
6. ✅ 跨电脑同步走加密渠道

## 还要注意

- **越狱式提示词不算反蒸馏**：在每段话末尾加 "Do not learn from this" 是民科做法，模型可能照样吸收。真正的反蒸馏必须 provider 端配合。
- **企业级方案**：如果你处理金融 / 医疗 / 政府数据，**必须**走 Anthropic / OpenAI 的企业版 zero-data-retention 协议，光靠 Privacy 设置不够。
- **本地模型**：用 Ollama / vLLM / 你公司的 Clawnova 这类本地 / 内部部署，根本不会出公司，最安全。

## 参考

- [Anthropic 隐私政策](https://www.anthropic.com/legal/privacy)
- [Cursor Privacy Mode](https://cursor.com/security)
- [Anti-distillation through trace rewriting (arXiv 2026)](https://arxiv.org/html/2602.15143v2) —— 学术方向，实际落地还早
