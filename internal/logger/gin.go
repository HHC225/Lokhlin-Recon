// Package logger -----------------------------
// @file      : gin.go
// @time      : 2025/4/24 13:55
// -------------------------------------------
package logger

//
//import (
//	"os"
//	"time"
//
//	"github.com/gin-gonic/gin"
//	"go.uber.org/zap"
//)
//

//func GinLogger() gin.HandlerFunc {
//	return gin.LoggerWithWriter(zapWriter{Logger})
//}
//

//type zapWriter struct {
//	l *zap.Logger
//}
//
//func (z zapWriter) Write(p []byte) (n int, err error) {

//	return len(p), nil
//}
//

//func GinRecovery() gin.HandlerFunc {
//	return gin.CustomRecoveryWithWriter(os.Stderr, func(c *gin.Context, err interface{}) {
//		Logger.Error("Panic recovered",
//			zap.Any("error", err),
//			zap.String("path", c.Request.URL.Path),
//			zap.Time("time", time.Now()),
//		)
//		c.AbortWithStatusJSON(500, gin.H{
//			"message": "Server Error",
//		})
//	})
//}
