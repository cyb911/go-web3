package router

import (
	"go-web3/internal/constants"
	"go-web3/internal/infra/eth"
	"go-web3/internal/infra/redis"
	"go-web3/internal/utils"
	"log"
	"runtime/debug"
	"time"

	"github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {
	r := gin.New()
	r.Use(gin.Logger())
	r.Use(func(c *gin.Context) {
		defer func() {
			if rec := recover(); rec != nil {
				log.Printf("[STACK] %s", debug.Stack())
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

	r.GET("/health/eth", func(c *gin.Context) {
		key, _ := eth.GetPrivateKeyTemp()
		utils.OkData(c, key)
	})

	// 账户模块
	accountGroup := r.Group("/account")
	registerAccountRoutes(accountGroup)

	// 合约交互
	contractGroup := r.Group("/contract/nft/auction")
	registerContractRoutes(contractGroup)

	return r
}
