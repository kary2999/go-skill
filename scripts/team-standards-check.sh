#!/usr/bin/env bash
# Team Standards · 提交前规范检查（pre-commit / CI 用）
# 版本：v1.0.0
#
# 用法：
#   bash team-standards-check.sh                         # pre-commit 模式：扫 git staged 文件
#   bash team-standards-check.sh file1.go file2.sql ...  # 显式扫指定文件
#   bash team-standards-check.sh --all                   # 扫整个 git 仓库（除 .gitignore 内）
#
# 退出码：
#   0  = 无 P0/P1 违规（通过）
#   1  = 有 P0/P1 违规（拒绝 commit）
#   2  = 脚本错误（无 git / 参数错）
#
# 跳过单次检查：
#   git commit --no-verify
#
# 检查的规则（按 14 铁律 + 命名规范）：
#   - 铁律 #1  · 硬编码密钥
#   - 铁律 #2  · errors.New 直接调
#   - 铁律 #3  · _ = err 丢弃
#   - 铁律 #6  · 金额 float64
#   - 铁律 #9  · SELECT *
#   - 铁律 #10 · 数据库外键
#   - 铁律 #12 · 敏感数据进日志（手机 / 邮箱 / 身份证）
#   - 命名 §1.2 · user_id → uid
#   - 命名 §2.1 · 裸 time/timestamp/ts 列
#   - 命名 §2.2 · gmt_create/create_time → created_at
#   - 命名 §4.2 · 裸 amount → 业务前缀
#   - 命名 §4.3 · vol/size/amt → quantity/_qty
#   - 命名 §5.1 · 布尔字段非 is_ 前缀
#   - 命名 §5.2 · is_deleted BOOLEAN → deleted_at
#   - 命名 §1.5 · txid → tx_hash
#   - 命名 §8   · ip_addr/login_ip → client_ip

set -uo pipefail

# ---------------- 颜色 ----------------
if [ -t 1 ]; then
  RED=$'\033[31m'
  YEL=$'\033[33m'
  GRN=$'\033[32m'
  GRY=$'\033[90m'
  BLD=$'\033[1m'
  RST=$'\033[0m'
else
  RED='' ; YEL='' ; GRN='' ; GRY='' ; BLD='' ; RST=''
fi

# ---------------- 收集要检查的文件 ----------------
FILES=()
MODE="staged"

if [ "$#" -gt 0 ]; then
  if [ "$1" = "--all" ]; then
    MODE="all"
    while IFS= read -r f; do
      FILES+=("$f")
    done < <(git ls-files 2>/dev/null)
  else
    MODE="explicit"
    FILES=("$@")
  fi
else
  if ! git rev-parse --is-inside-work-tree >/dev/null 2>&1; then
    echo "${RED}✗ 当前目录不在 git 仓库内${RST}" >&2
    exit 2
  fi
  while IFS= read -r f; do
    FILES+=("$f")
  done < <(git diff --cached --name-only --diff-filter=ACMR 2>/dev/null)
fi

if [ "${#FILES[@]}" -eq 0 ]; then
  echo "${GRY}（没要检查的文件）${RST}"
  exit 0
fi

# ---------------- 违规记录 ----------------
P0_FILE=$(mktemp)
P1_FILE=$(mktemp)
P2_FILE=$(mktemp)
trap 'rm -f "$P0_FILE" "$P1_FILE" "$P2_FILE"' EXIT

record() {
  # record <P0|P1|P2> <规则编号> <文件:行号> <匹配片段> <修复建议>
  local level="$1" rule="$2" loc="$3" snippet="$4" fix="$5"
  local target
  case "$level" in
    P0) target="$P0_FILE" ;;
    P1) target="$P1_FILE" ;;
    *)  target="$P2_FILE" ;;
  esac
  printf '%s\n  位置: %s\n  代码: %s\n  改成: %s\n\n' "$rule" "$loc" "$snippet" "$fix" >> "$target"
}

# ---------------- 检查函数 ----------------

