// student_request_review_service
package main

import (
	"student_request_review_service/db"
	"student_request_review_service/routes"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

func main() {

	db.InitDB()
	defer db.CloseDB()

	router := gin.Default()

	routes.PostNewReviewRequest(router)
	routes.GetAvailCourcesList(router)
	routes.GetReviewStatus(router)

	router.Run(":8087")
}
