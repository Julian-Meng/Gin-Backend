package errorx

import (
	"fmt"
	"log"
	"runtime/debug"
	"strings"
	"sync/atomic"
	"time"

	"github.com/gin-gonic/gin"
)

var requestSeq uint64

func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := strings.TrimSpace(c.GetHeader(HeaderRequestID))
		if requestID == "" {
			requestID = nextRequestID()
		}

		c.Set(ContextRequestIDKey, requestID)
		c.Writer.Header().Set(HeaderRequestID, requestID)
		c.Next()
	}
}

func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if rec := recover(); rec != nil {
				requestID := GetRequestID(c)
				log.Printf("[panic] request_id=%s method=%s path=%s panic=%v\n%s",
					requestID,
					c.Request.Method,
					c.Request.URL.Path,
					rec,
					debug.Stack(),
				)

				Abort(c, 500, CodeInternalError, "服务器内部错误", fmt.Errorf("%v", rec), nil)
			}
		}()

		c.Next()
	}
}

func nextRequestID() string {
	seq := atomic.AddUint64(&requestSeq, 1)
	return fmt.Sprintf("req-%d-%06d", time.Now().UnixMilli(), seq%1000000)
}
