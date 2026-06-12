# 断言：testify assert vs require

## 决策树

| 断言失败后继续跑还有意义吗？ | 用 |
|---|---|
| 有意义（记录所有失败点） | `assert.*` |
| 没意义（后续必 panic，如 nil pointer） | `require.*` |

## 典型搭配

```go
// ✅ 错误 / nil 判断 → require（失败就停，避免 nil panic）
require.NoError(t, err)
require.NotNil(t, got)

// ✅ 字段断言 → assert（多个字段失败一次性报全）
assert.Equal(t, "foo", got.Name)
assert.Equal(t, int64(42), got.ID)
assert.True(t, got.Amount.Equal(decimal.NewFromInt(100)))
```

## 错误断言

```go
// ❌ 不要比字符串
assert.Equal(t, "not found", err.Error())

// ❌ 不要 Contains
assert.Contains(t, err.Error(), "not found")

// ✅ 用 errors.Is（检查错误链里是否包含目标错误）
require.Error(t, err)
assert.True(t, errors.Is(err, ErrNotFound))

// ✅ 用 errors.As（需要拆包装取字段）
var apiErr *APIError
require.ErrorAs(t, err, &apiErr)
assert.Equal(t, 404, apiErr.Code)
```

## 金额

```go
// ❌ float 或 decimal 直接 assert.Equal
assert.Equal(t, decimal.NewFromFloat(1.10), got.Amount)  // 精度字段可能差异

// ✅ 用 decimal.Equal
assert.True(t, got.Amount.Equal(decimal.NewFromInt(1).Div(decimal.NewFromInt(10)).Mul(...)))

// 或写 helper
func assertDecimalEq(t *testing.T, want, got decimal.Decimal) {
    t.Helper()
    assert.True(t, want.Equal(got), "want %s got %s", want, got)
}
```

## 时间

```go
// ❌ 精确比对
assert.Equal(t, time.Now(), got.CreatedAt)

// ✅ 容差
assert.WithinDuration(t, time.Now(), got.CreatedAt, time.Second)

// ✅ 注入 fake clock 后精确比对
```

## JSON

```go
assert.JSONEq(t, `{"a":1,"b":2}`, got)
```

## ElementsMatch（顺序无关）

```go
assert.ElementsMatch(t, []int{1, 2, 3}, got)  // got 可以是 {3,1,2}
```

## 自定义失败消息

```go
assert.Equal(t, want, got, "case %s: inputs were %+v", tt.name, tt.input)
```

## 什么时候用 `t.Fatal` 裸调用

几乎不需要。写了就用 require 替代。唯一例外：写全局 TestMain 里的 setup 失败。
