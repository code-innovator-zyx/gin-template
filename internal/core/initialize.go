package core

import (
	"github.com/code-innovator-zyx/gin-template/internal/config"
	"github.com/code-innovator-zyx/gin-template/pkg/logger"
	"github.com/code-innovator-zyx/gin-template/pkg/orm"
	"log"
)

/*
* @Author: zouyx
* @Email: 1003941268@qq.com
* @Date:   2025/10/30 下午4:19
* @Package:
 */

func Init() {
	// 初始化配置
	var err error
	Config, err = config.Init()
	if err != nil {
		log.Fatal(err)
	}
	// 初始化日志
	logger.Init(Config.Logger)
	// 初始化数据库 .......
	if Config.Database != nil {
		db, err = orm.Init(*Config.Database)
		if err != nil {
			log.Fatal(err)
		}
	}
}
