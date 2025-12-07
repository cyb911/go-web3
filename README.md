# GO-WEB3

**作者**: Amor Inksmith
**创建时间**: 2025-12-03
**最后更新**: 2025-12-03
**版本**: v1.0.0

## 📋 项目特性

### 功能清单
- ✅ 账户余额查询
- ✅ 指定区块查询
- ✅ 以太币转账交易
- ✅ 合约交互-拍卖合约结算


### 基础设施建设
- ✅ 本地 NONCE 统一管理
- ✅ 幂等性中间件
- ✅ 链上事件监听（通用型，不与具体的合约，事件耦合。现实可插拔式的链上数据监听）
- ✅ 链上事件周期性扫描

## 🛠 技术栈

- **编程语言**: Go 1.24+
- **Web框架**: Gin
- **文档**: Swagger/OpenAPI 3.0

## 📦 项目结构
```
    ├── cmd
        ├── server                      (命令行启动)
    ├── contract                        (合约绑定代码)
        ├── constants                   (合约地址常量)
        ├── nftauction                  (拍买合约)
    ├── internal                        (本项目内部代码)
        ├── config                      (配置包)
        ├── constants                   (常量)
        ├── handlers                    (处理器层)
        ├── infra                       (基础设施-不关心业务逻辑)
            ├── eth                     (ethclient)
                ├── event               (链上数据处理)
                    ├── abi_registry.go (ABI注册)
                    ├── context.go      (事件上下文)
                    ├── handler.go      (事件处理器接口)
                    ├── middleware.go   (中间件)
                    ├── route.go        (路由)
                    ├── router.go       (路由执行)
            ├── redis                   (Redis)
        ├── middleware                  (gin 中间件)
            ├── idempotency.go          (幂等性)                                                                                                                             
        ├── router                      (路由层)
        ├── service                     (业务逻辑层)                                                   
        └── utils                       (工具包)                                   

```

## 🚀 **快速开始**

### 环境要求

- Go 1.24 或更高版本
- Git
- Gin