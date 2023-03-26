package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"
	"unicode"

	"github.com/SharinganAi/recipes-api/models"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type RecipesHandler struct {
	collection  *mongo.Collection
	redisClient *redis.Client
	ctx         context.Context
}

func NewRecipesHandler(ctx context.Context, collection *mongo.Collection, redisClient *redis.Client) *RecipesHandler {
	return &RecipesHandler{
		collection:  collection,
		ctx:         ctx,
		redisClient: redisClient,
	}
}

// swagger:operation POST /recipes recipes createRecipe
// Creates a new recipe
// ---
// produces:
// - application/json
// responses:
// '200':
//   - description: Successful operation
func (h *RecipesHandler) NewRecipesHandler(c *gin.Context) {
	var recipe models.Recipe
	if err := c.BindJSON(&recipe); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	recipe.ID = primitive.NewObjectID()
	recipe.PublishdAt = time.Now()
	_, err := h.collection.InsertOne(h.ctx, recipe)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	h.redisClient.Del(h.ctx, "recipes")
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
func (h *RecipesHandler) ListRecipesHandler(c *gin.Context) {
	val, err := h.redisClient.Get(h.ctx, "recipes").Result()
	if err == redis.Nil {
		log.Println("request to MongoDb")
		curr, err := h.collection.Find(h.ctx, bson.M{})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}
		defer curr.Close(h.ctx)
		recipes := []models.Recipe{}
		for curr.Next(h.ctx) {
			var recipe models.Recipe
			err = curr.Decode(&recipe)
			if err != nil {
				log.Println("Error decoding recipe in recipe list:", err)
			}
			recipes = append(recipes, recipe)
		}
		data, _ := json.Marshal(recipes)
		h.redisClient.Set(h.ctx, "recipes", string(data), 0)
		c.JSON(http.StatusOK, recipes)
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	} else {
		log.Printf("Request server from redis")
		recipes := []models.Recipe{}
		json.Unmarshal([]byte(val), &recipes)
		c.JSON(http.StatusOK, recipes)
	}
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
func (h *RecipesHandler) GetRecipeHandler(c *gin.Context) {
	id := c.Param("id")
	objId, _ := primitive.ObjectIDFromHex(id)
	res := h.collection.FindOne(h.ctx, bson.M{
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
func (h *RecipesHandler) UpdateRecipeHandler(c *gin.Context) {
	id := c.Param("id")
	var recipe models.Recipe
	if err := c.BindJSON(&recipe); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
	}
	objId, _ := primitive.ObjectIDFromHex(id)
	_, err := h.collection.UpdateOne(h.ctx, bson.M{
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
func (h *RecipesHandler) DeleteRecipeHandler(c *gin.Context) {
	id := c.Param("id")
	objId, _ := primitive.ObjectIDFromHex(id)
	_, err := h.collection.DeleteOne(h.ctx, bson.M{
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
func (h *RecipesHandler) SearchRecipeHandler(c *gin.Context) {
	tag := c.Query("tag")
	cur, err := h.collection.Find(h.ctx, bson.M{"tags": bson.M{"$in": []string{tag, strings.ToUpperSpecial(unicode.SpecialCase{}, tag), strings.ToLower(tag), strings.ToUpper(tag)}}})
	defer func() {
		err = cur.Close(h.ctx)
	}()
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	}
	recipes := []models.Recipe{}
	for cur.Next(h.ctx) {
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
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}

}
