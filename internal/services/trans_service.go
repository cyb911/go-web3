package services

import (
	"context"
	"crypto/ecdsa"
	"errors"
	"go-web3/internal/config"
	"go-web3/internal/infra/eth"
	"go-web3/internal/utils"
	"math/big"
	"strings"

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
	nonce, err := eth.NonceMgr.GetNextNonce(ctx, from)
	if err != nil {
		return "", err
	}

	// 金额转换 ETH → Wei（避免 big.Float）
	amountWei, ok := new(big.Int).SetString(utils.ParseEthToWei(amountEth), 10)
	if !ok {
		return "", errors.New("invalid amount")
	}

	// EIP-1559 推荐 gas 参数
	tipCap, err := eth.EthClient.SuggestGasTipCap(ctx) // maxPriorityFeePerGas
	if err != nil {
		return "", err
	}

	// 获取最新区块 → baseFee
	header, err := eth.EthClient.HeaderByNumber(ctx, nil)
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
	chainID, err := eth.EthClient.ChainID(ctx)
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
	err = eth.EthClient.SendTransaction(ctx, signTx)
	if err != nil {
		if shouldRollbackNonce(err) {
			_ = eth.NonceMgr.RevertNonce(ctx, from)
		}
		return "", err
	}

	return signTx.Hash().Hex(), nil

}

func shouldRollbackNonce(err error) bool {
	msg := err.Error()

	// 这些说明 tx 已进入 mempool，不需要回滚
	if strings.Contains(msg, "already known") ||
		strings.Contains(msg, "nonce too low") ||
		strings.Contains(msg, "replacement transaction underpriced") ||
		strings.Contains(msg, "known transaction") {
		return false
	}
	return true
}
