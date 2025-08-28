package api

import (
	core "project/internal"

	"github.com/gin-gonic/gin"
)

func Build(app *core.App) *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery(), gin.Logger())

	r.GET("/health", func(c *gin.Context) { c.JSON(200, gin.H{"status": "ok"}) })

	// Auth routes
	authH := &AuthHandler{App: app}
	r.POST("/v1/login/request", authH.LoginRequest)
	r.POST("/v1/login", authH.Login)
	r.POST("/v1/auth/refresh", authH.RefreshToken)

	// Protected routes
	auth := r.Group("/v1", AuthRequired(app.Cfg.JWTSecret))
	{
		auth.GET("/companies", authH.ListCompanies)
	}
	return r
}
