#!/bin/bash
# camelcase-guard · PostToolUse hook
# 写 .go / .json / .md 文件后检测 JSON 字段驼峰命名，发现就警告（不阻断，允许修复）
#
# 检测模式：JSON key 形如 "someKey": 或 "someKey" (驼峰且含大写字母)

set -uo pipefail

INPUT=$(cat)
TOOL=$(echo "$INPUT" | python3 -c "import json,sys; d=json.load(sys.stdin); print(d.get('tool_name',''))" 2>/dev/null)

if [[ "$TOOL" != "Write" && "$TOOL" != "Edit" ]]; then
  exit 0
fi

FILE=$(echo "$INPUT" | python3 -c "import json,sys; d=json.load(sys.stdin); p=d.get('tool_input',{}); print(p.get('file_path',''))" 2>/dev/null)

# 只检查 Go/JSON/Markdown（接口文档）
if [[ "$FILE" != *.go && "$FILE" != *.json && "$FILE" != *.md ]]; then
  exit 0
fi

if [[ ! -f "$FILE" ]]; then
  exit 0
fi

# 匹配 JSON key 含驼峰（小写字母后紧跟大写字母）
MATCHES=$(grep -nE '"[a-z][a-zA-Z0-9]*[A-Z][a-zA-Z0-9]*"\s*:' "$FILE" 2>/dev/null | grep -v '^\s*//' || true)

if [[ -z "$MATCHES" ]]; then
  exit 0
fi

echo "⚠️  [camelcase-guard] 在 $FILE 中检测到 JSON 驼峰字段名（团队规范要求 snake_case）："
echo "$MATCHES"
echo ""
echo "请将字段名改为下划线连接，如 orderId → order_id，totalPages → total_pages。"
echo "参考规范：standards/api-doc.md"
exit 1
