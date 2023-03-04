// Recipes API
//
// This is a sample recipes API.
//
//	Schemes: http
//	Host: localhost:8080
//	BasePath: /
//	Version: 1.0.0
//	Contact: Dilip Kumar Singh<dilipkr.18@gmail.com> https://google.com
//
// Consumes:
//   - application/json
//
// Produces:
//   - application/json
//
// swagger: meta
package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
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
	file, _ := os.ReadFile("recipes.json")
	_ = json.Unmarshal([]byte(file), &recipes)
}

// swagger:operation POST /recipes recipes createRecipe
// Creates a new recipe
// ---
// produces:
// - application/json
// responses:
// '200':
//   - description: Successful operation
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

// swagger:operation GET /recipes recipes listRecipes
// Returns list of recipes
// ---
// produces:
// - application/json
// responses:
// '200':
//
//   - description: Successful operation
func ListRecipesHandler(c *gin.Context) {
	c.JSON(http.StatusOK, recipes)
}

// swagger:operation GET /recipes/{id} recipes GetRecipe
// Returns recipe with an ID
// ---
// parameters:
//   - name: id
//     in: path
//     description: ID of the recipe
//     required: true
//     type: string
//
// produces:
// - application/json
// responses:
// '200':
//
//   - description: Successful operation
func GetRecipeHandler(c *gin.Context) {
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

// Update an existing recipe
// ---
// parameters:
//   - name: id
//     in: path
//     description: ID of the recipe
//     required: true
//     type: string
//
// produces:
//   - application/json
//
// responses:
//
//		'200':
//	   description: Successful operation
//	 '400':
//	   description: user error
//	 '404':
//	   description: Recipe id not found
//
//swagger:operation PUT /recipes/{id} recipes updateRecipe
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

// Delete an existing recipe
// ---
// parameters:
//   - name: id
//     in: path
//     description: ID of the recipe
//     required: true
//     type: string
//
// produces:
//   - application/json
//
// responses:
//
//		'200':
//	   description: Successful operation
//	 '400':
//	   description: user error
//	 '404':
//	   description: Recipe id not found
//
//swagger:operation DELETE /recipes/{id} recipes DeleteRecipe
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

// Search recipes based on tag passed as query
// ---
// parameters:
//   - name: tag
//     in: query
//     description: tag of the recipe
//     required: true
//     type: string
//
// produces:
//   - application/json
//
// responses:
//
//		'200':
//	   description: Successful operation
//	 '400':
//	   description: user error
//	 '404':
//	   description: Recipe id not found
//
//swagger:operation GET /recipes/search recipes searchRecipe
func SearchRecipeHandler(c *gin.Context) {
	tag := c.Query("tag")
	recipeList := []models.Recipe{}
	for i, v := range recipes {
		for _, t := range v.Tags {
			if strings.EqualFold(t, tag) {
				recipeList = append(recipeList, recipes[i])
				break
			}
		}
	}
	if len(recipeList) == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "No recipes found",
		})
	} else {
		c.JSON(http.StatusOK, recipeList)
	}
}

func main() {
	router := gin.Default()
	router.POST("/recipes", NewRecipesHandler)
	router.GET("/recipes", ListRecipesHandler)
	router.GET("/recipes/:id", GetRecipeHandler)
	router.PUT("/recipes/:id", UpdateRecipeHandler)
	router.DELETE("/recipes/:id", DeleteRecipeHandler)
	router.GET("/recipes/search/", SearchRecipeHandler)
	router.Run()
}
