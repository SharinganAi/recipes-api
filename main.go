package main

import (
	"models"

	"github.com/gin-gonic/gin"
)

var (
	recipes []models.Recipe
)

func init() {
	recipes := []models.Recipe{}
}

func NewRecipesHandler(c *gin.Context) {

}

func main() {
	router := gin.Default()
	router.POST("/recipes", NewRecipesHandler)
	router.Run()

}
