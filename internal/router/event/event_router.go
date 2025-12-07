package router_event

import (
	"go-web3/contracts/constants"
	"go-web3/contracts/nftauction"
	"go-web3/internal/handlers"
	"go-web3/internal/infra/eth"
	"go-web3/internal/infra/eth/event"
	"log"
	"os"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
)

func SetupRouter() *event.Router {
	logger := log.New(os.Stdout, "[eth-event] ", log.LstdFlags)
	eventRouter := event.NewRouter(eth.EthWssClient, logger)
	parsedABI, _ := abi.JSON(strings.NewReader(nftauction.NftauctionMetaData.ABI))
	event.RegisterABI("NftAuctionV1", parsedABI, constants.ADDRESS_NFT_AUCTION)
	eventRouter.Use(event.Recover(), event.Logger())
	eventRouter.Event("NftAuctionV1", "AuctionCreated").
		Use(handlers.ListenerAuctionCreated)
	return eventRouter
}
