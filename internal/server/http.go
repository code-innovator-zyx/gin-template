package server

import (
	"context"
	"errors"
	"fmt"
	"gin-admin/internal/handler"
	"gin-admin/internal/services"
	"github.com/sirupsen/logrus"
	"net/http"
	"os"
	"os/signal"
	"time"
)

/*
* @Author: zouyx
* @Email: 1003941268@qq.com
* @Date:   2025 2025/12/3 下午4:20
* @Package:
 */

func StartHttpServer(ctx *services.ServiceContext) {
	// 初始化路由
	svr := &http.Server{
		Addr:         fmt.Sprintf(":%d", ctx.Config.Server.Port),
		Handler:      handler.Init(ctx),
		ReadTimeout:  ctx.Config.Server.ReadTimeout,
		WriteTimeout: ctx.Config.Server.WriteTimeout,
		IdleTimeout:  ctx.Config.Server.IdleTimeout,
	}

	go func() {
		logrus.Infof("服务器启动成功，监听端口: %d", ctx.Config.Server.Port)
		if err := svr.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logrus.Fatalf("启动服务器失败: %v", err)
		}
	}()
	// 捕获系统信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	sign := <-quit
	logrus.Infof("接收到系统信号: %v, 正在关闭服务器...", sign)
	timeoutCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	svr.SetKeepAlivesEnabled(false) // 拒绝新的请求连接
	if err := svr.Shutdown(timeoutCtx); err != nil {
		logrus.Errorf("服务器关闭异常: %v", err)
		logrus.Info("强制关闭服务器")
		if err := svr.Close(); err != nil {
			logrus.Fatalf("强制关闭服务器失败: %v", err)
		}
	}

	logrus.Info("服务器已成功关闭")
}
