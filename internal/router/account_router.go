package router

import (
	"go-web3/internal/handlers"
	"go-web3/internal/handlers/eth-block"
	"go-web3/internal/middleware"

	"github.com/gin-gonic/gin"
)

func registerAccountRoutes(router *gin.RouterGroup) {
	// 获取账户地址余额信息
	router.GET("/balance/:address", handlers.GetBalance)

	// 转账
	router.POST("/trans", middleware.Idempotency(), handlers.Trans)

	// 查询交易收据
	router.GET("/trans/receipt/:txHash", handlers.GetTxReceipt)

	//查询指定区块号的区块信息
	router.GET("/block/:number", eth_block.GetBlockInfo)
}
