package api

import (
	"net/http"
	core "project/internal"

	"github.com/gin-gonic/gin"
)

// AdminRequired middleware checks if the user has admin privileges
func AdminRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		isAdmin, exists := c.Get("is_admin")
		if !exists || !isAdmin.(bool) {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "admin access required"})
			return
		}
		c.Next()
	}
}

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

		// User management routes (admin only)
		userH := NewUserHandler(app)
		users := auth.Group("/users", AdminRequired())
		{
			users.GET("", userH.ListUsers)
			users.POST("", userH.CreateUser)
			users.DELETE("/:id", userH.DeleteUser)
		}
	}
	return r
}
