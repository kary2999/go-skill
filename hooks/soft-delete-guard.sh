#!/usr/bin/env bash
# soft-delete-guard · PostToolUse
# 铁律：软删除必须用 deleted_at TIMESTAMPTZ，禁止 is_deleted BOOLEAN
# 来源：database.md §字段语义后缀规范

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
  *.sql|*.go|*.proto) ;;
  *) exit 0 ;;
esac

case "$FILE_PATH" in
  */vendor/*) exit 0 ;;
esac

[[ ! -f "$FILE_PATH" ]] && exit 0

# 检测 is_deleted（字段定义或结构体中）
HITS=$(grep -niE '\bis_deleted\b' "$FILE_PATH" \
  | grep -v '^\s*//' | head -5 || true)

if [[ -n "$HITS" ]]; then
  echo "🚫 [soft-delete-guard] 铁律违反：禁止使用 is_deleted，软删除统一用 deleted_at"
  echo ""
  echo "文件：$FILE_PATH"
  echo "违规行："
  echo "$HITS" | while IFS= read -r line; do echo "  $line"; done
  echo ""
  echo "📖 依据：database.md — 软删除字段一律 deleted_at TIMESTAMPTZ(6)，禁止 is_deleted BOOLEAN"
  echo "✅ 修复："
  echo "   SQL:  deleted_at TIMESTAMPTZ(6) DEFAULT NULL"
  echo "   Go:   DeletedAt *time.Time \`json:\"deleted_at\"\`"
  echo "   查询: WHERE deleted_at IS NULL"
  exit 1
fi

exit 0
