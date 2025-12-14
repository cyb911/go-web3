package trans

import (
	"context"
	"crypto/ecdsa"
	"errors"
	"go-web3/internal/config"
	"go-web3/internal/infra/eth"
	"go-web3/internal/infra/eth/nonce"
	"go-web3/internal/utils"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
)

type TxReceiptResp struct {
	TxHash          string       `json:"txHash"`
	BlockHash       string       `json:"blockHash"`
	BlockNumber     string       `json:"blockNumber"`
	From            string       `json:"from"`
	To              string       `json:"to"`     // 交易接受方地址，如果为 nil 表示合约部署交易
	Status          uint64       `json:"status"` // 交易执行状态（成功/失败）PS:失败也会被打包到区块中！只是结果为失败。
	GasUsed         uint64       `json:"gasUsed"`
	ContractAddress string       `json:"contractAddress"` // 新部署的合约地址
	Logs            []*types.Log `json:"logs"`            // 合约 emit 的所有事件
}

func GetTxReceipt(txHash string) (any, error) {
	ctx := context.Background()
	hash := common.HexToHash(txHash)

	receipt, err := eth.EthClient.TransactionReceipt(ctx, hash)
	if err != nil {
		return nil, err
	}

	// 获取 from / to
	tx, _, err := eth.EthClient.TransactionByHash(ctx, hash)
	if err != nil {
		return nil, err
	}

	// 根据chainID 获取 签名者
	signer := types.LatestSignerForChainID(tx.ChainId())
	fromAddr, err := types.Sender(signer, tx)
	if err != nil {
		return nil, err
	}

	from := fromAddr.Hex()

	to := ""
	if tx.To() != nil {
		to = tx.To().Hex()
	}

	resp := &TxReceiptResp{
		TxHash:          receipt.TxHash.Hex(),
		BlockHash:       receipt.BlockHash.Hex(),
		BlockNumber:     receipt.BlockNumber.String(),
		From:            from,
		To:              to,
		Status:          receipt.Status,
		GasUsed:         receipt.GasUsed,
		ContractAddress: receipt.ContractAddress.Hex(),
		Logs:            receipt.Logs,
	}
	return resp, nil

}

func Trans(to string, amountEth string) (string, error) {
	ctx := context.Background()
	// 加载私钥
	cfg := config.Get().EthConfig()
	privateKeyHex := cfg.Private
	privateKey, err := crypto.HexToECDSA(privateKeyHex)
	if err != nil {
		return "", errors.New("invalid private key")
	}
	publicKey := privateKey.Public().(*ecdsa.PublicKey) // 转换成具体的公钥结构体对象
	from := crypto.PubkeyToAddress(*publicKey)

	// nonce
	_nonce, err := eth.NonceMgr.GetNextNonce(ctx, from)
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
		Nonce:     _nonce,
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
		// 判断 nonce 是否与链上数据不一致。不一致强制同步链上 nonce
		if nonce.IsNonceError(err) {
			_ = eth.NonceMgr.ForceSyncNonce(ctx, from)
		}
		return "", err
	}

	return signTx.Hash().Hex(), nil

}
