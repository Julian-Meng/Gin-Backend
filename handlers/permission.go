package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// PermissionItem 用于前端展示的单条权限信息
type PermissionItem struct {
	Key          string   `json:"key"`           // 内部标识，如 "admin_person_list"
	Group        string   `json:"group"`         // 所属功能模块，如 "员工管理"
	Name         string   `json:"name"`          // 展示名称，如 "查看员工列表"
	Description  string   `json:"description"`   // 可选，更详细说明
	Method       string   `json:"method"`        // HTTP 方法，如 GET/POST/PUT/DELETE
	Path         string   `json:"path"`          // 接口路径，如 "/api/admin/persons"
	AllowedRoles []string `json:"allowed_roles"` // 允许访问的角色列表
	HasAccess    bool     `json:"has_access"`    // 当前登录用户是否具备该权限
}

// roleInList 简单判断角色是否在允许列表中
func roleInList(role string, list []string) bool {
	for _, r := range list {
		if r == role {
			return true
		}
	}
	return false
}

// GetPermissions 返回一个“权限矩阵”，用于前端展示当前用户能干什么
// 挂在：
//
//	user.GET("/permissions", handlers.GetPermissions)
//
// 要求：已通过 JWTAuthMiddleware，所以 ctx 里有 "role"
func GetPermissions(c *gin.Context) {
	roleVal, ok := c.Get("role")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code": 1,
			"msg":  "未获取到用户角色信息",
		})
		return
	}
	role, _ := roleVal.(string)

	// 可选：把当前用户信息也放回去
	usernameVal, _ := c.Get("username")
	username, _ := usernameVal.(string)

	// 约定角色：
	//   - "superadmin"
	//   - "admin"
	//   - "staff"（普通员工）
	// 其中：
	//   - /api/admin/** 只允许 ["admin", "superadmin"]
	//   - /api/user/** 允许 ["staff", "admin", "superadmin"]

	adminRoles := []string{"admin", "superadmin"}
	userRoles := []string{"staff", "admin", "superadmin"}

	perms := []PermissionItem{
		// 仪表盘
		{
			Key:          "admin_dashboard_view",
			Group:        "仪表盘",
			Name:         "管理员仪表盘",
			Description:  "查看管理员仪表盘数据",
			Method:       "GET",
			Path:         "/api/admin/dashboard",
			AllowedRoles: adminRoles,
		},
		{
			Key:          "user_dashboard_view",
			Group:        "仪表盘",
			Name:         "个人仪表盘",
			Description:  "查看当前用户的仪表盘数据",
			Method:       "GET",
			Path:         "/api/user/dashboard",
			AllowedRoles: userRoles,
		},

		// 员工管理
		{
			Key:          "admin_person_list",
			Group:        "员工管理",
			Name:         "查看员工列表",
			Description:  "分页查询所有员工信息",
			Method:       "GET",
			Path:         "/api/admin/persons",
			AllowedRoles: adminRoles,
		},
		{
			Key:          "admin_person_create",
			Group:        "员工管理",
			Name:         "创建员工",
			Description:  "新增员工记录",
			Method:       "POST",
			Path:         "/api/admin/person",
			AllowedRoles: adminRoles,
		},
		{
			Key:          "admin_person_update",
			Group:        "员工管理",
			Name:         "修改员工信息",
			Description:  "编辑员工基本信息",
			Method:       "PUT",
			Path:         "/api/admin/person/:id",
			AllowedRoles: adminRoles,
		},
		{
			Key:          "admin_person_delete_by_id",
			Group:        "员工管理",
			Name:         "删除员工（按ID）",
			Description:  "通过 ID 删除员工记录",
			Method:       "DELETE",
			Path:         "/api/admin/person/:id",
			AllowedRoles: adminRoles,
		},
		{
			Key:          "admin_person_delete_by_empid",
			Group:        "员工管理",
			Name:         "删除员工（按emp_id）",
			Description:  "通过 emp_id 删除员工记录",
			Method:       "DELETE",
			Path:         "/api/admin/person/emp/:emp_id",
			AllowedRoles: adminRoles,
		},
		{
			Key:          "admin_person_detail",
			Group:        "员工管理",
			Name:         "查看员工详情",
			Description:  "通过 ID 查看员工详细信息",
			Method:       "GET",
			Path:         "/api/admin/person/:id",
			AllowedRoles: adminRoles,
		},
		{
			Key:          "admin_person_change_job",
			Group:        "员工管理",
			Name:         "调整员工职位",
			Description:  "修改员工职位字段",
			Method:       "PUT",
			Path:         "/api/admin/person/job",
			AllowedRoles: adminRoles,
		},
		{
			Key:          "admin_person_change_state",
			Group:        "员工管理",
			Name:         "修改员工在职状态",
			Description:  "标记员工在职/离职",
			Method:       "PUT",
			Path:         "/api/admin/person/state",
			AllowedRoles: adminRoles,
		},
		{
			Key:          "admin_person_change_department",
			Group:        "员工管理",
			Name:         "调整员工部门",
			Description:  "变更员工所属部门",
			Method:       "PUT",
			Path:         "/api/admin/person/change-dept",
			AllowedRoles: adminRoles,
		},

		// 普通用户相关个人档案（user 路由）
		{
			Key:          "user_profile_view_id",
			Group:        "个人档案",
			Name:         "查看员工资料（按ID）",
			Description:  "通过 ID 查看员工资料（通常由前端控制只能看自己的）",
			Method:       "GET",
			Path:         "/api/user/profile/:id",
			AllowedRoles: userRoles,
		},
		{
			Key:          "user_profile_update_id",
			Group:        "个人档案",
			Name:         "修改员工资料（按ID）",
			Description:  "通过 ID 修改员工资料（通常是修改个人信息）",
			Method:       "PUT",
			Path:         "/api/user/profile/:id",
			AllowedRoles: userRoles,
		},
		{
			Key:          "user_profile_me",
			Group:        "个人档案",
			Name:         "查看个人档案",
			Description:  "基于当前登录 emp_id 查询完整档案信息",
			Method:       "GET",
			Path:         "/api/user/profile",
			AllowedRoles: userRoles,
		},
		{
			Key:          "admin_person_profile",
			Group:        "个人档案",
			Name:         "查看任意员工档案",
			Description:  "管理员查看指定 emp_id 的档案详情",
			Method:       "GET",
			Path:         "/api/admin/person/profile/:emp_id",
			AllowedRoles: adminRoles,
		},

		// 部门管理
		{
			Key:          "admin_department_list",
			Group:        "部门管理",
			Name:         "查看部门列表",
			Description:  "查询所有部门信息",
			Method:       "GET",
			Path:         "/api/admin/departments",
			AllowedRoles: adminRoles,
		},
		{
			Key:          "admin_department_detail",
			Group:        "部门管理",
			Name:         "查看部门详情",
			Description:  "按 ID 查询单个部门",
			Method:       "GET",
			Path:         "/api/admin/department/:id",
			AllowedRoles: adminRoles,
		},
		{
			Key:          "admin_department_create",
			Group:        "部门管理",
			Name:         "创建部门",
			Description:  "新增部门记录",
			Method:       "POST",
			Path:         "/api/admin/department",
			AllowedRoles: adminRoles,
		},
		{
			Key:          "admin_department_update",
			Group:        "部门管理",
			Name:         "修改部门",
			Description:  "编辑部门信息",
			Method:       "PUT",
			Path:         "/api/admin/department/:id",
			AllowedRoles: adminRoles,
		},
		{
			Key:          "admin_department_delete",
			Group:        "部门管理",
			Name:         "删除部门",
			Description:  "删除部门记录",
			Method:       "DELETE",
			Path:         "/api/admin/department/:id",
			AllowedRoles: adminRoles,
		},
		{
			Key:          "user_department_detail",
			Group:        "部门管理",
			Name:         "用户查看部门详情",
			Description:  "普通用户通过 ID 查看部门信息",
			Method:       "GET",
			Path:         "/api/user/department/:id",
			AllowedRoles: userRoles,
		},

		// 人事变更（Personnel）
		{
			Key:          "admin_personnel_list",
			Group:        "人事变更",
			Name:         "查看人事变更列表",
			Description:  "管理员查看所有人事变更记录",
			Method:       "GET",
			Path:         "/api/admin/changes",
			AllowedRoles: adminRoles,
		},
		{
			Key:          "admin_personnel_detail",
			Group:        "人事变更",
			Name:         "查看人事变更详情",
			Description:  "管理员查看单条人事变更记录",
			Method:       "GET",
			Path:         "/api/admin/change/:id",
			AllowedRoles: adminRoles,
		},
		{
			Key:          "admin_personnel_create",
			Group:        "人事变更",
			Name:         "创建人事变更记录",
			Description:  "管理员直接创建人事变更",
			Method:       "POST",
			Path:         "/api/admin/change",
			AllowedRoles: adminRoles,
		},
		{
			Key:          "admin_personnel_approve",
			Group:        "人事变更",
			Name:         "审批人事变更",
			Description:  "管理员审批人事变更申请",
			Method:       "PUT",
			Path:         "/api/admin/change/approve",
			AllowedRoles: adminRoles,
		},
		{
			Key:          "user_personnel_request",
			Group:        "人事变更",
			Name:         "提交人事变更申请",
			Description:  "普通用户提交变更请求（如调岗申请等）",
			Method:       "POST",
			Path:         "/api/user/change/request",
			AllowedRoles: userRoles,
		},
		{
			Key:          "user_personnel_list",
			Group:        "人事变更",
			Name:         "查看我的人事变更记录",
			Description:  "普通用户查看自己的申请与审批状态",
			Method:       "GET",
			Path:         "/api/user/changes",
			AllowedRoles: userRoles,
		},

		// 账号管理
		{
			Key:          "admin_account_list",
			Group:        "账号管理",
			Name:         "查看账号列表",
			Description:  "管理员查看所有系统账号",
			Method:       "GET",
			Path:         "/api/admin/accounts",
			AllowedRoles: adminRoles,
		},
		{
			Key:          "admin_account_create",
			Group:        "账号管理",
			Name:         "创建账号",
			Description:  "管理员新建账号（通常关联员工）",
			Method:       "POST",
			Path:         "/api/admin/account",
			AllowedRoles: adminRoles,
		},
		{
			Key:          "admin_account_update",
			Group:        "账号管理",
			Name:         "修改账号信息",
			Description:  "管理员调整账号角色、状态等",
			Method:       "PUT",
			Path:         "/api/admin/account/:id",
			AllowedRoles: adminRoles,
		},
		{
			Key:          "admin_account_delete",
			Group:        "账号管理",
			Name:         "删除账号",
			Description:  "管理员删除账号",
			Method:       "DELETE",
			Path:         "/api/admin/account/:id",
			AllowedRoles: adminRoles,
		},

		// 公告管理
		{
			Key:          "admin_notice_create",
			Group:        "公告管理",
			Name:         "发布公告",
			Description:  "创建系统公告",
			Method:       "POST",
			Path:         "/api/admin/notice",
			AllowedRoles: adminRoles,
		},
		{
			Key:          "admin_notice_update",
			Group:        "公告管理",
			Name:         "修改公告",
			Description:  "更新公告内容",
			Method:       "PUT",
			Path:         "/api/admin/notice/:id",
			AllowedRoles: adminRoles,
		},
		{
			Key:          "admin_notice_delete",
			Group:        "公告管理",
			Name:         "删除公告",
			Description:  "删除公告记录",
			Method:       "DELETE",
			Path:         "/api/admin/notice/:id",
			AllowedRoles: adminRoles,
		},
		{
			Key:          "admin_notice_detail",
			Group:        "公告管理",
			Name:         "查看公告详情（管理端）",
			Description:  "管理员查看单条公告详情",
			Method:       "GET",
			Path:         "/api/admin/notice/:id",
			AllowedRoles: adminRoles,
		},
		{
			Key:          "public_notice_list",
			Group:        "公告管理",
			Name:         "查看公告列表（公开）",
			Description:  "无需登录即可查看公告列表",
			Method:       "GET",
			Path:         "/api/notice",
			AllowedRoles: []string{}, // 代表不限角色（公开接口）
		},

		// 考勤管理
		{
			Key:          "user_attendance_checkin",
			Group:        "考勤管理",
			Name:         "签到",
			Description:  "当前用户签到打卡",
			Method:       "POST",
			Path:         "/api/user/attendance/checkin",
			AllowedRoles: userRoles,
		},
		{
			Key:          "user_attendance_checkout",
			Group:        "考勤管理",
			Name:         "签退",
			Description:  "当前用户签退打卡",
			Method:       "POST",
			Path:         "/api/user/attendance/checkout",
			AllowedRoles: userRoles,
		},
		{
			Key:          "user_attendance_my",
			Group:        "考勤管理",
			Name:         "查看个人考勤记录",
			Description:  "当前用户查询自己的考勤记录",
			Method:       "GET",
			Path:         "/api/user/attendance/my",
			AllowedRoles: userRoles,
		},
		{
			Key:          "admin_attendance_search",
			Group:        "考勤管理",
			Name:         "管理员查询考勤",
			Description:  "按条件查询所有员工的考勤记录",
			Method:       "GET",
			Path:         "/api/admin/attendance",
			AllowedRoles: adminRoles,
		},
		{
			Key:          "admin_attendance_update",
			Group:        "考勤管理",
			Name:         "管理员修改考勤记录",
			Description:  "修正员工考勤（状态/备注/时间等）",
			Method:       "PUT",
			Path:         "/api/admin/attendance/:id",
			AllowedRoles: adminRoles,
		},
		{
			Key:          "admin_attendance_delete",
			Group:        "考勤管理",
			Name:         "管理员删除考勤记录",
			Description:  "删除指定的考勤记录",
			Method:       "DELETE",
			Path:         "/api/admin/attendance/:id",
			AllowedRoles: adminRoles,
		},
	}

	// 给每条权限标记 HasAccess
	for i := range perms {
		// 公开接口（AllowedRoles 为空）可以按需要决定是否视为所有人都有
		if len(perms[i].AllowedRoles) == 0 {
			perms[i].HasAccess = true
		} else {
			perms[i].HasAccess = roleInList(role, perms[i].AllowedRoles)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"code":         0,
		"msg":          "ok",
		"current_role": role,
		"current_user": username,
		"permissions":  perms,
	})
}
