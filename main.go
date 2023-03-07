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
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/SharinganAi/recipes-api/models"
	"github.com/SharinganAi/recipes-api/utils"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

var (
	recipes           []models.Recipe
	ctx               context.Context
	err               error
	client            *mongo.Client
	RecipesCollection *mongo.Collection
)

func init() {
	// recipes = []models.Recipe{}
	// file, _ := os.ReadFile("recipes.json")
	// _ = json.Unmarshal([]byte(file), &recipes)
	ctx := context.Background()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(utils.GetMongoURI()))
	if err != nil {
		log.Fatal("Error connecting to MongoDB: ", err)
	}
	if err = client.Ping(context.TODO(), readpref.Primary()); err != nil {
		log.Fatal("Error while pinging MongoDB: ", err)
	}
	log.Println("Connected to MongoDb.")

	// var listOfRecipes []interface{}
	// for _, recipe := range recipes {
	// 	listOfRecipes = append(listOfRecipes, recipe)
	// }
	RecipesCollection = client.Database(os.Getenv("MONGO_DATABASE_NAME")).Collection("recipes")
	// insertManyResult, err := collection.InsertMany(ctx, listOfRecipes)
	// if err != nil {
	// 	log.Fatal("Error while inserting recipes in MongoDB:", err)
	// }
	// log.Printf("No of records inserted in MongoDB: %d", insertManyResult)
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
	recipe.ID = primitive.NewObjectID()
	recipe.PublishdAt = time.Now()
	_, err := RecipesCollection.InsertOne(ctx, recipe)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
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
	cur, err := RecipesCollection.Find(ctx, bson.M{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
	defer cur.Close(ctx)
	recipes := []models.Recipe{}
	for cur.Next(ctx) {
		var recipe *models.Recipe
		err = cur.Decode(&recipe)
		if err != nil {
			log.Println("Error while decoding recipe from database", err)
			continue
		}
		recipes = append(recipes, *recipe)
	}
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
	objId, _ := primitive.ObjectIDFromHex(id)
	res := RecipesCollection.FindOne(ctx, bson.M{
		"_id": objId,
	})
	var recipe *models.Recipe
	err := res.Decode(&recipe)
	if err != nil {
		c.JSON(http.StatusNotFound, []models.Recipe{})
		return
	}
	c.JSON(http.StatusOK, recipe)
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
	objId, _ := primitive.ObjectIDFromHex(id)
	_, err := RecipesCollection.UpdateOne(ctx, bson.M{
		"_id": objId,
	}, bson.D{{Key: "$set", Value: bson.D{
		{Key: "name", Value: recipe.Name},
		{Key: "tags", Value: recipe.Tags},
		{Key: "ingredients", Value: recipe.Ingredients},
		{Key: "instructions", Value: recipe.Instructions},
	}}})
	if err != nil {
		log.Println("Error while updating recipe", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"Message": "Recipe updated successfully"})
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
	objId, _ := primitive.ObjectIDFromHex(id)
	_, err := RecipesCollection.DeleteOne(ctx, bson.M{
		"_id": objId,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
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
	cur, err := RecipesCollection.Find(ctx, bson.M{"tags": bson.M{"$in": []string{tag, strings.Title(tag), strings.ToLower(tag), strings.ToUpper(tag)}}})
	defer cur.Close(ctx)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	}
	recipes := []models.Recipe{}
	for cur.Next(ctx) {
		var recipe *models.Recipe
		err = cur.Decode(&recipe)
		if err != nil {
			log.Println("Error while decoding recipe from database while searching", tag, ": ", err)
			continue
		}
		recipes = append(recipes, *recipe)
	}
	if len(recipes) == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "No recipes found",
		})
	} else {
		c.JSON(http.StatusOK, recipes)
	}
}

func main() {
	router := gin.Default()
	router.POST("/recipes/", NewRecipesHandler)
	router.GET("/recipes/", ListRecipesHandler)
	router.GET("/recipes/:id/", GetRecipeHandler)
	router.PUT("/recipes/:id/", UpdateRecipeHandler)
	router.DELETE("/recipes/:id/", DeleteRecipeHandler)
	router.GET("/recipes/search/", SearchRecipeHandler)
	router.Run()
}
