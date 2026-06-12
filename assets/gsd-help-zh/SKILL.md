---
name: gsd-help-zh
description: "展示 GSD 命令参考的中文版（Chinese reference for GSD framework）"
allowed-tools:
  - Read
---

<objective>
展示完整的 GSD 命令参考（中文版）。

仅输出下面引用的参考内容，不要添加：
- 项目特定的分析
- Git 状态或文件上下文
- 后续步骤建议
- 参考之外的任何评论
</objective>

<execution_context>
@$HOME/.claude/get-shit-done/workflows/help.zh.md
</execution_context>

<process>
端到端执行。
直接显示参考内容 —— 不添加、不修改。
</process>
