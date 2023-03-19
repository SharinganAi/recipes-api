package handlers

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/SharinganAi/recipes-api/models"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
}

type Claims struct {
	UserName string `json:"user_name"`
	jwt.StandardClaims
}

type JWTOutput struct {
	Token   string    `json:"token"`
	Expires time.Time `json:"expires"`
}

func (handler *AuthHandler) SignInHandler(c *gin.Context) {
	var user models.User
	//check parameters binding in request object
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	//check creedentials
	if user.UserName != "Admin" || user.Password != "password" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
		return
	}
	//generate JWT token
	expirationTime := time.Now().Add(15 * time.Minute)
	claims := &Claims{
		UserName: user.UserName,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	fmt.Println("JWT secret generated:", os.Getenv("JWT_SECRET"))
	tokenString, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	JWTOutput := JWTOutput{
		Token:   tokenString,
		Expires: expirationTime,
	}
	c.JSON(http.StatusOK, JWTOutput)
}
