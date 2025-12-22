package handlers

import (
	"backend/dao"
	"backend/middlewares"
	"backend/models"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
)

type SuperAdminConfig struct {
	Enabled  bool
	Username string
	Password string
	Role     string
}

var (
	superAdminOnce sync.Once
	superAdminCfg  SuperAdminConfig
	superAdminErr  error
)

// MustInitAuthConfig 应当在 main() 中、godotenv.Load() 之后调用一次。
// 目的：fail-fast 校验必要的 env 配置，避免“运行起来才发现配置缺失”。
func MustInitAuthConfig() {
	_ = mustLoadSuperAdminConfig()
}

func mustLoadSuperAdminConfig() SuperAdminConfig {
	superAdminOnce.Do(func() {
		cfg, err := loadSuperAdminConfig()
		if err != nil {
			superAdminErr = err
			return
		}
		superAdminCfg = cfg
	})

	if superAdminErr != nil {
		log.Fatal("❌ Superadmin权限配置出错: ", superAdminErr)
	}
	return superAdminCfg
}

func loadSuperAdminConfig() (SuperAdminConfig, error) {
	enabledRaw := strings.TrimSpace(os.Getenv("SUPERADMIN_ENABLED"))
	if enabledRaw == "" {
		return SuperAdminConfig{}, fmt.Errorf("缺少所需的env变量: SUPERADMIN_ENABLED (必须是 true/false)")
	}
	enabled, err := strconv.ParseBool(enabledRaw)
	if err != nil {
		return SuperAdminConfig{}, fmt.Errorf("无效的 SUPERADMIN_ENABLED=%q (必须是 true/false)", enabledRaw)
	}

	// 禁用时：不要求其他字段
	if !enabled {
		return SuperAdminConfig{Enabled: false}, nil
	}

	username := strings.TrimSpace(os.Getenv("SUPERADMIN_USERNAME"))
	if username == "" {
		return SuperAdminConfig{}, fmt.Errorf("缺少所需的env变量: SUPERADMIN_USERNAME")
	}

	password := strings.TrimSpace(os.Getenv("SUPERADMIN_PASSWORD"))
	if password == "" {
		return SuperAdminConfig{}, fmt.Errorf("缺少所需的env变量: SUPERADMIN_PASSWORD")
	}

	role := strings.TrimSpace(os.Getenv("SUPERADMIN_ROLE"))
	if role == "" {
		return SuperAdminConfig{}, fmt.Errorf("缺少所需的env变量: SUPERADMIN_ROLE")
	}

	return SuperAdminConfig{
		Enabled:  true,
		Username: username,
		Password: password,
		Role:     role,
	}, nil
}

// POST /api/login
func Login(c *gin.Context) {
	cfg := mustLoadSuperAdminConfig()

	var req models.Account
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":  1,
			"msg":   "请求格式错误",
			"error": err.Error(),
		})
		return
	}

	req.Username = strings.TrimSpace(req.Username)
	req.Password = strings.TrimSpace(req.Password)

	if req.Username == "" || req.Password == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 1,
			"msg":  "用户名和密码不能为空",
		})
		return
	}

	// 1) superadmin 兜底账号（仅启用时）
	if cfg.Enabled && req.Username == cfg.Username {
		if req.Password != cfg.Password {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code": 1,
				"msg":  "superadmin 密码错误",
			})
			return
		}

		token, err := middlewares.GenerateToken(cfg.Username, "", cfg.Role)
		if err != nil {
			log.Println("❌ 生成 superadmin Token 失败:", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"code": 1,
				"msg":  "生成 Token 失败",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"code":     0,
			"msg":      "登录成功",
			"token":    token,
			"username": cfg.Username,
			"role":     cfg.Role,
		})
		return
	}

	// 2) 普通用户登录
	account, ok := dao.ValidateLogin(req.Username, req.Password)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code": 1,
			"msg":  "用户名或密码错误",
		})
		return
	}

	token, err := middlewares.GenerateToken(account.Username, account.EmpID, account.Role)
	if err != nil {
		log.Println("❌ 生成 Token 失败:", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":  1,
			"msg":   "Token 生成失败",
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":     0,
		"msg":      "登录成功",
		"token":    token,
		"username": account.Username,
		"role":     account.Role,
	})
}

// POST /api/register
func Register(c *gin.Context) {
	cfg := mustLoadSuperAdminConfig()

	var acc models.Account
	if err := c.ShouldBindJSON(&acc); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":  1,
			"msg":   "请求格式错误",
			"error": err.Error(),
		})
		return
	}

	acc.Username = strings.TrimSpace(acc.Username)

	if acc.Username == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 1,
			"msg":  "用户名不能为空",
		})
		return
	}

	// 启用 superadmin 时：保留用户名不可注册
	if cfg.Enabled && acc.Username == cfg.Username {
		c.JSON(http.StatusForbidden, gin.H{
			"code": 1,
			"msg":  "不允许注册保留用户名",
		})
		return
	}

	// 检查重名
	if _, exists := dao.GetAccountByUsername(acc.Username); exists {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 1,
			"msg":  "用户名已存在",
		})
		return
	}

	if err := dao.InsertAccount(acc); err != nil {
		log.Println("❌ 注册失败:", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":  1,
			"msg":   "注册失败",
			"error": fmt.Sprintf("%v", err),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"success": true,
		"msg":     "注册成功",
	})
}
