package middleware

import (
	"gin-template/pkg/errcode"
	"gin-template/pkg/response"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// RateLimiter 限流器
type RateLimiter struct {
	rate     int                       // 每秒允许的请求数
	capacity int                       // 桶容量
	buckets  map[string]*TokenBucket   // 每个IP一个令牌桶
	mu       sync.RWMutex
}

// TokenBucket 令牌桶
type TokenBucket struct {
	tokens    int
	capacity  int
	rate      int
	lastTime  time.Time
	mu        sync.Mutex
}

// NewRateLimiter 创建限流器
// rate: 每秒生成的令牌数
// capacity: 桶容量
func NewRateLimiter(rate, capacity int) *RateLimiter {
	return &RateLimiter{
		rate:     rate,
		capacity: capacity,
		buckets:  make(map[string]*TokenBucket),
	}
}

// Allow 检查是否允许请求
func (rl *RateLimiter) Allow(key string) bool {
	rl.mu.RLock()
	bucket, exists := rl.buckets[key]
	rl.mu.RUnlock()

	if !exists {
		rl.mu.Lock()
		// 双重检查
		bucket, exists = rl.buckets[key]
		if !exists {
			bucket = &TokenBucket{
				tokens:   rl.capacity,
				capacity: rl.capacity,
				rate:     rl.rate,
				lastTime: time.Now(),
			}
			rl.buckets[key] = bucket
		}
		rl.mu.Unlock()
	}

	return bucket.Take()
}

// Take 尝试获取一个令牌
func (tb *TokenBucket) Take() bool {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(tb.lastTime).Seconds()
	
	// 根据时间流逝添加令牌
	tokensToAdd := int(elapsed * float64(tb.rate))
	if tokensToAdd > 0 {
		tb.tokens += tokensToAdd
		if tb.tokens > tb.capacity {
			tb.tokens = tb.capacity
		}
		tb.lastTime = now
	}

	// 尝试消费一个令牌
	if tb.tokens > 0 {
		tb.tokens--
		return true
	}

	return false
}

// RateLimit 限流中间件
// 用法: router.Use(middleware.RateLimit(100, 200)) // 每秒100个请求，桶容量200
func RateLimit(rate, capacity int) gin.HandlerFunc {
	limiter := NewRateLimiter(rate, capacity)

	return func(c *gin.Context) {
		// 使用IP作为限流key
		key := c.ClientIP()
		
		if !limiter.Allow(key) {
			response.FailWithStatus(c, 429, errcode.TooManyRequests, "请求过于频繁，请稍后再试")
			c.Abort()
			return
		}

		c.Next()
	}
}

// RateLimitByUser 基于用户ID的限流
// 需要在JWT中间件之后使用
func RateLimitByUser(rate, capacity int) gin.HandlerFunc {
	limiter := NewRateLimiter(rate, capacity)

	return func(c *gin.Context) {
		// 从上下文获取用户ID
		userID, exists := c.Get("user_id")
		if !exists {
			// 如果没有用户ID，使用IP限流
			userID = c.ClientIP()
		}

		key := userID.(string)
		if !limiter.Allow(key) {
			response.FailWithStatus(c, 429, errcode.TooManyRequests, "请求过于频繁，请稍后再试")
			c.Abort()
			return
		}

		c.Next()
	}
}

// CleanupExpired 定期清理过期的限流记录
func (rl *RateLimiter) CleanupExpired(interval time.Duration) {
	ticker := time.NewTicker(interval)
	go func() {
		for range ticker.C {
			rl.mu.Lock()
			now := time.Now()
			for key, bucket := range rl.buckets {
				bucket.mu.Lock()
				// 如果超过1小时未使用，删除该限流记录
				if now.Sub(bucket.lastTime) > time.Hour {
					delete(rl.buckets, key)
				}
				bucket.mu.Unlock()
			}
			rl.mu.Unlock()
		}
	}()
}

