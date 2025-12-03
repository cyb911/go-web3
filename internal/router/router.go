package router

import (
	"go-web3/internal/constants"
	"go-web3/internal/handlers"
	"go-web3/internal/infra/redis"
	"go-web3/internal/utils"
	"log"
	"time"

	"github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {
	r := gin.New()
	r.Use(gin.Logger())
	r.Use(func(c *gin.Context) {
		defer func() {
			if rec := recover(); rec != nil {
				log.Printf("panic: %v", rec)
				utils.FailMsg(c, constants.InternalServerError, "系统内部错误！")
			}
		}()
		c.Next()
	})

	// 心跳检测
	r.GET("/health", func(c *gin.Context) {
		utils.Ok(c)
	})

	r.GET("/health/redis", func(c *gin.Context) {
		err := redis.Rdb.Set(redis.Ctx, "health", "health", 10*time.Minute).Err()
		if err != nil {
			panic(err)
		}
		result, err := redis.Rdb.Get(redis.Ctx, "health").Result()
		if err != nil {
			return
		}
		utils.OkData(c, result)
	})

	accountGroup := r.Group("/account")
	{
		// 获取账户地址余额信息
		accountGroup.GET("/:address", handlers.GetBalance)
		// 转账
		accountGroup.POST("/trans", handlers.Trans)

		// 查询交易收据

		//查询指定区块号的区块信息
		accountGroup.GET("/block/:number", handlers.GetBlockInfo)
	}

	// 合约交互
	contractGroup := r.Group("/contract/nft/auction")
	{
		// 结算拍卖
		contractGroup.GET("/settle/:auctionId", handlers.SettleAuction)
	}

	return r
}
