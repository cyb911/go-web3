package event

// Route 事件路由
type Route struct {
	Contract    string         // 合约
	Event       string         // 事件
	handlers    []EventHandler // 最终执行的 handler 链
	middlewares []Middleware   // 中间件列表
}

func (r *Route) Use(handler interface{}) *Route {
	switch v := handler.(type) {

	case Middleware:
		// 注册局部中间件
		r.middlewares = append(r.middlewares, v)

	case EventHandler:
		// 注册业务处理器
		r.handlers = append(r.handlers, v)

	case func(ctx *Context) error:
		r.handlers = append(r.handlers, EventHandlerFunc(v))

	default:
		panic("invalid route Use(): must be Middleware, EventHandler or func(*Context) error")
	}
	return r
}
