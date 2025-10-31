package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/code-innovator-zyx/gin-template/core"
	"github.com/code-innovator-zyx/gin-template/router"
	"github.com/sirupsen/logrus"
	"log"
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
	core.Init()
	r := router.Init()

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", core.Config.Server.Port),
		Handler: r,
	}
	// 启动HTTP服务器
	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("启动服务器失败: %v", err)
		}
	}()

	logrus.Infof("服务器启动成功，监听端口: %d", core.Config.Server.Port)

	// 优雅关闭
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logrus.Info("正在关闭服务器...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("服务器关闭异常: %v", err)
	}

	logrus.Info("服务器已成功关闭")
}
