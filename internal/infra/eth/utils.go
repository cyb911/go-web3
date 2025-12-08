package eth

import (
	"crypto/ecdsa"
	"errors"
	"go-web3/internal/config"
	"sort"
	"strings"
	"sync"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
)

var (
	cachePrivateKey *ecdsa.PrivateKey
	loadOnce        sync.Once
	loadErr         error
)

// GetPrivateKeyTemp 弃用，基础设施模块下的代码，不能依赖外部代码
func GetPrivateKeyTemp() (*ecdsa.PrivateKey, error) {
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

// GetPrivateKey 私钥解析
func GetPrivateKey(privateKey string) (*ecdsa.PrivateKey, error) {
	loadOnce.Do(func() {
		privateKeyHex := strings.TrimSpace(privateKey)
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

func SortLogs(logs []types.Log) {
	sort.Slice(logs, func(i, j int) bool {
		if logs[i].BlockNumber == logs[j].BlockNumber {
			return logs[i].Index < logs[j].Index
		}
		return logs[i].BlockNumber < logs[j].BlockNumber
	})
}
