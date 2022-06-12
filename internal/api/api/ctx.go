package api

import (
	"context"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

const (
	ctxKey_context = "__ctx__context__"
	ctxKey_logger  = "__ctx__logger__"
	ctxKey_reqID   = "__ctx__reqid__"
)

func setReqID(c *gin.Context, reqID string) {
	c.Set(ctxKey_reqID, reqID)
}
func getReqID(c *gin.Context) string {
	return c.GetString(ctxKey_reqID)
}

func setLogger(c *gin.Context, logger *zap.Logger) {
	c.Set(ctxKey_logger, logger)
}
func getLogger(c *gin.Context) *zap.Logger {
	return c.MustGet(ctxKey_logger).(*zap.Logger)
}

func setCtx(c *gin.Context, ctx context.Context) {
	c.Set(ctxKey_context, ctx)
}
func getCtx(c *gin.Context) context.Context {
	return c.MustGet(ctxKey_context).(context.Context)
}
