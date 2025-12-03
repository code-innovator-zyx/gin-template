package main

import (
	"context"
	"errors"
	"fmt"
	"gin-admin/internal/core"
	"gin-admin/internal/handler"
	_ "gin-admin/internal/model"
	"gin-admin/internal/services"
	"github.com/sirupsen/logrus"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	// 初始化应用配置和依赖
	core.Init()
	cfg := core.MustGetConfig()
	// 初始化内部服务
	ctx := services.MustInitServiceContext(cfg)
	// 初始化路由
	r := handler.Init(ctx)

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:      r,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	go func() {
		logrus.Infof("服务器启动成功，监听端口: %d", cfg.Server.Port)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logrus.Fatalf("启动服务器失败: %v", err)
		}
	}()

	// 优雅关闭
	gracefulShutdown(srv)
}

// gracefulShutdown
func gracefulShutdown(srv *http.Server) {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	sig := <-quit
	logrus.Infof("接收到系统信号: %v, 正在关闭服务器...", sig)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		logrus.Errorf("服务器关闭异常: %v", err)
		logrus.Info("强制关闭服务器")
		if err := srv.Close(); err != nil {
			logrus.Fatalf("强制关闭服务器失败: %v", err)
		}
	}

	logrus.Info("服务器已成功关闭")
}
