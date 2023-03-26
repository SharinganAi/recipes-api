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
// SecurityDefinitions:
// api_key:
//
//	 type: apiKey
//		name: Authorization
//		in: header
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
	"log"
	"os"

	"github.com/SharinganAi/recipes-api/handlers"
	"github.com/SharinganAi/recipes-api/middlewares"
	"github.com/SharinganAi/recipes-api/utils"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

var (
	RecipesCollection *mongo.Collection
	recipesHandler    *handlers.RecipesHandler
	authHandler       *handlers.AuthHandler
	UsersCollection   *mongo.Collection
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
	log.Println("Connecting to Redis")
	redisAddr := fmt.Sprintf("%s:%s", os.Getenv("REDIS_SERVER"), os.Getenv("REDIS_PORT"))
	redisClient := redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: "",
		DB:       0,
	})
	status := redisClient.Ping(ctx)
	fmt.Println("Redis connection status", status)
	// var listOfRecipes []interface{}
	// for _, recipe := range recipes {
	// 	listOfRecipes = append(listOfRecipes, recipe)
	// }
	RecipesCollection = client.Database(os.Getenv("MONGO_DATABASE_NAME")).Collection("recipes")
	UsersCollection = client.Database(os.Getenv("MONGO_DATABASE_NAME")).Collection("users")
	recipesHandler = handlers.NewRecipesHandler(ctx, RecipesCollection, redisClient)
	authHandler = handlers.NewAuthHandler(ctx, UsersCollection)
	// insertManyResult, err := collection.InsertMany(ctx, listOfRecipes)
	// if err != nil {
	// 	log.Fatal("Error while inserting recipes in MongoDB:", err)
	// }
	// log.Printf("No of records inserted in MongoDB: %d", insertManyResult)
}

func main() {
	router := gin.Default()
	prefixRouter := router.Group("/api/v0")
	//Make endpoints with authorized credentials related to recipes and defined prefix group
	prefixRouter.GET("/recipes/", recipesHandler.ListRecipesHandler)
	prefixRouter.POST("/signin/", authHandler.SignInHandler)
	prefixRouter.POST("/signup/", authHandler.SignupHandler)
	authorized := prefixRouter.Group("/")
	authorized.Use(middlewares.AuthMiddlewareNew())
	authorized.POST("recipes/", recipesHandler.NewRecipesHandler)
	authorized.GET("recipes/:id/", recipesHandler.GetRecipeHandler)
	authorized.PUT("recipes/:id/", recipesHandler.UpdateRecipeHandler)
	authorized.DELETE("recipes/:id/", recipesHandler.DeleteRecipeHandler)
	authorized.GET("recipes/search/", recipesHandler.SearchRecipeHandler)
	authorized.POST("refresh/", authHandler.Refreshhandler)

	//Run the web server
	if os.Getenv("IS_SSL") == "true" {
		port := os.Getenv("SSL_PORT")
		router.RunTLS(":"+port, "certs/localhost.crt", "certs/localhost.key")
	} else {
		port := os.Getenv("SERVER_PORT")
		router.Run(":" + port)
	}
}
