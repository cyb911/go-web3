package handlers

import (
	"go-web3/internal/constants"
	"go-web3/internal/services"
	"go-web3/internal/utils"
	"math/big"

	"github.com/gin-gonic/gin"
)

// SettleAuction 拍卖结算
func SettleAuction(c *gin.Context) {
	auctionIdStr := c.Param("auctionId")
	if auctionIdStr == "" {
		utils.FailMsg(c, constants.ParamError, "auctionId is required")
		return
	}
	auctionId := new(big.Int)
	auctionId, ok := auctionId.SetString(auctionIdStr, 10)
	if !ok {
		utils.FailMsg(c, constants.ParamError, "invalid auctionId")
		return
	}
	err := services.SettleAuction(auctionId)
	if err != nil {
		utils.FailMsg(c, constants.ContractError, err.Error())
		return
	}

	utils.Ok(c)
}

func CancelAuction(c *gin.Context) {
	auctionIdStr := c.Param("auctionId")
	if auctionIdStr == "" {
		utils.FailMsg(c, constants.ParamError, "auctionId is required")
		return
	}
	auctionId := new(big.Int)
	auctionId, ok := auctionId.SetString(auctionIdStr, 10)
	if !ok {
		utils.FailMsg(c, constants.ParamError, "invalid auctionId")
		return
	}
	err := services.CancelAuction(auctionId)
	if err != nil {
		utils.FailMsg(c, constants.ContractError, err.Error())
		return
	}

	utils.Ok(c)
}
