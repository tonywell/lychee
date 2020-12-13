package main

import (
	"fmt"
	"lychee/config"
	"lychee/logger"
	"lychee/router"
	"lychee/service"
)

func main() {
	config.Init()
	logger.InitLogger()
	//装载路由
	r := router.Routers()
	go service.WsManager.Start()
	r.Run(fmt.Sprintf("%s:%s", config.AllConfig.Server.Address, config.AllConfig.Server.Port))
}
