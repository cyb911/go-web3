package event

import (
	"errors"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
)

// ABI 元数据管理中心

// ABIInfo 合约的完整描述元数据
type ABIInfo struct {
	ContractName string
	ABI          abi.ABI
	Address      common.Address
}

var ABIRegistry = map[string]*ABIInfo{}

func RegisterABI(name string, a abi.ABI, addr string) {
	ABIRegistry[name] = &ABIInfo{
		ContractName: name,
		ABI:          a,
		Address:      common.HexToAddress(addr),
	}
}

func GetABIByContract(name string) (*ABIInfo, error) {
	v, ok := ABIRegistry[name]
	if !ok {
		return nil, errors.New("ABI not registered: " + name)
	}
	return v, nil
}
