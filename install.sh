#!/usr/bin/env bash
# 命令行 fallback 安装器 —— 零参数，默认"全局安装 Claude+Cursor"，覆盖已有。
# 推荐使用 GUI：./team-standards-installer（双击也行）
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
VERSION="$(cat "$REPO_ROOT/VERSION")"

CLAUDE_DIR="$HOME/.claude/skills/go-team-standards"
CURSOR_DIR="$HOME/.cursor/rules"

echo "→ team-standards v$VERSION 全局安装（覆盖已有）"

# Claude Skill
rm -rf "$CLAUDE_DIR"
mkdir -p "$CLAUDE_DIR/references" "$CLAUDE_DIR/assets" "$CLAUDE_DIR/demos"
cp "$REPO_ROOT/claude/go-team-standards/SKILL.md" "$CLAUDE_DIR/"
cp -RL "$REPO_ROOT/claude/go-team-standards/references/." "$CLAUDE_DIR/references/"
cp "$REPO_ROOT/assets/.golangci.yml" "$CLAUDE_DIR/assets/"
cp -R "$REPO_ROOT/demos/." "$CLAUDE_DIR/demos/"
echo "$VERSION" > "$CLAUDE_DIR/.installed-version"
echo "  ✓ Claude → $CLAUDE_DIR"

# Cursor rules
mkdir -p "$CURSOR_DIR"
find "$CURSOR_DIR" -maxdepth 1 -name '[0-9][0-9]-*.mdc' -delete 2>/dev/null || true
cp "$REPO_ROOT"/cursor/rules/*.mdc "$CURSOR_DIR/"
echo "$VERSION" > "$CURSOR_DIR/.installed-version"
echo "  ✓ Cursor → $CURSOR_DIR"

echo "✓ 完成。推荐改用 GUI：./team-standards-installer"
