package event

import (
	"sync"

	"github.com/ethereum/go-ethereum/common"
)

// 使用 map 做 set
var (
	watchedAddrs = map[common.Address]struct{}{}
	mu           sync.RWMutex
)

// AddWatchedAddress 添加一个合约地址到扫描列表
func AddWatchedAddress(addr common.Address) {
	mu.Lock()
	defer mu.Unlock()
	watchedAddrs[addr] = struct{}{}
}

// AllWatchedAddresses 返回所有监控的合约地址列表
func AllWatchedAddresses() []common.Address {
	mu.RLock()
	defer mu.RUnlock()

	addrs := make([]common.Address, 0, len(watchedAddrs))
	for addr := range watchedAddrs {
		addrs = append(addrs, addr)
	}
	return addrs
}
