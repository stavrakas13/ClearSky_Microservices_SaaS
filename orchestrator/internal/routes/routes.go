// file: orchestrator/internal/routes/router.go
package routes

import (
	"orchestrator/internal/handlers"
	mw "orchestrator/internal/middleware"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	amqp "github.com/rabbitmq/amqp091-go"
)

func SetupRouter(ch *amqp.Channel) *gin.Engine {
	r := gin.Default()

	// DEVELOPMENT CORS: allow all origins and methods, handle preflight automatically
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	// 16 MiB in‐memory before spilling to /tmp
	r.MaxMultipartMemory = 16 << 20

	// Student only - with proper JWT middleware and student validation
	std := r.Group("/")
	std.Use(mw.JWTAuthMiddleware()) // Add JWT middleware
	std.Use(func(c *gin.Context) {  // Add student role validation
		role := c.GetString("role")
		if role != "student" {
			c.JSON(403, gin.H{"error": "Access restricted to students only"})
			c.Abort()
			return
		}
		c.Next()
	})
	{
		std.GET("/personal/grades", func(c *gin.Context) {
			handlers.HandleGetPersonalGrades(c, ch)
		})
		std.PATCH("/student/reviewRequest", func(c *gin.Context) {
			handlers.HandlePostNewRequest(c, ch)
		})
		std.PATCH("/student/status", func(c *gin.Context) {
			handlers.HandleGetRequestStatus(c, ch)
		})
	}

	// Instructor only - add proper middleware
	instr := r.Group("/")
	instr.Use(mw.JWTAuthMiddleware())
	instr.Use(func(c *gin.Context) {
		role := c.GetString("role")
		if role != "instructor" {
			c.JSON(403, gin.H{"error": "Access restricted to instructors only"})
			c.Abort()
			return
		}
		c.Next()
	})
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
		r.PATCH("/user/change-password", func(c *gin.Context) {
			handlers.HandleUserChangePassword(c, ch)
		})
	}

	// Shared stats endpoints (all roles)
	stats := r.Group("/stats")
	stats.Use(mw.JWTAuthMiddleware()) // Add JWT middleware
	{
		stats.GET("/available", func(c *gin.Context) {
			handlers.HandleSubmissionLogs(c, ch)
		})
		stats.GET("/courses", func(c *gin.Context) {
			handlers.HandleSubmissionLogs(c, ch)
		})
	}

	return r
}
