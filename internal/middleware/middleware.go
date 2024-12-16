package middleware

import (
	"fmt"
	"net/http"
	"strings"
	"github.com/asifrahaman13/bhagabad_gita/internal/helper"
	"github.com/gin-gonic/gin"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		fmt.Println("Inside the Auth middleware", c.GetHeader("Authorization"))
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			helper.JSONResponse(c, http.StatusUnauthorized, "Unauthorized", nil)
			c.Abort() 
			return
		}
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			helper.JSONResponse(c, http.StatusUnauthorized, "Unauthorized", nil)
			c.Abort() 
			return
		}
		accessToken := parts[1]
		fmt.Println("Access token: ", accessToken)
		userName, err := helper.VerifyToken(accessToken)
		if err != nil {
			helper.JSONResponse(c, http.StatusUnauthorized, "Unauthorized", nil)
			c.Abort() 
			return
		}
		fmt.Println("User email: ", userName)
		c.Set("username", userName)
		c.Next()
	}
}
