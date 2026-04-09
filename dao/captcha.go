package dao

import (
	"fmt"
	"strings"
	"sync"
)

var (
	loginFailMu      sync.RWMutex
	loginFailCounter = make(map[string]int)
)

func loginFailKey(username, ip string) string {
	u := strings.ToLower(strings.TrimSpace(username))
	i := strings.TrimSpace(ip)
	return fmt.Sprintf("%s|%s", u, i)
}

// GetLoginFailCount 获取指定用户名+IP的连续登录失败次数
func GetLoginFailCount(username, ip string) int {
	key := loginFailKey(username, ip)
	loginFailMu.RLock()
	defer loginFailMu.RUnlock()
	return loginFailCounter[key]
}

// IncrementLoginFailCount 递增指定用户名+IP的连续登录失败次数并返回新值
func IncrementLoginFailCount(username, ip string) int {
	key := loginFailKey(username, ip)
	loginFailMu.Lock()
	defer loginFailMu.Unlock()
	loginFailCounter[key]++
	return loginFailCounter[key]
}

// ClearLoginFailCount 清除指定用户名+IP的连续登录失败次数
func ClearLoginFailCount(username, ip string) {
	key := loginFailKey(username, ip)
	loginFailMu.Lock()
	defer loginFailMu.Unlock()
	delete(loginFailCounter, key)
}
