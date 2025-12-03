package middleware

import (
	"errors"
	"gin-admin/internal/services"
	"gin-admin/pkg/components/jwt"
	"gin-admin/pkg/response"
	"github.com/gin-gonic/gin"
	jwtv4 "github.com/golang-jwt/jwt/v4"
	"github.com/sirupsen/logrus"
	"strings"
)

const (
	ACCESSTOKEN_KEY  = "Authorization"
	REFRESHTOKEN_KEY = "X-Refresh-Token"
)

// JWT 认证中间件
func JWT(svrCtx *services.ServiceContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader(ACCESSTOKEN_KEY)
		if authHeader == "" {
			response.Unauthorized(c, "请先登录")
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if !(len(parts) == 2 && parts[0] == jwt.TokenPrefix) {
			response.Unauthorized(c, "无效的Token格式")
			c.Abort()
			return
		}

		token := parts[1]
		claims, err := svrCtx.Jwt.ParseAccessToken(c.Request.Context(), token)
		if err == nil {
			// 没过期，直接放行
			c.Set("uid", claims.UserID)
			c.Set("username", claims.Username)
			c.Set("sessionId", claims.SessionID)
			c.Next()
			return
		}
		// 若是其他错误（签名错误、格式错误）
		if !errors.Is(err, jwtv4.ErrTokenExpired) {
			logrus.Error("failed to parse jwt token :" + err.Error())
			response.Unauthorized(c, err.Error())
			c.Abort()
			return
		}
		// 超时了，刷新token
		refreshToken, errCookie := c.Cookie(REFRESHTOKEN_KEY)
		if errCookie != nil {
			response.Unauthorized(c, "登录已过期，请重新登录")
			c.Abort()
			return
		}
		tokenPair, errRefresh := svrCtx.Jwt.RefreshToken(c.Request.Context(), refreshToken)
		if errRefresh != nil {
			logrus.Error("failed to refresh jwt token :" + errRefresh.Error())
			response.Unauthorized(c, errRefresh.Error())
			c.Abort()
			return
		}
		c.Header("X-Set-Access-Token", tokenPair.AccessToken)
		c.SetCookie(REFRESHTOKEN_KEY,
			tokenPair.RefreshToken,
			int(svrCtx.Config.Jwt.RefreshTokenExpire.Seconds()),
			"/",
			"",
			false,
			true)
		claims, _ = svrCtx.Jwt.ParseAccessToken(c.Request.Context(), tokenPair.AccessToken)
		c.Set("uid", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("sessionId", claims.SessionID)
		c.Next()
	}
}
