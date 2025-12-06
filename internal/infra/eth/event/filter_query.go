package event

import (
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
)

func buildFilterQuery(contract common.Address, event string, a abi.ABI) ethereum.FilterQuery {
	id := a.Events[event].ID
	return ethereum.FilterQuery{
		Addresses: []common.Address{contract},
		Topics:    [][]common.Hash{{id}},
	}
}
