package main

import (
	"go-web3/internal/config"
	"go-web3/internal/infra/eth"
	"go-web3/internal/infra/redis"
	"go-web3/internal/router"
	ethevent "go-web3/internal/router/event"
	"log"
)

func main() {
	var err error
	cfg := config.Get()

	// redis 初始化
	redis.InitRedis()

	// 初始化 ETH client
	eth.InitEthClient()

	eth.InitNonce(redis.Rdb)

	// ETH 事件处理器
	eventRouter := ethevent.SetupRouter()
	// 异步执行，不要阻塞main导致gin无法启动
	go eventRouter.Listen()
	// --- 启动 Scanner
	//scanner := ethevent.SetupScanner()
	//go scanner.Start()

	// 设置路由
	r := router.SetupRouter()

	log.Printf("Server listening on :%s", cfg.AppPort())

	err = r.Run(":" + cfg.AppPort())
	if err != nil {
		panic(err)
	}
}
