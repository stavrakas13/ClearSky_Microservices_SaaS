package controllers

import (
	"github.com/gin-gonic/gin"
)

func GetAvailCourcesList(c *gin.Context) {
	c.JSON(200, "MY COURCES")

}
