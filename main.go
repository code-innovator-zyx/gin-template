package main

import (
	"context"
	"errors"
	"fmt"
	"gin-template/internal/core"
	"gin-template/internal/handler"
	"github.com/sirupsen/logrus"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// @title 标题
// @version 1.0 版本
// @description 描述
// @termsOfService http://swagger.io/terms/
// @contact.name 联系人
// @contact.url http://www.swagger.io/support
// @contact.email 584807419@qq.com
// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html
// @host videotools.cn
// @BasePath /freeapi/v1
// @query.collection.format multi

func main() {
	// 初始化应用配置和依赖
	core.Init()

	// 初始化路由
	r := handler.Init()

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", core.Config.Server.Port),
		Handler:      r,
		ReadTimeout:  core.Config.Server.ReadTimeout,
		WriteTimeout: core.Config.Server.WriteTimeout,
		IdleTimeout:  core.Config.Server.IdleTimeout,
	}

	go func() {
		logrus.Infof("服务器启动成功，监听端口: %d", core.Config.Server.Port)
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
