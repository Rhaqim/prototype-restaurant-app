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
	ut "github.com/Rhaqim/thedutchapp/pkg/utils"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

/*
Get User by ID
*/
var usersCollection = config.UserCollection

func CreatNewUser(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	defer database.ConnectMongoDB().Disconnect(context.TODO())

	var user = hp.CreatUser{}
	var response = hp.MongoJsonResponse{
		Date: time.Now(),
	}
	if err := c.BindJSON(&user); err != nil {
		// config.Logs("error", err.Error())
		response.Type = "error"
		response.Message = "fullname, username, email, password are required"
		c.JSON(http.StatusBadRequest, response)
		return
	}
	// config.Logs("info", "User: "+user.Fullname+" "+user.Username+" "+user.Email)

	checkEmail, err := hp.CheckIfEmailExists(user.Email) // check if email exists
	if err != nil {
		// config.Logs("error", err.Error())
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
		// config.Logs("error", err.Error())
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

	ok := hp.RoleIsValid(user.Role)

	if !ok {
		user.Role = "user"
	}

	filter := bson.M{
		"fullname":      user.Fullname,
		"username":      user.Username,
		"avatar":        user.Avatar,
		"email":         user.Email,
		"password":      password,
		"social":        user.Social,
		"role":          user.Role,
		"refreshToken":  user.RefreshToken,
		"emailverified": user.EmailVerified,
		"createdAt":     user.CreatedAt,
		"updatedAt":     user.UpdatedAt,
	}
	insertResult, err := usersCollection.InsertOne(ctx, filter)
	if err != nil {
		// config.Logs("error", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	log.Println("insertResult: ", insertResult)
	response.Type = "success"
	response.Message = "User created"
	c.JSON(http.StatusOK, response)
}

func GetUserByID(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	defer database.ConnectMongoDB().Disconnect(context.TODO())

	var user bson.M
	request := hp.GetUserById{}
	var response = hp.MongoJsonResponse{}

	if err := c.BindJSON(&request); err != nil {
		// config.Logs("error", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	log.Print("Request ID sent by client:", request.ID)

	id, err := primitive.ObjectIDFromHex(request.ID.Hex())
	config.CheckErr(err)

	// config.Logs("info", "ID: "+id.Hex())

	filter := bson.M{"_id": id}
	if err := usersCollection.FindOne(ctx, filter).Decode(&user); err != nil {
		// config.Logs("error", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response.Type = "success"
	response.Data = user
	response.Message = "User found"

	c.JSON(http.StatusOK, response)
}

func GetUserByEmail(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	defer database.ConnectMongoDB().Disconnect(context.TODO())

	var user bson.M
	request := hp.GetUserByEmailStruct{}
	var response = hp.MongoJsonResponse{}

	if err := c.BindJSON(&request); err != nil {
		// config.Logs("error", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	log.Print("Request ID sent by client:", request.Email)

	// config.Logs("info", "Email: "+request.Email)

	filter := bson.M{"email": request.Email}
	if err := usersCollection.FindOne(ctx, filter).Decode(&user); err != nil {
		// config.Logs("error", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response.Type = "success"
	response.Data = user
	response.Message = "User found"

	c.JSON(http.StatusOK, response)
}

func UpdateAvatar(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	defer database.ConnectMongoDB().Disconnect(context.TODO())

	request := hp.UpdateUserAvatar{}
	response := hp.MongoJsonResponse{}

	if err := c.BindJSON(&request); err != nil {
		// config.Logs("error", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	log.Print("Request ID sent by client:", request.ID)

	id, err := primitive.ObjectIDFromHex(request.ID.Hex())
	config.CheckErr(err)

	// config.Logs("info", "ID: "+id.Hex())

	request.UpdatedAt = primitive.NewDateTimeFromTime(time.Now())

	filter := bson.M{"_id": id}

	update := bson.M{
		"$set": bson.M{
			"avatar":    request.Avatar,
			"updatedAt": request.UpdatedAt,
		},
	}

	updateResult, err := usersCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		// config.Logs("error", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	log.Println("updateResult: ", updateResult)
	response.Type = "success"
	response.Message = "User updated"
	c.JSON(http.StatusOK, response)

}

func DeleteUser(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	defer database.ConnectMongoDB().Disconnect(context.TODO())

	request := hp.GetUserById{}
	response := hp.MongoJsonResponse{}

	if err := c.BindJSON(&request); err != nil {
		// config.Logs("error", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	log.Print("Request ID sent by client:", request.ID)

	id, err := primitive.ObjectIDFromHex(request.ID.Hex())
	config.CheckErr(err)

	// config.Logs("info", "ID: "+id.Hex())

	filter := bson.M{"_id": id}

	updateResult, err := usersCollection.DeleteOne(ctx, filter)
	if err != nil {
		// config.Logs("error", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	log.Println("updateResult: ", updateResult)
	response.Type = "success"
	response.Message = "User updated"
	c.JSON(http.StatusOK, response)

}

func UpdateUsersKYC(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	defer database.ConnectMongoDB().Disconnect(context.TODO())

	funcName := ut.GetFunctionName()

	request := hp.KYC{}

	user, err := hp.GetUserFromToken(c) // get user from token
	if err != nil {
		respons := hp.SetError(err, "User not found", funcName)
		c.JSON(http.StatusBadRequest, respons)
		return
	}

	if err := c.BindJSON(&request); err != nil {
		response := hp.SetError(err, "Error binding JSON", funcName)
		c.JSON(http.StatusBadRequest, response)
		return
	}

	request.UpdatedAt = primitive.NewDateTimeFromTime(time.Now())

	// update user kyc
	filter := bson.M{"_id": user.ID}
	update := bson.M{
		"$set": request,
	}

	updateResult, err := usersCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		// config.Logs("error", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	log.Println("updateResult: ", updateResult)
	response := hp.SetSuccess("success", "User updated", funcName)
	c.JSON(http.StatusOK, response)

}
