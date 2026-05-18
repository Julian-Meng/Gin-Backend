package middlewares

import (
	"backend/middlewares/errorx"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func decodeMiddlewareBody(t *testing.T, rec *httptest.ResponseRecorder) map[string]interface{} {
	t.Helper()

	var body map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response body failed: %v", err)
	}
	return body
}

func TestJWTAuthMiddleware_MissingAuthorization_UsesUnifiedErrorProtocol(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.Use(errorx.RequestID(), JWTAuthMiddleware())
	r.GET("/protected", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"code": 0})
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("unexpected status: got=%d want=%d body=%s", rec.Code, http.StatusUnauthorized, rec.Body.String())
	}

	body := decodeMiddlewareBody(t, rec)
	if got := body["error_code"]; got != errorx.CodeUnauthorized {
		t.Fatalf("unexpected error_code: got=%v want=%s", got, errorx.CodeUnauthorized)
	}
	requestID, _ := body["request_id"].(string)
	if requestID == "" {
		t.Fatalf("request_id should not be empty, body=%s", rec.Body.String())
	}
}

func TestAdminOnly_NonAdmin_UsesUnifiedErrorProtocol(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.Use(errorx.RequestID(), func(c *gin.Context) {
		c.Set("role", "staff")
		c.Next()
	}, AdminOnly())
	r.GET("/admin-only", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"code": 0})
	})

	req := httptest.NewRequest(http.MethodGet, "/admin-only", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("unexpected status: got=%d want=%d body=%s", rec.Code, http.StatusForbidden, rec.Body.String())
	}

	body := decodeMiddlewareBody(t, rec)
	if got := body["error_code"]; got != errorx.CodeForbidden {
		t.Fatalf("unexpected error_code: got=%v want=%s", got, errorx.CodeForbidden)
	}
}
