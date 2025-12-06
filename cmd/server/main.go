package main

import (
	"go-web3/contracts/constants"
	"go-web3/contracts/nftauction"
	"go-web3/internal/config"
	"go-web3/internal/handlers"
	"go-web3/internal/infra/eth"
	"go-web3/internal/infra/eth/event"
	"go-web3/internal/infra/redis"
	"go-web3/internal/router"
	"log"
	"os"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
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
	parsedABI, _ := abi.JSON(strings.NewReader(nftauction.NftauctionMetaData.ABI))
	event.RegisterABI("NftAuctionV1", parsedABI, constants.ADDRESS_NFT_AUCTION)
	logger := log.New(os.Stdout, "[event] ", log.LstdFlags)
	eventRouter := event.NewRouter(eth.EthWssClient, logger)
	eventRouter.Use(event.Recover(), event.Logger())
	eventRouter.Event("NftAuctionV1", "AuctionCreated").
		Use(handlers.ListenerAuctionCreated)

	// 异步执行，不要阻塞main导致gin无法启动
	go eventRouter.Listen()

	// 设置路由
	r := router.SetupRouter()

	log.Printf("Server listening on :%s", cfg.AppPort())

	err = r.Run(":" + cfg.AppPort())
	if err != nil {
		panic(err)
	}
}
