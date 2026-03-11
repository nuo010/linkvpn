package router

import (
	"net/http"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"singleOpenVpn/internal/config"
	"singleOpenVpn/internal/handler"
)

func Setup(db *gorm.DB, cfg *config.Config) *gin.Engine {
	r := gin.Default()
	r.Use(corsMiddleware())

	api := r.Group("/api")
	api.POST("/login", handler.Login(db, cfg))

	auth := api.Group("")
	auth.Use(handler.AuthMiddleware(cfg))
	{
		auth.GET("/home", handler.GetHome(db, cfg))
		auth.GET("/users", handler.ListUsers(db, cfg))
		auth.GET("/logs/login", handler.GetLoginLogs(db))
		auth.DELETE("/logs/login", handler.ClearLoginLogs(db))
		auth.GET("/logs/vpn", handler.GetVPNConnectionLogs(db, cfg))
		auth.DELETE("/logs/vpn", handler.ClearVPNConnectionLogs(db, cfg))
		auth.GET("/logs/vpn-file", handler.GetVPNLogFile(cfg))
		auth.POST("/users", handler.CreateUser(db, cfg))
		auth.PUT("/users/:id", handler.UpdateUser(db, cfg))
		auth.DELETE("/users/:id", handler.DeleteUser(db, cfg))
		auth.GET("/users/:id/ccd", handler.GetUserCCD(db, cfg))
		auth.PUT("/users/:id/ccd", handler.SetUserCCD(db, cfg))
		auth.GET("/users/:id/download", handler.DownloadClientConfig(db, cfg))

		auth.GET("/config", handler.GetServerConfig(db))
		auth.GET("/config/need-initial-setup", handler.NeedInitialClientConfig(db))
		auth.POST("/config", handler.SetServerConfig(db, cfg))
		auth.GET("/config/file", handler.GetServerConfigFile(cfg))
		auth.PUT("/config/file", handler.PutServerConfigFile(db, cfg))
		auth.GET("/config/default", handler.GetDefaultServerConfig(cfg))
		auth.POST("/config/restart", handler.RestartVPNService(cfg))
		auth.GET("/config/params", handler.GetOpenVPNParams(db))
		auth.POST("/config/params", handler.SetOpenVPNParams(db, cfg))
		auth.POST("/config/params/apply", handler.ApplyOpenVPNParams(db, cfg))
		auth.POST("/vpn/init", handler.InitVPN(cfg))
		auth.GET("/vpn/status", handler.GetVPNStatus(cfg))
		auth.GET("/stats/usage", handler.GetUsageStats(cfg))
	}

	if cfg.StaticDir != "" {
		dir := filepath.Clean(cfg.StaticDir)
		r.NoRoute(staticHandler(dir))
	}

	return r
}

func staticHandler(dir string) gin.HandlerFunc {
	fileServer := http.FileServer(http.Dir(dir))
	return func(c *gin.Context) {
		p := c.Request.URL.Path
		if p == "/" {
			p = "/index.html"
		}
		f, err := http.Dir(dir).Open(p)
		if err == nil {
			f.Close()
			fileServer.ServeHTTP(c.Writer, c.Request)
			return
		}
		c.Request.URL.Path = "/index.html"
		fileServer.ServeHTTP(c.Writer, c.Request)
	}
}

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	}
}
