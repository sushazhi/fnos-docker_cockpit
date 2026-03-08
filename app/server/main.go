package main

import (
	"dockpit/internal/config"
	"dockpit/internal/handler"
	"dockpit/internal/middleware"
	"dockpit/internal/service"
	"dockpit/pkg/docker"
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
)

func main() {
	config.Init()

	if err := docker.Init(); err != nil {
		log.Printf("璀﹀憡: Docker杩炴帴澶辫触: %v", err)
	}

	service.InitSessionService()
	service.InitAuthService()
	service.InitAuditService()
	middleware.InitRateLimiter()

	gin.SetMode(gin.ReleaseMode)
	r := gin.New()

	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	r.Use(middleware.HTTPSRedirect())
	r.Use(middleware.SecurityHeaders())
	r.Use(middleware.CORS())
	r.Use(middleware.RateLimit()) // 搴旂敤閫熺巼闄愬埗涓棿浠?

	r.Static("/assets", "./ui/assets")
	r.StaticFile("/", "./ui/index.html")
	r.StaticFile("/favicon.ico", "./ui/favicon.ico")

	auth := handler.NewAuthHandler()
	containers := handler.NewContainerHandler()
	images := handler.NewImageHandler()
	networks := handler.NewNetworkHandler()
	volumes := handler.NewVolumeHandler()
	system := handler.NewSystemHandler()
	compose := handler.NewComposeHybridHandler()
	files := handler.NewFileHandler()
	audit := handler.NewAuditHandler()

	authGroup := r.Group("/api/auth")
	{
		authGroup.GET("/check", auth.CheckInit)
		authGroup.POST("/setup", middleware.AuthRateLimit(), auth.Setup)
		authGroup.POST("/login", middleware.AuthRateLimit(), auth.Login)
	}

	api := r.Group("/api")
	api.Use(middleware.AuthRequired())
	{
		api.GET("/me", auth.Me)
		api.POST("/logout", auth.Logout)
		api.POST("/change-password", middleware.AuthRateLimit(), middleware.CSRFRequired(), auth.ChangePassword)
		api.GET("/csrf-token", auth.GetCSRFToken)

		api.GET("/containers", containers.List)
		api.GET("/container/:id", containers.Get)
		api.GET("/container/:id/stats", containers.Stats)
		api.GET("/container/:id/logs", containers.Logs)
		api.POST("/container/:id/start", middleware.CSRFRequired(), containers.Start)
		api.POST("/container/:id/stop", middleware.CSRFRequired(), containers.Stop)
		api.POST("/container/:id/restart", middleware.CSRFRequired(), containers.Restart)
		api.POST("/container/:id/pause", middleware.CSRFRequired(), containers.Pause)
		api.POST("/container/:id/unpause", middleware.CSRFRequired(), containers.Unpause)
		api.POST("/container/:id/remove", middleware.CSRFRequired(), containers.Remove)
		api.POST("/container/:id/exec", middleware.CSRFRequired(), containers.Exec)
		api.POST("/container/:id/rename", middleware.CSRFRequired(), containers.Rename)
		api.POST("/container/:id/update", middleware.CSRFRequired(), containers.Update)
		api.POST("/container/:id/commit", middleware.CSRFRequired(), containers.Commit)
		api.POST("/container/create", middleware.CSRFRequired(), containers.Create)
		api.POST("/container/:id/terminal", middleware.CSRFRequired(), containers.ExecCreate)
		api.POST("/exec/:execId/resize", middleware.CSRFRequired(), containers.ExecResize)
		api.GET("/exec/:execId/ws", containers.ExecWebSocket)

		api.GET("/images", images.List)
		api.GET("/image/:id", images.Get)
		api.GET("/image/:id/history", images.History)
		api.GET("/image/:id/check-update", images.CheckUpdate)
		api.POST("/image/:id/update", middleware.CSRFRequired(), images.Update)
		api.POST("/image/pull", middleware.CSRFRequired(), images.Pull)
		api.GET("/image/pull-stream", middleware.CSRFRequired(), images.PullStream)
		api.POST("/image/push", middleware.CSRFRequired(), images.Push)
		api.POST("/image/:id/remove", middleware.CSRFRequired(), images.Remove)
		api.POST("/image/tag", middleware.CSRFRequired(), images.Tag)
		api.GET("/image/search/check", images.CheckSearchAvailable)
		api.GET("/image/search", images.Search)
		api.POST("/image/prune", middleware.CSRFRequired(), images.Prune)
		api.POST("/image/build", middleware.CSRFRequired(), images.Build)

		api.GET("/networks", networks.List)
		api.GET("/network/:id", networks.Get)
		api.POST("/network/create", middleware.CSRFRequired(), networks.Create)
		api.POST("/network/:id/remove", middleware.CSRFRequired(), networks.Remove)
		api.POST("/network/:id/connect", middleware.CSRFRequired(), networks.Connect)
		api.POST("/network/:id/disconnect", middleware.CSRFRequired(), networks.Disconnect)
		api.POST("/network/prune", middleware.CSRFRequired(), networks.Prune)

		api.GET("/volumes", volumes.List)
		api.GET("/volume/:name", volumes.Get)
		api.POST("/volume/create", middleware.CSRFRequired(), volumes.Create)
		api.POST("/volume/:name/remove", middleware.CSRFRequired(), volumes.Remove)
		api.POST("/volume/prune", middleware.CSRFRequired(), volumes.Prune)

		api.GET("/system/info", system.Info)
		api.GET("/system/version", system.Version)
		api.GET("/system/df", system.DiskUsage)
		api.GET("/system/check", system.Check)
		api.GET("/system/events", system.Events)
		api.GET("/system/app-info", system.AppInfo)
		api.POST("/system/prune", middleware.CSRFRequired(), system.Prune)

		api.GET("/compose", compose.List)
		api.GET("/compose/:name", compose.Get)
		api.POST("/compose/save", middleware.CSRFRequired(), compose.Save)
		api.POST("/compose/:name/delete", middleware.CSRFRequired(), compose.Delete)
		api.POST("/compose/:name/up", middleware.CSRFRequired(), compose.Up)
		api.POST("/compose/:name/down", middleware.CSRFRequired(), compose.Down)
		api.GET("/compose/:name/ps", compose.Ps)
		api.GET("/compose/:name/logs", compose.Logs)
		api.POST("/compose/:name/pull", middleware.CSRFRequired(), compose.Pull)
		api.POST("/compose/:name/build", middleware.CSRFRequired(), compose.Build)
		api.POST("/compose/:name/restart", middleware.CSRFRequired(), compose.Restart)
		api.POST("/compose/:name/stop", middleware.CSRFRequired(), compose.Stop)
		api.POST("/compose/:name/start", middleware.CSRFRequired(), compose.Start)

		api.GET("/files", files.List)
		api.GET("/file/read", files.Read)
		api.POST("/file/write", middleware.CSRFRequired(), files.Write)
		api.DELETE("/file", middleware.CSRFRequired(), files.Delete)
		api.POST("/file/mkdir", middleware.CSRFRequired(), files.Mkdir)
		api.POST("/file/rename", middleware.CSRFRequired(), files.Rename)
		api.POST("/file/upload", middleware.CSRFRequired(), files.Upload)
		api.GET("/file/download", files.Download)
		api.POST("/file/copy-to-container", middleware.CSRFRequired(), files.CopyToContainer)

		api.GET("/audit", audit.List)
	}

	r.NoRoute(func(c *gin.Context) {
		c.File("./ui/index.html")
	})

	addr := fmt.Sprintf("0.0.0.0:%d", config.Get().Port)
	log.Printf("Docker瀹瑰櫒绠＄悊鏈嶅姟宸插惎鍔紝绔彛: %d", config.Get().Port)
	
	if err := r.Run(addr); err != nil {
		log.Fatalf("鍚姩鏈嶅姟澶辫触: %v", err)
	}
}

