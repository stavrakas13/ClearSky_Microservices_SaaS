package routes

import (
	"stats_service/services"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func SetupRoutes(r *gin.Engine, db *gorm.DB) {
	r.GET("/distributions", func(c *gin.Context) {
		classID := c.Query("class_id")
		examDate := c.Query("exam_date")

		err := services.CalculateDistributions(db, classID, examDate)
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}

		c.JSON(200, gin.H{"message": "Distributions calculated successfully"})
	})
	r.POST("/upload", services.PostData(db))
}
