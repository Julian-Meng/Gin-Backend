package errorx

import (
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

const (
	ContextRequestIDKey = "request_id"
	HeaderRequestID     = "X-Request-ID"
)

func GetRequestID(c *gin.Context) string {
	if c == nil {
		return ""
	}
	v, ok := c.Get(ContextRequestIDKey)
	if !ok {
		return ""
	}
	requestID, _ := v.(string)
	return strings.TrimSpace(requestID)
}

func Abort(c *gin.Context, status int, errorCode, msg string, detail error, extra gin.H) {
	body := gin.H{
		"code":       1,
		"msg":        msg,
		"error_code": errorCode,
		"request_id": GetRequestID(c),
	}

	if detail != nil && exposeDetail() {
		body["detail"] = detail.Error()
	}

	for k, v := range extra {
		body[k] = v
	}

	c.AbortWithStatusJSON(status, body)
}

func BadRequest(c *gin.Context, msg string, detail error) {
	Abort(c, http.StatusBadRequest, CodeInvalidParam, msg, detail, nil)
}

func Unauthorized(c *gin.Context, msg string, detail error) {
	Abort(c, http.StatusUnauthorized, CodeUnauthorized, msg, detail, nil)
}

func Forbidden(c *gin.Context, msg string, detail error) {
	Abort(c, http.StatusForbidden, CodeForbidden, msg, detail, nil)
}

func NotFound(c *gin.Context, msg string, detail error) {
	Abort(c, http.StatusNotFound, CodeNotFound, msg, detail, nil)
}

func Conflict(c *gin.Context, msg string, detail error) {
	Abort(c, http.StatusConflict, CodeConflict, msg, detail, nil)
}

func Internal(c *gin.Context, msg string, detail error) {
	Abort(c, http.StatusInternalServerError, CodeInternalError, msg, detail, nil)
}

func exposeDetail() bool {
	if strings.EqualFold(strings.TrimSpace(os.Getenv("ERROR_DETAIL_ENABLED")), "true") {
		return true
	}

	return strings.EqualFold(strings.TrimSpace(os.Getenv("GIN_MODE")), gin.DebugMode)
}
