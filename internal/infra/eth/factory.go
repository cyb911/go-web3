package eth

import (
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/redis/go-redis/v9"
)

type EthFactory struct {
	Client       *ethclient.Client
	NonceManager *NonceManager
}

func NewEthFactory(client *ethclient.Client, rdb *redis.Client) *EthFactory {
	// 先判断 nonce 管理器 是否已经构建出
	var nonceManager *NonceManager

	if NonceMgr == nil {
		nonceManager = NewNonceManager(rdb, client)
	} else {
		nonceManager = NonceMgr
	}
	return &EthFactory{
		Client:       client,
		NonceManager: nonceManager,
	}
}

// NewTransactor —— 工厂为指定私钥产生一个“交易器”
func (f *EthFactory) NewTransactor(privateKeyHex string) *Transactor {
	pk, err := crypto.HexToECDSA(privateKeyHex)
	if err != nil {
		panic("invalid private key")
	}
	transactor, err := NewTransactor(f.Client, f.NonceManager, pk)
	if err != nil {
		panic("failed to create transactor")
	}
	return transactor
}
