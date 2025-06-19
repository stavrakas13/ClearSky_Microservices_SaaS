package routes

import (
	"orchestrator/internal/handlers"

	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	amqp "github.com/rabbitmq/amqp091-go"
)

func SetupRouter(ch *amqp.Channel) *gin.Engine {
	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins: []string{"http://localhost:5173", "https://frontend.example.com"},
		AllowMethods: []string{"POST", "OPTIONS", "PATCH", "PUT"},
		AllowHeaders: []string{"Content-Type", "Authorization"},
		MaxAge:       12 * time.Hour,
	}))

	// 16 MiB in-memory before spilling to /tmp
	r.MaxMultipartMemory = 16 << 20 // 16 MiB

	// r.PATCH("/purchase", func(c *gin.Context) {
	// 	handlers.HandleCreditsPurchased(c, ch)
	// })

	r.GET("/mycredits", func(c *gin.Context) {
		handlers.HandleCreditsAvail(c, ch)
	})

	r.PATCH("/postFinalGrades", func(c *gin.Context) {
		handlers.UploadExcelFinal(c, ch)
	})

	r.POST("/registration", func(c *gin.Context) {
		handlers.HandleInstitutionRegistered(c, ch)
	})

	r.POST("/upload_init", func(c *gin.Context) {
		handlers.UploadExcelInit(c, ch)
	})

	r.POST("/stats/persist", func(c *gin.Context) {
		handlers.HandlePersistAndCalculate(c, ch)
	})
	r.POST("/stats/distributions", func(c *gin.Context) {
		handlers.HandleGetDistributions(c, ch)
	})

	r.POST("/personal/courses", func(c *gin.Context) {
		handlers.HandleGetStudentCourses(c, ch)
	})
	r.POST("/personal/grades", func(c *gin.Context) {
		handlers.HandleGetPersonalGrades(c, ch)
	})

	// Student and Instructor API calls.

	r.PATCH("/student/reviewRequest", func(c *gin.Context) {
		// Calls HandlePostNewRequest:
		//   Expects JSON body {course_id, user_id, student_message, exam_period}
		//   Returns JSON: {"data": map[string]interface{}} where data is the service response
		handlers.HandlePostNewRequest(c, ch)
	})
	r.PATCH("/student/status", func(c *gin.Context) {
		// Calls HandleGetRequestStatus:
		//   Expects JSON body {course_id, user_id, exam_period}
		//   Returns JSON: {"data": map[string]interface{}} containing status details
		handlers.HandleGetRequestStatus(c, ch)
	})
	r.PATCH("/instructor/review-list", func(c *gin.Context) {
		// Calls HandleGetRequestList:
		//   Expects JSON body {course_id, exam_period}
		//   Returns JSON: {"data": map[string]interface{}} listing pending reviews
		handlers.HandleGetRequestList(c, ch)
	})
	r.PATCH("/instructor/reply", func(c *gin.Context) {
		// Calls HandlePostResponse:
		//   Expects JSON body {course_id, user_id, exam_period, instructor_reply_message, instructor_action}
		//   Returns JSON: {"data": map[string]interface{}} acknowledgments from services
		handlers.HandlePostResponse(c, ch)
	})

	// User Management Endpoints
	r.POST("/user/register", func(c *gin.Context) {
		handlers.HandleUserRegister(c, ch)
	})
	r.POST("/user/login", func(c *gin.Context) {
		handlers.HandleUserLogin(c, ch)
	})
	r.DELETE("/user/delete", func(c *gin.Context) {
		handlers.HandleUserDelete(c, ch)
	})
	r.POST("/user/google-login", func(c *gin.Context) {
		handlers.HandleUserGoogleLogin(c, ch)
	})

	return r
}
