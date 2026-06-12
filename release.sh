#!/usr/bin/env bash
# 一键发布：build .app + 打 DMG + 直接落到 ../version/ 归档目录
# 用法：bash release.sh
#
# v1.7.22 起加这个脚本，避免每次 build 出来的 DMG 堆在项目根目录用户找不到。
set -euo pipefail

cd "$(dirname "${BASH_SOURCE[0]}")"

VERSION=$(cat VERSION | tr -d '[:space:]')
STAMP=$(date +%Y%m%d-%H%M)
DMG_NAME="TeamStandards-v${VERSION}-${STAMP}.dmg"

ARCHIVE_DIR="$(cd .. && pwd)/version"
mkdir -p "$ARCHIVE_DIR"

DMG_PATH="$ARCHIVE_DIR/$DMG_NAME"

echo "==> 1/3 编译 .app"
rm -rf "Team Standards.app"
bash build-mac-app.sh > /tmp/build-app.log 2>&1 || {
  echo "✗ build-mac-app.sh 失败，详见 /tmp/build-app.log"
  tail -10 /tmp/build-app.log
  exit 1
}
echo "  ✓ Team Standards.app 完成"

echo "==> 2/3 制作 DMG"
STAGING=$(mktemp -d)
cp -R "Team Standards.app" "$STAGING/"
ln -s /Applications "$STAGING/Applications"
hdiutil create -volname "Team Standards ${VERSION}" \
  -srcfolder "$STAGING" \
  -ov -format UDZO "$DMG_PATH" > /tmp/hdiutil.log 2>&1 || {
  echo "✗ hdiutil 失败，详见 /tmp/hdiutil.log"
  tail -10 /tmp/hdiutil.log
  rm -rf "$STAGING"
  exit 1
}
rm -rf "$STAGING"
echo "  ✓ DMG 制作完成"

echo "==> 3/3 归档到 version/"
SIZE=$(ls -lh "$DMG_PATH" | awk '{print $5}')

echo ""
echo "✅ 发布完成"
echo "   $DMG_PATH ($SIZE)"
echo ""
echo "在 Finder 里打开归档目录："
echo "   open '$ARCHIVE_DIR'"
