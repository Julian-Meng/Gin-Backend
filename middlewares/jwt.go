package middlewares

import (
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
)

// =============================
//  JWT 配置与结构
// =============================

// 优先使用环境变量（默认值仅供开发环境测试）
var jwtSecret = []byte(getSecret())

func getSecret() string {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		// 开发环境默认值
		secret = "ChangeThisToYourOwnSecret"
	}
	return secret
}

// Claims 定义 JWT Payload
type Claims struct {
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

// =============================
//  生成 Token
// =============================

func GenerateToken(username, role string) (string, error) {
	claims := Claims{
		Username: username,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)), // 默认 24 小时有效
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "gin-backend",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

// =============================
//  解析 Token
// =============================

func ParseToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})
	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}
	return nil, err
}

// =============================
//  登录验证中间件
// =============================

func JWTAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code": 1,
				"msg":  "缺少 Authorization Header",
			})
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code": 1,
				"msg":  "Token 格式错误，应为 'Bearer {token}'",
			})
			c.Abort()
			return
		}

		claims, err := ParseToken(parts[1])
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code": 1,
				"msg":  "Token 无效或已过期，请重新登录",
			})
			c.Abort()
			return
		}

		// 角色合法性验证
		if claims.Role != "admin" && claims.Role != "staff" && claims.Role != "superadmin" {
			c.JSON(http.StatusForbidden, gin.H{
				"code": 1,
				"msg":  "无效角色访问",
			})
			c.Abort()
			return
		}

		// 自动续签：如果剩余时间小于 6 小时，则返回新 token
		if time.Until(claims.ExpiresAt.Time) < 6*time.Hour {
			newToken, _ := GenerateToken(claims.Username, claims.Role)
			c.Header("X-Refresh-Token", newToken)
		}

		// 写入上下文，供后续 handler 使用
		c.Set("username", claims.Username)
		c.Set("role", claims.Role)
		c.Next()
	}
}

// =============================
//  管理员专用中间件
// =============================

func AdminOnly() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("role")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code": 1,
				"msg":  "缺少用户身份信息",
			})
			c.Abort()
			return
		}

		if role == "admin" || role == "superadmin" {
			c.Next()
			return
		}

		c.JSON(http.StatusForbidden, gin.H{
			"code": 1,
			"msg":  "权限不足，仅管理员可访问",
		})
		c.Abort()
	}
}
