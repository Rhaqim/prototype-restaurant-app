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

var authCollection = config.UserCollection

func Signup(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	defer database.ConnectMongoDB().Disconnect(context.TODO())

	var user = hp.CreatUser{}

	if err := c.BindJSON(&user); err != nil {
		config.Logs("error", err.Error())
		response := hp.SetError(err, "fullname, username, email, password are required ")
		c.JSON(http.StatusBadRequest, response)
		return
	}
	config.Logs("info", "User: "+user.Fullname+" "+user.Username+" "+user.Email)

	checkEmail, err := hp.CheckIfEmailExists(user.Email) // check if email exists
	if err != nil {
		config.Logs("error", err.Error())
	}
	if checkEmail {
		response := hp.SetError(nil, "Email already exists")
		c.JSON(http.StatusBadRequest, response)
		return
	}

	checkUsername, err := hp.CheckIfUsernameExists(user.Username) // check if username exists
	if err != nil {
		config.Logs("error", err.Error())
	}
	if checkUsername {
		response := hp.SetError(nil, "Username already exists")
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
		"refreshtoken":  user.RefreshToken,
		"emailverified": user.EmailVerified,
		"createdAt":     user.CreatedAt,
		"updatedAt":     user.UpdatedAt,
	}
	insertResult, err := authCollection.InsertOne(ctx, filter)
	if err != nil {
		config.Logs("error", err.Error())
		response := hp.SetError(err, "Error creating user")
		c.JSON(http.StatusBadRequest, response)
		return
	}
	log.Println("insertResult: ", insertResult)

	t, rt, err := auth.GenerateJWT(user.Email, user.Username, insertResult.InsertedID.(primitive.ObjectID))

	if err != nil {
		config.Logs("error", err.Error()+" Error generating token")
		response := hp.SetError(err, "Error creating user")
		c.JSON(http.StatusBadRequest, response)
		return
	}

	err = hp.UpdateRefreshToken(ctx, insertResult.InsertedID.(primitive.ObjectID), rt)
	if err != nil {
		config.Logs("error", err.Error()+" Error updating refresh token")
		response := hp.SetError(err, "Error creating user")
		c.JSON(http.StatusBadRequest, response)
		return
	}

	userResponse := hp.UserResponse{
		ID:        insertResult.InsertedID.(primitive.ObjectID),
		Fullname:  user.Fullname,
		Username:  user.Username,
		Avatar:    user.Avatar,
		Email:     user.Email,
		Role:      user.Role,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}

	data := gin.H{
		"accessToken":  t,
		"refreshToken": rt,
		"user":         userResponse,
		"expiresAt":    time.Now().Add(time.Hour * 24).Unix(),
	}

	response := hp.SetSuccess("User created successfully", data)
	c.JSON(http.StatusOK, response)
}

func SignIn(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	defer database.ConnectMongoDB().Disconnect(context.TODO())

	var request = hp.SignIn{}
	var user = hp.UserStruct{}

	if err := c.BindJSON(&request); err != nil {
		config.Logs("error", err.Error())
		response := hp.SetError(err, "email and password are required ")
		c.JSON(http.StatusBadRequest, response)
		return
	}
	log.Print("Request ID sent by client:", request.Username)

	filter := bson.M{"username": request.Username}
	if err := usersCollection.FindOne(ctx, filter).Decode(&user); err != nil {
		config.Logs("error", err.Error())
		response := hp.SetError(err, "User not found")
		c.JSON(http.StatusBadRequest, response)
		return
	}
	log.Println("User: ", user)
	if auth.CheckPasswordHash(request.Password, user.Password) {

		t, rt, err := auth.GenerateJWT(user.Email, user.Username, user.ID)

		if err != nil {
			config.Logs("error", err.Error()+" Error generating token")
			response := hp.SetError(err, "Error Signing in, please try again later")
			c.JSON(http.StatusBadRequest, response)
			return
		}

		userResponse := hp.UserResponse{
			ID:        user.ID,
			Fullname:  user.Fullname,
			Username:  user.Username,
			Email:     user.Email,
			Avatar:    user.Avatar,
			Role:      user.Role,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
		}

		var data = gin.H{
			"accessToken":  t,
			"refreshToken": rt,
			"user":         userResponse,
			"expiresAt":    time.Now().Add(time.Hour * 24).Unix(),
		}
		response := hp.SetSuccess("User signed in successfully", data)
		c.JSON(http.StatusOK, response)
	} else {
		response := hp.SetError(nil, "Invalid password")
		c.JSON(http.StatusBadRequest, response)
	}
}

func Signout(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	defer database.ConnectMongoDB().Disconnect(context.TODO())

	user := hp.UserResponse{}

	check, ok := c.Get("user") //check if user is logged in
	if !ok {
		response := hp.SetError(nil, "User not logged in")
		c.JSON(http.StatusBadRequest, response)
		return
	}

	request := check.(hp.UserResponse)
	filter := bson.M{"username": request.Username}

	if err := usersCollection.FindOne(ctx, filter).Decode(&user); err != nil {
		config.Logs("error", err.Error())
		response := hp.SetError(err, "User not found")
		c.JSON(http.StatusBadRequest, response)
		return
	}

	err := hp.UpdateRefreshToken(ctx, user.ID, "")
	if err != nil {
		config.Logs("error", err.Error())
		response := hp.SetError(err, "Error signing out")
		c.JSON(http.StatusBadRequest, response)
		return
	}

	response := hp.SetSuccess("User signed out successfully", nil)
	c.JSON(http.StatusOK, response)
}

