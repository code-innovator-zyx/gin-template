package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestCors(t *testing.T) {
	router := gin.New()
	router.Use(Cors())
	router.GET("/test", func(c *gin.Context) {
		c.String(200, "OK")
	})

	t.Run("Regular Request", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)
		assert.Equal(t, "*", w.Header().Get("Access-Control-Allow-Origin"))
		assert.Equal(t, "GET, POST, PUT, DELETE, OPTIONS", w.Header().Get("Access-Control-Allow-Methods"))
		assert.Equal(t, "true", w.Header().Get("Access-Control-Allow-Credentials"))
	})

	t.Run("OPTIONS Request", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("OPTIONS", "/test", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)
		assert.Equal(t, "*", w.Header().Get("Access-Control-Allow-Origin"))
	})
}

func TestRequestID(t *testing.T) {
	router := gin.New()
	router.Use(RequestID())
	router.GET("/test", func(c *gin.Context) {
		requestID := c.GetString("request_id")
		c.String(200, requestID)
	})

	t.Run("Generate New Request ID", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)
		requestID := w.Header().Get("X-Request-ID")
		assert.NotEmpty(t, requestID, "Request ID should be generated")
		assert.Len(t, requestID, 36, "UUID should be 36 characters")
	})

	t.Run("Use Existing Request ID", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		existingID := "existing-request-id-123"
		req.Header.Set("X-Request-ID", existingID)
		router.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)
		requestID := w.Header().Get("X-Request-ID")
		assert.Equal(t, existingID, requestID, "Should use existing request ID")
	})
}

func TestRecovery(t *testing.T) {
	router := gin.New()
	router.Use(Recovery())
	router.GET("/panic", func(c *gin.Context) {
		panic("test panic")
	})
	router.GET("/normal", func(c *gin.Context) {
		c.String(200, "OK")
	})

	t.Run("Recover From Panic", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/panic", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, 500, w.Code)
		assert.Contains(t, w.Body.String(), "服务器内部错误")
	})


	t.Run("Normal Request", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/normal", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)
		assert.Equal(t, "OK", w.Body.String())
	})
}

func TestLogger(t *testing.T) {
	router := gin.New()
	router.Use(Logger())
	router.GET("/test", func(c *gin.Context) {
		c.String(200, "OK")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	assert.Equal(t, "OK", w.Body.String())
}
