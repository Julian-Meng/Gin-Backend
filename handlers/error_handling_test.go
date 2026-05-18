package handlers

import (
	"backend/middlewares/errorx"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/gin-gonic/gin"
)

func newJSONTestContext(method, target, body, requestID string) (*gin.Context, *httptest.ResponseRecorder) {
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(method, target, strings.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	if strings.TrimSpace(requestID) != "" {
		c.Set(errorx.ContextRequestIDKey, requestID)
	}
	return c, rec
}

func decodeBody(t *testing.T, rec *httptest.ResponseRecorder) map[string]interface{} {
	t.Helper()

	var body map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response body failed: %v", err)
	}
	return body
}

func resetAuthGlobalsForTest() {
	superAdminOnce = sync.Once{}
	superAdminCfg = SuperAdminConfig{}
	superAdminErr = nil

	captchaOnce = sync.Once{}
	captchaSvc = nil
	captchaTTL = 0
	loginCapThld = 0
}

func assertErrorResponse(t *testing.T, rec *httptest.ResponseRecorder, wantStatus int, wantErrorCode string) map[string]interface{} {
	t.Helper()

	if rec.Code != wantStatus {
		t.Fatalf("unexpected status: got=%d want=%d body=%s", rec.Code, wantStatus, rec.Body.String())
	}

	body := decodeBody(t, rec)
	if got := body["code"]; got != float64(1) {
		t.Fatalf("unexpected code: got=%v want=1", got)
	}
	if got := body["error_code"]; got != wantErrorCode {
		t.Fatalf("unexpected error_code: got=%v want=%s", got, wantErrorCode)
	}
	return body
}

func TestLogin_InvalidJSON_ReturnsUnifiedBadRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	resetAuthGlobalsForTest()
	t.Setenv("SUPERADMIN_ENABLED", "false")

	c, rec := newJSONTestContext(http.MethodPost, "/api/login", `{"username":`, "req-login-invalid-json")
	Login(c)

	body := assertErrorResponse(t, rec, http.StatusBadRequest, errorx.CodeInvalidParam)
	if got := body["request_id"]; got != "req-login-invalid-json" {
		t.Fatalf("unexpected request_id: got=%v", got)
	}
}

func TestGetPersonByID_InvalidID_ReturnsBadRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)

	c, rec := newJSONTestContext(http.MethodGet, "/api/admin/person/abc", "", "req-person-invalid-id")
	c.Params = gin.Params{{Key: "id", Value: "abc"}}

	GetPersonByID(c)

	assertErrorResponse(t, rec, http.StatusBadRequest, errorx.CodeInvalidParam)
}

func TestGetMyPersonnelList_MissingEmpID_ReturnsUnauthorized(t *testing.T) {
	gin.SetMode(gin.TestMode)

	c, rec := newJSONTestContext(http.MethodGet, "/api/user/changes", "", "req-personnel-no-empid")
	GetMyPersonnelList(c)

	assertErrorResponse(t, rec, http.StatusUnauthorized, errorx.CodeUnauthorized)
}

func TestAdminUpdateAttendance_InvalidID_ReturnsBadRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)

	c, rec := newJSONTestContext(http.MethodPut, "/api/admin/attendance/abc", `{}`, "req-attendance-invalid-id")
	c.Params = gin.Params{{Key: "id", Value: "abc"}}

	AdminUpdateAttendance(c)

	assertErrorResponse(t, rec, http.StatusBadRequest, errorx.CodeInvalidParam)
}

func TestAdminClaimWaitingSessions_InvalidJSON_ReturnsBadRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)

	c, rec := newJSONTestContext(http.MethodPost, "/api/chat/admin/sessions/claim", `{"limit":`, "req-chat-bad-json")
	c.Set("role", "admin")
	c.Set("emp_id", "EMP0001")

	AdminClaimWaitingSessions(c)

	assertErrorResponse(t, rec, http.StatusBadRequest, errorx.CodeInvalidParam)
}

func TestParseAdminClaimRequest_EmptyBody_UsesDefaultLimit(t *testing.T) {
	gin.SetMode(gin.TestMode)

	c, _ := newJSONTestContext(http.MethodPost, "/api/chat/admin/sessions/claim", "", "")
	req, err := parseAdminClaimRequest(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if req.Limit != 20 {
		t.Fatalf("unexpected default limit: got=%d want=20", req.Limit)
	}
}
