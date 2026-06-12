---
title: "我的开发档案"
version: "1.0.0"
last_modified: "2026-04-27"
---

# 🧬 我的开发档案（dev-dna profile）

> 这是默认模板。在 Team Standards App「🧬 我的开发 DNA」卡片里编辑后会覆盖此文件。
> AI 读这份档案后，回答任何技术问题时**必须**遵循下方偏好。

---

## 1. 基本信息

- **角色**：（如：Go 后端工程师 / 微服务架构师 / Tech Lead）___
- **主栈**：（如：Go 1.22 · Kratos v2 · Wire · Buf · PostgreSQL · Kafka · Redis）___
- **次栈**：（次要语言 / 框架，如：TypeScript · React）___
- **团队**：___
- **业务领域**：___（如：交易系统 / 数据中台 / 风控）

## 2. 编码偏好（强 → 弱）

> 描述你**主动选择**的风格。AI 应该按此产出代码。

- **错误处理**：必须包装 + 走 xerror.New(errno.X)；禁裸 errors.New
- **命名**：snake_case for SQL columns / camelCase for Go vars / kebab-case for k8s resources
- **测试**：表驱动 + mockgen，禁手写 mock
- **注释**：godoc 风格（`// 函数名 does ...`）+ 中文业务注释
- **代码组织**：interface 先行；Wire 注入；usecase / data / service 三层
- **金额**：必须 decimal.Decimal；DB DECIMAL(28,8)
- **日志**：slog + snake_case key + trace_id

（其他自由补充：）
- ___
- ___

## 3. 反偏好（不喜欢看到的代码）

> AI 看到这些应该主动指出 + 改成你偏好的写法。

- ❌ panic 在业务代码
- ❌ SELECT * / OFFSET 大翻页
- ❌ fmt.Println 当日志
- ❌ float64 存金额
- ❌ 注释掉的死代码 / TODO 不带日期 + 责任人
- ❌ 单测里直连真 DB

（其他自由补充：）
- ❌ ___

## 4. 常用模式（"按我的习惯写"对应的具体套路）

> 用户说"按我习惯实现 X"时，AI 直接照这些套路产出，不要再问。

### 4.1 新建一个 Kratos 微服务
- cmd/main.go 走 Wire 装配
- proto 定义在 api/<biz>/v1/<biz>.proto
- biz / data / service 三层
- 错误码在 internal/errno/<biz>.go

### 4.2 新增数据表
- 先写 migration 文件 `migrations/YYYYMMDD_NN_<action>.sql`
- 主键 BIGSERIAL；`id`、`created_at`、`updated_at`、`deleted_at` 标配
- 时间字段 TIMESTAMPTZ(6)
- 禁外键

### 4.3 新增 HTTP 接口
- 先在 .proto 加 RPC + http binding
- handler 走 service 层转 usecase
- 错误统一返回 xerror，前端按 errno 国际化

（自由补充更多模式：）
### 4.x ___

## 5. 工具链 / 环境

- **IDE**：（Cursor / Claude Code / GoLand / VS Code）___
- **终端**：（iTerm2 / Warp / kitty）___
- **shell**：（zsh / bash / fish）___
- **关键 alias**：（`gst=git status` / `k=kubectl` 等）___
- **测试命令**：`go test -race -cover ./...`
- **lint**：`golangci-lint run --timeout=5m`

## 6. 决策口味

- **性能 vs 可读性**：默认___（如：可读性优先，热路径再优化）
- **新依赖引入**：默认___（如：保守，先看是否能用 stdlib）
- **上线节奏**：偏好___（如：周二 / 周四，避开周一周五）
- **评审风格**：___（如：先指铁律，再指细节，最后给修正代码）

## 7. 技术债 / 我自己常犯的错

> 让 AI 主动提醒你别再犯。

- ___（如：我经常忘记 ctx.Done() 的 select）
- ___（如：写 SQL 容易漏 LIMIT）

## 8. 学习方向 / 兴趣

> 让 AI 在合适的话题下推荐进阶资料。

- ___（如：分布式一致性 / Raft / OLAP）
- ___

---

## ✏️ 编辑指南

- 每次重大变化（换栈、换团队、风格演进）→ 更新顶部 `last_modified` + 提 `version`
- 内容尽量**具体可执行**（"我喜欢 xerror.New(errno.X)"），别太抽象（"我喜欢好的错误处理"）
- 跨电脑迁移：**整个 `~/.claude/skills/dev-dna/` 目录 tar 走**就行
