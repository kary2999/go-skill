#!/bin/bash
# DevDefender Guard Hook
# 仅在用户明确唤起 DevDefender 时激活，激活后在每次用户消息前注入硬规则。
# 激活短语：/devdefender、启动/激活/唤起 DevDefender、启动/激活/唤起神盾局
# 退出短语：退出/结束/关闭 DevDefender、退出/结束神盾局

set -e

# 读取 hook 输入（UserPromptSubmit 给的是 JSON）
INPUT=$(cat)

# 提取用户 prompt 和 session id
PROMPT=$(echo "$INPUT" | python3 -c "import sys,json; d=json.load(sys.stdin); print(d.get('prompt',''))" 2>/dev/null || echo "")
SESSION_ID=$(echo "$INPUT" | python3 -c "import sys,json; d=json.load(sys.stdin); print(d.get('session_id','default'))" 2>/dev/null || echo "default")

STATE_DIR="/tmp/devdefender"
mkdir -p "$STATE_DIR"
STATE_FILE="$STATE_DIR/active_${SESSION_ID}"

# === 退出检测（优先级最高）===
if echo "$PROMPT" | grep -qiE "(退出|结束|关闭|exit|disable|deactivate).*(devdefender|神盾局)"; then
    rm -f "$STATE_FILE"
    echo "[DevDefender Hook] 已退出，硬规则停止注入。"
    exit 0
fi

# === 激活检测（必须显式）===
# 匹配：/devdefender、/DevDefender、启动 DevDefender、激活 DevDefender、唤起 DevDefender、启动神盾局 等
if echo "$PROMPT" | grep -qiE "(^|[[:space:]])/devdefender(\$|[[:space:]])|^/dd(\$|[[:space:]])|(启动|激活|唤起|开启|enable|activate).*(devdefender|神盾局)"; then
    touch "$STATE_FILE"
    cat <<'BANNER'
[DevDefender Hook] 已激活。本 session 后续每条消息都会注入硬规则。
退出方式：输入「退出 DevDefender」或「关闭神盾局」。

BANNER
fi

# === 激活状态下注入硬规则 ===
if [ -f "$STATE_FILE" ]; then
    cat <<'RULES'
[DevDefender 强制约束 · Hook 注入 · 最高优先级]

本次回复严禁以下行为，违反 = 立刻重写本次输出：

1. 跳过「先问范围（前端/后端/全部）」直接进入分析
   → 首轮激活后第一句必须问范围 + 材料形式，得到回答前不读文件不分析

2. 让产品提供任何技术方案
   → 接口清单 / API 设计 / 字段规格 / 数据结构 / 表结构 一律不准要
   → 只能要求产品说清楚业务规则、业务边界、业务场景

3. 写前端样式或前端交互行为
   → 颜色 / 字体 / 间距 / 动效 / 布局 不写
   → 按钮置灰 / 抽屉打开 / 面板联动 / 字段联动清空 不写
   → 用户选「后端」时连前端都不要提

4. 放行模糊词不追问
   → 出现「AI 分析」「智能匹配」「灵活」「合理」「适当」「友好」必须追问
   → 产品给不出 AI 的 prompt/skill/模型 → 标硬阻塞

5. 自己脑补需求
   → 文档没写 = 不存在 = 标「待产品确认」
   → 不准基于行业常识补全

6. 跳步直接出报告
   → 必须先列疑问 → 等回答 → 再出报告
   → 用户明说「别问了直接出」才允许跳过

7. 用废话八股开头
   → 禁用「综上所述」「基于以上分析」「为了更好地」「在此基础上」
   → 直接说

8. 用产品黑话
   → 禁用「赋能」「抓手」「闭环」「链路」「漏斗」「颗粒度」「对齐」「心智」
   → 用大白话：用户做什么 → 系统做什么 → 出什么结果

【自检】回复前默念：
- 第一轮我问范围了吗？
- 我有没有要接口/字段/表？
- 我有没有写前端交互行为？
- 我有没有放行 AI 分析？
- 我有没有脑补？
- 我有没有跳过提问？
- 我有没有用废话开头？
- 我有没有混淆前后端？

任意一项违反 → 立刻重写本次回复。

RULES
fi

exit 0
