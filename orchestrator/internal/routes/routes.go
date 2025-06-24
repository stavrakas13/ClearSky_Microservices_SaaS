// orchestrator/internal/routes/routes.go
package routes

import (
	"orchestrator/internal/handlers"
	mw "orchestrator/internal/middleware"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	amqp "github.com/rabbitmq/amqp091-go"
)

// SetupRouter configures all HTTP endpoints and returns the Gin engine.
func SetupRouter(ch *amqp.Channel) *gin.Engine {
	r := gin.Default()

	// Allow CORS in development
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	r.MaxMultipartMemory = 16 << 20 // 16 MiB

	// ────────────────────────────────────────────────────────────────────────
	//  Public endpoints (no JWT)
	// ────────────────────────────────────────────────────────────────────────
	{
		r.POST("/user/register", func(c *gin.Context) { handlers.HandleUserRegister(c, ch) })
		r.POST("/user/login", func(c *gin.Context) { handlers.HandleUserLogin(c, ch) })
		r.DELETE("/user/delete", func(c *gin.Context) { handlers.HandleUserDelete(c, ch) })
		r.POST("/user/google-login", func(c *gin.Context) { handlers.HandleUserGoogleLogin(c, ch) })
		r.PATCH("/user/change-password", func(c *gin.Context) { handlers.HandleUserChangePassword(c, ch) })
		r.GET("/institutions", func(c *gin.Context) {
			handlers.GetInstitutions(c)
		})
		// NEW: purchase credits endpoint
		// front-end does: PATCH /purchase { name, amount }

	}

	repr := r.Group("/")
	repr.Use(mw.JWTAuthMiddleware())
	repr.Use(func(c *gin.Context) {
		if c.GetString("role") != "institution_representative" {
			c.JSON(403, gin.H{"error": "Access restricted to tuinstitution_representative only"})
			c.Abort()
			return
		}
		c.Next()
	})
	{
		repr.PATCH("/purchase", func(c *gin.Context) {
			handlers.HandleCreditsPurchased(c, ch)
		})
		repr.GET("/mycredits", func(c *gin.Context) {
			handlers.HandleCreditsAvail(c, ch)
		})
		repr.POST("/registration", func(c *gin.Context) {
			handlers.HandleInstitutionRegistered(c, ch)
		})
	}
	// ────────────────────────────────────────────────────────────────────────
	//  Student‐only endpoints
	// ────────────────────────────────────────────────────────────────────────
	std := r.Group("/")
	std.Use(mw.JWTAuthMiddleware())
	std.Use(func(c *gin.Context) {
		if c.GetString("role") != "student" {
			c.JSON(403, gin.H{"error": "Access restricted to students only"})
			c.Abort()
			return
		}
		c.Next()
	})
	{
		std.GET("/personal/grades", func(c *gin.Context) { handlers.HandleGetPersonalGrades(c, ch) })
		std.PATCH("/student/reviewRequest", func(c *gin.Context) { handlers.HandlePostNewRequest(c, ch) })
		std.PATCH("/student/status", func(c *gin.Context) { handlers.HandleGetRequestStatus(c, ch) })
	}

	// ────────────────────────────────────────────────────────────────────────
	//  Instructor‐only endpoints
	// ────────────────────────────────────────────────────────────────────────
	instr := r.Group("/")
	instr.Use(mw.JWTAuthMiddleware())
	instr.Use(func(c *gin.Context) {
		if c.GetString("role") != "instructor" {
			c.JSON(403, gin.H{"error": "Access restricted to instructors only"})
			c.Abort()
			return
		}
		c.Next()
	})
	{
		instr.POST("/upload_init", func(c *gin.Context) { handlers.UploadExcelInit(c, ch) })
		instr.PATCH("/postFinalGrades", func(c *gin.Context) { handlers.UploadExcelFinal(c, ch) })
		instr.PATCH("/instructor/review-list", func(c *gin.Context) { handlers.HandleGetRequestList(c, ch) })
		instr.PATCH("/instructor/reply", func(c *gin.Context) { handlers.HandlePostResponse(c, ch) })
	}

	// ────────────────────────────────────────────────────────────────────────
	//  Shared stats endpoints (all roles)
	// ────────────────────────────────────────────────────────────────────────
	stats := r.Group("/stats")
	stats.Use(mw.JWTAuthMiddleware())
	{
		stats.GET("/available", func(c *gin.Context) { handlers.HandleSubmissionLogs(c, ch) })
		stats.GET("/courses", func(c *gin.Context) { handlers.HandleSubmissionLogs(c, ch) })
		stats.POST("/distributions", handlers.HandleGetDistributions(ch))
	}

	return r
}
