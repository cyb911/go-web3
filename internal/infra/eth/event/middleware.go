package event

// Middleware 事件监听中间件
type Middleware func(EventHandler) EventHandler

// Recover 捕获 panic
func Recover() Middleware {
	return func(next EventHandler) EventHandler {
		return EventHandlerFunc(func(ctx *Context) error {
			defer func() {
				if r := recover(); r != nil {
					ctx.Logger.Printf("panic recovered: %v", r)
				}
			}()
			return next.OnEvent(ctx)
		})
	}
}

// Logger 打印事件日志
func Logger() Middleware {
	return func(next EventHandler) EventHandler {
		return EventHandlerFunc(func(ctx *Context) error {
			ctx.Logger.Printf("Event received: Contract=%s Event=%s Tx=%s",
				ctx.ContractName,
				ctx.EventName,
				ctx.Log.TxHash.Hex(),
			)
			return next.OnEvent(ctx)
		})
	}
}
