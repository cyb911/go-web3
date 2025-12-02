package handlers

import (
	"go-web3/internal/constants"
	"go-web3/internal/services"
	"go-web3/internal/utils"
	"strconv"

	"github.com/gin-gonic/gin"
)

func GetBlockInfo(c *gin.Context) {
	number := c.Param("number")
	if number == "" {
		utils.FailMsg(c, constants.ParamError, "number is required")
		return
	}

	blockNumber, err := strconv.ParseUint(number, 10, 64)
	if err != nil {
		utils.FailMsg(c, "E00002", "区块号必须是整数")
		return
	}

	result, err := services.GetBlockInfo(blockNumber)
	if err != nil {
		utils.FailMsg(c, constants.AccountError, err.Error())
		return
	}

	utils.OkData(c, result)
}
