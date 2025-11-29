package services

import "github.com/gin-gonic/gin"

type Meta struct {
	Page       int `json:"page,omitempty"`
	PerPage    int `json:"per_page,omitempty"`
	Total      int `json:"total,omitempty"`
	TotalPages int `json:"total_pages,omitempty"`
}

func SuccessResponse(c *gin.Context, data interface{}) {
	c.JSON(200, gin.H{
		"success": true,
		"data":    data,
	})
}

func CreatedResponse(c *gin.Context, data interface{}) {
	c.JSON(201, gin.H{
		"success": true,
		"data":    data,
	})
}

func ErrorResponse(c *gin.Context, status int, message string) {
	c.JSON(status, gin.H{
		"success": false,
		"error": gin.H{
			"message": message,
		},
	})
}

func UnauthorizedResponse(c *gin.Context, message string) {
	ErrorResponse(c, 401, message)
}

func PaginatedResponse(c *gin.Context, data interface{}, meta Meta) {
	c.JSON(200, gin.H{
		"success": true,
		"data":    data,
		"meta":    meta,
	})
}
