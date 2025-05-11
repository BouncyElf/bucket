package bucket

import (
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func BenchmarkRateLimiter(b *testing.B) {
	conf := NewDefaultConfig()
	conf.TokenNumber = 10000
	router := gin.New()
	router.Use(BucketHandler(conf))

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/", nil)
			req.RemoteAddr = "192.168.1.1:8080" // 固定IP
			router.ServeHTTP(w, req)
		}
	})
}

func TestAtomicBucket(t *testing.T) {
	conf := NewDefaultConfig()
	key := "test_ip"

	bucket := conf.Storage.GetOrCreate(key, func() *AtomicBucket {
		return &AtomicBucket{
			token:     conf.TokenNumber,
			updatedAt: time.Now().UnixNano(),
		}
	})

	t.Run("normal consume", func(t *testing.T) {

		// 首次消费应成功
		if atomic.AddInt64(&bucket.token, -1) < 0 {
			t.Fatal("正常消费失败")
		}
	})

	t.Run("rate limiting", func(t *testing.T) {
		// 耗尽令牌
		for i := int64(0); i < conf.TokenNumber; i++ {
			atomic.AddInt64(&bucket.token, -1)
		}

		// 应触发限流
		if atomic.AddInt64(&bucket.token, -1) >= 0 {
			t.Fatal("限流未生效")
		}
	})
}

func TestHTTPMiddleware(t *testing.T) {
	conf := NewDefaultConfig()
	conf.TokenNumber = 10
	// 10s
	conf.RefillMicrosecond = 10000000
	router := gin.New()
	router.GET("/api", BucketHandler(conf), func(c *gin.Context) {
		c.String(200, "ok")
	})

	t.Run("single IP", func(t *testing.T) {
		// 模拟合法请求
		for i := 0; int64(i) < conf.TokenNumber; i++ {
			w := performRequest(router, "GET", "/api", "192.168.1.1")
			assert.Equal(t, 200, w.Code)
		}

		// 触发限流
		w := performRequest(router, "GET", "/api", "192.168.1.1")
		assert.Equal(t, 429, w.Code)
	})
}

func performRequest(r http.Handler, method, path, ip string) *httptest.ResponseRecorder {
	req, _ := http.NewRequest(method, path, nil)
	req.RemoteAddr = ip + ":8080"
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}
