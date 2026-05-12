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

func HashPassword(password string) string{
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	if err != nil{
		log.Panic(err)
	}
	return string(bytes)
}

func VerifyPassword(userPassword string, providedPassword string)(bool,string){
	err := bcrypt.CompareHashAndPassword([]byte(providedPassword),[]byte(userPassword))
	check := true
	msg := ""
	if err != nil{
		msg = fmt.Sprintf("email or password is incorrect")
		check = false
	}
	return check,msg
}

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

		password := HashPassword(*user.Password)
		user.Password = &password

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

func Login() gin.HandlerFunc{
	return func(c *gin.Context){
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		var user models.User

		var foundUser models.User

		if err := c.BindJSON(&user); err != nil{
			c.JSON(http.StatusBadRequest, gin.H{"error":err.Error()})
			return
		}

		err := userCollection.FindOne(ctx, bson.M{"email":user.Email}).Decode(&foundUser)

		if err != nil{
			c.JSON(http.StatusInternalServerError, gin.H{"error":"email or password is incorrect"})
			return
		}

		passwordIsValid, msg := VerifyPassword(*user.Password, *foundUser.Password)
		if !passwordIsValid {
    c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
    return
	}

	if foundUser.Email == nil{
		c.JSON(http.StatusInternalServerError, gin.H{"error":"user not found"})
	}

	token, refreshToken, _ := helper.GenerateAllTokens(*foundUser.Email, *foundUser.FirstName, *foundUser.LastName, *foundUser.UserType, foundUser.UserID)
	helper.UpdateAllTokens(token,refreshToken, foundUser.UserID)
	err = userCollection.FindOne(ctx, bson.M{"user_id":foundUser.UserID}).Decode(&foundUser)
	if err != nil{
		c.JSON(http.StatusInternalServerError, gin.H{"error":err.Error()})
	}
	c.JSON(http.StatusOK, foundUser)

	}
}

func GetUsers() gin.HandlerFunc{
	return func(c *gin.Context){
		if err := helper.CheckUserType(c, "ADMIN"); err != nil{
			c.JSON(http.StatusBadRequest, gin.H{"error":err.Error()})
			return
		}
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		recordPerPage, err := strconv.Atoi(c.Query("recordPerPage"))
		if err != nil || recordPerPage < 1{
			recordPerPage = 10
		}
		page, err1 := strconv.Atoi(c.Query("page"))
		if err1 != nil || page < 1{
			page = 1
		}

		startIndex := (page - 1)*recordPerPage

		matchStage := bson.D{{"$match", bson.D{{}}}}
		groupStage := bson.D{{"$group", bson.D{
				{"_id", bson.D{{"_id", "null"}}},
				{"total_count", bson.D{{"$sum", 1}}},
				{"data", bson.D{{"$push", "$$ROOT"}}},
		}}}
	}
}

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