#!/usr/bin/env bash
# ╔══════════════════════════════════════════════════════════════════════════╗
# ║  team-guard · 团队规范一体化守卫                                        ║
# ║  一个文件覆盖所有时机，无需分开安装                                     ║
# ║                                                                          ║
# ║  触发时机（在 settings.json 注册三个事件指向同一脚本）：                 ║
# ║    UserPromptSubmit → DevDefender 需求防御                              ║
# ║    PreToolUse(Bash) → git commit 前：message 格式 + 全量铁律扫描        ║
# ║    PostToolUse(Write|Edit) → 写文件后：铁律逐项检测                     ║
# ║                                                                          ║
# ║  Claude Code settings.json 配置示例：                                   ║
# ║  {                                                                       ║
# ║    "hooks": {                                                            ║
# ║      "UserPromptSubmit": [{                                              ║
# ║        "hooks": [{"type":"command",                                      ║
# ║          "command":"bash ~/.claude/hooks/team-guard.sh"}]               ║
# ║      }],                                                                 ║
# ║      "PreToolUse": [{                                                    ║
# ║        "matcher": "Bash",                                                ║
# ║        "hooks": [{"type":"command",                                      ║
# ║          "command":"bash ~/.claude/hooks/team-guard.sh"}]               ║
# ║      }],                                                                 ║
# ║      "PostToolUse": [{                                                   ║
# ║        "matcher": "Write|Edit",                                          ║
# ║        "hooks": [{"type":"command",                                      ║
# ║          "command":"bash ~/.claude/hooks/team-guard.sh"}]               ║
# ║      }]                                                                  ║
# ║    }                                                                     ║
# ║  }                                                                       ║
# ╚══════════════════════════════════════════════════════════════════════════╝

set -uo pipefail

INPUT=$(cat)

# ── 识别触发时机 ─────────────────────────────────────────────────────────────
TOOL=$(python3 -c "import json,sys; d=json.load(sys.stdin); print(d.get('tool_name',''))" 2>/dev/null <<< "$INPUT" || echo "")
HOOK_TYPE=$(python3 -c "import json,sys; d=json.load(sys.stdin); print(d.get('hook_event_name', d.get('type','')))" 2>/dev/null <<< "$INPUT" || echo "")

# 没有 tool_name → UserPromptSubmit
if [[ -z "$TOOL" ]]; then
  HOOK_TYPE="UserPromptSubmit"
fi

# ══════════════════════════════════════════════════════════════════════════════
# 模块 A · UserPromptSubmit — DevDefender 需求防御
# ══════════════════════════════════════════════════════════════════════════════
if [[ "$HOOK_TYPE" == "UserPromptSubmit" || -z "$TOOL" ]]; then

  PROMPT=$(python3 -c "import sys,json; d=json.load(sys.stdin); print(d.get('prompt',''))" 2>/dev/null <<< "$INPUT" || echo "")
  SESSION_ID=$(python3 -c "import sys,json; d=json.load(sys.stdin); print(d.get('session_id','default'))" 2>/dev/null <<< "$INPUT" || echo "default")

  STATE_DIR="/tmp/devdefender"
  mkdir -p "$STATE_DIR"
  STATE_FILE="$STATE_DIR/active_${SESSION_ID}"

  # 退出
  if echo "$PROMPT" | grep -qiE "(退出|结束|关闭|exit|disable|deactivate).*(devdefender|神盾局)"; then
    rm -f "$STATE_FILE"
    echo "[DevDefender] 已退出，硬规则停止注入。"
    exit 0
  fi

  # 激活
  if echo "$PROMPT" | grep -qiE "(^|[[:space:]])/devdefender(\$|[[:space:]])|^/dd(\$|[[:space:]])|(启动|激活|唤起|开启|enable|activate).*(devdefender|神盾局)"; then
    touch "$STATE_FILE"
    echo "[DevDefender] 已激活。退出：输入「退出 DevDefender」。"
    echo ""
  fi

  # 注入硬规则
  if [[ -f "$STATE_FILE" ]]; then
    cat <<'RULES'
