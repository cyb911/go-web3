package event

// EventHandler 事件处理接口
type EventHandler interface {
	OnEvent(ctx *Context) error
}

// EventHandlerFunc 适配函数式处理器(装饰器模式)
type EventHandlerFunc func(ctx *Context) error

// OnEvent 函数类型 EventHandlerFunc 实现接口的 OnEvent，并包装成 EventHandler 接口类型
func (f EventHandlerFunc) OnEvent(ctx *Context) error {
	return f(ctx)
}