func RefreshToken(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	defer database.ConnectMongoDB().Disconnect(context.TODO())

	check, ok := c.Get("user") //check if user is logged in
	if !ok {
		response := hp.SetError(nil, "User not logged in")
		c.JSON(http.StatusBadRequest, response)
		return
	}

	request := check.(hp.UserResponse)
	var user = hp.UserResponse{}

	id, err := primitive.ObjectIDFromHex(request.ID.Hex())
	if err != nil {
		config.Logs("error", err.Error())
		response := hp.SetError(err, "Error refreshing token")
		c.JSON(http.StatusBadRequest, response)
		return
	}

	filter := bson.M{"_id": id}
	if err := usersCollection.FindOne(ctx, filter).Decode(&user); err != nil {
		config.Logs("error", err.Error())
		response := hp.SetError(err, "User not found")
		c.JSON(http.StatusBadRequest, response)
		return
	}

	if user.RefreshToken != request.RefreshToken {
		response := hp.SetError(nil, "Invalid refresh token")
		c.JSON(http.StatusBadRequest, response)
		return
	}

	t, rt, err := auth.GenerateJWT(user.Email, user.Username, user.ID)

	if err != nil {
		config.Logs("error", err.Error()+" Error generating token")
		response := hp.SetError(err, "Error refreshing token")
		c.JSON(http.StatusBadRequest, response)
		return
	}

	err = hp.UpdateRefreshToken(ctx, id, rt)
	if err != nil {
		config.Logs("error", err.Error()+" Error updating refresh token")
		response := hp.SetError(err, "Error refreshing token")
		c.JSON(http.StatusBadRequest, response)
		return
	}

	var data = gin.H{
		"token":     t,
		"refresh":   rt,
		"user":      user.Username,
		"expiresAt": time.Now().Add(time.Hour * 24).Unix(),
	}
	response := hp.SetSuccess("Token refreshed successfully", data)
	c.JSON(http.StatusOK, response)
}

func ForgotPassword(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	defer database.ConnectMongoDB().Disconnect(context.TODO())

	request := hp.ForgotPassword{}

	if err := c.BindJSON(&request); err != nil {
		config.Logs("error", err.Error())
		response := hp.SetError(err, "Error binding request")
		c.JSON(http.StatusBadRequest, response)
		return
	}
	log.Print("Request ID sent by client:", request.Email)

	var user = hp.UserResponse{}
	filter := bson.M{"email": request.Email}
	if err := usersCollection.FindOne(ctx, filter).Decode(&user); err != nil {
		config.Logs("error", err.Error()+" User not found")
		response := hp.SetError(err, "User not found")
		c.JSON(http.StatusBadRequest, response)
		return
	}

	t, rt, err := auth.GenerateJWT(user.Email, user.Username, user.ID)
	if err != nil {
		config.Logs("error", err.Error()+" Error generating token")
		response := hp.SetError(err, "Error generating token")
		c.JSON(http.StatusBadRequest, response)
		return
	}

	err = hp.UpdateRefreshToken(ctx, user.ID, rt)
	if err != nil {
		config.Logs("error", err.Error()+" Error updating refresh token")
		response := hp.SetError(err, "Error updating refresh token")
		c.JSON(http.StatusBadRequest, response)
		return
	}

	var data = gin.H{
		"token": t,
	}
	response := hp.SetSuccess("Token generated successfully", data)
	c.JSON(http.StatusOK, response)
}

func ResetPassword(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	defer database.ConnectMongoDB().Disconnect(context.TODO())

	request := hp.ResetPassword{}

	if err := c.BindJSON(&request); err != nil {
		config.Logs("error", err.Error()+" Error binding request")
		response := hp.SetError(err, "Error binding request")
		c.JSON(http.StatusBadRequest, response)
		return
	}
	log.Print("Request ID sent by client:", request.ID)

	id, err := primitive.ObjectIDFromHex(request.ID.Hex())
	if err != nil {
		config.Logs("error", err.Error()+" Error converting id")
		response := hp.SetError(err, "Error converting id")
		c.JSON(http.StatusBadRequest, response)
		return
	}

	var user = hp.UserStruct{}
	filter := bson.M{"_id": id}
	if err := usersCollection.FindOne(ctx, filter).Decode(&user); err != nil {
		config.Logs("error", err.Error()+" User not found")
		response := hp.SetError(err, "User not found")
		c.JSON(http.StatusBadRequest, response)
		return
	}

	if user.RefreshToken != request.RefreshToken {
		config.Logs("error", "Invalid refresh token")
		response := hp.SetError(nil, "Invalid refresh token")
		c.JSON(http.StatusBadRequest, response)
		return
	}

	hashedPassword, err := auth.HashPassword(request.NewPassword)
	if err != nil {
		config.Logs("error", err.Error()+" Error hashing password")
		response := hp.SetError(err, "Error hashing password")
		c.JSON(http.StatusBadRequest, response)
		return
	}

	update := bson.M{
		"$set": bson.M{
			"password": hashedPassword,
		},
	}

	_, err = usersCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		config.Logs("error", err.Error()+" Error updating password")
		response := hp.SetError(err, "Error updating password")
		c.JSON(http.StatusBadRequest, response)
		return
	}

	var response = hp.SetSuccess("Password updated successfully", nil)
	c.JSON(http.StatusOK, response)
}
