package handlers

import (
	"backend/models"
	"time"
)

type APISuccessResponse struct {
	Code int    `json:"code" example:"0"`
	Msg  string `json:"msg" example:"ok"`
}

type APIErrorResponse struct {
	Code        int    `json:"code" example:"1"`
	Msg         string `json:"msg" example:"请求失败"`
	ErrorCode   string `json:"error_code,omitempty" example:"INVALID_PARAM"`
	RequestID   string `json:"request_id,omitempty" example:"req-1715842301123-000001"`
	Detail      string `json:"detail,omitempty" example:"json: cannot unmarshal number into Go struct field ..."`
	NeedCaptcha bool   `json:"need_captcha,omitempty" example:"true"`
	FailCount   int    `json:"fail_count,omitempty" example:"3"`
}

type SwaggerLoginRequest struct {
	Username    string `json:"username" example:"admin"`
	Password    string `json:"password" example:"123456"`
	CaptchaID   string `json:"captcha_id,omitempty" example:"rY8M8P4KQd"`
	CaptchaCode string `json:"captcha_code,omitempty" example:"A3D7"`
}

type SwaggerRegisterRequest struct {
	Username    string `json:"username" example:"staff01"`
	Password    string `json:"password" example:"123456"`
	Role        string `json:"role" example:"staff"`
	CaptchaID   string `json:"captcha_id" example:"rY8M8P4KQd"`
	CaptchaCode string `json:"captcha_code" example:"A3D7"`
}

type PersonDepartmentChangeRequest struct {
	EmpID string `json:"emp_id" example:"EMP0001"`
	Dept  string `json:"dept" example:"技术部"`
}

type PersonStateChangeRequest struct {
	EmpID string `json:"emp_id" example:"EMP0001"`
	State int    `json:"state" example:"1"`
}

type PersonJobChangeRequest struct {
	EmpID string `json:"emp_id" example:"EMP0001"`
	Job   string `json:"job" example:"后端工程师"`
}

type CaptchaData struct {
	Scene         string `json:"scene" example:"register"`
	CaptchaID     string `json:"captcha_id" example:"rY8M8P4KQd"`
	ImageBase64   string `json:"image_base64" example:"data:image/png;base64,..."`
	ExpireSeconds int    `json:"expire_seconds" example:"180"`
}

type CaptchaSuccessResponse struct {
	Code int         `json:"code" example:"0"`
	Msg  string      `json:"msg" example:"ok"`
	Data CaptchaData `json:"data"`
}