[DevDefender 强制约束 · Hook 注入 · 最高优先级]

本次回复严禁以下行为，违反 = 立刻重写本次输出：

1. 跳过「先问范围（前端/后端/全部）」直接进入分析
   → 首轮激活后第一句必须问范围 + 材料形式，得到回答前不读文件不分析

2. 让产品提供任何技术方案
   → 接口清单 / API 设计 / 字段规格 / 数据结构 / 表结构 一律不准要
   → 只能要求产品说清楚业务规则、业务边界、业务场景

3. 写前端样式或前端交互行为
   → 颜色 / 字体 / 间距 / 动效 / 布局 不写
   → 按钮置灰 / 抽屉打开 / 面板联动 / 字段联动清空 不写
   → 用户选「后端」时连前端都不要提

4. 放行模糊词不追问
   → 出现「AI 分析」「智能匹配」「灵活」「合理」「适当」「友好」必须追问
   → 产品给不出 AI 的 prompt/skill/模型 → 标硬阻塞

5. 自己脑补需求
   → 文档没写 = 不存在 = 标「待产品确认」
   → 不准基于行业常识补全

6. 跳步直接出报告
   → 必须先列疑问 → 等回答 → 再出报告
   → 用户明说「别问了直接出」才允许跳过

7. 用废话八股开头
   → 禁用「综上所述」「基于以上分析」「为了更好地」「在此基础上」
   → 直接说

8. 用产品黑话
   → 禁用「赋能」「抓手」「闭环」「链路」「漏斗」「颗粒度」「对齐」「心智」
   → 用大白话：用户做什么 → 系统做什么 → 出什么结果

【自检】回复前默念：我问范围了吗？我要接口/字段/表了吗？我写前端了吗？
任意一项违反 → 立刻重写本次回复。

RULES
  fi

  exit 0
fi

