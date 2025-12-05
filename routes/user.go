package routes

import (
	"github/ahmedhamed993/go-auth/controllers"
	"github/ahmedhamed993/go-auth/middlewares"

	"github.com/gin-gonic/gin"
)

func UserRoutes(rg *gin.RouterGroup) {
	users := rg.Group("/users")
	// Apply optional authentication middleware to all user routes
	// This sets user context if token is provided, but doesn't abort if missing
	users.Use(middlewares.AuthMiddleware())
	{
		// GET /users - List all users (requires users:read permission)
		users.GET("",
			middlewares.RequirePermission("users:read"),
			controllers.GetAllUsers)

		// GET /users/:id - Get user by ID (requires users:read or users:read:own with ownership check)
		users.GET("/:id",
			middlewares.RequirePermission("users:read"),
			controllers.GetUserByID)

		// POST /users - Create new user (requires users:create permission)
		users.POST("",
			middlewares.RequirePermission("users:create"),
			controllers.CreateUser)

		// PUT /users/:id - Update user (requires users:update or users:update:own with ownership check)
		users.PUT("/:id",
			middlewares.RequirePermission("users:update"),
			controllers.UpdateUser)

		// DELETE /users/:id - Delete user (requires users:delete permission)
		users.DELETE("/:id",
			middlewares.RequirePermission("users:delete"),
			controllers.DeleteUser)
	}
}
