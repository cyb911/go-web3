package eth

import (
	"context"
	"go-web3/internal/config"
	"log"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/redis/go-redis/v9"
)

var EthClient *ethclient.Client
var NonceMgr *NonceManager

func InitEthClient() {
	cfg := config.Get().EthConfig()
	client, err := ethclient.Dial(cfg.RpcUrl)
	if err != nil {
		log.Fatalf("Failed to connect to Ethereum RPC: %v", err)
	}

	EthClient = client
}

func InitNonce(redis *redis.Client) {
	NonceMgr = NewNonceManager(redis, EthClient)

	ctx := context.Background()
	cfg := config.Get().EthConfig()

	privateKey, _ := crypto.HexToECDSA(cfg.Private)
	addr := crypto.PubkeyToAddress(privateKey.PublicKey)

	// 程序启动自动强制同步链上 nonce
	_ = NonceMgr.ForceSyncNonce(ctx, addr)
}
