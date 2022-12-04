package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
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
	file, _ := ioutil.ReadFile("recipes.json")
	_ = json.Unmarshal([]byte(file), &recipes)
}

func NewRecipesHandler(c *gin.Context) {
	var recipe models.Recipe
	if err := c.BindJSON(&recipe); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	recipe.ID = xid.New().String()
	recipe.PublishdAt = time.Now()
	recipes = append(recipes, recipe)
	c.JSON(http.StatusOK, recipe)
}

func ListRecipesHandler(c *gin.Context) {
	c.JSON(http.StatusOK, recipes)
}

func GetRecipesHandler(c *gin.Context) {
	id := c.Param("id")
	index := -1
	for i, val := range recipes {
		if val.ID == id {
			index = i
		}
	}
	if index == -1 {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "recipe not found",
		})
		return
	}
	c.JSON(http.StatusOK, recipes[index])
}

// update an existing recipe. check first if Recipe.ID is already present in recipes list.
// if not present return error else update the the recipe object
func UpdateRecipeHandler(c *gin.Context) {
	id := c.Param("id")
	var recipe models.Recipe
	if err := c.BindJSON(&recipe); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
	}
	index := -1
	for i, val := range recipes {
		if val.ID == id {
			index = i
		}
	}
	if index == -1 {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "recipe not found",
		})
		return
	}
	recipes[index] = recipe
	c.JSON(http.StatusOK, recipe)
}

func DeleteRecipeHandler(c *gin.Context) {
	id := c.Param("id")
	index := -1
	for i, val := range recipes {
		if val.ID == id {
			index = i
		}
	}
	if index == -1 {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "recipe not found",
		})
		return
	}
	recipes = append(recipes[:index], recipes[index+1:]...)
	c.JSON(http.StatusOK, gin.H{
		"Message": fmt.Sprintf("recipe with id %s deleted", id),
	})
}

func main() {
	router := gin.Default()
	router.POST("/recipes", NewRecipesHandler)
	router.GET("/recipes", ListRecipesHandler)
	router.GET("/recipes/:id", GetRecipesHandler)
	router.PUT("/recipes/:id", UpdateRecipeHandler)
	router.DELETE("/recipes/:id", DeleteRecipeHandler)
	router.Run()
}
