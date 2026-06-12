#!/usr/bin/env bash
# Team Standards · 自升级 detached helper（v1.7.38）
#
# 由 App 后端在升级前 spawn，detach 运行。App 自己 quit 后由本脚本接力：
#   1. 等当前 App PID 退出
#   2. 备份当前 .app
#   3. mount DMG
#   4. 拷新 .app 到 /Applications/
#   5. detach DMG
#   6. xattr -cr
#   7. 启动新 App
#   8. 失败任何一步 → 回滚旧 .app
#
# 用法（被 App 自动调用，不用人跑）：
#   bash upgrade-helper.sh <parent-pid> <dmg-path> <target-app-path> <log-path>

set -uo pipefail

PARENT_PID="${1:-}"
DMG="${2:-}"
APP="${3:-}"
LOG="${4:-/tmp/team-standards-upgrade.log}"

if [ -z "$PARENT_PID" ] || [ -z "$DMG" ] || [ -z "$APP" ]; then
  echo "usage: $0 <pid> <dmg> <app> [log]" >&2
  exit 2
fi

# 全部输出到 log（detached 后没 TTY）
exec >> "$LOG" 2>&1

echo ""
echo "============================================"
echo "$(date '+%Y-%m-%d %H:%M:%S') · upgrade helper starting"
echo "  parent_pid: $PARENT_PID"
echo "  dmg:        $DMG"
echo "  app:        $APP"
echo "============================================"

# --- 1. 等父进程退出（最多 60s） ---
for i in $(seq 1 60); do
  if ! kill -0 "$PARENT_PID" 2>/dev/null; then
    echo "[$i s] parent quit ✓"
    break
  fi
  sleep 1
done
if kill -0 "$PARENT_PID" 2>/dev/null; then
  echo "✗ parent didn't quit in 60s, abort"
  exit 1
fi

# --- 2. 备份当前 .app ---
APP_BAK=""
if [ -d "$APP" ]; then
  APP_BAK="${APP}.bak-$(date +%Y%m%d-%H%M%S)"
  if mv "$APP" "$APP_BAK"; then
    echo "✓ backed up to: $APP_BAK"
  else
    echo "✗ backup failed (permission?)"
    exit 1
  fi
fi

# 失败回滚函数
rollback() {
  if [ -n "$APP_BAK" ] && [ -d "$APP_BAK" ]; then
    rm -rf "$APP"
    mv "$APP_BAK" "$APP" 2>/dev/null
    echo "↩ rolled back from $APP_BAK"
  fi
}

# --- 3. mount DMG（优先使用 Go 预挂载路径，避免含空格解析失败）---
PRE_MOUNT="${5:-}"
if [ -n "$PRE_MOUNT" ] && [ -d "$PRE_MOUNT" ]; then
  MOUNT="$PRE_MOUNT"
  echo "✓ using pre-mounted (from Go): $MOUNT"
else
  # 兼容旧版：Go 没传或路径不存在时自己 mount
  MOUNT=$(hdiutil attach -nobrowse -readonly -plist "$DMG" 2>/dev/null \
    | python3 -c "
import plistlib, sys
d = plistlib.loads(sys.stdin.buffer.read())
mps = [e.get('mount-point','') for e in d.get('system-entities',[]) if e.get('mount-point','')]
print(mps[-1] if mps else '', end='')
" 2>/dev/null)
  if [ -z "$MOUNT" ] || [ ! -d "$MOUNT" ]; then
    echo "✗ hdiutil mount failed (MOUNT='$MOUNT')"
    rollback
    exit 1
  fi
  echo "✓ mounted (self): $MOUNT"
fi

# --- 4. 找 DMG 里的 .app ---
NEW_APP=$(find "$MOUNT" -maxdepth 2 -name "*.app" -type d 2>/dev/null | head -1)
if [ -z "$NEW_APP" ] || [ ! -d "$NEW_APP" ]; then
  echo "✗ no .app found in DMG"
  hdiutil detach -force "$MOUNT" 2>/dev/null
  rollback
  exit 1
fi
echo "✓ found new app: $NEW_APP"

# --- 5. 拷贝到 /Applications/（或者父目录） ---
APP_PARENT=$(dirname "$APP")
if ! cp -R "$NEW_APP" "$APP_PARENT/"; then
  echo "✗ copy failed"
  hdiutil detach -force "$MOUNT" 2>/dev/null
  rollback
  exit 1
fi

# 验证拷贝结果
if [ ! -d "$APP" ]; then
  # 可能源 .app 文件名跟目标不一样（比如 Team Standards.app vs team-standards.app）
  NEW_APP_NAME=$(basename "$NEW_APP")
  COPIED_APP="$APP_PARENT/$NEW_APP_NAME"
  if [ -d "$COPIED_APP" ] && [ "$COPIED_APP" != "$APP" ]; then
    mv "$COPIED_APP" "$APP" 2>/dev/null
  fi
fi

if [ ! -d "$APP" ]; then
  echo "✗ post-copy .app missing at $APP"
  hdiutil detach -force "$MOUNT" 2>/dev/null
  rollback
  exit 1
fi
echo "✓ copied to: $APP"

# --- 6. detach DMG ---
hdiutil detach -force "$MOUNT" 2>/dev/null
echo "✓ unmounted"

# --- 7. xattr 解隔离 ---
xattr -cr "$APP" 2>/dev/null
echo "✓ xattr cleaned"

# --- 8. 拷贝成功，可以删备份了 ---
if [ -n "$APP_BAK" ] && [ -d "$APP_BAK" ]; then
  # 保留 1 个版本备份以防新版有 bug 立即回滚
  # 用户可手动 rm -rf 清理
  echo "↩ backup kept at: $APP_BAK"
  echo "  （新版稳定后可手动删: rm -rf '$APP_BAK'）"
fi

# --- 9. 启动新 App ---
sleep 1
if open "$APP"; then
  echo "✓ relaunched new app"
else
  echo "⚠ open command failed, 请手工打开: $APP"
fi

# --- 10. 删 DMG 临时文件 ---
case "$DMG" in
  /tmp/*|/var/folders/*)
    rm -f "$DMG"
    echo "✓ removed temp dmg: $DMG"
    ;;
esac

echo "$(date '+%Y-%m-%d %H:%M:%S') · upgrade helper done ✓"
echo ""
