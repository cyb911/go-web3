package services

import (
	"context"
	"errors"
	"go-web3/internal/config"
	"math/big"
)

type BlockInfo struct {
	Number       uint64 `json:"number"`
	Hash         string `json:"hash"`
	ParentHash   string `json:"parentHash"`
	Timestamp    uint64 `json:"timestamp"`
	Transactions int    `json:"transactions"`
	GasLimit     uint64 `json:"gasLimit"`
	GasUsed      uint64 `json:"gasUsed"`
}

func GetBlockInfo(blockNumber uint64) (*BlockInfo, error) {
	ctx := context.Background()
	num := new(big.Int).SetUint64(blockNumber)
	block, err := config.EthClient.BlockByNumber(ctx, num)

	if err != nil {
		return nil, errors.New("failed to get block: " + err.Error())
	}

	info := &BlockInfo{
		Number:       block.NumberU64(),
		Hash:         block.Hash().Hex(),
		ParentHash:   block.ParentHash().Hex(),
		Timestamp:    block.Time(),
		Transactions: len(block.Transactions()),
		GasLimit:     block.GasLimit(),
		GasUsed:      block.GasUsed(),
	}

	return info, nil

}