# 是否跳过这个文件（例：测试 fixture / migrations 旧文件 / 自动生成）
should_skip() {
  local f="$1"
  case "$f" in
    *vendor/*)          return 0 ;;
    *node_modules/*)    return 0 ;;
    *.gen.go|*.pb.go)   return 0 ;;
    *_mock.go|*mocks/*) return 0 ;;
    *.lock|*.sum)       return 0 ;;
  esac
  # 不存在 / 不是普通文件 / 二进制？跳过
  [ ! -f "$f" ] && return 0
  return 1
}

is_test_file() {
  case "$1" in
    *_test.go|*testdata/*|*fixtures/*|tests/*) return 0 ;;
  esac
  return 1
}

is_go_file() { [[ "$1" == *.go ]]; }
is_sql_file() { [[ "$1" == *.sql ]]; }
is_proto_file() { [[ "$1" == *.proto ]]; }

# ---------- 铁律 #1 · 硬编码密钥 ----------
check_hardcoded_secret() {
  local f="$1"
  # 经典密钥前缀（OpenAI / clawnova / Anthropic 等）
  grep -nE '(sk-[A-Za-z0-9_-]{16,}|kgb_[A-Za-z0-9_-]{12,}|Bearer [A-Za-z0-9_.-]{20,}|password\s*[:=]\s*"[^"]{6,}")' "$f" 2>/dev/null \
    | head -3 \
    | while IFS=: read -r line content; do
        # 跳过 .example / .template 文件里的演示
        case "$f" in *.example|*.template) continue ;; esac
        record P0 "铁律 #1 · 硬编码密钥" "$f:$line" \
          "$(echo "$content" | head -c 80)" \
          "os.Getenv(\"XXX_KEY\") 或 config.GetString(\"xxx.key\")"
      done
}

# ---------- 铁律 #2 · errors.New 直接调（业务代码）----------
check_bare_errors_new() {
  local f="$1"
  is_go_file "$f" || return
  is_test_file "$f" && return
  grep -nE 'errors\.New\(' "$f" 2>/dev/null \
    | head -3 \
    | while IFS=: read -r line content; do
        record P1 "铁律 #2 · errors.New 不走 xerror" "$f:$line" \
          "$(echo "$content" | head -c 80)" \
          "xerror.New(errno.XxxCode)"
      done
}

# ---------- 铁律 #3 · _ = err 丢弃 ----------
check_discard_err() {
  local f="$1"
  is_go_file "$f" || return
  is_test_file "$f" && return
  grep -nE '_\s*=\s*err\b' "$f" 2>/dev/null \
    | head -3 \
    | while IFS=: read -r line content; do
        record P1 "铁律 #3 · err 丢弃" "$f:$line" \
          "$(echo "$content" | head -c 80)" \
          "return fmt.Errorf(\"<ctx>: %w\", err)"
      done
}

# ---------- 铁律 #6 · 金额 float ----------
check_money_float() {
  local f="$1"
  is_go_file "$f" || is_sql_file "$f" || return
  if is_go_file "$f"; then
    # Go 结构体 / 变量
    grep -nE '\b(balance|amount|fee|price|premium|interest|cost)\w*\s+(float32|float64)\b' "$f" 2>/dev/null \
      | head -3 \
      | while IFS=: read -r line content; do
          record P0 "铁律 #6 · 金额 float（精度丢失）" "$f:$line" \
            "$(echo "$content" | head -c 80)" \
            "decimal.Decimal（github.com/shopspring/decimal）"
        done
  fi
  if is_sql_file "$f"; then
    # SQL DDL FLOAT 类型 + 金额名
    grep -niE '\b(balance|amount|fee|price|premium)\w*\s+(FLOAT|DOUBLE|REAL)' "$f" 2>/dev/null \
      | head -3 \
      | while IFS=: read -r line content; do
          record P0 "铁律 #6 · SQL 金额 FLOAT/DOUBLE" "$f:$line" \
            "$(echo "$content" | head -c 80)" \
            "DECIMAL(28,8) 或加密币场景 NUMERIC(38,18)"
        done
  fi
}

# ---------- 铁律 #9 · SELECT * ----------
check_select_star() {
  local f="$1"
  # Go / SQL 都查
  grep -nE 'SELECT\s+\*' "$f" 2>/dev/null \
    | head -3 \
    | while IFS=: read -r line content; do
        # 跳过注释里的（粗略：行首是 -- 或 //）
        case "$content" in
          *"--"*"SELECT"*) continue ;;
          *"//"*"SELECT"*) continue ;;
        esac
        record P1 "铁律 #9 · 禁 SELECT *" "$f:$line" \
          "$(echo "$content" | head -c 80)" \
          "明确列出字段：SELECT id, uid, ..."
      done
}

# ---------- 铁律 #10 · 数据库外键 ----------
check_foreign_key() {
  local f="$1"
  if is_sql_file "$f"; then
    grep -niE 'FOREIGN\s+KEY|REFERENCES\s+[a-z_]+\s*\(' "$f" 2>/dev/null \
      | head -3 \
      | while IFS=: read -r line content; do
          record P1 "铁律 #10 · 禁数据库级 FOREIGN KEY" "$f:$line" \
            "$(echo "$content" | head -c 80)" \
            "关联关系由应用层维护（影响分库分表 / DDL 灵活性）"
        done
  fi
  if is_go_file "$f"; then
    grep -nE '`gorm:"[^"]*foreignKey:' "$f" 2>/dev/null \
      | head -3 \
      | while IFS=: read -r line content; do
          record P1 "铁律 #10 · GORM foreignKey tag" "$f:$line" \
            "$(echo "$content" | head -c 80)" \
            "去掉 foreignKey tag，应用层维护关联"
        done
  fi
}

# ---------- 铁律 #12 · 敏感数据进日志 / 测试 fixture ----------
check_sensitive_in_log() {
  local f="$1"
  is_go_file "$f" || return
  # 形如 slog.Info(..., "phone", "13800138000")
  grep -nE '(slog|log)\.[A-Z][a-z]+\(.*"(phone|mobile|email|id_card|身份证|手机号)"' "$f" 2>/dev/null \
    | head -3 \
    | while IFS=: read -r line content; do
        record P0 "铁律 #12 · 敏感数据进日志（合规事故）" "$f:$line" \
          "$(echo "$content" | head -c 80)" \
          "脱敏后入日志（mask.Phone(x) / mask.Email(x)）或不入日志"
      done
}

# ---------- 命名 §1.2 · user_id → uid ----------
check_naming_uid() {
  local f="$1"
  if is_sql_file "$f"; then
    # SQL 列：user_id BIGINT
    grep -nE '\buser_id\b' "$f" 2>/dev/null \
      | head -3 \
      | while IFS=: read -r line content; do
          # 跳过注释
          case "$content" in *"--"*"user_id"*) continue ;; esac
          record P1 "命名 §1.2 · 用户字段统一 uid" "$f:$line" \
            "$(echo "$content" | head -c 80)" \
            "uid VARCHAR(64) NOT NULL"
        done
  fi
  if is_go_file "$f"; then
    # Go struct field：UserID 或 user_id JSON tag
    grep -nE '`json:"user_id"|UserID\s+(int64|uint64|string|int)\b' "$f" 2>/dev/null \
      | head -3 \
      | while IFS=: read -r line content; do
          record P1 "命名 §1.2 · 用户字段统一 uid" "$f:$line" \
            "$(echo "$content" | head -c 80)" \
            "字段名 UID/Uid 或 json tag uid"
        done
  fi
  if is_proto_file "$f"; then
    grep -nE '\buser_id\b' "$f" 2>/dev/null \
      | head -3 \
      | while IFS=: read -r line content; do
          record P1 "命名 §1.2 · proto 字段 uid" "$f:$line" \
            "$(echo "$content" | head -c 80)" \
            "string uid = X;"
        done
  fi
}

# ---------- 命名 §2.2 · gmt_create / create_time → created_at ----------
check_naming_time_field() {
  local f="$1"
  is_sql_file "$f" || is_go_file "$f" || is_proto_file "$f" || return
  grep -nE '\b(gmt_create|gmt_modified|create_time|update_time|ctime|utime|created_time|updated_time)\b' "$f" 2>/dev/null \
    | head -3 \
    | while IFS=: read -r line content; do
        case "$content" in *"--"*) continue ;; esac
        record P1 "命名 §2.2 · 时间字段统一 _at 后缀" "$f:$line" \
          "$(echo "$content" | head -c 80)" \
          "created_at / updated_at TIMESTAMPTZ(6) UTC"
      done
}

# ---------- 命名 §2.1 · 裸 time/timestamp/ts 列 ----------
check_naming_bare_time() {
  local f="$1"
  is_sql_file "$f" || return
  # SQL DDL 列名为 time / timestamp / ts（业务列，非元数据）
  grep -niE '^\s*(time|timestamp|ts)\s+(TIMESTAMP|DATETIME|INT|BIGINT)' "$f" 2>/dev/null \
    | head -3 \
    | while IFS=: read -r line content; do
        record P2 "命名 §2.1 · 裸 time/timestamp/ts 列" "$f:$line" \
          "$(echo "$content" | head -c 80)" \
          "用动作语义：executed_at / settled_at / verified_at"
      done
}

# ---------- 命名 §4.2 · 裸 amount 列 ----------
check_naming_bare_amount() {
  local f="$1"
  is_sql_file "$f" || return
  grep -niE '^\s*amount\s+(NUMERIC|DECIMAL)' "$f" 2>/dev/null \
    | head -3 \
    | while IFS=: read -r line content; do
        record P1 "命名 §4.2 · 裸 amount 列需业务前缀" "$f:$line" \
          "$(echo "$content" | head -c 80)" \
          "order_amount / fee_amount / withdrawal_amount"
      done
}

# ---------- 命名 §4.3 · vol/size/amt 缩写 ----------
check_naming_abbrev() {
  local f="$1"
  is_sql_file "$f" || return
  grep -niE '^\s*(vol|size|amt)\s+(INT|BIGINT|NUMERIC|DECIMAL)' "$f" 2>/dev/null \
    | head -3 \
    | while IFS=: read -r line content; do
        record P2 "命名 §4.3 · 缩写命名 vol/size/amt" "$f:$line" \
          "$(echo "$content" | head -c 80)" \
          "quantity / order_qty / page_size"
      done
}

# ---------- 命名 §5.1 · 布尔非 is_/has_/can_ 前缀 ----------
check_naming_bool_prefix() {
  local f="$1"
  is_sql_file "$f" || return
  grep -niE '^\s*[a-z_]+\s+(BOOLEAN|BOOL|TINYINT\(1\))' "$f" 2>/dev/null \
    | grep -viE '^\s*\b(is_|has_|can_|allow_|deleted_at)' \
    | head -3 \
    | while IFS=: read -r line content; do
        # 跳过审计字段
        case "$content" in
          *"deleted_at"*) continue ;;
          *"id"*"PRIMARY"*) continue ;;
        esac
        record P2 "命名 §5.1 · 布尔字段需 is_ / has_ / can_ 前缀" "$f:$line" \
          "$(echo "$content" | head -c 80)" \
          "is_enabled / has_verified / can_withdraw"
      done
}

# ---------- 命名 §5.2 · is_deleted → deleted_at ----------
check_naming_soft_delete() {
  local f="$1"
  is_sql_file "$f" || return
  grep -niE 'is_deleted\s+(BOOLEAN|BOOL|TINYINT)' "$f" 2>/dev/null \
    | head -3 \
    | while IFS=: read -r line content; do
        record P1 "命名 §5.2 · 软删除统一 deleted_at" "$f:$line" \
          "$(echo "$content" | head -c 80)" \
          "deleted_at TIMESTAMPTZ(6) （NULL 表示未删）"
      done
}

# ---------- 命名 §1.5 · txid / transaction_hash ----------
check_naming_tx_hash() {
  local f="$1"
  grep -nE '\b(txid|transaction_hash)\b' "$f" 2>/dev/null \
    | head -3 \
    | while IFS=: read -r line content; do
        case "$content" in *"--"*) continue ;; esac
        record P2 "命名 §1.5 · 区块链交易哈希用 tx_hash" "$f:$line" \
          "$(echo "$content" | head -c 80)" \
          "tx_hash VARCHAR(128)"
      done
}

# ---------- 命名 §8 · ip_addr / login_ip ----------
check_naming_client_ip() {
  local f="$1"
  grep -nE '\b(ip_addr|login_ip|user_ip|register_ip)\b' "$f" 2>/dev/null \
    | head -3 \
    | while IFS=: read -r line content; do
        case "$content" in *"--"*) continue ;; esac
        record P2 "命名 §8 · 客户端 IP 统一 client_ip" "$f:$line" \
          "$(echo "$content" | head -c 80)" \
          "client_ip INET"
      done
}

# ---------------- 主流程 ----------------
echo "${BLD}🔍 Team Standards 规范检查${RST}（模式: $MODE，文件数: ${#FILES[@]}）"
echo ""

CHECKED=0
for f in "${FILES[@]}"; do
  should_skip "$f" && continue
  CHECKED=$((CHECKED+1))

  check_hardcoded_secret "$f"
  check_bare_errors_new "$f"
  check_discard_err "$f"
  check_money_float "$f"
  check_select_star "$f"
  check_foreign_key "$f"
  check_sensitive_in_log "$f"
  check_naming_uid "$f"
  check_naming_time_field "$f"
  check_naming_bare_time "$f"
  check_naming_bare_amount "$f"
  check_naming_abbrev "$f"
  check_naming_bool_prefix "$f"
  check_naming_soft_delete "$f"
  check_naming_tx_hash "$f"
  check_naming_client_ip "$f"
done

P0_COUNT=$(grep -c '^铁律\|^命名' "$P0_FILE" 2>/dev/null || echo 0)
P1_COUNT=$(grep -c '^铁律\|^命名' "$P1_FILE" 2>/dev/null || echo 0)
P2_COUNT=$(grep -c '^铁律\|^命名' "$P2_FILE" 2>/dev/null || echo 0)

if [ -s "$P0_FILE" ]; then
  echo "${RED}${BLD}🔴 P0 阻断（必改才能 commit）· ${P0_COUNT} 条${RST}"
  echo ""
  cat "$P0_FILE"
fi
if [ -s "$P1_FILE" ]; then
  echo "${YEL}${BLD}🟡 P1 重要（强烈建议改）· ${P1_COUNT} 条${RST}"
  echo ""
  cat "$P1_FILE"
fi
if [ -s "$P2_FILE" ]; then
  echo "${GRY}${BLD}🔵 P2 改进 · ${P2_COUNT} 条${RST}"
  echo ""
  cat "$P2_FILE"
fi

TOTAL=$((P0_COUNT + P1_COUNT + P2_COUNT))
if [ "$TOTAL" -eq 0 ]; then
  echo "${GRN}✓ 通过 · 已检查 $CHECKED 个文件，无违规${RST}"
  exit 0
fi

echo "──────────────────────────────────────"
echo "${BLD}已检查 $CHECKED 个文件，发现违规：${RST}"
echo "  ${RED}P0 ${P0_COUNT} 条${RST}（阻断）"
echo "  ${YEL}P1 ${P1_COUNT} 条${RST}（建议）"
echo "  ${GRY}P2 ${P2_COUNT} 条${RST}（改进）"
echo ""

if [ "$P0_COUNT" -gt 0 ] || [ "$P1_COUNT" -gt 0 ]; then
  echo "${RED}✗ commit 拒绝（紧急情况可 git commit --no-verify 跳过）${RST}"
  exit 1
fi

echo "${GRN}✓ 仅 P2，放行${RST}"
exit 0
