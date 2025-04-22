package controllers

import (
	"net/http"

	"credits_service/dbService"

	"github.com/gin-gonic/gin"
)

type SpendReq struct {
	Name   string `json:"name"`
	Amount int    `json:"amount"` // Capitalized & correct type
}

func SpendController(c *gin.Context) {
	var req SpendReq

	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON body"})
		return
	}

	// Call Diminish from dbService
	success, err := dbService.Diminish(req.Name, req.Amount)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	if success {
		c.JSON(http.StatusOK, gin.H{
			"message": "Credits deducted successfully",
			"name":    req.Name,
			"amount":  req.Amount,
		})
	} else {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Unknown error occurred while deducting credits",
		})
	}
}
