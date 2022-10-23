package controllers

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/Rhaqim/thedutchapp/pkg/auth"
	"github.com/Rhaqim/thedutchapp/pkg/config"
	"github.com/Rhaqim/thedutchapp/pkg/database"
	hp "github.com/Rhaqim/thedutchapp/pkg/helpers"

	"github.com/gin-gonic/gin"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var authCollection = config.AuthCollection

func Signup(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	defer database.ConnectMongoDB().Disconnect(context.TODO())

	var user = hp.CreatUser{}
	var response = hp.MongoJsonResponse{}
	if err := c.BindJSON(&user); err != nil {
		config.Logs("error", err.Error())
		response.Type = "error"
		response.Message = "fullname, username, email, password are required"
		c.JSON(http.StatusBadRequest, response)
		return
	}
	config.Logs("info", "User: "+user.Fullname+" "+user.Username+" "+user.Email)

	checkEmail, err := hp.CheckIfEmailExists(user.Email) // check if email exists
	if err != nil {
		config.Logs("error", err.Error())
		response.Type = "error"
		response.Message = err.Error()
		c.JSON(http.StatusBadRequest, response)
		return
	}
	if checkEmail {
		response.Type = "error"
		response.Message = "Email already exists"
		c.JSON(http.StatusBadRequest, response)
		return
	}

	checkUsername, err := hp.CheckIfUsernameExists(user.Username) // check if username exists
	if err != nil {
		config.Logs("error", err.Error())
		response.Type = "error"
		response.Message = err.Error()
		c.JSON(http.StatusBadRequest, response)
		return
	}
	if checkUsername {
		response.Type = "error"
		response.Message = "Username already exists"
		c.JSON(http.StatusBadRequest, response)
		return
	}

	user.CreatedAt = primitive.NewDateTimeFromTime(time.Now())
	user.UpdatedAt = primitive.NewDateTimeFromTime(time.Now())
	password, err := auth.HashAndSalt(user.Password)
	config.CheckErr(err)
	filter := bson.M{
		"fullname":  user.Fullname,
		"username":  user.Username,
		"avatar":    user.Avatar,
		"email":     user.Email,
		"password":  password,
		"social":    user.Social,
		"role":      user.Role,
		"createdAt": user.CreatedAt,
		"updatedAt": user.UpdatedAt,
	}
	insertResult, err := authCollection.InsertOne(ctx, filter)
	if err != nil {
		config.Logs("error", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	log.Println("insertResult: ", insertResult)

	t, rt, err := auth.GenerateJWT(user.Email, user.Username, insertResult.InsertedID.(primitive.ObjectID))

	if err != nil {
		config.Logs("error", err.Error())
		response.Type = "error"
		response.Message = err.Error()
		c.JSON(http.StatusBadRequest, response)
		return
	}

	err = hp.UpdateRefreshToken(ctx, insertResult.InsertedID.(primitive.ObjectID), rt)
	if err != nil {
		config.Logs("error", err.Error())
		response.Type = "error"
		response.Message = err.Error()
		c.JSON(http.StatusBadRequest, response)
		return
	}

	response.Type = "success"
	response.Message = "User created"
	response.Data = gin.H{
		"token":     t,
		"refresh":   rt,
		"user":      user.Username,
		"expiresAt": time.Now().Add(time.Hour * 24).Unix(),
	}
	c.JSON(http.StatusOK, response)
}

func SignIn(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	defer database.ConnectMongoDB().Disconnect(context.TODO())

	var request = hp.SignIn{}
	var user = hp.UserStruct{}
	var response = hp.MongoJsonResponse{}

	if err := c.BindJSON(&request); err != nil {
		config.Logs("error", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	log.Print("Request ID sent by client:", request.Username)

	filter := bson.M{"username": request.Username}
	if err := usersCollection.FindOne(ctx, filter).Decode(&user); err != nil {
		config.Logs("error", err.Error())
		response.Type = "error"
		response.Message = err.Error()
		c.JSON(http.StatusBadRequest, response)
		return
	}
	log.Println("User: ", user)
	if auth.CheckPasswordHash(request.Password, user.Password) {

		t, rt, err := auth.GenerateJWT(user.Email, user.Username, user.ID)

		if err != nil {
			config.Logs("error", err.Error())
			response.Type = "error"
			response.Message = err.Error()
			c.JSON(http.StatusBadRequest, response)
			return
		}

		response.Type = "success"
		response.Message = "User signed in"
		response.Data = gin.H{
			"token":     t,
			"refresh":   rt,
			"user":      user.Username,
			"expiresAt": time.Now().Add(time.Hour * 24).Unix(),
		}
		c.JSON(http.StatusOK, response)
	} else {
		response.Type = "error"
		response.Message = "Invalid Credentials"
		c.JSON(http.StatusBadRequest, response)
	}
}

func Signout(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	defer database.ConnectMongoDB().Disconnect(context.TODO())

	request := hp.GetUserById{}

	if err := c.BindJSON(&request); err != nil {
		config.Logs("error", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	log.Print("Request ID sent by client:", request.ID)

	id, err := primitive.ObjectIDFromHex(request.ID.Hex())
	if err != nil {
		config.Logs("error", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err = hp.UpdateRefreshToken(ctx, id, "")
	if err != nil {
		config.Logs("error", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var response = hp.MongoJsonResponse{}
	response.Type = "success"
	response.Message = "User signed out"
	c.JSON(http.StatusOK, response)
}
