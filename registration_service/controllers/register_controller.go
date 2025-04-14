// controllers/register_controller.go
package controllers

import (
	"net/http"

	"registration_service/dbService"

	"github.com/gin-gonic/gin"
)

type UserRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Director string `json:"director"`
}

func RegisterController(c *gin.Context) {
	var req UserRequest

	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON body"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"name":     req.Name,
		"email":    req.Email,
		"director": req.Director,
	})

	result, err := dbService.AddInstitution(req.Name, req.Email, req.Director)

	if err != nil {
		if result == 2 {
			c.JSON(http.StatusConflict, gin.H{"error": "Institution already registered"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Institution registered successfully"})
}
