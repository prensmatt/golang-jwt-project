package controllers

import(
	"context"
	"fmt"
	"log"
	"strconv"
	"net/http"
	"time"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"golang-jwt-project/models"
	helper "golang-jwt-project/helpers"
	"golang-jwt-project/database"
	"golang.org/x/crypto/bcrypt"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var userCollection *mongo.Collection = database.OpenCollection(database.Client, "user")

var validate = validator.New()

func HashPassword()

func VerifyPassword()

func Signup() gin.HandlerFunc{
	return func(c *gin.Context){
		var ctx,cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()
		var user models.User
		if err := c.BindJSON(&user); err != nil{
			c.JSON(http.StatusBadRequest, gin.H{"error":err.Error()})
			return
		}
		validationErr := validate.Struct(user)
		if validationErr != nil{
			c.JSON(http.StatusBadRequest, gin.H{"error":validationErr.Error()})
			return
		}

		count, err := userCollection.CountDocuments(ctx, bson.M{"email":user.Email})
		if err != nil{
			log.Panic(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error":"error occurred while checking for email"})
		}
		count, err = userCollection.CountDocuments(ctx, bson.M{"phone":user.Phone})
		if err != nil{
			log.Panic(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error":"error occurred while checking for phone number"})
		}

		if count > 0{
			c.JSON(http.StatusInternalServerError, gin.H{"error":"this email or phone number already exist"})
		}

		user.CreatedAt, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		user.UpdatedAt, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		user.ID = primitive.NewObjectID()
		user.UserID = user.ID.Hex()

		token, refreshToken, _ := helper.GenerateAllTokens(*user.Email, *user.FirstName, *user.LastName, *user.UserType, user.UserID)
		
		user.Token = &token
		user.RefreshToken = &refreshToken

		resultInsertionNumber, insertErr := userCollection.InsertOne(ctx, user)
		if insertErr != nil{
			msg := fmt.Sprintf("user item was not created")
			c.JSON(http.StatusInternalServerError, gin.H{"error":msg})
			return
		}
		c.JSON(http.StatusOK, resultInsertionNumber)
	}
}

func Login()

func GetUsers()

func GetUser() gin.HandlerFunc{
	return func(c *gin.Context){
		userId := c.Param("user_id")

		if err := helper.MatchUserTypeToUid(c, userId); err != nil{
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		var ctx,cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()
		var user models.User

		err := userCollection.FindOne(ctx,bson.M{"user_id":userId}).Decode(&user)

		if err != nil{
			c.JSON(http.StatusInternalServerError,gin.H{"error":err.Error()})
			return
		}

		c.JSON(http.StatusOK, user)
	}
}