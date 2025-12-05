package main

import (
	"errors"
	"fmt"
	"github/ahmedhamed993/go-auth/database"
	"github/ahmedhamed993/go-auth/middlewares"
	"github/ahmedhamed993/go-auth/routes"
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	fmt.Println("Go Auth")
	router := gin.Default()
	router.RedirectTrailingSlash = false

	router.Use(middlewares.ErrorHandler())
	v1 := router.Group("/api/v1")

	database.Connect()

	routes.AuthRoutes(v1)

	v1.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "HOME ",
		})
	})

	// health check
	v1.GET("/ok", func(c *gin.Context) {
		somethingWentWrong := false

		if somethingWentWrong {
			c.Error(errors.New("something went wrong"))
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "Everything is fine!",
		})
	})
	// to test error handler
	v1.GET("/error", func(c *gin.Context) {
		somethingWentWrong := true

		if somethingWentWrong {
			c.Error(errors.New("something went wrong"))
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "Everything is fine!",
		})
	})

	router.Run()
}
