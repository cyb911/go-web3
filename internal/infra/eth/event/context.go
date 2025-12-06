package event

import (
	"context"
	"log"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

// Context 事件上下文。为事件处理器提供“事件处理所需全部信息”
type Context struct {
	Ctx    context.Context // 用于取消事件处理、超时控制等。
	Log    types.Log       // go-ethereum 返回的链上原始日志
	Client *ethclient.Client

	ContractName string
	EventName    string

	ABIEventUnpack func(out interface{}, log types.Log) error

	Logger *log.Logger
}

// BindEvent 自动解析事件结构体
func (c *Context) BindEvent(out interface{}) error {
	return c.ABIEventUnpack(out, c.Log)
}
