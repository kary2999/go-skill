#!/usr/bin/env bash
# error-ignore-guard · PostToolUse
# 铁律：永远检查 error，禁止用 _ 忽略
# 来源：go-style.md §4.1 错误处理基本原则

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

# 检测 _ = someFunc() 或 _, err = / if _, err 无关的裸 _ 赋值给 error 返回
# 重点：xxx, _ := / xxx, _ = （第二个返回值常是 error）
HITS=$(grep -nE '^[^/]*,\s*_\s*[:=]=' "$FILE_PATH" \
  | grep -v '^\s*//' \
  | grep -vE ',\s*_\s*=\s*(os\.Setenv|os\.Unsetenv|os\.Remove|os\.MkdirAll|fmt\.Fprintln\(os\.Stderr)' \
  | head -5 || true)

if [[ -n "$HITS" ]]; then
  echo "🚫 [error-ignore-guard] 铁律违反：禁止用 _ 忽略 error 返回值"
  echo ""
  echo "文件：$FILE_PATH"
  echo "违规行："
  echo "$HITS" | while IFS= read -r line; do echo "  $line"; done
  echo ""
  echo "📖 依据：go-style.md §4.1 — 永远检查 error，禁止使用 _ 忽略"
  echo "✅ 修复：显式处理 error，或用 fmt.Errorf(\"操作失败: %w\", err) 向上透传"
  exit 1
fi

exit 0
