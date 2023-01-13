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

var authCollection = config.UserCollection

/*
	Signup

Creates a new user
Checks if email and username already exists
Hashes password
Checks if the role is valid otherwise assigns it to user
Generates a JWT access token and refresh token
Stores the refresh token in the database
Sends a verification code to the email
Sends the access token and refresh token in the response
*/
var Signup = AbstractConnection(signUp)

func signUp(c *gin.Context, ctx context.Context) {
	var funcName = ut.GetFunctionName()

	var user = hp.UserStruct{}

	if err := c.BindJSON(&user); err != nil {
		response := hp.SetError(err, "Error Validating request", funcName)
		c.AbortWithStatusJSON(http.StatusBadRequest, response)
		return
	}

	checkEmail, err := hp.CheckIfEmailExists(user.Email) // check if email exists
	if err != nil {
		response := hp.SetError(err, ", Error checking email", funcName)
		c.AbortWithStatusJSON(http.StatusInternalServerError, response)
		return
	}
	if checkEmail {
		response := hp.SetError(nil, ", Email already exists", funcName)
		c.AbortWithStatusJSON(http.StatusBadRequest, response)
		return
	}

	checkUsername, err := hp.CheckIfUsernameExists(user.Username) // check if username exists
	if err != nil {
		response := hp.SetError(err, "Error checking username", funcName)
		c.AbortWithStatusJSON(http.StatusInternalServerError, response)
		return
	}
	if checkUsername {
		response := hp.SetError(nil, "Username already exists", funcName)
		c.AbortWithStatusJSON(http.StatusBadRequest, response)
		return
	}

	// Set default values
	user.Transactions = []hp.Transactions{}
	user.Role = hp.Roles(hp.Roles(user.Role).String())
	user.EmailVerified = false
	user.CreatedAt, user.UpdatedAt = hp.CreatedAtUpdatedAt()
	user.Account.CreatedAt, user.Account.UpdatedAt = hp.CreatedAtUpdatedAt()

	// Hash password
	password, err := auth.HashAndSalt(user.Password)
	if err != nil {
		response := hp.SetError(err, "Error hashing password", funcName)
		c.AbortWithStatusJSON(http.StatusInternalServerError, response)
		return
	}
	user.Password = password

	// Add User to database
	insertResult, err := authCollection.InsertOne(ctx, user)
	if err != nil {
		response := hp.SetError(err, "Error creating user", funcName)
		c.AbortWithStatusJSON(http.StatusInternalServerError, response)
		return
	}

	// Generate JWT
	t, rt, err := auth.GenerateJWT(user.Email, user.Username, insertResult.InsertedID.(primitive.ObjectID))

	if err != nil {
		response := hp.SetError(err, "Error creating user", funcName)
		c.AbortWithStatusJSON(http.StatusInternalServerError, response)
		return
	}

	// Store refresh token in database
	err = hp.UpdateRefreshToken(ctx, insertResult.InsertedID.(primitive.ObjectID), rt)
	if err != nil {
		response := hp.SetError(err, "Error updating refresh token", funcName)
		c.AbortWithStatusJSON(http.StatusInternalServerError, response)
		return
	}

	// Send email verification link
	if err = hp.SendVerificationEmail(ctx, user.Email); err != nil {
		response := hp.SetError(err, "Error sending email verification email", funcName)
		c.AbortWithStatusJSON(http.StatusInternalServerError, response)
		return
	}

	// Remove verification code from database after 5 minutes
	go func() {
		time.Sleep(5 * time.Minute)
		c := context.Background()
		err := hp.RemoveVerificationCode(c, user.Email)
		if err != nil {
			hp.SetDebug("Error removing verification code: "+err.Error(), funcName)
		}
	}()

	userResponse := user

	userResponse.ID = insertResult.InsertedID.(primitive.ObjectID)
	userResponse.Password = ""

	data := gin.H{
		"accessToken":  t,
		"refreshToken": rt,
		"user":         userResponse,
		"expires_at":   config.AccessTokenExpireTime.Unix(),
	}

	response := hp.SetSuccess("User created successfully", data, funcName)
	c.JSON(http.StatusOK, response)
}

/*
	Verify email

Gets the token and email from the query
Verifies the email with the token sent to the email
*/
var VerifyEmail = AbstractConnection(verifyEmail)

func verifyEmail(c *gin.Context, ctx context.Context) {
	var funcName = ut.GetFunctionName()

	token := c.Query("token")
	email := c.Query("email")

	if token == "" || email == "" {
		response := hp.SetError(nil, "Token and email are required", funcName)
		c.AbortWithStatusJSON(http.StatusBadRequest, response)
		return
	}

	err := hp.VerifyEmail(ctx, email, token)
	if err != nil {
		response := hp.SetError(err, "Error verifying email", funcName)
		c.AbortWithStatusJSON(http.StatusInternalServerError, response)
		return
	}

	response := hp.SetSuccess("Email verified successfully", nil, funcName)
	c.JSON(http.StatusOK, response)
}

