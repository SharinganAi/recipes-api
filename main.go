package main

import (
	"net/http"
	"time"

	"github.com/SharinganAi/recipes-api/models"
	"github.com/gin-gonic/gin"
	"github.com/rs/xid"
)

var (
	recipes []models.Recipe
)

func init() {
	recipes = []models.Recipe{}
}

func NewRecipesHandler(c *gin.Context) {
	var recipe models.Recipe
	if err := c.BindJSON(&recipe); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
	}
	recipe.ID = xid.New().String()
	recipe.PublishdAt = time.Now()
	recipes = append(recipes, recipe)
	c.JSON(http.StatusOK, recipe)
}

func main() {
	router := gin.Default()
	router.POST("/recipes", NewRecipesHandler)
	router.Run()
}
