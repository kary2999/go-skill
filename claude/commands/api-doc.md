---
description: 写接口文档（按 api-doc-example.md 格式）
argument-hint: <接口名 + 用途，如 "POST /api/v1/orders 创建订单">
---

# 写接口文档：$ARGUMENTS

## 步骤

1. **必读**：`~/.claude/skills/go-team-standards/references/api-doc-example.md`
2. **必读**：`~/.claude/skills/go-team-standards/references/api-design.md`
3. **必读**：`~/.claude/skills/go-team-standards/references/error-codes.md`

## 接口文档结构（强制）

```markdown
# POST /api/v1/<path>

> 一句话描述用途

## 通用约定（如未在通用文档定义）
| 项 | 约定 |
|---|---|
| 基础路径 | /api/v1 |
| 认证 | Bearer Token |
| Content-Type | application/json |

## 请求

### URL
`POST /api/v1/<path>`

### Header
| Key | 必需 | 说明 |
|---|---|---|
| Authorization | ✓ | Bearer Token |
| X-Trace-ID | ○ | 客户端可选，服务端会生成 |

### Body
```json
{
  "field_a": "string",
  "field_b": 100
}
```

| 字段 | 类型 | 必需 | 说明 | 取值范围 |
|---|---|---|---|---|
| field_a | string | ✓ | ... | 1-64 字符 |
| field_b | int | ✓ | ... | 1-1e9 |

## 响应

### 成功 (HTTP 200)
```json
{
  "code": 0,
  "message": "success",
  "data": { ... }
}
```

### 错误码
| code | HTTP | 含义 | 触发条件 |
|---|---|---|---|
| 0 | 200 | 成功 | - |
| 40001 | 400 | 参数错 | field_a 缺失 |
| 40901 | 409 | 业务冲突 | 重复幂等键 |
| 50001 | 500 | 服务内部错 | 下游异常 |

## 调用示例

```bash
curl -X POST 'https://api.xxx.com/api/v1/<path>' \
  -H 'Authorization: Bearer xxx' \
  -H 'Content-Type: application/json' \
  -d '{"field_a":"...","field_b":100}'
```

## 边界与注意

- 幂等性：是否幂等？依赖什么 key？
- 限流：每秒上限？
- 超时：上游建议设置多久？
- 兼容性：旧版本表现？字段是否可选？
```

## 命名约束

字段名遵守全局统一命名规范（uid / _at / 金额带前缀）。

## 输出格式

文档头加：
```
<!-- [skill: go-team-standards · 接口文档] $ARGUMENTS -->
```

末尾单独一行：🌟
