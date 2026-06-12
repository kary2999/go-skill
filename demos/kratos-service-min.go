// Demo: 最小 Kratos 服务骨架 — cmd/main.go
// 使用 mask-go-common-lib 统一初始化 logging / tracing / alarm / config
package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/log"

	"github.com/company/mask-go-common-lib/alarm"
	"github.com/company/mask-go-common-lib/config"
	"github.com/company/mask-go-common-lib/logging"
	"github.com/company/mask-go-common-lib/tracing"
)

// Name 是服务在服务发现 / 日志 / 追踪中的唯一标识。
// 命名约定：kebab-case，语义清晰，避免 util/common。
const Name = "order-service"

func main() {
	// 1. 配置（ConfigMap 友好 + 热加载）
	cfg, err := config.NewFromFile("configs/dev/" + Name + ".yaml")
	if err != nil {
		panic(err) // init 阶段 panic 允许；业务代码禁用
	}

	// 2. 结构化日志（自动提取 trace_id / span_id）
	logger := logging.New(Name, cfg.Env)
	log.SetLogger(logger) // Kratos 全局适配

	// 3. OTel tracing（dev 全量 / prod 10% 采样）
	shutdown, err := tracing.Init(context.Background(), Name, cfg.Env)
	if err != nil {
		panic(err)
	}
	defer shutdown(context.Background())

	// 4. 告警管理（异步去重 + 多渠道）
	alarmMgr := alarm.NewManager(cfg.Alarm)
	defer alarmMgr.Close()

	// 5. Wire 装配的 App（服务 / 中间件 / 路由在 wire_gen.go 中自动拼装）
	app, cleanup, err := wireApp(cfg, logger, alarmMgr)
	if err != nil {
		panic(err)
	}
	defer cleanup()

	// 6. 优雅退出
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	if err := app.Run(); err != nil {
		log.Errorw("msg", "app run failed", "err", err)
		os.Exit(1)
	}
	<-ctx.Done()
}

// buildApp 是 Kratos 应用工厂，wire.go 里会把它声明为 Provider。
func buildApp(logger log.Logger, hs *httpServer, gs *grpcServer) *kratos.App {
	return kratos.New(
		kratos.Name(Name),
		kratos.Logger(logger),
		kratos.Server(hs, gs),
	)
}

// httpServer / grpcServer 的具体初始化在 internal/server/ 下，
// 使用 middleware 包挂载 recovery / tracing / metrics / validate 中间件。
type httpServer = kratos.Server
type grpcServer = kratos.Server
