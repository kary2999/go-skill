#!/bin/bash
# go-panic-guard · PostToolUse hook
# 写 .go 文件后扫描裸 panic()，非 main.go / cmd/ 里的 panic 直接阻断并提示修成 xerror
#
# Claude Code settings.json 配置：
# "hooks": {
#   "PostToolUse": [{"matcher": "Write|Edit", "hooks": [{"type": "command", "command": "bash ~/.claude/hooks/go-panic-guard.sh"}]}]
# }

set -uo pipefail

INPUT=$(cat)
TOOL=$(echo "$INPUT" | python3 -c "import json,sys; d=json.load(sys.stdin); print(d.get('tool_name',''))" 2>/dev/null)

# 只处理 Write / Edit 工具
if [[ "$TOOL" != "Write" && "$TOOL" != "Edit" ]]; then
  exit 0
fi

FILE=$(echo "$INPUT" | python3 -c "import json,sys; d=json.load(sys.stdin); p=d.get('tool_input',{}); print(p.get('file_path',''))" 2>/dev/null)

# 只检查 .go 文件
if [[ "$FILE" != *.go ]]; then
  exit 0
fi

# main.go / cmd/ 目录允许 panic
if [[ "$FILE" == */main.go || "$FILE" == */cmd/* ]]; then
  exit 0
fi

# 检查文件是否存在（Write 已落盘）
if [[ ! -f "$FILE" ]]; then
  exit 0
fi

MATCHES=$(grep -n '\bpanic(' "$FILE" 2>/dev/null || true)
if [[ -z "$MATCHES" ]]; then
  exit 0
fi

echo "⛔ [go-panic-guard] 在 $FILE 中发现裸 panic："
echo "$MATCHES"
echo ""
echo "团队规范：业务代码禁止裸 panic，改用 xerror.New / xerror.Wrap 返回错误。"
echo "只有 main.go 和 cmd/ 下允许 panic。"
exit 1
