package router

import (
	"go-web3/internal/handlers"

	"github.com/gin-gonic/gin"
)

func registerContractRoutes(router *gin.RouterGroup) {
	// TODO 创建拍卖功能
	// 结算拍卖
	router.GET("/settle/:auctionId", handlers.SettleAuction)
}
