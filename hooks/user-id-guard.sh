#!/usr/bin/env bash
# user-id-guard · PostToolUse
# 铁律：全平台禁止 user_id 字段，必须用 uid
# 来源：field-naming.md §1.2 多租户三件套
#
# 检测范围：.go / .sql / .proto 文件中出现 user_id（字段定义、结构体、JSON tag）
# 排除：注释行、测试文件（*_test.go）、vendor/

set -euo pipefail

TOOL_NAME="${TOOL_NAME:-}"
TOOL_INPUT="${TOOL_INPUT:-}"

# 只看写文件类工具
case "$TOOL_NAME" in
  write_file|edit_file|str_replace_editor|create_file) ;;
  *) exit 0 ;;
esac

# 提取被写入的文件路径
FILE_PATH=$(echo "$TOOL_INPUT" | python3 -c "
import sys, json
try:
    d = json.load(sys.stdin)
    print(d.get('path') or d.get('file_path') or d.get('new_path') or '')
except:
    print('')
" 2>/dev/null || echo "")

[[ -z "$FILE_PATH" ]] && exit 0

# 过滤：只检查 .go .sql .proto
case "$FILE_PATH" in
  *_test.go) exit 0 ;;
  *.go|*.sql|*.proto) ;;
  *) exit 0 ;;
esac

[[ ! -f "$FILE_PATH" ]] && exit 0

# 跳过 vendor / migrations（存量）
case "$FILE_PATH" in
  */vendor/*|*/migrations/*) exit 0 ;;
esac

# 检测：user_id（字段名 / json tag / 结构体字段），排除注释行
HITS=$(grep -n '\buser_id\b' "$FILE_PATH" \
  | grep -v '^\s*//' \
  | grep -v '^\s*#' \
  | grep -v '// TODO\|// FIXME\|// HACK' \
  | head -5 || true)

if [[ -n "$HITS" ]]; then
  echo "🚫 [user-id-guard] 铁律违反：禁止使用 user_id，全平台统一用 uid"
  echo ""
  echo "文件：$FILE_PATH"
  echo "违规行："
  echo "$HITS" | while IFS= read -r line; do echo "  $line"; done
  echo ""
  echo "📖 依据：field-naming.md §1.2 — 全平台禁止 user_id / userId / u_id，统一为 uid"
  echo "✅ 修复：将所有 user_id 替换为 uid（同表多角色用 operator_uid / from_uid / to_uid）"
  exit 1
fi

exit 0
