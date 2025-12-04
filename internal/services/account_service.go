package services

import (
	"context"
	"errors"
	"go-web3/internal/config"
	"go-web3/internal/infra/eth"
	"math"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

type Balance struct {
	Address     string `json:"address"`
	BalanceWei  string `json:"balanceWei"`
	BalanceETH  string `json:"balanceETH"`
	NetworkName string `json:"network"`
}

func GetEthBalance(address string) (*Balance, error) {
	// 校验地址格式
	if !common.IsHexAddress(address) {
		return nil, errors.New("invalid wallet address")
	}

	// 查询余额（单位：Wei）
	account := common.HexToAddress(address)
	balanceWei, err := eth.EthClient.BalanceAt(context.Background(), account, nil)
	if err != nil {
		return nil, err
	}

	// 转换成 ETH（浮点）字符串
	bf := new(big.Float).SetInt(balanceWei)
	ethValue := new(big.Float).Quo(bf, big.NewFloat(math.Pow10(18)))
	ethStr := ethValue.Text('f', 18)

	return &Balance{
		Address:     address,
		BalanceWei:  balanceWei.String(),
		BalanceETH:  ethStr,
		NetworkName: config.Get().EthConfig().NetworkName,
	}, nil
}
