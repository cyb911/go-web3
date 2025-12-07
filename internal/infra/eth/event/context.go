package event

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"reflect"

	"github.com/ethereum/go-ethereum/common"
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

	ABIInfo        *ABIInfo
	ABIEventUnpack func(out interface{}, log types.Log) error

	Logger *log.Logger
}

// BindEvent 自动解析事件结构体
func (c *Context) BindEvent(out interface{}) error {
	if err := c.ABIEventUnpack(out, c.Log); err != nil {
		return err
	}

	// 动解析 indexed 参数（topics 部分）
	eventABI, ok := c.ABIInfo.ABI.Events[c.EventName]
	if !ok {
		return fmt.Errorf("event %s not found in ABI", c.EventName)
	}

	topics := c.Log.Topics
	topicIndex := 1

	v := reflect.ValueOf(out)
	elem := v.Elem()

	for i, arg := range eventABI.Inputs {
		if !arg.Indexed {
			continue // 跳过非 indexed
		}

		if topicIndex >= len(topics) {
			return fmt.Errorf("missing topic for %s", arg.Name)
		}

		field := elem.Field(i)
		//Solidity ABI，indexed 参数只可能是 static types
		switch arg.Type.String() {

		case "uint256":
			field.Set(reflect.ValueOf(new(big.Int).SetBytes(topics[topicIndex].Bytes())))

		case "address":
			field.Set(reflect.ValueOf(common.BytesToAddress(topics[topicIndex].Bytes())))

		//case "uint64":
		//	bi := new(big.Int).SetBytes(topics[topicIndex].Bytes())
		//	field.SetUint(bi.Uint64())

		default:
			return fmt.Errorf("unsupported indexed type %s", arg.Type.String())
		}

		topicIndex++
	}

	return nil
}
