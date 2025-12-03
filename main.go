package main

import (
	"gin-admin/internal/config"
	"gin-admin/internal/migrates"
	"gin-admin/internal/server"
	"gin-admin/internal/services"
	"gin-admin/pkg/components/logger"
)

func main() {
	// 初始化配置文件
	cfg, err := config.Init()
	if err != nil {
		panic(err)
	}
	// 初始化日志
	logger.Init(cfg.Logger)
	// 初始化内部服务
	ctx := services.MustInitServiceContext(cfg)
	// migrate
	err = migrates.Do(ctx)
	if err != nil {
		panic(err)
	}
	server.StartHttpServer(ctx)
}
