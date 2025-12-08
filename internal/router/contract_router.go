package router

import (
	"go-web3/internal/handlers"

	"github.com/gin-gonic/gin"
)

func registerContractRoutes(router *gin.RouterGroup) {
	// 结算拍卖
	router.GET("/settle/:auctionId", handlers.SettleAuction)
	// 取消拍卖
	router.GET("/cancel/:auctionId", handlers.CancelAuction)
}
