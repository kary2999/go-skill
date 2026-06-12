#!/usr/bin/env bash
# float-amount-guard · PostToolUse
# 铁律：金额/价格/费用字段禁止用 float，必须用 decimal
# 来源：database.md §字段语义 / go-style.md §10
#
# 检测范围：.go 文件中 float32/float64 变量名包含 amount/price/fee/balance/total/cost

set -euo pipefail

TOOL_NAME="${TOOL_NAME:-}"
TOOL_INPUT="${TOOL_INPUT:-}"

case "$TOOL_NAME" in
  write_file|edit_file|str_replace_editor|create_file) ;;
  *) exit 0 ;;
esac

FILE_PATH=$(echo "$TOOL_INPUT" | python3 -c "
import sys, json
try:
    d = json.load(sys.stdin)
    print(d.get('path') or d.get('file_path') or d.get('new_path') or '')
except:
    print('')
" 2>/dev/null || echo "")

[[ -z "$FILE_PATH" ]] && exit 0

case "$FILE_PATH" in
  *.go) ;;
  *) exit 0 ;;
esac

case "$FILE_PATH" in
  *_test.go|*/vendor/*) exit 0 ;;
esac

[[ ! -f "$FILE_PATH" ]] && exit 0

# 检测 float32/float64 后接含金额语义的标识符
HITS=$(grep -nEi '(float32|float64)[[:space:]]+[a-zA-Z_]*(amount|price|fee|balance|total|cost|value)[a-zA-Z_]*' "$FILE_PATH" \
  | grep -v '^\s*//' | head -5 || true)

# 也检测结构体字段：Amount/Price float64
HITS2=$(grep -nEi '[[:space:]]+(Amount|Price|Fee|Balance|Total|Cost|OrderAmount|FeeAmount|TradeAmount|WithdrawAmount)[[:space:]]+(float32|float64)' "$FILE_PATH" \
  | grep -v '^\s*//' | head -5 || true)

ALL_HITS="${HITS}${HITS2}"

if [[ -n "$ALL_HITS" ]]; then
  echo "🚫 [float-amount-guard] 铁律违反：金额/价格字段禁止使用 float32/float64"
  echo ""
  echo "文件：$FILE_PATH"
  echo "违规行："
  echo "$ALL_HITS" | while IFS= read -r line; do echo "  $line"; done
  echo ""
  echo "📖 依据：database.md — 金额必须用 DECIMAL(18,8)；Go 代码用 github.com/shopspring/decimal"
  echo "✅ 修复：import \"github.com/shopspring/decimal\"，将 float64 改为 decimal.Decimal"
  exit 1
fi

exit 0
