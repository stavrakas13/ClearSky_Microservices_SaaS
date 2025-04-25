package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func Validate() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetString("user_id")
		email := c.GetString("email")
		role := c.GetString("role")

		c.JSON(http.StatusOK, gin.H{
			"user_id": userID,
			"email":   email,
			"role":    role,
		})
	}
}
