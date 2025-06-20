package routes

import (
	"orchestrator/internal/handlers"
	mw "orchestrator/internal/middleware"

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

	// Shared stats endpoints (all roles)
	stats := r.Group("/stats")
	stats.Use(mw.JWTAuthMiddleware())
	{
		stats.POST("/persist", func(c *gin.Context) {
			handlers.HandlePersistAndCalculate(c, ch)
		})
		stats.POST("/distributions", func(c *gin.Context) {
			handlers.HandleGetDistributions(c, ch)
		})
	}

	// Institution‐representative only
	inst := r.Group("/")
	inst.Use(mw.JWTAuthMiddleware(), RoleCheck("institution_representative"))
	{
		inst.POST("/registration", func(c *gin.Context) {
			handlers.HandleInstitutionRegistered(c, ch)
		})
		inst.PATCH("/purchase", func(c *gin.Context) {
			handlers.HandleCreditsPurchased(c, ch)
		})
		inst.GET("/mycredits", func(c *gin.Context) {
			handlers.HandleCreditsAvail(c, ch)
		})
	}

	// Instructor only
	instr := r.Group("/")
	instr.Use(mw.JWTAuthMiddleware(), RoleCheck("instructor"))
	{
		instr.POST("/upload_init", func(c *gin.Context) {
			handlers.UploadExcelInit(c, ch)
		})
		instr.PATCH("/postFinalGrades", func(c *gin.Context) {
			handlers.UploadExcelFinal(c, ch)
		})
		instr.PATCH("/instructor/review-list", func(c *gin.Context) {
			handlers.HandleGetRequestList(c, ch)
		})
		instr.PATCH("/instructor/reply", func(c *gin.Context) {
			handlers.HandlePostResponse(c, ch)
		})
	}

	// Student only
	std := r.Group("/")
	std.Use(mw.JWTAuthMiddleware(), RoleCheck("student"))
	{
		std.POST("/personal/courses", func(c *gin.Context) {
			handlers.HandleGetStudentCourses(c, ch)
		})
		std.POST("/personal/grades", func(c *gin.Context) {
			handlers.HandleGetPersonalGrades(c, ch)
		})
		std.PATCH("/student/reviewRequest", func(c *gin.Context) {
			handlers.HandlePostNewRequest(c, ch)
		})
		std.PATCH("/student/status", func(c *gin.Context) {
			handlers.HandleGetRequestStatus(c, ch)
		})
	}

	// Public User‐management (no JWT)
	{
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
	}

	return r
}

// RoleCheck ensures the JWT role claim matches
func RoleCheck(role string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.GetString("role") != role {
			c.JSON(403, gin.H{"error": "forbidden"})
			c.Abort()
			return
		}
		c.Next()
	}
}
