package middlewares

import (
	"net/http"
	"os"

	"github.com/SharinganAi/recipes-api/handlers"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.GetHeader("X-API-KEY") != os.Getenv("X_API_KEY") {
			c.AbortWithStatus(401)
		}
		c.Next()
	}
}

func AuthMiddlewareNew() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenValue := c.GetHeader("Authorization")
		claims := &handlers.Claims{}
		tkn, err := jwt.ParseWithClaims(tokenValue, claims,
			func(t *jwt.Token) (interface{}, error) {
				return []byte(os.Getenv("JWT_SECRET")), nil
			})
		if err != nil {
			c.AbortWithStatus(http.StatusUnauthorized)
		}
		if tkn == nil || !tkn.Valid {
			c.AbortWithStatus(http.StatusUnauthorized)
		}
		c.Next()
	}
}