# ══════════════════════════════════════════════════════════════════════════════
# 模块 B · PreToolUse(Bash) — git commit 前全量守卫
# ══════════════════════════════════════════════════════════════════════════════
if [[ "$TOOL" == "Bash" ]]; then

  CMD=$(python3 -c "import json,sys; d=json.load(sys.stdin); print(d.get('tool_input',{}).get('command',''))" 2>/dev/null <<< "$INPUT" || echo "")

  # 只拦截 git commit
  if ! echo "$CMD" | grep -qE '^\s*git\s+commit'; then
    exit 0
  fi

  ERRORS=()
  WARNINGS=()

  # ── B1. Commit Message 格式 ────────────────────────────────────────────────
  MSG=$(echo "$CMD" | sed -nE "s/.*-m[[:space:]]+['\"]([^'\"]+)['\"].*/\1/p" | head -1 || true)
  if [[ -z "$MSG" ]]; then
    MSG=$(echo "$CMD" | awk '/cat <</{found=1;next} /^EOF/{found=0} found && NF{print;exit}' | xargs 2>/dev/null || true)
  fi

  if [[ -n "$MSG" ]]; then
    TYPES="feat|fix|docs|style|refactor|perf|test|chore|ci|revert|build"
    if ! echo "$MSG" | grep -qE "^($TYPES)(\([a-z0-9/_-]+\))?: .{1,72}"; then
      ERRORS+=("⛔ [commit-format] message 不符合 Conventional Commits 规范
   实际：$MSG
   格式：<type>(<scope>): <subject>
   type：feat|fix|docs|style|refactor|perf|test|chore|ci|revert|build
   示例：feat(order): 新增批量取消接口")
    fi
  fi

  # ── B2. 暂存文件铁律扫描 ───────────────────────────────────────────────────
  STAGED=$(git diff --cached --name-only 2>/dev/null || true)

  if [[ -z "$STAGED" ]]; then
    if [[ ${#ERRORS[@]} -gt 0 ]]; then
      printf '%s\n\n' "${ERRORS[@]}"
      exit 1
    fi
    exit 0
  fi

  _check() {
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
    case "$FILE" in */vendor/*|*/node_modules/*|*.pb.go) continue ;; esac

    # Go 检测
    if [[ "$FILE" == *.go && "$FILE" != *_test.go ]]; then
      case "$FILE" in
        */cmd/*|*/main.go) ;;
        *) _check "$FILE" '\bpanic\(' '^\s*//' "go-panic" "业务代码禁止裸 panic，改用 xerror.New/xerror.Wrap" ;;
      esac
      _check "$FILE" 'fmt\.(Println|Printf|Print)\(' '^\s*//' \
        "fmt-println" "禁用 fmt.Print*，改用 slog 结构化日志（go-style §7）"
      _check "$FILE" ',\s*_\s*:?=.*\(.*\)' '^\s*//' \
        "error-ignore" "禁止丢弃 error，必须显式处理（go-style §4.1）"
      _check "$FILE" '^\s*go\s+func\s*\(' '^\s*//' \
        "goroutine-naked" "禁止裸 go func()，使用 errgroup 管理（go-style §5）"
      _check "$FILE" '(float32|float64).*(amount|price|fee|balance|cost)|(amount|price|fee|balance|cost).*(float32|float64)' '^\s*//' \
        "float-amount" "金额字段禁用 float，改用 decimal.Decimal"
      _check "$FILE" '\buser_id\b' '^\s*//' \
        "user-id" "禁止 user_id，统一用 uid（field-naming §1.2）"
      _check "$FILE" '\bis_deleted\b' '^\s*//' \
        "soft-delete" "禁止 is_deleted，软删除用 deleted_at TIMESTAMPTZ"
      # camelCase JSON tag
      CAMEL=$(grep -nE 'json:"[a-z][a-zA-Z]*[A-Z][a-zA-Z]*"' "$FILE" | grep -v '^\s*//' | head -5 || true)
      [[ -n "$CAMEL" ]] && ERRORS+=("⛔ [camelcase-json] $FILE JSON tag 使用了驼峰命名
$(echo "$CAMEL" | sed 's/^/   /')
   → JSON key 必须 snake_case")
      # 时间字段命名
      TW=$(grep -nE '\b(created|updated|deleted|expired)(Time|Date|Timestamp)\b' "$FILE" | grep -v '^\s*//' | head -3 || true)
      [[ -n "$TW" ]] && WARNINGS+=("⚠️  [time-suffix] $FILE 时间字段建议用 _at 后缀（created_at 而非 createdTime）
$(echo "$TW" | sed 's/^/   /')")
    fi

    # SQL 检测
    if [[ "$FILE" == *.sql ]]; then
      _check "$FILE" '\buser_id\b' '^\s*--' "user-id(sql)" "禁止 user_id 列，统一用 uid"
      _check "$FILE" '\bis_deleted\b' '^\s*--' "soft-delete(sql)" "禁止 is_deleted，用 deleted_at TIMESTAMPTZ"
      _check "$FILE" '\b(FLOAT|DOUBLE|REAL)\b.*(amount|price|fee|balance)|(amount|price|fee|balance).*(FLOAT|DOUBLE|REAL)' '^\s*--' \
        "float-amount(sql)" "金额列禁用 FLOAT/DOUBLE/REAL，改用 NUMERIC(18,8)"
      if grep -qiE 'CREATE\s+TABLE' "$FILE"; then
        grep -qiE '\bcreated_at\b' "$FILE" || ERRORS+=("⛔ [sql-timestamps] $FILE CREATE TABLE 缺少 created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()")
        grep -qiE '\bupdated_at\b' "$FILE" || ERRORS+=("⛔ [sql-timestamps] $FILE CREATE TABLE 缺少 updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()")
      fi
      TS=$(grep -inE '\bTIMESTAMP\b' "$FILE" | grep -ivE '\bTIMESTAMPTZ\b|\bWITH TIME ZONE\b' | grep -v '^\s*--' | head -5 || true)
      [[ -n "$TS" ]] && ERRORS+=("⛔ [timestamptz(sql)] $FILE 时间列用 TIMESTAMPTZ，不要裸 TIMESTAMP
$(echo "$TS" | sed 's/^/   /')")
      VC=$(grep -inE '\bVARCHAR\b' "$FILE" | grep -ivE '\bVARCHAR\s*\(' | grep -v '^\s*--' | head -5 || true)
      [[ -n "$VC" ]] && ERRORS+=("⛔ [varchar-len(sql)] $FILE VARCHAR 必须声明长度，如 VARCHAR(255)
$(echo "$VC" | sed 's/^/   /')")
      PK=$(grep -niE 'PRIMARY KEY' "$FILE" | grep -ivE '(bigint|bigserial|uuid)' | head -3 || true)
      [[ -n "$PK" ]] && WARNINGS+=("⚠️  [sql-pk-type] $FILE 主键建议使用 BIGINT/BIGSERIAL/UUID
$(echo "$PK" | sed 's/^/   /')")
    fi

    # Proto 检测
    if [[ "$FILE" == *.proto ]]; then
      _check "$FILE" '\buser_id\b' '^\s*//' "user-id(proto)" "禁止 user_id，统一用 uid"
      _check "$FILE" '\bis_deleted\b' '^\s*//' "soft-delete(proto)" "禁止 is_deleted，软删除用 deleted_at"
      _check "$FILE" '\bfloat\b.*(amount|price|fee)|(amount|price|fee).*\bfloat\b' '^\s*//' \
        "float-amount(proto)" "金额字段禁用 float，改用 string + 服务层 decimal 转换"
    fi

  done

  # 汇总
  if [[ ${#WARNINGS[@]} -gt 0 ]]; then
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo "  📋 提交前警告（不阻断）"
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
fi

# ══════════════════════════════════════════════════════════════════════════════
# 模块 C · PostToolUse(Write|Edit) — 写文件后铁律检测
# ══════════════════════════════════════════════════════════════════════════════
if [[ "$TOOL" == "Write" || "$TOOL" == "Edit" ]]; then

  FILE=$(python3 -c "import json,sys; d=json.load(sys.stdin); p=d.get('tool_input',{}); print(p.get('file_path') or p.get('path') or '')" 2>/dev/null <<< "$INPUT" || echo "")
  [[ -z "$FILE" || ! -f "$FILE" ]] && exit 0

  case "$FILE" in */vendor/*|*/node_modules/*|*.pb.go) exit 0 ;; esac

  ERRORS=()

  _post_check() {
    local pattern="$1" exclude="$2" label="$3" advice="$4"
    local hits
    if [[ -n "$exclude" ]]; then
      hits=$(grep -nE "$pattern" "$FILE" | grep -vE "$exclude" | head -5 || true)
    else
      hits=$(grep -nE "$pattern" "$FILE" | head -5 || true)
    fi
    if [[ -n "$hits" ]]; then
      ERRORS+=("⛔ [$label] $FILE
$(echo "$hits" | sed 's/^/   /')
   → $advice")
    fi
  }

  # Go 检测
  if [[ "$FILE" == *.go && "$FILE" != *_test.go ]]; then
    case "$FILE" in
      */cmd/*|*/main.go) ;;
      *) _post_check '\bpanic\(' '^\s*//' "go-panic" "业务代码禁止裸 panic，改用 xerror.New/xerror.Wrap" ;;
    esac
    _post_check 'fmt\.(Println|Printf|Print)\(' '^\s*//' \
      "fmt-println" "禁用 fmt.Print*，改用 slog 结构化日志"
    _post_check ',\s*_\s*:?=.*\(.*\)' '^\s*//' \
      "error-ignore" "禁止丢弃 error"
    _post_check '^\s*go\s+func\s*\(' '^\s*//' \
      "goroutine-naked" "禁止裸 go func()，使用 errgroup"
    _post_check '(float32|float64).*(amount|price|fee|balance|cost)|(amount|price|fee|balance|cost).*(float32|float64)' '^\s*//' \
      "float-amount" "金额字段禁用 float，改用 decimal.Decimal"
    _post_check '\buser_id\b' '^\s*//' \
      "user-id" "禁止 user_id，统一用 uid"
    _post_check '\bis_deleted\b' '^\s*//' \
      "soft-delete" "禁止 is_deleted，软删除用 deleted_at TIMESTAMPTZ"
    CAMEL=$(grep -nE 'json:"[a-z][a-zA-Z]*[A-Z][a-zA-Z]*"' "$FILE" | grep -v '^\s*//' | head -5 || true)
    [[ -n "$CAMEL" ]] && ERRORS+=("⛔ [camelcase-json] $FILE JSON tag 驼峰命名
$(echo "$CAMEL" | sed 's/^/   /')
   → JSON key 必须 snake_case")
  fi

  # SQL 检测
  if [[ "$FILE" == *.sql ]]; then
    _post_check '\buser_id\b' '^\s*--' "user-id(sql)" "禁止 user_id 列，统一用 uid"
    _post_check '\bis_deleted\b' '^\s*--' "soft-delete(sql)" "禁止 is_deleted，用 deleted_at TIMESTAMPTZ"
    _post_check '\b(FLOAT|DOUBLE|REAL)\b.*(amount|price|fee|balance)|(amount|price|fee|balance).*(FLOAT|DOUBLE|REAL)' '^\s*--' \
      "float-amount(sql)" "金额列禁用 FLOAT/DOUBLE/REAL，改用 NUMERIC(18,8)"
    if grep -qiE 'CREATE\s+TABLE' "$FILE"; then
      grep -qiE '\bcreated_at\b' "$FILE" || ERRORS+=("⛔ [sql-timestamps] $FILE CREATE TABLE 缺少 created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()")
      grep -qiE '\bupdated_at\b' "$FILE" || ERRORS+=("⛔ [sql-timestamps] $FILE CREATE TABLE 缺少 updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()")
    fi
    TS=$(grep -inE '\bTIMESTAMP\b' "$FILE" | grep -ivE '\bTIMESTAMPTZ\b|\bWITH TIME ZONE\b' | grep -v '^\s*--' | head -5 || true)
    [[ -n "$TS" ]] && ERRORS+=("⛔ [timestamptz(sql)] $FILE 时间列请用 TIMESTAMPTZ
$(echo "$TS" | sed 's/^/   /')")
    VC=$(grep -inE '\bVARCHAR\b' "$FILE" | grep -ivE '\bVARCHAR\s*\(' | grep -v '^\s*--' | head -5 || true)
    [[ -n "$VC" ]] && ERRORS+=("⛔ [varchar-len(sql)] $FILE VARCHAR 需声明长度
$(echo "$VC" | sed 's/^/   /')")
  fi

  # Proto 检测
  if [[ "$FILE" == *.proto ]]; then
    _post_check '\buser_id\b' '^\s*//' "user-id(proto)" "禁止 user_id，统一用 uid"
    _post_check '\bis_deleted\b' '^\s*//' "soft-delete(proto)" "禁止 is_deleted，软删除用 deleted_at"
    _post_check '\bfloat\b.*(amount|price|fee)|(amount|price|fee).*\bfloat\b' '^\s*//' \
      "float-amount(proto)" "金额字段禁用 float，改用 string"
  fi

  if [[ ${#ERRORS[@]} -gt 0 ]]; then
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo "  🚫 铁律违反 — ${#ERRORS[@]} 处问题"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    printf '%s\n\n' "${ERRORS[@]}"
    exit 1
  fi

  exit 0
fi

exit 0
