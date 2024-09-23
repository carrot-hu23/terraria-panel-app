package api

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"strings"
)

func Logger() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		if !strings.HasPrefix(param.Path, "/swagger/") && !strings.HasPrefix(param.Path, "/assets/") {
			statusColor := param.StatusCodeColor()
			methodColor := param.MethodColor()
			resetColor := param.ResetColor()
			return fmt.Sprintf("[GIN] %v |%s %3d %s| %13v | %15s |%s %-7s %s %#v\n%s",
				param.TimeStamp.Format("2006/01/02 - 15:04:05"),
				statusColor, param.StatusCode, resetColor,
				param.Latency,
				param.ClientIP,
				methodColor, param.Method, resetColor,
				param.Path,
				param.ErrorMessage,
			)
		}
		return ""
	})
}

func RegisterRouter(r *gin.Engine) {

	r.Use(Logger(), gin.Recovery())

	// r.POST("/api/login", loginHandler)

	//apiGroup := r.Group("/api")
	//anonymousGroup := apiGroup.Group("")
	//{
	//
	//}

	authGroup := r.Group("/api")
	// authGroup.Use(auth.JWTAuthMiddleware())
	{
		authGroup.POST("/game/config", editServerConfig)
		authGroup.GET("/game/config", getServerConfig)
		authGroup.GET("/game/start", startGame)
		authGroup.GET("/game/stop", stopGame)
		authGroup.POST("/game/cmd", sendCmd)
		authGroup.GET("/game/log", gameLogs)
		authGroup.GET("/game/status", gameStatus)
		authGroup.GET("/game/backup", gameBackup)
		authGroup.GET("/game/backup/restore", restoreBackup)
		authGroup.DELETE("/game/backup", deleteBackup)
	}

	r.LoadHTMLGlob("dist/index.html")                  // 添加入口index.html
	r.Static("/static", "./dist/static")               // 添加资源路径
	r.Static("/assets", "./dist/assets")               // 添加资源路径
	r.StaticFile("/favicon.ico", "./dist/favicon.ico") // 添加资源路径
	r.StaticFile("/terraria", "./dist/terraria")
	r.StaticFile("/", "./dist/index.html")
}
