#!/usr/bin/env bash
# git-commit-guard · PreToolUse · 全量提交前守卫
#
# 触发时机：git commit 执行前
# 检测内容：
#   1. Commit Message 格式（Conventional Commits）
#   2. 所有暂存文件的铁律扫描：
#      Go   — panic / fmt.Println / error丢弃 / 裸goroutine / float金额 / user_id / is_deleted / camelCase JSON tag
#      SQL  — user_id / is_deleted / float金额 / 缺少 timestamps / 缺少 deleted_at（软删除）
#      Proto— user_id / is_deleted
#
# Claude Code settings.json 配置：
# "hooks": {
#   "PreToolUse": [{"matcher": "Bash", "hooks": [{"type": "command", "command": "bash ~/.claude/hooks/git-commit-guard.sh"}]}]
# }

set -uo pipefail

# ── 读取输入 ──────────────────────────────────────────────────────────────────
INPUT=$(cat)
TOOL=$(echo "$INPUT" | python3 -c "import json,sys; d=json.load(sys.stdin); print(d.get('tool_name',''))" 2>/dev/null || echo "")
[[ "$TOOL" != "Bash" ]] && exit 0

CMD=$(echo "$INPUT" | python3 -c "import json,sys; d=json.load(sys.stdin); print(d.get('tool_input',{}).get('command',''))" 2>/dev/null || echo "")

# 只拦截 git commit（排除 git commit --amend 里纯修 message 无新文件的情形也一起检）
if ! echo "$CMD" | grep -qE '^\s*git\s+commit'; then
  exit 0
fi

ERRORS=()
WARNINGS=()

# ── 1. Commit Message 格式 ──────────────────────────────────────────────────
MSG=""
# 优先从 -m "..." 或 -m '...' 提取（兼容 macOS sed）
MSG=$(echo "$CMD" | sed -nE "s/.*-m[[:space:]]+['\"]([^'\"]+)['\"].*/\1/p" | head -1 || true)
# HEREDOC：取 cat <<'EOF' ... EOF 之间第一非空行
if [[ -z "$MSG" ]]; then
  MSG=$(echo "$CMD" | awk '/cat <</{found=1;next} /^EOF/{found=0} found && NF{print;exit}' | xargs 2>/dev/null || true)
fi

