package routes

import (
	"instructor_review_reply_service/controllers"

	"github.com/gin-gonic/gin"
)

func PostReply(router *gin.Engine) {
	router.POST("/review_request/:review_id", controllers.PostReply)
}

func GetReviewReqeustList(router *gin.Engine) {
	router.GET("/allrequests", controllers.GetReviewReqeustList)
}

func GetRequestInfo(router *gin.Engine) {
	router.GET("/review_info/:review_id", controllers.GetRequestInfo)
}
