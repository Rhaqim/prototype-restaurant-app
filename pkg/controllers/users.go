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
var (
	usersCollection = config.UserCollection
	SetUsersCache   = AbstractConnection(setUsersCache)
	GetUsersCache   = AbstractConnection(getUsersCache)
)

func CreatNewUser(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), config.ContextTimeout)
	defer cancel()
	defer database.DisconnectMongoDB()

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
	user.Role = hp.Roles(hp.Roles(user.Role).String())
	password, err := auth.HashAndSalt(user.Password)
	if err != nil {
		// config.Logs("error", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
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
		"created_at":    user.CreatedAt,
		"updated_at":    user.UpdatedAt,
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

func GetUser(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), config.ContextTimeout)
	defer cancel()
	defer database.ConnectMongoDB().Disconnect(context.TODO())

	funcName := ut.GetFunctionName()

	var user hp.UserResponse

	// Get ID from Query
	id := c.Query("id")

	// Get email from Query
	email := c.Query("email")

	// Get username from Query
	username := c.Query("username")

	var filter bson.M

	switch {
	case id != "":
		id, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			response := hp.SetError(err, "Invalid ID", funcName)
			c.AbortWithStatusJSON(http.StatusBadRequest, response)
			return
		}
		filter = bson.M{"_id": id}
	case email != "":
		filter = bson.M{"email": email}
	case username != "":
		filter = bson.M{"username": username}
	default:
		response := hp.SetError(nil, "Invalid Query", funcName)
		c.AbortWithStatusJSON(http.StatusBadRequest, response)
		return
	}

	user, err := hp.GetUser(ctx, filter)
	if err != nil {
		response := hp.SetError(err, "User not found", funcName)
		c.AbortWithStatusJSON(http.StatusBadRequest, response)
		return
	}

	response := hp.SetSuccess("User found", user, funcName)
	c.JSON(http.StatusOK, response)
}

func UpdateUser(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), config.ContextTimeout)
	defer cancel()

	defer database.ConnectMongoDB().Disconnect(context.TODO())

	request := hp.UserUpdate{}

	funcName := ut.GetFunctionName()

	err := c.BindJSON(&request)
	if err != nil {
		response := hp.SetError(err, "Invalid request", funcName)
		c.AbortWithStatusJSON(http.StatusBadRequest, response)
		return
	}

	user, err := hp.GetUserFromToken(c) // get user from token
	if err != nil {
		respons := hp.SetError(err, "User not found", funcName)
		c.AbortWithStatusJSON(http.StatusBadRequest, respons)
		return
	}

	filter := bson.M{"_id": user.ID}
	update := bson.M{
		"$set": request,
	}

	_, err = usersCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		response := hp.SetError(err, "User not found", funcName)
		c.AbortWithStatusJSON(http.StatusBadRequest, response)
		return
	}

	response := hp.SetSuccess("User updated", request, funcName)
	c.JSON(http.StatusOK, response)

}

func DeleteUser(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), config.ContextTimeout)
	defer cancel()

	defer database.ConnectMongoDB().Disconnect(context.TODO())

	funcName := ut.GetFunctionName()

	_id := c.Query("id")

	// Get ID from Query
	id, err := primitive.ObjectIDFromHex(_id)
	if err != nil {
		response := hp.SetError(err, "Invalid ID", funcName)
		c.AbortWithStatusJSON(http.StatusBadRequest, response)
		return
	}

	user, err := hp.GetUserFromToken(c) // get user from token
	if err != nil {
		respons := hp.SetError(err, "User not found", funcName)
		c.AbortWithStatusJSON(http.StatusUnauthorized, respons)
		return
	}

	if user.Role != hp.Admin || user.ID != id {
		respons := hp.SetError(err, "You are not authorized to delete this user", funcName)
		c.AbortWithStatusJSON(http.StatusUnauthorized, respons)
		return
	}

	filter := bson.M{"_id": id}

	deleteResult, err := usersCollection.DeleteOne(ctx, filter)
	if err != nil {
		response := hp.SetError(err, "User not found", funcName)
		c.AbortWithStatusJSON(http.StatusBadRequest, response)
		return
	}

	response := hp.SetSuccess("User deleted", deleteResult, funcName)
	c.JSON(http.StatusOK, response)
}

/* Customer KYC
@params: KYC struct
@returns: success, error
*/

func UpdateUsersKYC(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), config.ContextTimeout)
	defer cancel()

	defer database.ConnectMongoDB().Disconnect(context.TODO())

	funcName := ut.GetFunctionName()

	request := hp.KYC{}

	user, err := hp.GetUserFromToken(c) // get user from token
	if err != nil {
		respons := hp.SetError(err, "User not found", funcName)
		c.AbortWithStatusJSON(http.StatusUnauthorized, respons)
		return
	}

	if err := c.BindJSON(&request); err != nil {
		response := hp.SetError(err, "Error binding JSON", funcName)
		c.AbortWithStatusJSON(http.StatusBadRequest, response)
		return
	}

	request.UpdatedAt = primitive.NewDateTimeFromTime(time.Now())
	request.KYCStatus = hp.KYCStatus(hp.KYCStatus(request.KYCStatus).String())
	request.IdentityType = hp.IdentityType(hp.IdentityType(request.IdentityType).String())
	request.MapInfo.Lat, request.MapInfo.Long, request.MapInfo.PlaceID, _ = hp.GetLatLong(request.Address)

	// check valid year of birth
	okDOB := hp.ValidateKYCDOB(request.DOB)
	if !okDOB {
		response := hp.SetError(err, "Invalid year of birth", funcName)
		c.AbortWithStatusJSON(http.StatusBadRequest, response)
		return
	}

	// update user kyc
	filter := bson.M{"_id": user.ID}
	update := bson.M{
		"$set": request,
	}

	_, err = usersCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		// config.Logs("error", err.Error())
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Send request to KYC service

	response := hp.SetSuccess("success", "User updated", funcName)
	c.JSON(http.StatusOK, response)

}

func setUsersCache(c *gin.Context, ctx context.Context) {
	funcName := ut.GetFunctionName()

	err := hp.SetUsersCache(ctx)
	if err != nil {
		response := hp.SetError(err, "Error setting cache", funcName)
		c.AbortWithStatusJSON(http.StatusBadRequest, response)
		return
	}

	response := hp.SetSuccess("Cache set", "success", funcName)
	c.JSON(http.StatusOK, response)
}

/* Get Users from Cache
@params: none
@returns: success, error
*/

func getUsersCache(c *gin.Context, ctx context.Context) {
	funcName := ut.GetFunctionName()

	users, err := hp.GetUsersFromCache(ctx)
	if err != nil {
		response := hp.SetError(err, "Error getting cache", funcName)
		c.AbortWithStatusJSON(http.StatusBadRequest, response)
		return
	}

	response := hp.SetSuccess("Cache found", users, funcName)
	c.JSON(http.StatusOK, response)
}
