#!/usr/bin/env bash
# goroutine-naked-guard · PostToolUse
# 铁律：禁止裸 go func()，必须通过 errgroup 或封装函数管理生命周期
# 来源：go-style.md §5 并发编程

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
  *_test.go|*/vendor/*|*/cmd/main.go) exit 0 ;;
esac

[[ ! -f "$FILE_PATH" ]] && exit 0

# 检测裸 go func() 启动（不在 g.Go / errgroup 上下文内）
# 简单模式：行内直接 go func(
HITS=$(grep -nE '^\s*go\s+func\s*\(' "$FILE_PATH" \
  | grep -v '^\s*//' | head -5 || true)

if [[ -n "$HITS" ]]; then
  echo "🚫 [goroutine-naked-guard] 铁律违反：禁止裸启 goroutine，必须用 errgroup 管理生命周期"
  echo ""
  echo "文件：$FILE_PATH"
  echo "违规行："
  echo "$HITS" | while IFS= read -r line; do echo "  $line"; done
  echo ""
  echo "📖 依据：go-style.md §5 — 裸 goroutine 无退出机制，panic/泄漏无法追踪"
  echo "✅ 修复示例："
  echo "   g, ctx := errgroup.WithContext(ctx)"
  echo "   g.Go(func() error {"
  echo "       return doWork(ctx)"
  echo "   })"
  echo "   if err := g.Wait(); err != nil { ... }"
  exit 1
fi

exit 0
