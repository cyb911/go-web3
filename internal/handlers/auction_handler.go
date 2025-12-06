package handlers

import (
	"fmt"
	"go-web3/contracts/nftauction"
	"go-web3/internal/infra/eth/event"
)

// 拍卖合约事件处理器

func ListenerAuctionCreated(ctx *event.Context) error {
	evt := &nftauction.NftauctionAuctionCreated{}
	err := ctx.BindEvent(evt)
	if err != nil {
		return err
	}

	fmt.Println("---- AuctionCreated ----")
	fmt.Println("AuctionId:", evt.AuctionId.String())
	fmt.Println("Seller:", evt.Seller.Hex())
	fmt.Println("NFT:", evt.Nft.Hex())
	fmt.Println("TokenId:", evt.TokenId.String())
	fmt.Println("MinBid:", evt.MinBid.String())
	return nil
}
