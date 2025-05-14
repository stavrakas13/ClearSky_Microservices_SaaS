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

import (
	"encoding/json"
	"fmt"
	"student_request_review_service/controllers"
)

type Message struct {
	Action string                 `json:"action"`
	Params map[string]string      `json:"params"`
	Body   map[string]interface{} `json:"body"`
}

func Routing(messageBody []byte) (string, error) {
	var msg Message
	err := json.Unmarshal(messageBody, &msg)
	if err != nil {
		return "", fmt.Errorf("failed to parse message: %w", err)
	}

	switch msg.Action {
	case "PostNewReviewRequest":
		return controllers.PostNewReviewRequest(msg.Params, msg.Body)
	case "GetReviewStatus":
		return controllers.GetReviewStatus(msg.Params)
	default:
		return "", fmt.Errorf("unknown action: %s", msg.Action)
	}
}
