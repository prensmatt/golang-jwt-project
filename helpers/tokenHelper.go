package helpers

import(
	"context"
	"fmt"
	"log"
	"os"
	"time"
	"golang-jwt-project/database"
	jwt "github.com/dgrijalva/jwt-go"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type SignedDetails struct{
	Email string
	FirstName string
	LastName string
	Uid string
	UserType string
	jwt.StandardClaims
}

var userCollection *mongo.Collection = database.OpenCollection(database.Client,"user")

var SECRET_KEY string = os.Getenv("SECRET_KEY")


func GenerateAllTokens(email, firstName, lastName, userType, uid string)(signedToken string, signedRefreshToken string, err error){
	claims := &SignedDetails{
		Email:email,
		FirstName:firstName,
		LastName:lastName,
		Uid:uid,
		UserType:userType,
		StandardClaims:jwt.StandardClaims{
			ExpiresAt:time.Now().Local().Add(time.Hour*time.Duration(24)).Unix(),
		},
	}

	refreshClaims := &SignedDetails{
		StandardClaims:jwt.StandardClaims{
			ExpiresAt: time.Now().Local().Add(time.Duration(168)).Unix(),
		},
	}

	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(SECRET_KEY))
	refreshToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims).SignedString([]byte(SECRET_KEY))
	if err != nil{
		log.Panic(err)
		return
	}

	return token, refreshToken, err
}