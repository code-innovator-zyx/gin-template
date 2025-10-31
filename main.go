package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/code-innovator-zyx/gin-template/core"
	"github.com/code-innovator-zyx/gin-template/internal/router"
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
	r := router.Init()

	// 配置HTTP服务器
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", core.Config.Server.Port),
		Handler:      r,
		ReadTimeout:  core.Config.Server.ReadTimeout,
		WriteTimeout: core.Config.Server.WriteTimeout,
		IdleTimeout:  core.Config.Server.IdleTimeout,
	}

	// 启动HTTP服务器
	go func() {
		logrus.Infof("服务器启动成功，监听端口: %d", core.Config.Server.Port)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logrus.Fatalf("启动服务器失败: %v", err)
		}
	}()

	// 优雅关闭
	gracefulShutdown(srv)
}

// gracefulShutdown 处理优雅关闭服务器
func gracefulShutdown(srv *http.Server) {
	// 创建一个接收信号的通道
	quit := make(chan os.Signal, 1)
	
	// 监听中断信号和终止信号
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	
	// 阻塞直到接收到信号
	sig := <-quit
	logrus.Infof("接收到系统信号: %v, 正在关闭服务器...", sig)

	// 创建一个带超时的上下文
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	// 尝试优雅关闭服务器
	if err := srv.Shutdown(ctx); err != nil {
		logrus.Errorf("服务器关闭异常: %v", err)
		logrus.Info("强制关闭服务器")
		if err := srv.Close(); err != nil {
			logrus.Fatalf("强制关闭服务器失败: %v", err)
		}
	}

	logrus.Info("服务器已成功关闭")
}
