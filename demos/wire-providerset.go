// Demo: Wire ProviderSet — 编译期依赖注入
// 每层（biz / data / service / server）导出自己的 ProviderSet，
// 在 cmd/wire.go 里聚合。
//
// 典型目录：
//   internal/biz/biz.go        — ProviderSet = wire.NewSet(NewOrderUsecase)
//   internal/data/data.go      — ProviderSet = wire.NewSet(NewDB, NewOrderRepo)
//   internal/service/service.go— ProviderSet = wire.NewSet(NewOrderService)
//   cmd/wire.go                — 聚合 + 生成

// ================ internal/data/data.go ================

package data

import (
	"github.com/google/wire"
	"gorm.io/gorm"

	"github.com/company/mask-go-common-lib/logging"
)

// ProviderSet 是 data 层对外暴露的依赖集合。
var ProviderSet = wire.NewSet(NewDB, NewOrderRepo)

// NewDB 构造 PostgreSQL 连接（生产应从 config 读取 DSN 并设连接池上限）。
func NewDB(cfg *Config, logger *logging.Logger) (*gorm.DB, func(), error) {
	db, err := gorm.Open( /* postgres.Open(cfg.DSN) */ nil)
	if err != nil {
		return nil, nil, err
	}
	cleanup := func() {
		sqlDB, _ := db.DB()
		_ = sqlDB.Close()
	}
	return db, cleanup, nil
}

// ================ cmd/wire.go ================
// //go:build wireinject
// // +build wireinject
//
// package main
//
// import (
// 	"github.com/google/wire"
// 	"yourorg/yoursvc/internal/biz"
// 	"yourorg/yoursvc/internal/data"
// 	"yourorg/yoursvc/internal/server"
// 	"yourorg/yoursvc/internal/service"
// )
//
// func wireApp(cfg *config.Config, logger *logging.Logger, alarm *alarm.Manager) (*kratos.App, func(), error) {
// 	panic(wire.Build(
// 		data.ProviderSet,
// 		biz.ProviderSet,
// 		service.ProviderSet,
// 		server.ProviderSet,
// 		buildApp,
// 	))
// }

// 运行 `wire ./cmd/` 生成 wire_gen.go。
// 永远不要手写 wire_gen.go，也不要把它从 .gitignore 去掉 —— 它就是签过字的依赖图，入库。
