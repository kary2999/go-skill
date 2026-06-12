#!/usr/bin/env bash
# 在 git 项目根目录跑：把 team-standards-check.sh 装为 pre-commit hook
#
# 用法：
#   cd ~/your-go-service
#   bash <(cat /path/to/install-precommit.sh)
#
#   或 App 一键装（通过 /api/commit-guard/install 后端）

set -euo pipefail

# 必须在 git 项目内
if [ ! -d ".git" ]; then
    echo "✗ 当前目录不是 git 仓库（缺 .git/）"
    exit 1
fi

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
CHECK_SCRIPT="${SCRIPT_DIR}/team-standards-check.sh"

if [ ! -f "${CHECK_SCRIPT}" ]; then
    echo "✗ 找不到 team-standards-check.sh"
    echo "  确认本脚本和 team-standards-check.sh 同目录"
    exit 1
fi

PROJ_SCRIPTS_DIR="scripts"
PROJ_CHECK="${PROJ_SCRIPTS_DIR}/team-standards-check.sh"
HOOK=".git/hooks/pre-commit"

mkdir -p "${PROJ_SCRIPTS_DIR}"
cp "${CHECK_SCRIPT}" "${PROJ_CHECK}"
chmod +x "${PROJ_CHECK}"
echo "✓ 检查脚本落地: ${PROJ_CHECK}"

# 写 hook
cat > "${HOOK}" <<'HOOK_EOF'
#!/usr/bin/env bash
# 团队规范 pre-commit hook（v1.0）· 由 Team Standards App 装
# 跳过单次检查：git commit --no-verify

set -uo pipefail

SCRIPT="scripts/team-standards-check.sh"
if [ ! -x "${SCRIPT}" ]; then
    echo "⚠ ${SCRIPT} 不存在或不可执行，跳过团队规范检查"
    exit 0
fi

bash "${SCRIPT}"
HOOK_EOF

chmod +x "${HOOK}"
echo "✓ Pre-commit hook 装好: ${HOOK}"
echo ""
echo "下次 git commit 时会自动跑规范检查。"
echo "紧急情况跳过：git commit --no-verify"
echo ""
echo "💡 建议把 scripts/team-standards-check.sh 一起提交，让团队所有人都生效。"