/*
	Signin
Signs in a user
Uses the username and password to sign in a user
Gets the user from the database
Checks the password with the one in the database
Generates a JWT and refresh token
Sends the JWT and refresh token to the client
Stores the refresh token in the database
Checks if the user has verified their email otherwise sends a verification email
Returns the user data and the JWT and refresh token
*/

var SignIn = AbstractConnection(signIn)

func signIn(c *gin.Context, ctx context.Context) {
	funcName := ut.GetFunctionName()

	var request = hp.SignIn{}
	var user = hp.UserStruct{}

	if err := c.BindJSON(&request); err != nil {
		response := hp.SetError(err, "email and password are required ", funcName)
		c.AbortWithStatusJSON(http.StatusBadRequest, response)
		return
	}
	log.Print("Request ID sent by client:", request.Username)

	filter := bson.M{"username": request.Username}

	err := authCollection.FindOne(ctx, filter).Decode(&user)
	if err != nil {
		response := hp.SetError(err, "Error finding user", funcName)
		c.AbortWithStatusJSON(http.StatusBadRequest, response)
		return
	}

	if auth.CheckPasswordHash(request.Password, user.Password) {

		t, rt, err := auth.GenerateJWT(user.Email, user.Username, user.ID)

		if err != nil {
			response := hp.SetError(err, " Error generating token: Error Signing in, please try again later", funcName)
			c.AbortWithStatusJSON(http.StatusInternalServerError, response)
			return
		}

		err = hp.UpdateRefreshToken(ctx, user.ID, rt)
		if err != nil {
			response := hp.SetError(err, "Error updating refresh token", funcName)
			c.AbortWithStatusJSON(http.StatusInternalServerError, response)
			return
		}

		// Check if email is verified
		if user.Role != hp.Admin {
			if !user.EmailVerified {
				err := hp.SendVerificationEmail(ctx, user.Email)
				if err != nil {
					response := hp.SetError(err, "Error sending email verification email", funcName)
					c.AbortWithStatusJSON(http.StatusInternalServerError, response)
					return
				}

				// Remove verification code from database after 5 minutes
				go func() {
					time.Sleep(5 * time.Minute)
					c := context.Background()
					err := hp.RemoveVerificationCode(c, user.Email)
					if err != nil {
						hp.SetDebug("Error removing verification code: "+err.Error(), funcName)
					}
				}()

				response := hp.SetError(nil, "Email not verified, please check your email for verification code", funcName)
				c.AbortWithStatusJSON(http.StatusUnauthorized, response)
				return
			}
		}

		userResponse := hp.UserResponse{
			ID:            user.ID,
			FirstName:     user.FirstName,
			LastName:      user.LastName,
			Username:      user.Username,
			Email:         user.Email,
			EmailVerified: user.EmailVerified,
			Friends:       user.Friends,
			Wallet:        user.Wallet,
			Transactions:  user.Transactions,
			Role:          user.Role,
			CreatedAt:     user.CreatedAt,
			UpdatedAt:     user.UpdatedAt,
		}

		var data = gin.H{
			"accessToken":  t,
			"refreshToken": rt,
			"user":         userResponse,
			"expires_at":   config.AccessTokenExpireTime.Unix(),
		}
		response := hp.SetSuccess("User signed in successfully", data, funcName)
		c.JSON(http.StatusOK, response)
	} else {
		response := hp.SetError(nil, "Invalid password", funcName)
		c.AbortWithStatusJSON(http.StatusBadRequest, response)
	}
}

var Signout = AbstractConnection(signOut)

func signOut(c *gin.Context, ctx context.Context) {
	funcName := ut.GetFunctionName()

	user, err := hp.GetUserFromToken(c)
	if err != nil {
		response := hp.SetError(err, "User not logged in", funcName)
		c.AbortWithStatusJSON(http.StatusUnauthorized, response)
		return
	}

	err = hp.UpdateRefreshToken(ctx, user.ID, "")
	if err != nil {
		response := hp.SetError(err, "Error signing out", funcName)
		c.AbortWithStatusJSON(http.StatusInternalServerError, response)
		return
	}

	response := hp.SetSuccess("User signed out successfully", nil, funcName)
	c.JSON(http.StatusOK, response)
}

// func Signout(c *gin.Context) {
// 	ctx, cancel := context.WithTimeout(c.Request.Context(), config.ContextTimeout)
// 	defer cancel()
// 	defer database.DisconnectMongoDB()

// 	funcName := ut.GetFunctionName()

// 	user, err := hp.GetUserFromToken(c)
// 	if err != nil {
// 		response := hp.SetError(err, "User not logged in", funcName)
// 		c.AbortWithStatusJSON(http.StatusUnauthorized, response)
// 		return
// 	}

