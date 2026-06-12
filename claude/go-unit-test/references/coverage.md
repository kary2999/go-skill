# 覆盖率

## 命令

```bash
# 生成
go test -cover -coverprofile=cover.out ./...

# 看总数
go tool cover -func=cover.out | tail -1

# HTML 详情（浏览器里看）
go tool cover -html=cover.out -o cover.html

# 只看某包
go test -cover ./internal/biz/...
```

## 目标

| 层 | 目标 |
|---|---|
| biz / usecase（核心业务） | ≥ 90% |
| service（HTTP/gRPC handler 层） | ≥ 85% |
| data / repo | ≥ 70%（复杂查询靠集成测试补） |
| pkg / util 基础工具 | ≥ 80% |
| cmd / main.go / wire_gen.go | 不计入 |

## 排除目录/文件

```bash
go test -cover -coverprofile=cover.out \
    -coverpkg=./internal/...,./pkg/... \
    ./...

# 或用 coverprofile 处理脚本剔除
grep -v "wire_gen.go\|_mock.go\|.pb.go" cover.out > cover.filtered.out
```

CI 脚本里推荐：

```bash
go test -race -cover -coverprofile=cover.out -coverpkg=./internal/...,./pkg/... ./...
COV=$(go tool cover -func=cover.out | grep total | awk '{print $3}' | tr -d '%')
echo "coverage: $COV%"
# 低于阈值 fail
awk -v cov="$COV" 'BEGIN{if (cov+0 < 80) exit 1}'
```

## 查未覆盖的行

```bash
go tool cover -html=cover.out
# 红色 = 未覆盖；灰色 = 不可达或声明
```

## 不要追求 100%

以下代码**不值得**测：
- `wire_gen.go` / `.pb.go`（生成代码）
- 简单 getter/setter
- 纯日志打印路径
- `panic("unreachable")` 分支

把精力放在**业务分支**和**错误处理**。

## 新增代码必须带测试

PR 审查原则：**新增代码覆盖率 ≥ 90%**，不靠存量代码拉高均值。
