package event

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

var (
	routeTable = map[common.Address]map[common.Hash]*Route{}
	routeMu    sync.RWMutex
)

type Router struct {
	Client      *ethclient.Client
	Middlewares []Middleware
	Routes      []*Route
	Logger      *log.Logger
}

func NewRouter(client *ethclient.Client, logger *log.Logger) *Router {
	return &Router{
		Client: client,
		Logger: logger,
	}
}

func (r *Route) BuildHandler() EventHandler {
	// Step1: 至少要有一个 handler
	if len(r.handlers) == 0 {
		panic("no handler registered for route " + r.Event)
	}

	// 如果多个 handler，合并成一个顺序执行的 handler
	var final = r.handlers[0]
	for i := 1; i < len(r.handlers); i++ {
		h := r.handlers[i]
		prev := final
		final = EventHandlerFunc(func(ctx *Context) error {
			if err := prev.OnEvent(ctx); err != nil {
				return err
			}
			return h.OnEvent(ctx)
		})
	}

	// 应用 Route 局部中间件
	for i := len(r.middlewares) - 1; i >= 0; i-- {
		final = r.middlewares[i](final)
	}
	return final
}

func (r *Router) Use(m ...Middleware) {
	r.Middlewares = append(r.Middlewares, m...)
}

func (r *Router) Event(contract string, event string) *Route {
	abiInfo, err := GetABIByContract(contract)
	if err != nil {
		panic("ABI not registered: " + contract)
	}
	AddWatchedAddress(abiInfo.Address)
	rt := &Route{
		Contract: contract,
		Event:    event,
	}
	r.Routes = append(r.Routes, rt)

	// 找到事件签名 topic0（Keccak256）
	ev, ok := abiInfo.ABI.Events[event]
	if !ok {
		panic("event " + event + " not found in ABI")
	}

	// 注册到全局路由表，让扫描器能够找到
	registerRoute(abiInfo.Address, ev.ID, rt)

	return rt
}

func (r *Router) Listen() {
	for _, rt := range r.Routes {
		go r.listenRoute(rt)
	}

	select {} // 阻塞主线程
}

func (r *Router) listenRoute(rt *Route) {
	defer func() {
		if rec := recover(); rec != nil {
			r.Logger.Printf("[panic recovered in listenRoute] %v", rec)
			// 自动重启事件监听
			go r.listenRoute(rt)
		}
	}()

	abiInfo, err := GetABIByContract(rt.Contract)
	if err != nil {
		panic(err)
	}

	query := buildFilterQuery(abiInfo.Address, rt.Event, abiInfo.ABI)

	logs := make(chan types.Log)
	sub, err := r.Client.SubscribeFilterLogs(context.Background(), query, logs)
	if err != nil {
		panic(err)
	}

	// 构建 handlerChain (route 中间件 + 全局中间件)
	handler := rt.BuildHandler()
	for i := len(r.Middlewares) - 1; i >= 0; i-- {
		handler = r.Middlewares[i](handler)
	}

	for {
		select {
		case logData := <-logs:
			ctx := &Context{
				Ctx:          context.Background(),
				Log:          logData,
				Client:       r.Client,
				ContractName: rt.Contract,
				EventName:    rt.Event,
				Logger:       r.Logger,
				ABIInfo:      abiInfo,
				ABIEventUnpack: func(out interface{}, log types.Log) error {
					return abiInfo.ABI.UnpackIntoInterface(out, rt.Event, log.Data)
				},
			}

			go handler.OnEvent(ctx)

		case err := <-sub.Err():
			r.Logger.Println("订阅错误:", err)
			time.Sleep(2 * time.Second)
			go r.listenRoute(rt)
			return
		}
	}
}

func registerRoute(addr common.Address, topic common.Hash, rt *Route) {
	routeMu.Lock()
	defer routeMu.Unlock()

	if routeTable[addr] == nil {
		routeTable[addr] = map[common.Hash]*Route{}
	}
	routeTable[addr][topic] = rt
}

func FindRouteByAddressAndTopic(addr common.Address, topic common.Hash) *Route {
	routeMu.RLock()
	defer routeMu.RUnlock()

	if m, ok := routeTable[addr]; ok {
		if rt, ok := m[topic]; ok {
			return rt
		}
	}
	return nil
}
