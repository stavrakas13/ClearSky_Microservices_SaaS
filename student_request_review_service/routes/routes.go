package routes

import (
	"student_request_review_service/controllers"

	"github.com/gin-gonic/gin"
)

func PostNewReviewRequest(router *gin.Engine) {
	router.POST("/new_review_request/:course_id", controllers.PostNewReviewRequest)
}

func GetAvailCourcesList(router *gin.Engine) {
	router.GET("/mycources", controllers.GetAvailCourcesList)
}

func GetReviewStatus(router *gin.Engine) {
	router.GET("/review_info/:review_id", controllers.GetReviewStatus)
}
