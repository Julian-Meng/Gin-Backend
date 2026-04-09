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
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mojocn/base64Captcha"
)

const (
	defaultLoginCaptchaThreshold = 3
	defaultCaptchaTTLSeconds     = 180
	captchaTextSource            = "23456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz"
)

type loginRequest struct {
	Username    string `json:"username"`
	Password    string `json:"password"`
	CaptchaID   string `json:"captcha_id"`
	CaptchaCode string `json:"captcha_code"`
}

type registerRequest struct {
	Username    string `json:"username"`
	Password    string `json:"password"`
	Role        string `json:"role"`
	CaptchaID   string `json:"captcha_id"`
	CaptchaCode string `json:"captcha_code"`
}

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

	captchaOnce  sync.Once
	captchaSvc   *base64Captcha.Captcha
	captchaTTL   int
	loginCapThld int
)

// MustInitAuthConfig 应当在 main() 中、godotenv.Load() 之后调用一次。
// 目的：fail-fast 校验必要的 env 配置，避免“运行起来才发现配置缺失”。
func MustInitAuthConfig() {
	_ = mustLoadSuperAdminConfig()
	_ = mustLoadCaptchaConfig()
}

func mustLoadCaptchaConfig() *base64Captcha.Captcha {
	captchaOnce.Do(func() {
		captchaTTL = getEnvInt("CAPTCHA_TTL_SECONDS", defaultCaptchaTTLSeconds)
		if captchaTTL <= 0 {
			captchaTTL = defaultCaptchaTTLSeconds
		}

		loginCapThld = getEnvInt("LOGIN_CAPTCHA_THRESHOLD", defaultLoginCaptchaThreshold)
		if loginCapThld <= 0 {
			loginCapThld = defaultLoginCaptchaThreshold
		}

		driver := base64Captcha.NewDriverString(
			60,
			220,
			2,
			base64Captcha.OptionShowHollowLine,
			4,
			captchaTextSource,
			nil,
			nil,
			nil,
		)
		store := base64Captcha.NewMemoryStore(10240, time.Duration(captchaTTL)*time.Second)
		captchaSvc = base64Captcha.NewCaptcha(driver, store)
	})

	return captchaSvc
}

func getEnvInt(key string, defaultVal int) int {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return defaultVal
	}
	i, err := strconv.Atoi(v)
	if err != nil {
		log.Printf("\033[43m环境变量 %s=%q 不是合法 int，将使用默认值 %d\033[0m\n", key, v, defaultVal)
		return defaultVal
	}
	return i
}

func loginFailResponse(c *gin.Context, failCount int, msg string) {
	needCaptcha := failCount >= loginCapThld
	c.JSON(http.StatusUnauthorized, gin.H{
		"code":         1,
		"msg":          msg,
		"need_captcha": needCaptcha,
		"fail_count":   failCount,
	})
}

// GetCaptcha GET /api/captcha
func GetCaptcha(c *gin.Context) {
	service := mustLoadCaptchaConfig()
	scene := strings.ToLower(strings.TrimSpace(c.DefaultQuery("scene", "register")))
	if scene != "register" && scene != "login" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 1,
			"msg":  "无效的 scene，仅支持 register/login",
		})
		return
	}

	id, b64, _, err := service.Generate()
	if err != nil {
		log.Println("\033[31m生成验证码失败:\033[0m", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 1,
			"msg":  "生成验证码失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"msg":  "ok",
		"data": gin.H{
			"scene":          scene,
			"captcha_id":     id,
			"image_base64":   b64,
			"expire_seconds": captchaTTL,
		},
	})
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
		log.Fatal("\033[31mSuperadmin权限配置出错: \033[0m", superAdminErr)
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

// Login POST /api/login
func Login(c *gin.Context) {
	cfg := mustLoadSuperAdminConfig()
	mustLoadCaptchaConfig()

	ip := strings.TrimSpace(c.ClientIP())

	var req loginRequest
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

	needCaptcha := dao.GetLoginFailCount(req.Username, ip) >= loginCapThld
	if needCaptcha {
		req.CaptchaID = strings.TrimSpace(req.CaptchaID)
		req.CaptchaCode = strings.TrimSpace(req.CaptchaCode)
		if req.CaptchaID == "" || req.CaptchaCode == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":         1,
				"msg":          "需要输入验证码",
				"need_captcha": true,
				"fail_count":   dao.GetLoginFailCount(req.Username, ip),
			})
			return
		}

		if !captchaSvc.Verify(req.CaptchaID, req.CaptchaCode, true) {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":         1,
				"msg":          "验证码错误或已过期",
				"need_captcha": true,
				"fail_count":   dao.GetLoginFailCount(req.Username, ip),
			})
			return
		}
	}

	// 1) superadmin 兜底账号（仅启用时）
	if cfg.Enabled && req.Username == cfg.Username {
		if req.Password != cfg.Password {
			failCount := dao.IncrementLoginFailCount(req.Username, ip)
			loginFailResponse(c, failCount, "superadmin 密码错误")
			return
		}

		dao.ClearLoginFailCount(req.Username, ip)

		token, err := middlewares.GenerateToken(cfg.Username, "", cfg.Role)
		if err != nil {
			log.Println("\033[31m生成 superadmin Token 失败:\033[0m", err)
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
		failCount := dao.IncrementLoginFailCount(req.Username, ip)
		loginFailResponse(c, failCount, "用户名或密码错误")
		return
	}

	dao.ClearLoginFailCount(req.Username, ip)

	token, err := middlewares.GenerateToken(account.Username, account.EmpID, account.Role)
	if err != nil {
		log.Println("\033[31m生成 Token 失败:\033[0m", err)
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

// Register POST /api/register
func Register(c *gin.Context) {
	cfg := mustLoadSuperAdminConfig()
	mustLoadCaptchaConfig()

	var req registerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":  1,
			"msg":   "请求格式错误",
			"error": err.Error(),
		})
		return
	}

	req.Username = strings.TrimSpace(req.Username)
	req.CaptchaID = strings.TrimSpace(req.CaptchaID)
	req.CaptchaCode = strings.TrimSpace(req.CaptchaCode)

	if req.Username == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 1,
			"msg":  "用户名不能为空",
		})
		return
	}

	if req.CaptchaID == "" || req.CaptchaCode == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 1,
			"msg":  "注册需要验证码",
		})
		return
	}

	if !captchaSvc.Verify(req.CaptchaID, req.CaptchaCode, true) {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 1,
			"msg":  "验证码错误或已过期",
		})
		return
	}

	// 启用 superadmin 时：保留用户名不可注册
	if cfg.Enabled && req.Username == cfg.Username {
		c.JSON(http.StatusForbidden, gin.H{
			"code": 1,
			"msg":  "不允许注册保留用户名",
		})
		return
	}

	// 检查重名
	if _, exists := dao.GetAccountByUsername(req.Username); exists {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 1,
			"msg":  "用户名已存在",
		})
		return
	}

	acc := models.Account{
		Username: req.Username,
		Password: strings.TrimSpace(req.Password),
		Role:     strings.TrimSpace(req.Role),
	}

	if err := dao.InsertAccount(acc); err != nil {
		log.Println("\033[31m注册失败:\033[0m", err)
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
