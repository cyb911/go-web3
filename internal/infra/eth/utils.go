package eth

import (
	"crypto/ecdsa"
	"errors"
	"go-web3/internal/config"
	"strings"
	"sync"

	"github.com/ethereum/go-ethereum/crypto"
)

var (
	cachePrivateKey *ecdsa.PrivateKey
	loadOnce        sync.Once
	loadErr         error
)

func GetPrivateKey() (*ecdsa.PrivateKey, error) {
	loadOnce.Do(func() {
		cfg := config.Get().EthConfig()
		privateKeyHex := cfg.Private
		privateKeyHex = strings.TrimSpace(privateKeyHex)
		if privateKeyHex == "" {
			loadErr = errors.New("missing ETH_PRIVATE in config")
			return
		}

		if len(privateKeyHex) != 64 {
			loadErr = errors.New("invalid private key length: should be 64 hex characters")
			return
		}

		privateKey, err := crypto.HexToECDSA(privateKeyHex)
		if err != nil {
			loadErr = errors.New("failed to parse private key: " + err.Error())
			return
		}
		cachePrivateKey = privateKey
	})

	return cachePrivateKey, loadErr
}
