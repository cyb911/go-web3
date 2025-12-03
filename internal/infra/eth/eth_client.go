package eth

import (
	"go-web3/internal/config"
	"log"

	"github.com/ethereum/go-ethereum/ethclient"
)

var EthClient *ethclient.Client

func InitEthClient() {
	cfg := config.Get()
	client, err := ethclient.Dial(cfg.EthRpcUrl)
	if err != nil {
		log.Fatalf("Failed to connect to Ethereum RPC: %v", err)
	}

	EthClient = client
}
