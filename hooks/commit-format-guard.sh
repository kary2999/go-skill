#!/bin/bash
# commit-format-guard · PreToolUse hook
# 拦截 git commit，校验 Conventional Commits 格式
# 格式：<type>(<scope>): <subject>
# type 枚举：feat|fix|docs|style|refactor|perf|test|chore|ci|revert|build
#
# Claude Code settings.json 配置：
# "hooks": {
#   "PreToolUse": [{"matcher": "Bash", "hooks": [{"type": "command", "command": "bash ~/.claude/hooks/commit-format-guard.sh"}]}]
# }

set -uo pipefail

INPUT=$(cat)
TOOL=$(echo "$INPUT" | python3 -c "import json,sys; d=json.load(sys.stdin); print(d.get('tool_name',''))" 2>/dev/null)

if [[ "$TOOL" != "Bash" ]]; then
  exit 0
fi

CMD=$(echo "$INPUT" | python3 -c "import json,sys; d=json.load(sys.stdin); print(d.get('tool_input',{}).get('command',''))" 2>/dev/null)

# 只检查 git commit 命令
if ! echo "$CMD" | grep -qE 'git\s+commit'; then
  exit 0
fi

# 提取 -m 后的 message
MSG=$(echo "$CMD" | grep -oP '(?<=-m\s["\x27])[^"\x27]+' | head -1 || true)
if [[ -z "$MSG" ]]; then
  # 尝试提取 heredoc 中的第一行
  MSG=$(echo "$CMD" | grep -v '^EOF' | grep -v 'cat <<' | sed -n '2p' | xargs || true)
fi

if [[ -z "$MSG" ]]; then
  exit 0
fi

TYPES="feat|fix|docs|style|refactor|perf|test|chore|ci|revert|build"
PATTERN="^($TYPES)(\([a-z0-9/_-]+\))?: .{1,72}"

if echo "$MSG" | grep -qE "$PATTERN"; then
  exit 0
fi

echo "⛔ [commit-format-guard] commit message 不符合 Conventional Commits 规范："
echo "  实际：$MSG"
echo ""
echo "  格式：<type>(<scope>): <subject>"
echo "  type：feat | fix | docs | style | refactor | perf | test | chore | ci | revert | build"
echo "  示例：feat(order): 新增批量取消接口"
echo "        fix(payment): 修复金额精度丢失问题"
exit 1
