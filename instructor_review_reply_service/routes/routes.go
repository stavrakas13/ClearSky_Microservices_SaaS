package routes

/* import (
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
*/

import (
	"encoding/json"
	"fmt"
	"instructor_review_reply_service/controllers"
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
	case "instructor.postResponse":
		return controllers.PostReply(msg.Params, msg.Body)

	case "instructor.getRequestsList":
		return controllers.GetReviewReqeustList(msg.Params, msg.Body)

	case "instructor.getRequestInfo":
		return controllers.GetRequestInfo(msg.Params, msg.Body)

	case "instructor.insertStudentRequest":
		return controllers.InsertStudentRequest(msg.Params, msg.Body)

	default:
		return "", fmt.Errorf("unknown routing key: %s", routingKey)
	}
}
