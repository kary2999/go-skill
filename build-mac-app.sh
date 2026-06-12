#!/usr/bin/env bash
# 打包成 macOS .app bundle —— Finder 里双击图标启动，无终端窗口
set -euo pipefail

cd "$(dirname "${BASH_SOURCE[0]}")"

APP_NAME="Team Standards"
BUNDLE="${APP_NAME}.app"
BIN_NAME="team-standards-installer"
VERSION="$(cat VERSION)"

echo "→ 编译 Universal Binary（arm64 + x86_64，最低 macOS 11 Big Sur）"
# 显式指定 MACOSX_DEPLOYMENT_TARGET=11.0
# 不指定的话，2026+ Xcode 会隐式要求 macOS 15+，导致老系统（比如 14.5 Sonoma）起不来
# 11.0 覆盖所有 Apple Silicon 机器（M1 Pro/Max/Ultra 以上全兼容）
export MACOSX_DEPLOYMENT_TARGET=11.0

# arm64 覆盖所有 M 系列（M1 / M2 / M3 / M4 / 未来的 Mx）
CGO_ENABLED=1 GOOS=darwin GOARCH=arm64 \
  CC="clang -arch arm64 -mmacosx-version-min=11.0" \
  CGO_CFLAGS="-arch arm64 -mmacosx-version-min=11.0" \
  CGO_LDFLAGS="-arch arm64 -mmacosx-version-min=11.0" \
  go build -o /tmp/tsi-arm64 .

# amd64 覆盖 Intel Mac（2020 之前的 MBP/MBA/iMac）
CGO_ENABLED=1 GOOS=darwin GOARCH=amd64 \
  CC="clang -arch x86_64 -mmacosx-version-min=11.0" \
  CGO_CFLAGS="-arch x86_64 -mmacosx-version-min=11.0" \
  CGO_LDFLAGS="-arch x86_64 -mmacosx-version-min=11.0" \
  go build -o /tmp/tsi-amd64 .

# lipo 合并成 fat binary
lipo -create /tmp/tsi-arm64 /tmp/tsi-amd64 -output "$BIN_NAME"
rm -f /tmp/tsi-arm64 /tmp/tsi-amd64
echo "  ✓ Universal Binary：$(lipo -archs "$BIN_NAME")"

echo "→ 构建 $BUNDLE"
rm -rf "$BUNDLE"
mkdir -p "$BUNDLE/Contents/MacOS" "$BUNDLE/Contents/Resources"

# 把二进制放入 bundle，Launcher 脚本负责隐藏终端
cp "$BIN_NAME" "$BUNDLE/Contents/MacOS/$BIN_NAME"

# WebView 是 GUI 进程，必须前台运行（否则窗口会在 parent 退出时关闭）
# 把原生二进制直接作为 CFBundleExecutable 即可

cat > "$BUNDLE/Contents/Info.plist" <<EOF
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN"
 "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
  <key>CFBundleName</key>               <string>${APP_NAME}</string>
  <key>CFBundleDisplayName</key>        <string>${APP_NAME}</string>
  <key>CFBundleIdentifier</key>         <string>com.team.standards.installer</string>
  <key>CFBundleVersion</key>            <string>${VERSION}</string>
  <key>CFBundleShortVersionString</key> <string>${VERSION}</string>
  <key>CFBundleExecutable</key>         <string>${BIN_NAME}</string>
  <key>CFBundlePackageType</key>        <string>APPL</string>
  <key>LSMinimumSystemVersion</key>     <string>11.0</string>
  <key>NSHighResolutionCapable</key>    <true/>
  <key>LSUIElement</key>                <false/>
</dict>
</plist>
EOF

# 生成 App 图标：🦕 恐龙 logo，青蓝渐变背景
ICON_SRC="/tmp/ts-icon.png"
python3 - <<'PY' > /dev/null 2>&1 || true
try:
    from PIL import Image, ImageDraw, ImageFont
    size = 1024
    # 完全透明背景
    img = Image.new('RGBA', (size, size), (0, 0, 0, 0))
    d = ImageDraw.Draw(img)
    # 只画 🦕 emoji，不加任何背景
    try:
        # Apple Color Emoji 字体只支持固定位图尺寸（160 是最大的），画完后放大
        src = Image.new('RGBA', (256, 256), (0, 0, 0, 0))
        sd = ImageDraw.Draw(src)
        font = ImageFont.truetype("/System/Library/Fonts/Apple Color Emoji.ttc", 160)
        sd.text((128, 128), "🦕", font=font, anchor="mm", embedded_color=True)
        # 放大到 1024 保持清晰（emoji 位图自身已有足够分辨率）
        src = src.resize((size, size), Image.LANCZOS)
        img = Image.alpha_composite(img, src)
    except Exception:
        # 回退：画个 "TS" 文字（仍透明底）
        try:
            font = ImageFont.truetype("/System/Library/Fonts/Helvetica.ttc", 560)
        except Exception:
            font = ImageFont.load_default()
        d.text((size // 2, size // 2), "TS", fill=(45, 212, 191, 255), anchor="mm", font=font)
    img.save("/tmp/ts-icon.png")
except Exception as e:
    import sys
    sys.stderr.write(f"icon gen failed: {e}\n")
PY

if [[ -f "$ICON_SRC" ]]; then
  ICONSET="/tmp/TeamStandards.iconset"
  rm -rf "$ICONSET"
  mkdir -p "$ICONSET"
  for SIZE in 16 32 64 128 256 512; do
    sips -z $SIZE $SIZE "$ICON_SRC" --out "$ICONSET/icon_${SIZE}x${SIZE}.png" >/dev/null
    DOUBLE=$((SIZE*2))
    sips -z $DOUBLE $DOUBLE "$ICON_SRC" --out "$ICONSET/icon_${SIZE}x${SIZE}@2x.png" >/dev/null
  done
  iconutil -c icns "$ICONSET" -o "$BUNDLE/Contents/Resources/AppIcon.icns" 2>/dev/null || true
  rm -rf "$ICONSET"

  # 声明图标
  /usr/libexec/PlistBuddy -c "Add :CFBundleIconFile string AppIcon" "$BUNDLE/Contents/Info.plist" 2>/dev/null || true
fi

# Ad-hoc 代码签名（不花钱，但能让 macOS 不报"已损坏"）
echo "→ Ad-hoc 代码签名"
codesign --force --deep --sign - "$BUNDLE" 2>&1 | tail -3 || echo "  ⚠ 签名失败（可能已签）"

# 清理构建过程中残留的 quarantine xattr
xattr -cr "$BUNDLE" 2>/dev/null || true

# 验证签名
echo "→ 验证"
codesign --verify --deep --strict "$BUNDLE" 2>&1 | head -3
spctl --assess --verbose "$BUNDLE" 2>&1 | head -3 || true

echo ""
echo "✓ 打包完成：$PWD/$BUNDLE"
echo "  架构：$(lipo -archs "$BUNDLE/Contents/MacOS/$BIN_NAME")"
echo "  签名：ad-hoc (Apple 免费版)"
echo "  最低 macOS：11.0 Big Sur（2020+ 所有 Mac，含 14.5 Sonoma / 15.x Sequoia）"
echo ""
echo "如需分发给团队：在 App 里点「⬇️ 生成分发包」"
echo "（会自动用 ditto 打包，保证 .app 传输后仍可双击运行）"
