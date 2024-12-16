package routes

import (
	"github.com/asifrahaman13/bhagabad_gita/internal/handlers"
	"github.com/gin-gonic/gin"
)

func SetupV1Routes(router *gin.Engine) {

	v1 := router.Group("/auth")
	{
		v1.POST("/signup", handlers.UserHandler.Signup)
		v1.POST("/login", handlers.UserHandler.Login)
	}
}

func SetupPublicRoutes(router *gin.Engine) {
	public := router.Group("/v1")
	{
		public.GET("/public", handlers.UserHandler.PublicApi)
	}
}

func InitializeRoutes(router *gin.Engine) {
	SetupV1Routes(router)
	SetupPublicRoutes(router)
}
