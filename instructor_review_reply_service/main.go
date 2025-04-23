// instructor_review_reply_service
package main

import (
	"instructor_review_reply_service/db"
	"instructor_review_reply_service/routes"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

func main() {

	db.InitDB()
	defer db.CloseDB()

	router := gin.Default()

	routes.PostReply(router)
	routes.GetReviewReqeustList(router)
	routes.GetRequestInfo(router)

	router.Run(":8088")
}