if [[ -n "$MSG" ]]; then
  TYPES="feat|fix|docs|style|refactor|perf|test|chore|ci|revert|build"
  if ! echo "$MSG" | grep -qE "^($TYPES)(\([a-z0-9/_-]+\))?: .{1,72}"; then
    ERRORS+=("⛔ [commit-format] message 不符合 Conventional Commits 规范
   实际：$MSG
   格式：<type>(<scope>): <subject>
   type：feat | fix | docs | style | refactor | perf | test | chore | ci | revert | build
   示例：feat(order): 新增批量取消接口")
  fi
fi

# ── 获取暂存文件列表 ─────────────────────────────────────────────────────────
STAGED=$(git diff --cached --name-only 2>/dev/null || true)
[[ -z "$STAGED" ]] && {
  # 若无暂存文件只做 message 检查
  if [[ ${#ERRORS[@]} -gt 0 ]]; then
    printf '%s\n\n' "${ERRORS[@]}"
    exit 1
  fi
  exit 0
}

# ── 工具函数 ─────────────────────────────────────────────────────────────────
# check_file_pattern <file> <pattern> <exclude_pattern> <label> <advice>
check_staged() {
  local file="$1" pattern="$2" exclude="$3" label="$4" advice="$5"
  [[ ! -f "$file" ]] && return
  local hits
  if [[ -n "$exclude" ]]; then
    hits=$(grep -nE "$pattern" "$file" | grep -vE "$exclude" | head -5 || true)
  else
    hits=$(grep -nE "$pattern" "$file" | head -5 || true)
  fi
  if [[ -n "$hits" ]]; then
    ERRORS+=("⛔ [$label] $file
$(echo "$hits" | sed 's/^/   /')
   → $advice")
  fi
}

for FILE in $STAGED; do
  [[ ! -f "$FILE" ]] && continue

  # ── 跳过不需要检查的路径 ──
  case "$FILE" in
    */vendor/*|*/node_modules/*|*.pb.go) continue ;;
  esac

  # ══ Go 文件检测 ════════════════════════════════════════════════════════════
  if [[ "$FILE" == *.go && "$FILE" != *_test.go ]]; then

    # 排除 cmd/ main.go 的 panic 检测
    case "$FILE" in
      */cmd/*|*/main.go) ;;
      *)
        check_staged "$FILE" '\bpanic\(' '^\s*//' \
          "go-panic" "业务代码禁止裸 panic，改用 xerror.New / xerror.Wrap" ;;
    esac

    # fmt.Print* — 排除注释
    check_staged "$FILE" 'fmt\.(Println|Printf|Print)\(' '^\s*//' \
      "fmt-println" "禁用 fmt.Print*，改用 slog.Info / slog.Error 结构化日志（go-style §7）"

    # error 丢弃：_, _ := 或 _, err = 后 _ 覆盖
    check_staged "$FILE" ',\s*_\s*:?=.*\(.*\)' '^\s*//' \
      "error-ignore" "禁止丢弃 error，必须显式处理（go-style §4.1）"

    # 裸 goroutine
    check_staged "$FILE" '^\s*go\s+func\s*\(' '^\s*//' \
      "goroutine-naked" "禁止裸 go func()，使用 errgroup 管理生命周期（go-style §5）"

    # float 金额字段
    check_staged "$FILE" '(float32|float64).*\b(amount|price|fee|balance|cost|revenue|profit|payment)\b|\b(amount|price|fee|balance|cost|revenue|profit|payment)\b.*(float32|float64)' '^\s*//' \
      "float-amount" "金额/价格字段禁用 float，改用 decimal.Decimal（database §字段规范）"

    # user_id — 排除注释和测试
    check_staged "$FILE" '\buser_id\b' '^\s*//' \
      "user-id" "全平台禁止 user_id，统一用 uid（field-naming §1.2）"

    # is_deleted
    check_staged "$FILE" '\bis_deleted\b' '^\s*//' \
      "soft-delete" "软删除禁用 is_deleted，统一用 deleted_at TIMESTAMPTZ（database §字段规范）"

    # camelCase JSON tag（如 json:"userId" json:"createdAt"）
    CAMEL_HITS=$(grep -nE 'json:"[a-z][a-zA-Z]*[A-Z][a-zA-Z]*"' "$FILE" | grep -v '^\s*//' | head -5 || true)
    if [[ -n "$CAMEL_HITS" ]]; then
      ERRORS+=("⛔ [camelcase-json] $FILE 中 JSON tag 使用了驼峰命名
$(echo "$CAMEL_HITS" | sed 's/^/   /')
   → JSON key 必须 snake_case（如 user_name 而非 userName）")
    fi

    # 时间字段后缀检查：时间字段不含 _at 后缀
    TIME_HITS=$(grep -nE '\b(created|updated|deleted|expired|finished|started)(Time|Date|Timestamp)\b' "$FILE" | grep -v '^\s*//' | head -5 || true)
    if [[ -n "$TIME_HITS" ]]; then
      WARNINGS+=("⚠️  [time-suffix] $FILE 时间字段建议用 _at 后缀（created_at 而非 createdTime）
$(echo "$TIME_HITS" | sed 's/^/   /')")
    fi
  fi

  # ══ SQL 文件检测 ═══════════════════════════════════════════════════════════
  if [[ "$FILE" == *.sql ]]; then

    check_staged "$FILE" '\buser_id\b' '^\s*--' \
      "user-id(sql)" "全平台禁止 user_id 列，统一用 uid"

    check_staged "$FILE" '\bis_deleted\b' '^\s*--' \
      "soft-delete(sql)" "软删除禁用 is_deleted，统一用 deleted_at TIMESTAMPTZ NOT NULL DEFAULT NULL"

    # float 金额类型
    check_staged "$FILE" '\b(FLOAT|DOUBLE|REAL)\b.*(amount|price|fee|balance)|\b(amount|price|fee|balance)\b.*(FLOAT|DOUBLE|REAL)' '^\s*--' \
      "float-amount(sql)" "金额列禁用 FLOAT/DOUBLE/REAL，改用 NUMERIC(18,8) 或 DECIMAL"

    # CREATE TABLE 但缺少 created_at / updated_at
    if grep -qiE 'CREATE\s+TABLE' "$FILE"; then
      if ! grep -qiE '\bcreated_at\b' "$FILE"; then
        ERRORS+=("⛔ [sql-timestamps] $FILE 的 CREATE TABLE 缺少 created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()")
      fi
      if ! grep -qiE '\bupdated_at\b' "$FILE"; then
        ERRORS+=("⛔ [sql-timestamps] $FILE 的 CREATE TABLE 缺少 updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()")
      fi
    fi

    # TIMESTAMP 未带时区（应用 TIMESTAMPTZ）— 找含 TIMESTAMP 但不含 TIMESTAMPTZ 的行
    TS_HITS=$(grep -inE '\bTIMESTAMP\b' "$FILE" | grep -ivE '\bTIMESTAMPTZ\b|\bWITH TIME ZONE\b' | grep -v '^\s*--' | head -5 || true)
    if [[ -n "$TS_HITS" ]]; then
      ERRORS+=("⛔ [timestamptz(sql)] $FILE 时间列必须带时区，用 TIMESTAMPTZ 而非 TIMESTAMP（database §字段规范）
$(echo "$TS_HITS" | sed 's/^/   /')")
    fi

    # VARCHAR 无长度（找 VARCHAR 后面不跟括号的行）
    VC_HITS=$(grep -inE '\bVARCHAR\b' "$FILE" | grep -ivE '\bVARCHAR\s*\(' | grep -v '^\s*--' | head -5 || true)
    if [[ -n "$VC_HITS" ]]; then
      ERRORS+=("⛔ [varchar-len(sql)] $FILE VARCHAR 必须声明长度，如 VARCHAR(255)
$(echo "$VC_HITS" | sed 's/^/   /')")
    fi

    # 主键非 BIGINT / UUID
    PK_HITS=$(grep -niE 'PRIMARY KEY' "$FILE" | grep -ivE '(bigint|bigserial|uuid)' | head -3 || true)
    if [[ -n "$PK_HITS" ]]; then
      WARNINGS+=("⚠️  [sql-pk-type] $FILE 主键建议使用 BIGINT / BIGSERIAL / UUID
$(echo "$PK_HITS" | sed 's/^/   /')")
    fi
  fi

  # ══ Proto 文件检测 ══════════════════════════════════════════════════════════
  if [[ "$FILE" == *.proto ]]; then

    check_staged "$FILE" '\buser_id\b' '^\s*//' \
      "user-id(proto)" "Proto 禁止 user_id，统一用 uid"

    check_staged "$FILE" '\bis_deleted\b' '^\s*//' \
      "soft-delete(proto)" "Proto 禁止 is_deleted，软删除用 deleted_at"

    # float 金额
    check_staged "$FILE" '\bfloat\b.*(amount|price|fee|balance)|\b(amount|price|fee|balance)\b.*\bfloat\b' '^\s*//' \
      "float-amount(proto)" "Proto 金额字段禁用 float，改用 string + 服务层 decimal 转换"
  fi

done  # end for FILE

# ── 汇总输出 ─────────────────────────────────────────────────────────────────
if [[ ${#WARNINGS[@]} -gt 0 ]]; then
  echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
  echo "  📋 提交前警告（不阻断，但请关注）"
  echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
  printf '%s\n\n' "${WARNINGS[@]}"
fi

if [[ ${#ERRORS[@]} -gt 0 ]]; then
  echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
  echo "  🚫 提交被阻断 — 发现 ${#ERRORS[@]} 处铁律违反"
  echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
  printf '%s\n\n' "${ERRORS[@]}"
  echo "修复所有问题后重新提交。"
  exit 1
fi

exit 0
