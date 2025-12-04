package handlers

import (
	"go-web3/internal/constants"
	"go-web3/internal/services/account"
	"go-web3/internal/services/trans"
	"go-web3/internal/utils"

	"github.com/ethereum/go-ethereum/common"
	"github.com/gin-gonic/gin"
)

type TransInfoReq struct {
	To     string `json:"to" binding:"required,min=1,max=64"`
	Amount string `json:"amount" binding:"required"` // ETH
}

func GetBalance(c *gin.Context) {
	address := c.Param("address")
	if address == "" {
		utils.FailMsg(c, constants.ParamError, "address is required")
		return
	}
	result, err := account.GetEthBalance(address)
	if err != nil {
		utils.FailMsg(c, constants.AccountError, err.Error())
		return
	}

	utils.OkData(c, result)
}

func Trans(c *gin.Context) {
	var req TransInfoReq
	if err := c.ShouldBindJSON(&req); err != nil {
		if err.Error() == "EOF" {
			utils.FailMsg(c, constants.ParamError, "请填写参数！")
			return
		}
		utils.FailMsg(c, constants.ParamError, err.Error())
		return
	}
	// 校验地址格式
	if !common.IsHexAddress(req.To) {
		utils.FailMsg(c, constants.ParamError, "无效的账户地址！")
		return
	}
	result, err := trans.Trans(req.To, req.Amount)
	if err != nil {
		utils.FailMsg(c, constants.AccountError, err.Error())
		return
	}

	utils.OkData(c, result)
}

func GetTxReceipt(c *gin.Context) {
	txHash := c.Param("txHash")
	if txHash == "" {
		utils.FailMsg(c, constants.ParamError, "txHash is required")
		return
	}

	result, err := trans.GetTxReceipt(txHash)
	if err != nil {
		utils.FailMsg(c, constants.TransError, err.Error())
		return
	}

	utils.OkData(c, result)
}
