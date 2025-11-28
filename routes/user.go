package routes

import (
	"fmt"
	"net/http"

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
		users.GET("", func(c *gin.Context) {
			fmt.Println(("get all users"))
			c.JSON(http.StatusOK, gin.H{
				"success": true,
				"message": "all users route is fine!",
			})
		})

	}
}
