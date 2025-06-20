package routes

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func SetupRoutes(r *gin.Engine, db *gorm.DB) {
	//r.POST("/upload", services.PostData(db))
	r.GET("/Grades")
	r.POST("/intialGrades")
	r.PUT("/updateGrades")
}