// 	err = hp.UpdateRefreshToken(ctx, user.ID, "")
// 	if err != nil {
// 		response := hp.SetError(err, "Error signing out", funcName)
// 		c.AbortWithStatusJSON(http.StatusInternalServerError, response)
// 		return
// 	}

// 	response := hp.SetSuccess("User signed out successfully", nil, funcName)
// 	c.JSON(http.StatusOK, response)
// }

func RefreshToken(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), config.ContextTimeout)
	defer cancel()
	defer database.DisconnectMongoDB()

	funcName := ut.GetFunctionName()

	user, err := hp.GetUserFromToken(c)
	if err != nil {
		response := hp.SetError(err, "User not logged in", funcName)
		c.AbortWithStatusJSON(http.StatusUnauthorized, response)
		return
	}

	request := hp.RefreshToken{}

	if err := c.BindJSON(&request); err != nil {
		response := hp.SetError(err, "refresh token is required", funcName)
		c.AbortWithStatusJSON(http.StatusBadRequest, response)
		return
	}

	if user.RefreshToken != request.RefreshToken {
		response := hp.SetError(nil, "Invalid refresh token", funcName)
		c.AbortWithStatusJSON(http.StatusBadRequest, response)
		return
	}

	t, rt, err := auth.GenerateJWT(user.Email, user.Username, user.ID)

	if err != nil {
		response := hp.SetError(err, "Error generating token", funcName)
		c.AbortWithStatusJSON(http.StatusInternalServerError, response)
		return
	}

	err = hp.UpdateRefreshToken(ctx, user.ID, rt)
	if err != nil {
		response := hp.SetError(err, "Error updating refresh token", funcName)
		c.AbortWithStatusJSON(http.StatusInternalServerError, response)
		return
	}

	var data = gin.H{
		"token":      t,
		"refresh":    rt,
		"user":       user.Username,
		"expires_at": config.AccessTokenExpireTime.Unix(),
	}
	response := hp.SetSuccess("Token refreshed successfully", data, funcName)
	c.JSON(http.StatusOK, response)
}

func ForgotPassword(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), config.ContextTimeout)
	defer cancel()
	defer database.DisconnectMongoDB()

	funcName := ut.GetFunctionName()

	request := hp.ForgotPassword{}

	if err := c.BindJSON(&request); err != nil {
		response := hp.SetError(err, "Error binding request", funcName)
		c.AbortWithStatusJSON(http.StatusBadRequest, response)
		return
	}
	log.Print("Request ID sent by client:", request.Email)

	var user = hp.UserStruct{}
	options := hp.PasswordOpts
	filter := bson.M{"email": request.Email}
	if err := usersCollection.FindOne(ctx, filter, options).Decode(&user); err != nil {
		response := hp.SetError(err, "User not found", funcName)
		c.AbortWithStatusJSON(http.StatusBadRequest, response)
		return
	}

	t, rt, err := auth.GenerateJWT(user.Email, user.Username, user.ID)
	if err != nil {
		response := hp.SetError(err, "Error generating token", funcName)
		c.AbortWithStatusJSON(http.StatusInternalServerError, response)
		return
	}

	err = hp.UpdateRefreshToken(ctx, user.ID, rt)
	if err != nil {
		response := hp.SetError(err, "Error updating refresh token", funcName)
		c.AbortWithStatusJSON(http.StatusInternalServerError, response)
		return
	}

	var data = gin.H{
		"token": t,
	}
	response := hp.SetSuccess("Token generated successfully", data, funcName)
	c.JSON(http.StatusOK, response)
}

func ResetPassword(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), config.ContextTimeout)
	defer cancel()
	defer database.DisconnectMongoDB()

	funcName := ut.GetFunctionName()

	request := hp.ResetPassword{}

	if err := c.BindJSON(&request); err != nil {
		response := hp.SetError(err, "Error binding request", funcName)
		c.AbortWithStatusJSON(http.StatusBadRequest, response)
		return
	}

	user, err := hp.GetUserFromToken(c)
	if err != nil {
		response := hp.SetError(err, "User not logged in", funcName)
		c.AbortWithStatusJSON(http.StatusUnauthorized, response)
		return
	}

	filter := bson.M{"_id": user.ID}

	if user.RefreshToken != request.RefreshToken {
		response := hp.SetError(nil, "Invalid refresh token", funcName)
		c.AbortWithStatusJSON(http.StatusBadRequest, response)
		return
	}

	hashedPassword, err := auth.HashPassword(request.NewPassword)
	if err != nil {
		response := hp.SetError(err, "Error hashing password", funcName)
		c.AbortWithStatusJSON(http.StatusInternalServerError, response)
		return
	}

	update := bson.M{
		"$set": bson.M{
			"password": hashedPassword,
		},
	}

	_, err = usersCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		response := hp.SetError(err, "Error updating password", funcName)
		c.AbortWithStatusJSON(http.StatusInternalServerError, response)
		return
	}

	var response = hp.SetSuccess("Password updated successfully", nil, funcName)
	c.JSON(http.StatusOK, response)
}
