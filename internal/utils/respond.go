package utils

import (
	"go-web3/internal/constants"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type Response struct {
	Code      string      `json:"code"`
	Msg       string      `json:"msg"`
	Data      interface{} `json:"data,omitempty"`
	Timestamp int64       `json:"timestamp"`
}

func Ok(c *gin.Context) {
	c.JSON(http.StatusOK, Response{
		Code:      constants.SuccessCode,
		Msg:       "success",
		Timestamp: time.Now().UnixMilli(),
	})
}

func OkData(c *gin.Context, data any) {
	c.JSON(http.StatusOK, Response{
		Code:      constants.SuccessCode,
		Msg:       "success",
		Data:      data,
		Timestamp: time.Now().UnixMilli(),
	})
}

func Fail(c *gin.Context) {
	c.JSON(http.StatusBadRequest, Response{
		Code:      constants.FailCode,
		Msg:       "fail",
		Timestamp: time.Now().UnixMilli(),
	})
}

func FailMsg(c *gin.Context, errCode string, msg string) {
	c.JSON(http.StatusBadRequest, Response{
		Code:      errCode,
		Msg:       msg,
		Timestamp: time.Now().UnixMilli(),
	})
}
