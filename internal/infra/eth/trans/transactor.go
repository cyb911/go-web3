package trans

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"go-web3/internal/infra/eth/nonce"
	"math/big"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

// Transactor —— 链上交易发送器
type Transactor struct {
	client     *ethclient.Client
	nonceMgr   *nonce.NonceManager
	privateKey *ecdsa.PrivateKey
	from       common.Address
	chainID    *big.Int
}

func NewTransactor(client *ethclient.Client, nonceMgr *nonce.NonceManager, privateKey *ecdsa.PrivateKey) (*Transactor, error) {
	from := crypto.PubkeyToAddress(privateKey.PublicKey)
	chainID, err := client.ChainID(context.Background())
	if err != nil {
		return nil, err
	}

	return &Transactor{
		client:     client,
		nonceMgr:   nonceMgr,
		privateKey: privateKey,
		from:       from,
		chainID:    chainID,
	}, nil
}

// NewAuth 构建带 EIP-1559 的交易授权对象
func (t *Transactor) NewAuth(ctx context.Context) (*bind.TransactOpts, error) {
	auth, err := bind.NewKeyedTransactorWithChainID(t.privateKey, t.chainID)
	if err != nil {
		return nil, err
	}

	// EIP-1559 gas 设置
	tip, feeCap, err := SuggestEIP1559Fee(ctx, t.client)
	if err != nil {
		return nil, err
	}

	auth.GasTipCap = tip
	auth.GasFeeCap = feeCap
	auth.Value = big.NewInt(0)
	auth.From = t.from

	// nonce 设置
	nonce, err := t.nonceMgr.GetNextNonce(ctx, t.from)
	if err != nil {
		return nil, err
	}
	auth.Nonce = new(big.Int).SetUint64(nonce)

	return auth, nil
}

// EstimateGas —— 根据 dry-run 交易估算 gaslimit
func (t *Transactor) EstimateGas(ctx context.Context, dry *types.Transaction, auth *bind.TransactOpts) (uint64, error) {
	msg := ethereum.CallMsg{
		From:      t.from,
		To:        dry.To(),
		Data:      dry.Data(),
		Value:     auth.Value,
		GasFeeCap: auth.GasFeeCap,
		GasTipCap: auth.GasTipCap,
	}

	return t.client.EstimateGas(ctx, msg)
}

// SendTx —— 交易发送（自动处理 nonce 冲突 + 重试机制）
func (t *Transactor) SendTx(txFunc func(*bind.TransactOpts) (*types.Transaction, error)) (*types.Transaction, error) {
	ctx := context.Background()

retry:
	auth, err := t.NewAuth(ctx)
	if err != nil {
		return nil, err
	}

	// dry-run
	tmp := *auth
	tmp.NoSend = true

	dryTx, err := txFunc(&tmp)
	if err != nil {
		// revert reason
		return nil, err
	}

	// 模拟执行（eth-call）.避免失败扣除gas
	if err := t.SimulateCall(*dryTx.To(), dryTx.Data(), auth.Value); err != nil {
		return nil, err
	}

	// gas 估算
	gas, err := t.EstimateGas(ctx, dryTx, auth)
	if err != nil {
		return nil, err
	}
	auth.GasLimit = gas

	// 正式发送
	tx, err := txFunc(auth)
	if err != nil {
		// 自动处理 nonce 冲突
		if nonce.IsNonceError(err) {
			// 同步链上 nonce
			err := t.nonceMgr.ForceSyncNonce(ctx, t.from)
			if err != nil {
				return nil, err
			}
			time.Sleep(200 * time.Millisecond)
			goto retry
		}
		return nil, err
	}

	return tx, nil
}

func (t *Transactor) SimulateCall(to common.Address, data []byte, value *big.Int) error {
	msg := ethereum.CallMsg{
		From: t.from,
		To:   &to,
		Data: data,
		Value: func() *big.Int {
			if value == nil {
				return big.NewInt(0)
			}
			return value
		}(),
	}

	_, err := t.client.CallContract(context.Background(), msg, nil)
	if err != nil {
		// 解析 revert reason
		return decodeRevertReason(err)
	}
	return nil
}

func decodeRevertReason(err error) error {
	errStr := err.Error()

	// geth/erigon 统一格式：execution reverted: reason
	if strings.Contains(errStr, "execution reverted:") {
		parts := strings.SplitN(errStr, "execution reverted:", 2)
		if len(parts) == 2 {
			return fmt.Errorf(strings.TrimSpace(parts[1]))
		}
	}

	// 没有 reason（fallback）
	return fmt.Errorf("execution reverted")
}
