package middlewares

import (
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
)

// =============================
//  JWT 配置与结构（支持环境变量，缺省允许默认值）
// =============================

type jwtConfig struct {
	Secret                []byte
	ExpireHours           int
	Issuer                string
	RefreshThresholdHours int
}

func loadJWTConfig() jwtConfig {
	// 1) Secret（允许默认值：仅用于开发环境）
	secret := strings.TrimSpace(os.Getenv("JWT_SECRET"))
	if secret == "" {
		secret = "ChangeThisToYourOwnSecret"
	}

	// 2) ExpireHours（默认 24）
	expireHours := getenvInt("JWT_EXPIRE_HOURS", 24)

	// 3) Issuer（默认 gin-backend）
	issuer := strings.TrimSpace(os.Getenv("JWT_ISSUER"))
	if issuer == "" {
		issuer = "gin-backend"
	}

	// 4) Refresh threshold（默认 6）
	refreshThresholdHours := getenvInt("JWT_REFRESH_THRESHOLD_HOURS", 6)

	// 一些轻量校验，避免写 0 或负数导致奇怪行为
	if expireHours <= 0 {
		log.Printf("⚠️ JWT_EXPIRE_HOURS=%d 非法，将使用默认值 24", expireHours)
		expireHours = 24
	}
	if refreshThresholdHours < 0 {
		log.Printf("⚠️ JWT_REFRESH_THRESHOLD_HOURS=%d 非法，将使用默认值 6", refreshThresholdHours)
		refreshThresholdHours = 6
	}

	return jwtConfig{
		Secret:                []byte(secret),
		ExpireHours:           expireHours,
		Issuer:                issuer,
		RefreshThresholdHours: refreshThresholdHours,
	}
}

// 全局配置（一次加载）
var cfg = loadJWTConfig()

// Claims 定义 JWT Payload
type Claims struct {
	Username string `json:"username"`
	EmpID    string `json:"emp_id"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

// =============================
//  生成 Token
// =============================

func GenerateToken(username, empID, role string) (string, error) {
	now := time.Now()
	exp := now.Add(time.Duration(cfg.ExpireHours) * time.Hour)

	claims := Claims{
		Username: username,
		EmpID:    empID,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(exp),
			IssuedAt:  jwt.NewNumericDate(now),
			Issuer:    cfg.Issuer,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(cfg.Secret)
}

// =============================
//  解析 Token
// =============================

func ParseToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return cfg.Secret, nil
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
		if strings.TrimSpace(authHeader) == "" {
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

		// 自动续签：如果剩余时间小于阈值，则返回新 token
		// 注意：claims.ExpiresAt 可能为空（理论上不会）
		if claims.ExpiresAt != nil {
			threshold := time.Duration(cfg.RefreshThresholdHours) * time.Hour
			if time.Until(claims.ExpiresAt.Time) < threshold {
				newToken, err := GenerateToken(claims.Username, claims.EmpID, claims.Role)
				if err != nil {
					log.Printf("自动续签 Token 失败: %v", err)
				} else {
					c.Header("X-Refresh-Token", newToken)
				}
			}
		}

		// 写入上下文，供后续 handler 使用
		c.Set("username", claims.Username)
		c.Set("emp_id", claims.EmpID)
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

// =============================
//  环境变量加载校验
// =============================

func getenvInt(key string, def int) int {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return def
	}
	i, err := strconv.Atoi(v)
	if err != nil {
		log.Printf("⚠️ 环境变量 %s=%q 不是合法 int，将使用默认值 %d", key, v, def)
		return def
	}
	return i
}
