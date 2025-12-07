package router_event

import (
	"context"
	"go-web3/contracts/constants"
	"go-web3/contracts/nftauction"
	"go-web3/internal/handlers/eth-block"
	"go-web3/internal/infra/eth"
	"go-web3/internal/infra/eth/event"
	"go-web3/internal/infra/redis"
	"log"
	"os"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
)

func SetupRouter() *event.Router {
	logger := log.New(os.Stdout, "[eth-event-listener] ", log.LstdFlags)
	eventRouter := event.NewRouter(eth.EthWssClient, logger)
	parsedABI, _ := abi.JSON(strings.NewReader(nftauction.NftauctionMetaData.ABI))
	event.RegisterABI("NftAuctionV1", parsedABI, constants.ADDRESS_NFT_AUCTION)
	eventRouter.Use(event.Recover(), event.Logger())
	eventRouter.Event("NftAuctionV1", "AuctionCreated").
		Use(eth_block.ListenerAuctionCreated)
	return eventRouter
}

func SetupScanner() *event.Scanner {
	logger := log.New(os.Stdout, "[eth-event-scan] ", log.LstdFlags)
	blockStore := event.NewRedisBlockStore(redis.Rdb, 9787489)
	store := event.NewRedisDedupeStore(context.Background(), redis.Rdb)
	scanner := &event.Scanner{
		Client:        eth.EthClient,
		BlockStore:    blockStore,
		DedupeStore:   store,
		Chain:         "sepolia",
		Contracts:     []string{"NftAuctionV1"},
		ReorgDepth:    6,
		Confirmations: 6,
		Logger:        logger,
	}
	parsedABI, _ := abi.JSON(strings.NewReader(nftauction.NftauctionMetaData.ABI))
	event.RegisterABI("NftAuctionV1", parsedABI, constants.ADDRESS_NFT_AUCTION)
	logger.Println("Starting block scanner...")
	return scanner
}
