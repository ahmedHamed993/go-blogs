package routes

import (
	"github/ahmedhamed993/go-auth/controllers"

	"github.com/gin-gonic/gin"
)

func AuthRoutes(rg *gin.RouterGroup) {
	auth := rg.Group("/auth")
	{
		auth.POST("login", controllers.Login)

		auth.POST("register", controllers.Register)

	}
}
