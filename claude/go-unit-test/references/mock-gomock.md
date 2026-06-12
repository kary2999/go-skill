# mockgen + gomock 用法

## 生成 mock

源文件头加：
```go
//go:generate mockgen -source=order_repo.go -destination=mocks/mock_order_repo.go -package=mocks
```

跑：
```bash
go generate ./...
```

**禁手写 mock**。接口变了忘改就是假绿。

## 基本 EXPECT

```go
ctrl := gomock.NewController(t)
defer ctrl.Finish()

repo := mocks.NewMockOrderRepo(ctrl)

// 单次期望
repo.EXPECT().
    GetByID(gomock.Any(), int64(42)).
    Return(&Order{ID: 42}, nil)

// 多次期望
repo.EXPECT().Create(gomock.Any(), gomock.Any()).
    Return(&Order{}, nil).
    Times(3)

// 至少 N 次
repo.EXPECT().Update(gomock.Any(), gomock.Any()).
    Return(nil).
    MinTimes(1)

// 永不调用
repo.EXPECT().Delete(gomock.Any(), gomock.Any()).Times(0)
```

## 匹配器

| 匹配器 | 含义 |
|---|---|
| `gomock.Any()` | 任意值 |
| `gomock.Eq(x)` | 等于 x（深比较） |
| `gomock.Nil()` | nil |
| `gomock.Not(x)` | 不等于 x |
| `gomock.Len(3)` | 长度为 3 |
| 直接写值 `int64(42)` | 等于 42（同 Eq） |
| `gomock.AssignableToTypeOf(&Foo{})` | 类型匹配 |

## 自定义匹配器

```go
type amountGTE struct{ min decimal.Decimal }
func (m amountGTE) Matches(x any) bool {
    o, ok := x.(*Order)
    return ok && o.Amount.GreaterThanOrEqual(m.min)
}
func (m amountGTE) String() string { return "amount >= " + m.min.String() }

repo.EXPECT().Create(gomock.Any(), amountGTE{decimal.NewFromInt(100)})
```

## 动态返回

```go
repo.EXPECT().GetByID(gomock.Any(), gomock.Any()).
    DoAndReturn(func(ctx context.Context, id int64) (*Order, error) {
        if id == 404 {
            return nil, ErrNotFound
        }
        return &Order{ID: id}, nil
    })
```

## 顺序断言

```go
gomock.InOrder(
    repo.EXPECT().BeginTx(gomock.Any()),
    repo.EXPECT().Insert(gomock.Any(), gomock.Any()),
    repo.EXPECT().CommitTx(gomock.Any()),
)
```

## 常见坑

- ❌ 忘 `defer ctrl.Finish()` —— 未满足的期望不会报
- ❌ `EXPECT` 写在 `t.Run` 外 —— 所有子 case 共享，期望会被第一个吃掉
- ❌ `Times(0)` vs 不写 —— Times(0) 明确禁调，不写表示"未设置就 fail"
- ❌ 用 `gomock.Any()` 太激进 —— 等价于"只要被调到就过"，失去断言意义

## Return 的 nil 陷阱

```go
// ❌ 错：nil 的类型推断可能失败
repo.EXPECT().Get(...).Return(nil, nil)

// ✅ 对：显式类型
repo.EXPECT().Get(...).Return((*Order)(nil), error(nil))
```
