package services

import (
	"context"
	"crypto/ecdsa"
	"errors"
	"go-web3/internal/config"
	"go-web3/internal/utils"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
)

func Trans(to string, amountEth string) (string, error) {
	ctx := context.Background()
	// 加载私钥
	cfg := config.Get()
	privateKeyHex := cfg.EthPrivate
	privateKey, err := crypto.HexToECDSA(privateKeyHex)
	if err != nil {
		return "", errors.New("invalid private key")
	}
	publicKey := privateKey.Public().(*ecdsa.PublicKey) // 转换成具体的公钥结构体对象
	from := crypto.PubkeyToAddress(*publicKey)

	// nonce
	nonce, err := config.EthClient.PendingNonceAt(context.Background(), from)
	if err != nil {
		return "", err
	}

	// 金额转换 ETH → Wei（避免 big.Float）
	amountWei, ok := new(big.Int).SetString(utils.ParseEthToWei(amountEth), 10)
	if !ok {
		return "", errors.New("invalid amount")
	}

	// EIP-1559 推荐 gas 参数
	tipCap, err := config.EthClient.SuggestGasTipCap(ctx) // maxPriorityFeePerGas
	if err != nil {
		return "", err
	}

	// 获取最新区块 → baseFee
	header, err := config.EthClient.HeaderByNumber(ctx, nil)
	if err != nil {
		return "", err
	}
	baseFee := header.BaseFee

	// maxFeePerGas = baseFee + tipCap
	maxFeePerGas := new(big.Int).Add(baseFee, tipCap)

	// GasLimit：普通 ETH 转账 21000
	gasLimit := uint64(21000)

	// 构造交易信息
	toAddress := common.HexToAddress(to)
	chainID, err := config.EthClient.ChainID(ctx)
	if err != nil {
		return "", err
	}
	txData := &types.DynamicFeeTx{
		ChainID:   chainID,
		Nonce:     nonce,
		GasTipCap: tipCap,
		GasFeeCap: maxFeePerGas,
		Gas:       gasLimit,
		To:        &toAddress,
		Value:     amountWei,
		Data:      nil,
	}

	tx := types.NewTx(txData)

	// 自动匹配最新标准签名规则
	signTx, err := types.SignTx(tx, types.LatestSignerForChainID(chainID), privateKey)
	if err != nil {
		return "", err
	}

	// 广播交易
	err = config.EthClient.SendTransaction(ctx, signTx)
	if err != nil {
		return "", err
	}

	return signTx.Hash().Hex(), nil

}