type LoginSuccessResponse struct {
	Code     int    `json:"code" example:"0"`
	Msg      string `json:"msg" example:"登录成功"`
	Token    string `json:"token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	Username string `json:"username" example:"admin"`
	Role     string `json:"role" example:"admin"`
}

type RegisterSuccessResponse struct {
	Code    int    `json:"code" example:"0"`
	Success bool   `json:"success" example:"true"`
	Msg     string `json:"msg" example:"注册成功"`
}

type PersonListResponse struct {
	Code  int                   `json:"code" example:"0"`
	Msg   string                `json:"msg" example:"ok"`
	Data  []models.EmployeeInfo `json:"data"`
	Total int64                 `json:"total" example:"100"`
}

type PersonDetailResponse struct {
	Code int           `json:"code" example:"0"`
	Msg  string        `json:"msg" example:"ok"`
	Data models.Person `json:"data"`
}

type DepartmentListResponse struct {
	Code  int                          `json:"code" example:"0"`
	Msg   string                       `json:"msg" example:"ok"`
	Data  []models.DepartmentWithCount `json:"data"`
	Total int64                        `json:"total" example:"20"`
}

type DepartmentDetailResponse struct {
	Code int               `json:"code" example:"0"`
	Msg  string            `json:"msg" example:"ok"`
	Data models.Department `json:"data"`
}

type AttendanceListResponse struct {
	Code  int                       `json:"code" example:"0"`
	Msg   string                    `json:"msg" example:"ok"`
	Data  []models.AttendanceDetail `json:"data"`
	Total int64                     `json:"total" example:"31"`
}

type AttendanceUpdateRequest struct {
	Status   *int       `json:"status,omitempty" example:"2"`
	Remark   *string    `json:"remark,omitempty" example:"迟到 10 分钟"`
	CheckIn  *time.Time `json:"check_in,omitempty" swaggertype:"string" example:"2025-01-01T09:10:00+08:00"`
	CheckOut *time.Time `json:"check_out,omitempty" swaggertype:"string" example:"2025-01-01T18:02:00+08:00"`
}

type ChatUserSendResponseData struct {
	Session     models.ChatSession  `json:"session"`
	UserMessage models.ChatMessage  `json:"user_message"`
	AIMessage   *models.ChatMessage `json:"ai_message,omitempty"`
}

type ChatUserSendResponse struct {
	Code int                      `json:"code" example:"0"`
	Msg  string                   `json:"msg" example:"ok"`
	Data ChatUserSendResponseData `json:"data"`
}

type ChatAdminSendResponseData struct {
	Session models.ChatSession `json:"session"`
	Message models.ChatMessage `json:"message"`
}

type ChatAdminSendResponse struct {
	Code int                       `json:"code" example:"0"`
	Msg  string                    `json:"msg" example:"ok"`
	Data ChatAdminSendResponseData `json:"data"`
}

type ChatSessionsResponse struct {
	Code int                  `json:"code" example:"0"`
	Msg  string               `json:"msg" example:"ok"`
	Data []models.ChatSession `json:"data"`
}

type ChatMessagesResponse struct {
	Code int                  `json:"code" example:"0"`
	Msg  string               `json:"msg" example:"ok"`
	Data []models.ChatMessage `json:"data"`
}

type ChatSessionResponse struct {
	Code int                `json:"code" example:"0"`
	Msg  string             `json:"msg" example:"ok"`
	Data models.ChatSession `json:"data"`
}

type DashboardAdminResponse struct {
	Code int                       `json:"code" example:"0"`
	Msg  string                    `json:"msg" example:"success"`
	Data models.AdminDashboardData `json:"data"`
}

type DashboardUserResponse struct {
	Code int                      `json:"code" example:"0"`
	Msg  string                   `json:"msg" example:"success"`
	Data models.UserDashboardData `json:"data"`
}

type AIAnalyzeData struct {
	Analysis string `json:"analysis" example:"【概览】..."`
}

type AIAnalyzeResponse struct {
	Code int           `json:"code" example:"0"`
	Msg  string        `json:"msg" example:"success"`
	Data AIAnalyzeData `json:"data"`
}

type NoticeListResponse struct {
	Code  int             `json:"code" example:"0"`
	Msg   string          `json:"msg" example:"ok"`
	Data  []models.Notice `json:"data"`
	Total int64           `json:"total" example:"20"`
}

type NoticeDetailResponse struct {
	Code int           `json:"code" example:"0"`
	Msg  string        `json:"msg" example:"ok"`
	Data models.Notice `json:"data"`
}

type AccountListResponse struct {
	Code int              `json:"code" example:"0"`
	Msg  string           `json:"msg" example:"ok"`
	Data []models.Account `json:"data"`
}

type AccountCreateRequest struct {
	Username string `json:"username" example:"staff01"`
	Password string `json:"password" example:"123456"`
	Role     string `json:"role" example:"staff"`
}

type AccountUpdateRequest struct {
	Role   string `json:"role" example:"staff"`
	Status int    `json:"status" example:"1"`
}

type PersonnelListResponse struct {
	Code  int                `json:"code" example:"0"`
	Msg   string             `json:"msg" example:"ok"`
	Data  []models.Personnel `json:"data"`
	Total int64              `json:"total" example:"20"`
}

type PersonnelDetailResponse struct {
	Code int              `json:"code" example:"0"`
	Msg  string           `json:"msg" example:"ok"`
	Data models.Personnel `json:"data"`
}

type PersonnelCreateRequest struct {
	TargetDpt    uint   `json:"target_dpt" example:"2"`
	ChangeType   int    `json:"change_type" example:"1"`
	Description  string `json:"description" example:"调整至技术部"`
	LeaveStartAt string `json:"leave_start_at,omitempty" example:"2026-04-10"`
	LeaveEndAt   string `json:"leave_end_at,omitempty" example:"2026-04-11"`
	LeaveReason  string `json:"leave_reason,omitempty" example:"个人事务"`
	LeaveType    string `json:"leave_type,omitempty" example:"事假"`
	HandoverNote string `json:"handover_note,omitempty" example:"交接给张三"`
}

type PersonnelApproveRequest struct {
	ID           uint   `json:"id" example:"12"`
	Approver     string `json:"approver" example:"admin"`
	Approve      bool   `json:"approve" example:"true"`
	RejectReason string `json:"reject_reason,omitempty" example:"信息不完整"`
}

type ProfileDetailResponse struct {
	Code int             `json:"code" example:"0"`
	Msg  string          `json:"msg" example:"ok"`
	Data ProfileResponse `json:"data"`
}

type PermissionsResponse struct {
	Code        int              `json:"code" example:"0"`
	Msg         string           `json:"msg" example:"ok"`
	CurrentRole string           `json:"current_role" example:"staff"`
	CurrentUser string           `json:"current_user" example:"tom"`
	Permissions []PermissionItem `json:"permissions"`
}
