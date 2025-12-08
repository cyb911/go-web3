package eth

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum/ethclient"
)

// SuggestEIP1559Fee —— EIP-1559 动态 gas
func SuggestEIP1559Fee(ctx context.Context, client *ethclient.Client) (*big.Int, *big.Int, error) {
	tip, err := client.SuggestGasTipCap(ctx)
	if err != nil {
		return nil, nil, err
	}

	block, err := client.BlockByNumber(ctx, nil)
	if err != nil {
		return nil, nil, err
	}

	baseFee := block.BaseFee()

	// MaxFee = BaseFee + 2 * Tip
	maxFee := new(big.Int).Add(baseFee, new(big.Int).Mul(tip, big.NewInt(2)))

	return tip, maxFee, nil
}
