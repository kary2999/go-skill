#!/usr/bin/env bash
# fmt-println-guard · PostToolUse
# 铁律：禁止在业务代码中使用 fmt.Println / fmt.Printf，必须用 slog
# 来源：go-style.md §7 日志规范
#
# 排除：main.go（启动阶段允许）、*_test.go、cmd/

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

# 排除允许的场景
case "$FILE_PATH" in
  *_test.go|*/vendor/*|*/cmd/main.go|*/cmd/*.go) exit 0 ;;
esac

[[ ! -f "$FILE_PATH" ]] && exit 0

HITS=$(grep -nE 'fmt\.(Println|Printf|Print|Fprintf|Fprintln)\(' "$FILE_PATH" \
  | grep -v '^\s*//' | head -5 || true)

if [[ -n "$HITS" ]]; then
  echo "🚫 [fmt-println-guard] 铁律违反：业务代码禁止使用 fmt.Println/Printf，必须用结构化日志"
  echo ""
  echo "文件：$FILE_PATH"
  echo "违规行："
  echo "$HITS" | while IFS= read -r line; do echo "  $line"; done
  echo ""
  echo "📖 依据：go-style.md §7 — 统一使用 slog，禁止 fmt.Println 或字符串拼接日志"
  echo "✅ 修复示例："
  echo "   slog.InfoContext(ctx, \"描述\", slog.String(\"key\", val))"
  echo "   slog.ErrorContext(ctx, \"操作失败\", slog.String(\"err\", err.Error()))"
  exit 1
fi

exit 0
