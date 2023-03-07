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
	"log"
	"os"

	"github.com/SharinganAi/recipes-api/handlers"
	"github.com/SharinganAi/recipes-api/models"
	"github.com/SharinganAi/recipes-api/utils"
	"github.com/gin-gonic/gin"
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
	recipesHandler    handlers.RecipesHandler
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
	recipesHandler = handlers.NewRecipesHandler(ctx, RecipesCollection)
	// insertManyResult, err := collection.InsertMany(ctx, listOfRecipes)
	// if err != nil {
	// 	log.Fatal("Error while inserting recipes in MongoDB:", err)
	// }
	// log.Printf("No of records inserted in MongoDB: %d", insertManyResult)
}

func main() {
	router := gin.Default()
	router.POST("/recipes/", recipesHandler.NewRecipesHandler)
	router.GET("/recipes/", recipesHandler.ListRecipesHandler)
	router.GET("/recipes/:id/", recipesHandler.GetRecipeHandler)
	router.PUT("/recipes/:id/", recipesHandler.UpdateRecipeHandler)
	router.DELETE("/recipes/:id/", recipesHandler.DeleteRecipeHandler)
	router.GET("/recipes/search/", recipesHandler.SearchRecipeHandler)
	router.Run()
}
