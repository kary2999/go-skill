---
description: 主动触发提测报告生成（OrangeCat skill 显式调用）
argument-hint: <可选：标签或备注，如 "v1.0.0">
---

# 提测：$ARGUMENTS

主动调用 **orangecat skill** 生成提测报告。比模糊触发"提测"更可靠。

## 步骤

1. **必读**：`~/.claude/skills/orangecat/SKILL.md`（含 G1-G7 门禁）
2. 跑 git 命令收集证据：
   ```bash
   git diff --name-only origin/master...HEAD
   git log origin/master..HEAD --format='%H%n%s%n%n%b%n---'
   git log -1 --format='%an <%ae>'
   ```
3. 走 STEP 0-D 业务范围深度挖掘
4. 过 G1-G7 门禁（接口下游 / SQL 完整 / 脚本影响 / 自测 / 真实性 / 提测人 / 业务范围）
5. 生成两个文件到当前项目根：
   - `提测报告_<branch>_<时间>_QA版.md`
   - `提测报告_<branch>_<时间>_开发版.md`

## 门禁失败处理

任一门禁失败 → 拒绝生成，列出缺什么 + 让用户补充：

```
❌ 报告未生成。缺少：
  - G1: §3 接口变更第 2 行下游影响写了"无"，需具体调用方名字
  - G3: git diff 含 scripts/migrate.go，但 §4 脚本影响段没展开

请补充以上信息后重新执行 /tixuebj。
```

## 输出

每份报告头加：
```
<!-- [skill: orangecat] 提测报告 - $ARGUMENTS -->
```

聊天末尾单独一行：🐱
