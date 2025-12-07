package event

import (
	"context"
	"go-web3/internal/infra/eth"
	"log"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

type Scanner struct {
	Client        *ethclient.Client
	BlockStore    BlockStore // 接口：获取、更新 lastProcessedBlock
	DedupeStore   DedupeStore
	Chain         string   // 例如: "sepolia"
	Contracts     []string // 要扫描的合约名
	ReorgDepth    uint64   //Reorg 回滚保护用的值
	Confirmations uint64   //确认区块数
	Logger        *log.Logger
}

func (s *Scanner) Start() {
	ticker := time.NewTicker(2 * time.Second)

	for range ticker.C {
		if err := s.scanOnce(); err != nil {
			s.Logger.Printf("scan error: %v", err)
		}
	}
}

func (s *Scanner) scanOnce() error {
	ctx := context.Background()

	maxRange := uint64(10)

	latest, err := s.Client.BlockNumber(ctx)
	if err != nil {
		return err
	}

	// 扫描到可确认的区块
	if latest <= s.Confirmations {
		return nil
	}
	targetEnd := latest - s.Confirmations

	// 遍历每个合约独立存 lastBlock
	for _, contract := range s.Contracts {
		// 1. 获取上次扫描高度
		last, err := s.BlockStore.GetLastBlock(ctx, s.Chain, contract)
		if err != nil {
			return err
		}

		// 2. 计算 start（含 Reorg 回退）
		start := uint64(0)
		if last > s.ReorgDepth {
			start = last - s.ReorgDepth
		}

		if targetEnd <= start {
			continue
		}

		s.Logger.Printf("[%s] scan blocks %d → %d", contract, start, targetEnd)

		// 3. 扫描事件

		for from := start; from <= targetEnd; from += maxRange {
			to := from + maxRange - 1
			if to > targetEnd {
				to = targetEnd
			}

			s.Logger.Printf("  - batch %d → %d", from, to)

			logs, err := s.fetchLogs(ctx, from, to)
			if err != nil {
				return err
			}

			eth.SortLogs(logs)
			// 4. 处理事件
			for _, lg := range logs {
				route := FindRouteByAddressAndTopic(lg.Address, lg.Topics[0])
				if route == nil || route.Contract != contract {
					continue
				}

				if err := s.handleLog(ctx, lg, route); err != nil {
					s.Logger.Printf("handler error: %v", err)
				}
			}
		}

		// 5. 单独为该合约更新 lastBlock
		if err := s.BlockStore.SetLastBlock(ctx, s.Chain, contract, targetEnd); err != nil {
			return err
		}

	}
	return nil
}

// 扫描区间日志
func (s *Scanner) fetchLogs(ctx context.Context, start, end uint64) ([]types.Log, error) {
	query := ethereum.FilterQuery{
		FromBlock: new(big.Int).SetUint64(start),
		ToBlock:   new(big.Int).SetUint64(end),
	}

	// 添加所有注册过的合约地址
	query.Addresses = AllWatchedAddresses()

	return s.Client.FilterLogs(ctx, query)
}

// 处理事件（含路由 + BindEvent）
func (s *Scanner) handleLog(ctx context.Context, lg types.Log, route *Route) error {
	abiInfo, _ := GetABIByContract(route.Contract)

	c := &Context{
		Ctx:          ctx,
		Client:       s.Client,
		Log:          lg,
		ContractName: route.Contract,
		EventName:    route.Event,
		ABIInfo:      abiInfo,
		ABIEventUnpack: func(out interface{}, log types.Log) error {
			return abiInfo.ABI.UnpackIntoInterface(out, route.Event, log.Data)
		},
		Logger: s.Logger,
	}

	if s.DedupeStore.AlreadyHandled(lg) {
		return nil
	}
	s.DedupeStore.MarkHandled(lg)

	return route.Handler().OnEvent(c)
}
