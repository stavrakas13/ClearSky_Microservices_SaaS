package routes

/* import (
	"student_request_review_service/controllers"

	"github.com/gin-gonic/gin"
)

func PostNewReviewRequest(router *gin.Engine) {
	router.POST("/new_review_request/:course_id", controllers.PostNewReviewRequest)
}

func GetReviewStatus(router *gin.Engine) {
	router.GET("/review_info/:review_id", controllers.GetReviewStatus)
}
*/

// Routing with Rabbit mq -> takes routing key and calls different controller.

import (
	"encoding/json"
	"fmt"
	"student_request_review_service/controllers"
)

type Message struct {
	Params map[string]string      `json:"params"`
	Body   map[string]interface{} `json:"body"`
}

func Routing(routingKey string, messageBody []byte) (string, error) {
	var msg Message
	err := json.Unmarshal(messageBody, &msg)
	if err != nil {
		return "", fmt.Errorf("failed to parse message: %w", err)
	}

	switch routingKey {
	case "student.postNewRequest":
		return controllers.PostNewReviewRequest(msg.Params, msg.Body)

	case "student.getRequestStatus":
		return controllers.GetReviewStatus(msg.Params)

	case "student.updateInstructorResponse":
		return controllers.UpdateInstructorResponse(msg.Params, msg.Body)

	default:
		return "", fmt.Errorf("unknown routing key: %s", routingKey)
	}
}
