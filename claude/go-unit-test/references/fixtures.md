# Fixtures / testdata / golden file

## testdata/ 目录约定

Go 工具链会**自动忽略**叫 `testdata` 的目录（不参与 build，不参与 go list），放测试数据最安全。

```
internal/biz/
├── order.go
├── order_test.go
└── testdata/
    ├── create_order_request.json
    ├── create_order_response.golden.json
    └── malformed_request.json
```

## 读 fixture

```go
func loadJSON(t *testing.T, path string, dst any) {
    t.Helper()
    b, err := os.ReadFile(filepath.Join("testdata", path))
    require.NoError(t, err)
    require.NoError(t, json.Unmarshal(b, dst))
}

var req CreateOrderReq
loadJSON(t, "create_order_request.json", &req)
```

## Golden file 套路

测试输出难以手写（大段 JSON / HTML / SQL）时，用 golden file：

```go
var update = flag.Bool("update", false, "update golden files")

func TestRenderReport(t *testing.T) {
    got := RenderReport(input)
    goldenPath := filepath.Join("testdata", "report.golden.txt")

    if *update {
        require.NoError(t, os.WriteFile(goldenPath, []byte(got), 0644))
        return
    }

    want, err := os.ReadFile(goldenPath)
    require.NoError(t, err)
    assert.Equal(t, string(want), got)
}
```

跑法：
- 常规：`go test ./...`
- 更新 golden：`go test -update ./internal/report/`

## 什么时候用 golden

| 场景 | 用 |
|---|---|
| 短字符串 / 小结构 | 直接写 `want` |
| 大段 JSON / 模板渲染 / SQL builder 输出 | golden |
| 二进制 | golden + sha256 比对 |

## Diff 可读性

输出多行时用 `assert.Equal(t, strings.Split(want, "\n"), strings.Split(got, "\n"))` 让 diff 按行对。

或用 `go-cmp`：
```go
if diff := cmp.Diff(want, got); diff != "" {
    t.Errorf("mismatch (-want +got):\n%s", diff)
}
```
