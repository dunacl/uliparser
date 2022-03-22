package controller

import (
	"github.com/gin-gonic/gin"
)

func CheckHeaders(c *gin.Context) bool {
	//log.Print("Header ->", c.Request.Header["User-Agent"])
	return false
}
