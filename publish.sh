#!/usr/bin/env bash
# publish.sh —— 一键发版 pipeline
#
# 流程：
#   1. 校验 working tree 干净、CHANGELOG 有对应节、tag 不重复
#   2. git tag + push → origin (GitLab) + github (GitHub)
#   3. gh release create → GitHub Release（触发 Pages 自动部署）
#
# 用法：
#   1) 改代码 → 改 VERSION → 写 CHANGELOG 对应节 → git commit
#   2) bash publish.sh
#
# 前置：
#   - gh CLI 已登录 (gh auth login)

#   - remote origin = git@github.com:kary2999/go-skill.git

set -euo pipefail
cd "$(dirname "${BASH_SOURCE[0]}")"

VERSION=$(cat VERSION | tr -d '[:space:]')
TAG="v${VERSION}"
GITHUB_REPO="kary2999/go-skill"

echo "→ 准备发布 ${TAG}"
echo ""

# ---------- 1. working tree 必须干净 ----------
if [ -n "$(git status --porcelain)" ]; then
  echo "✗ 工作区有未提交改动，先 commit："
  git status --short
  exit 1
fi
echo "  ✓ 工作区干净"

# ---------- 2. CHANGELOG 必须有这一节 ----------
if ! grep -q "^## \[${VERSION}\]" CHANGELOG.md; then
  echo "✗ CHANGELOG.md 没有 [${VERSION}] 节"
  echo "  补一段 '## [${VERSION}] — $(date +%Y-%m-%d)' 在文件顶部"
  exit 1
fi
echo "  ✓ CHANGELOG 有 [${VERSION}] 节"

# ---------- 3. tag 不能重复 ----------
if git rev-parse "${TAG}" >/dev/null 2>&1; then
  echo "✗ 本地已存在 tag ${TAG}，先删: git tag -d ${TAG}"
  exit 1
fi
if git ls-remote --tags github "refs/tags/${TAG}" 2>/dev/null | grep -q "${TAG}"; then
  echo "✗ GitHub 远端已存在 tag ${TAG}"
  echo "  删远端: git push github :refs/tags/${TAG}"
  exit 1
fi
echo "  ✓ tag ${TAG} 未占用"

# ---------- 4. gh 必须登录 ----------
if ! gh auth status &>/dev/null; then
  echo "✗ gh CLI 未登录，跑: gh auth login"
  exit 1
fi
echo "  ✓ gh 已登录"

echo ""
echo "============================================================"
echo "  发布 ${TAG} —— commit=$(git rev-parse --short HEAD)"
echo "============================================================"
echo ""

# ---------- 5. git tag ----------
echo "==> 1/3 打 tag"
git tag -a "${TAG}" -m "Release ${TAG}"
echo "  ✓ tag ${TAG} 已打"
echo ""

# ---------- 6. push origin (GitLab) + github ----------
echo "==> 2/3 推送 tag"
git push origin "${TAG}" 2>&1 && echo "  ✓ ${TAG} → GitLab" || echo "  ⚠ GitLab push 失败（忽略）"
git push github main "${TAG}" 2>&1 && echo "  ✓ main + ${TAG} → GitHub"
echo ""

# ---------- 7. 提取 CHANGELOG 本版本内容 ----------
NOTES=$(awk "/^## \[${VERSION}\]/{flag=1; next} /^## \[/{flag=0} flag" CHANGELOG.md | sed '/^$/d')

# ---------- 8. GitHub Release ----------
echo "==> 3/3 创建 GitHub Release"
gh release create "${TAG}" \
  --repo "${GITHUB_REPO}" \
  --title "Team Standards ${TAG}" \
  --notes "${NOTES}" \
  2>&1

echo ""
echo "============================================================"
echo "✅ 发布完成 ${TAG}"
echo "   Release: https://github.com/${GITHUB_REPO}/releases/tag/${TAG}"
echo "   Pages:   https://kary2999.github.io/go-skill/"
echo "============================================================"
echo ""
echo "GitHub Actions 将自动部署最新版到 Pages（约 1 分钟）。"
