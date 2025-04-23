package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func PostReply(c *gin.Context) {

	c.JSON(http.StatusOK, "Reply Sent!")
}
