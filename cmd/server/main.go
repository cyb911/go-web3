package main

import (
	"go-web3/internal/config"
	"go-web3/internal/router"
	"log"
)

func main() {
	var err error
	cfg := config.Get()

	// 初始化 ETH client
	config.InitEthClient()

	// 设置路由
	r := router.SetupRouter()

	log.Printf("Server listening on :%s", cfg.AppPort)

	err = r.Run(":" + cfg.AppPort)
	if err != nil {
		panic(err)
	}
}
