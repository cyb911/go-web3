package services

import (
	"context"
	"crypto/ecdsa"
	"errors"
	"fmt"
	"go-web3/contracts/constants"
	"go-web3/contracts/nftauction"
	"go-web3/internal/config"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

// 拍卖结算
func SettleAuction(auctionId *big.Int) error {
	ctx := context.Background()
	// 加载私钥
	cfg := config.Get()
	privateKeyHex := cfg.EthPrivate
	privateKey, err := crypto.HexToECDSA(privateKeyHex)
	if err != nil {
		return errors.New("invalid private key")
	}

	// 创建授权对象 auth（用于发送交易）
	chainID, err := config.EthClient.ChainID(ctx)
	if err != nil {
		return err
	}

	auth, err := bind.NewKeyedTransactorWithChainID(privateKey, chainID)
	if err != nil {
		return err
	}

	auth.GasLimit = 0

	// 设置 nonce
	pubKey := privateKey.Public().(*ecdsa.PublicKey)
	fromAddr := crypto.PubkeyToAddress(*pubKey)
	nonce, err := config.EthClient.PendingNonceAt(ctx, fromAddr)
	if err != nil {
		return err
	}

	auth.Nonce = big.NewInt(int64(nonce))
	auth.Value = big.NewInt(0)

	gasPrice, err := config.EthClient.SuggestGasPrice(ctx)
	if err != nil {
		return err
	}
	auth.GasPrice = gasPrice

	// 绑定已经部署的合约(代理地址)
	address := common.HexToAddress(constants.ADDRESS_NFT_AUCTION)
	instance, err := nftauction.NewNftauctionTransactor(address, config.EthClient)
	if err != nil {
		return err
	}

	// 调用合约结算函数
	tx, err := instance.SettleAuction(auth, auctionId)
	if err != nil {
		return err
	}

	fmt.Println("tx sent:", tx.Hash().Hex())

	return nil

}
