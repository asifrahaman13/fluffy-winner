package helper

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func Response(c *gin.Context, code int, message string, body interface{}) {
	if message == "" {
		message = http.StatusText(code)
	}
	c.JSON(
		code,
		gin.H{
			"code":       code,
			"statusText": http.StatusText(code),
			"data":       message,
			"body":       body,
		},
	)
}

func JSONResponse(c *gin.Context, code int, message interface{}, body interface{}) {
	c.JSON(
		code,
		gin.H{
			"code": code,
			"data": message,
		},
	)
}
