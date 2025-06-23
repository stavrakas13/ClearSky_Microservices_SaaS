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
		return controllers.PostReply(msg.Body)

	case "instructor.getRequestsList":
		return controllers.GetReviewRequestList(msg.Body)

	case "instructor.getRequestInfo":
		return controllers.GetRequestInfo(msg.Body)

	case "instructor.insertStudentRequest":
		return controllers.InsertStudentRequest(msg.Body)

	// route for updating instructors table
	/* 	case "instructor.addCourse":
	courseID := msg.Params["course_id"]         // COURSE NAME FROM UPLOAD
	instructorID := msg.Params["instructor_id"] // LOGGED-IN INSTRUCTOR FROM JWT

	if courseID == "" || instructorID == "" {
		return "", fmt.Errorf("missing course_id or instructor_id in params")
	}

	err := controllers.AddCourse(courseID, instructorID)
	if err != nil {
		return "", fmt.Errorf("failed to add course: %w", err)
	}

	return `{"message": "Course successfully added"}`, nil */

	default:
		return "", fmt.Errorf("unknown routing key: %s", routingKey)
	}
}
