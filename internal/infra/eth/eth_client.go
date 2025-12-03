package eth

import (
	"go-web3/internal/config"
	"log"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/redis/go-redis/v9"
)

var EthClient *ethclient.Client
var NonceMgr *NonceManager

func InitEthClient() {
	cfg := config.Get()
	client, err := ethclient.Dial(cfg.EthRpcUrl)
	if err != nil {
		log.Fatalf("Failed to connect to Ethereum RPC: %v", err)
	}

	EthClient = client
}

func InitNonce(redis *redis.Client) {
	NonceMgr = NewNonceManager(redis, EthClient)
}
