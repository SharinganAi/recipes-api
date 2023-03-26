package handlers

import (
	"context"
	"crypto/sha256"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/SharinganAi/recipes-api/models"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

type AuthHandler struct {
	collection *mongo.Collection
	ctx        context.Context
}

func NewAuthHandler(ctx context.Context, collection *mongo.Collection) *AuthHandler {
	return &AuthHandler{
		collection: collection,
		ctx:        ctx,
	}
}

type Claims struct {
	UserName string `json:"user_name"`
	jwt.StandardClaims
}

type JWTOutput struct {
	Token   string    `json:"token"`
	Expires time.Time `json:"expires"`
}

// swagger:operation POST /signin auth signIn
// Login with username and password
// ---
// produces:
// - application/json
// responses:
//
//	 '200':
//
//		 description: Successful operation
//
//	 '401':
//
//		 description: Unauthorized
func (handler *AuthHandler) SignInHandler(c *gin.Context) {
	var user models.User
	var userResponse models.UserResponse
	//check parameters binding in request object
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	cur := handler.collection.FindOne(handler.ctx, bson.M{
		"user_name": user.UserName,
	})
	if cur.Err() != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": cur.Err().Error(),
		})
		return
	}
	err := cur.Decode(&userResponse)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	err = bcrypt.CompareHashAndPassword([]byte(userResponse.Password), []byte(user.Password))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid password"})
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

// swagger:operation POST /refresh auth refresh
// Get new token in exchange for an old one
// ---
// produces:
// - application/json
// responses:
//
//	 '200':
//
//		 description: Successful operation
//
//	 '400':
//
//		 description: Token is new and doesn't need a refresh
//
//	 '401':
//
//		 description: Invalid credentials
func (handler *AuthHandler) Refreshhandler(c *gin.Context) {
	tokenValue := c.GetHeader("Authorization")
	claims := &Claims{}
	tkn, err := jwt.ParseWithClaims(tokenValue, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("JWT_SECRET")), nil
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if tkn == nil || !tkn.Valid {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
		return
	}
	if time.Until(time.Unix(claims.ExpiresAt, 0)) > 30*time.Second {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Token is not expired yet"})
	}
	expirationTime := time.Now().Add(15 * time.Minute)
	claims.ExpiresAt = expirationTime.Unix()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
	jwtOutput := JWTOutput{
		Token:   tokenString,
		Expires: expirationTime,
	}
	c.JSON(http.StatusOK, jwtOutput)
}

// swagger:operation POST /signup auth signup
// signup a new user
// ---
// produces:
// - application/json
// responses:
//
//	'200':
//
//	 description: Successful operation
//
//	'400':
//
//	 description: User Id already exists
func (handler *AuthHandler) SignupHandler(c *gin.Context) {
	var user models.User
	//check parameters binding in request object
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	user.UserName = strings.ToLower(user.UserName)
	cur := handler.collection.FindOne(handler.ctx, bson.M{
		"user_name": user.UserName,
	})

	if cur.Err() != nil {
		if strings.Trim(cur.Err().Error(), " ") != "mongo: no documents in result" {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": cur.Err().Error(),
			})
			return
		} else {
			hashed, _ := bcrypt.GenerateFromPassword([]byte(user.Password), 8)
			h := sha256.New()
			handler.collection.InsertOne(handler.ctx, bson.M{
				"user_name": user.UserName,
				"password":  string(h.Sum([]byte(hashed))),
			})

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
			return
		}
	}
	c.JSON(http.StatusBadRequest, gin.H{"error": "User id already exist"})
}
