package controllers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

func GetAllUsers(c *gin.Context) {
	fmt.Println(("get all users"))
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "all users route is fine!",
	})
}
