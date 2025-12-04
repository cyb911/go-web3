package middleware

import (
	"bytes"
	"context"
	"encoding/json"
	"go-web3/internal/constants"
	"go-web3/internal/infra/redis"
	"go-web3/internal/utils"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

/*
responseRecorder 响应捕获器
-------------------------------------
Gin 在执行 Handler 时会向客户端写响应（Write/WriteString），
但我们需要在中间件中捕获业务返回的内容，以便：

1. 将第一次返回的响应缓存到 Redis
2. 后续重复请求直接复用缓存的响应（不用再次执行 Handler）

因此我们用一个包装器（Decorator），拦截所有写出的内容。
*/
type responseRecorder struct {
	gin.ResponseWriter               // 原始响应 writer
	body               *bytes.Buffer // 用于记录业务返回的 body
	statusCode         int           // 记录业务返回的 HTTP 状态码
}

// 拦截 Write（二进制写出）
func (r *responseRecorder) Write(b []byte) (int, error) {
	r.body.Write(b)                  // 把写出的内容复制到 buffer
	return r.ResponseWriter.Write(b) // 继续写给客户端
}

// WriteString 拦截 WriteString（字符串写出）
func (r *responseRecorder) WriteString(s string) (int, error) {
	r.body.WriteString(s) // 保存内容
	return r.ResponseWriter.WriteString(s)
}

// WriteHeader 拦截 WriteHeader（设置状态码）
func (r *responseRecorder) WriteHeader(code int) {
	r.statusCode = code
	r.ResponseWriter.WriteHeader(code)
}

/*
Idempotency 幂等中间件
--------------------------------------
功能：
 1. 任何需要“保证只执行一次”的接口必须携带 X-Idempotency-Key
 2. 第一次请求时执行 handler，并将完整响应缓存到 Redis
 3. 重复请求（相同的幂等 key）直接返回缓存的响应
*/
func Idempotency() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := context.Background()
		key := c.Request.Header.Get("X-Idempotency-Key")

		if key == "" {
			utils.FailMsg(c, constants.MissingFieldError, "X-Idempotency-Key is required")
			c.Abort()
			return
		}

		redisKey := "idem:" + key
		// 重复请求，已经有执行结果，则返回上次请求的结果
		if raw, err := redis.Rdb.Get(ctx, redisKey).Bytes(); err == nil {
			var cached struct {
				Status int               `json:"status"` // HTTP 状态码
				Header map[string]string `json:"header"` // Header 信息
				Body   string            `json:"body"`   // 原始 JSON Body
			}
			if json.Unmarshal(raw, &cached) == nil {
				for k, v := range cached.Header {
					c.Writer.Header().Set(k, v)
				}

				// 写入状态码
				c.Status(cached.Status)

				// 写入缓存的响应 Body
				_, _ = c.Writer.WriteString(cached.Body)
			}
			c.Abort()
			return
		}

		// 缓存 handlers 响应数据
		rec := &responseRecorder{
			ResponseWriter: c.Writer,
			body:           &bytes.Buffer{},
			statusCode:     http.StatusOK,
		}

		c.Writer = rec // 替换 gin writer，实现拦截
		// 执行业务 handler（例如转账）
		c.Next()

		resp := struct {
			Status int               `json:"status"`
			Header map[string]string `json:"header"`
			Body   string            `json:"body"`
		}{
			Status: rec.statusCode,
			Header: map[string]string{},
			Body:   rec.body.String(),
		}

		// 只记录必要的 Header
		for k, vals := range rec.Header() {
			if len(vals) > 0 {
				resp.Header[k] = vals[0]
			}
		}

		// 序列化缓存
		buf, _ := json.Marshal(resp)

		redis.Rdb.Set(ctx, redisKey, buf, 10*time.Minute)

	}
}
