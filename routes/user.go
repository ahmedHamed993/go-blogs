package routes

import (
	"github/ahmedhamed993/go-auth/controllers"

	"github.com/gin-gonic/gin"
)

func UserRoutes(rg *gin.RouterGroup) {
	users := rg.Group("/users")
	{
		// users.GET("/", controllers.GetAllUsers)
		// users.GET("/:id", controllers.GetUserByID)
		// users.POST("/", controllers.CreateUser)
		// users.PUT("/:id", controllers.UpdateUser)
		// users.DELETE("/:id", controllers.DeleteUser)
		users.GET("", controllers.GetAllUsers)

	}
}
