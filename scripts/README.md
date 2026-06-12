# Team Standards · 提交规范化（Layer 2 + Layer 3）

> 一句话：让违反 14 铁律 + 命名规范的代码**根本进不了 main 分支**。

---

## 三层防御架构

```
Layer 1 · AI Skill（go-team-standards / code-review）—— 软约束，写代码时建议
   ↓ 用户可绕过
Layer 2 · pre-commit hook —— 本地 commit 前 grep 检查，违规拒绝
   ↓ 用户可 git commit --no-verify 跳过
Layer 3 · GitLab CI Gate —— MR push 后 CI 跑检查，违规阻断 merge
```

Skill 是 Layer 1（advisory）。本目录的脚本是 Layer 2 + Layer 3。

---

## 文件清单

| 文件 | 用途 |
|---|---|
| `team-standards-check.sh` | 核心检查脚本（bash + grep 实现 16 条规则） |
| `install-precommit.sh` | 把 check.sh 装到 git 项目作 pre-commit hook |
| `gitlab-ci-snippet.yml` | 拷贝到项目 .gitlab-ci.yml 的 CI 模板 |

---

## 检查覆盖的规则（v1.0）

| Level | 规则 | 触发文件 |
|---|---|---|
| **P0** 阻断 | 铁律 #1 · 硬编码密钥 | 所有 |
| **P0** 阻断 | 铁律 #6 · 金额 float | .go / .sql |
| **P0** 阻断 | 铁律 #12 · 敏感数据进日志 | .go |
| **P1** 重要 | 铁律 #2 · errors.New 不走 xerror | .go |
| **P1** 重要 | 铁律 #3 · _ = err 丢弃 | .go |
| **P1** 重要 | 铁律 #9 · SELECT * | .go / .sql |
| **P1** 重要 | 铁律 #10 · 数据库外键 | .sql / .go |
| **P1** 重要 | 命名 §1.2 · user_id → uid | .go / .sql / .proto |
| **P1** 重要 | 命名 §2.2 · gmt_create → created_at | .go / .sql / .proto |
| **P1** 重要 | 命名 §4.2 · 裸 amount 列 | .sql |
| **P1** 重要 | 命名 §5.2 · is_deleted → deleted_at | .sql |
| **P2** 改进 | 命名 §2.1 · 裸 time/timestamp/ts 列 | .sql |
| **P2** 改进 | 命名 §4.3 · vol/size/amt 缩写 | .sql |
| **P2** 改进 | 命名 §5.1 · 布尔字段非 is_ 前缀 | .sql |
| **P2** 改进 | 命名 §1.5 · txid / transaction_hash | 所有 |
| **P2** 改进 | 命名 §8 · ip_addr / login_ip | 所有 |

**P0/P1 违规** → commit 拒绝 / MR 失败  
**P2 违规** → 仅警告，放行

---

## 装到自己项目（最快路径）

### 方式 A · 通过 Team Standards App（推荐）

打开 App → 「⚡ 安装」 →「更新与同步」 →「📋 提交规范化」 → 填项目路径 → 「📥 装 pre-commit hook」。

App 会：
1. 拷贝 `team-standards-check.sh` 到 `<你的项目>/scripts/`
2. 写 `.git/hooks/pre-commit` 调用它
3. chmod +x 二者

### 方式 B · 命令行手动装

```bash
# 拷贝脚本到自己项目
cp /Applications/Team\ Standards.app/Contents/Resources/scripts/* ~/your-project/scripts/
cd ~/your-project
bash scripts/install-precommit.sh
```

### 方式 C · CI 加 job

把 `gitlab-ci-snippet.yml` 内容拷贝到你项目的 `.gitlab-ci.yml`，确保 `scripts/team-standards-check.sh` 也提交到 git。

---

## 使用流程

### 本地 commit 时（Layer 2）

```bash
$ git add internal/wallet.go
$ git commit -m "feat: add balance query"

🔍 Team Standards 规范检查（模式: staged，文件数: 1）

🔴 P0 阻断（必改才能 commit）· 1 条

铁律 #6 · 金额 float（精度丢失）
  位置: internal/wallet.go:23
  代码:     Balance float64
  改成: decimal.Decimal（github.com/shopspring/decimal）

──────────────────────────────────────
已检查 1 个文件，发现违规：
  P0 1 条（阻断）
  P1 0 条（建议）
  P2 0 条（改进）

✗ commit 拒绝（紧急情况可 git commit --no-verify 跳过）
```

→ 改代码 → 重新 commit。

### MR push 时（Layer 3）

`team-standards-check` job 在 CI 跑：
- 没违规 → ✓ 绿灯，MR 可 merge
- 有 P0/P1 → ✗ 红灯，MR 阻断（哪怕 Approve 也不能 merge）

---

## 紧急跳过

某些场景必须强推（线上紧急 hotfix / 已知违规计划下版本修）：

```bash
# 本地跳过 pre-commit
git commit --no-verify -m "..."

# CI 跳过（**强烈不推荐**，违反团队规范）
# 在 MR 描述加 [skip lint] 也不行——只有 maintainer 手动 retry-skip 才行
```

---

## 反馈 / 加新规则

想加新规则到 `team-standards-check.sh`：

1. fork 公开 mirror `github.com/kary2999/standards`
2. 改 `scripts/team-standards-check.sh` 加 grep 函数
3. 提 MR 给团队规范 owner
4. 合并后 Team Standards App 下版本自动同步

或直接在你项目里改一份本地版本（不会被覆盖，但其他人没有）。
