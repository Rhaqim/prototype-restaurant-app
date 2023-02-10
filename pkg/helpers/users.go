package helpers

import (
	"context"
	"errors"

	em "github.com/Rhaqim/thedutchapp/pkg/email"
	ut "github.com/Rhaqim/thedutchapp/pkg/utils"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Type for Roles assigned to users
type Roles string

const (
	Admin    Roles = "admin"
	User     Roles = "user"
	Business Roles = "business"
)

// USER STRUCT all users
type UserStruct struct {
	ID                     primitive.ObjectID   `bson:"_id,omitempty" form:"_id,omitempty"`
	FirstName              string               `form:"first_name" bson:"first_name" binding:"required"`
	LastName               string               `form:"last_name" bson:"last_name" binding:"required"`
	Email                  string               `bson:"email" form:"email" binding:"required,email"`
	Username               string               `bson:"username" form:"username" binding:"required"`
	Password               string               `bson:"password" form:"password" binding:"required,min=8,max=32,alphanum"`
	Avatar                 Avatar               `bson:"avatar" form:"avatar" default:"{}"`
	Social                 SocialNetwork        `bson:"social" form:"social" default:"{}"`
	Friends                []primitive.ObjectID `bson:"friends" form:"friends" default:"[]"`
	Location               string               `bson:"location" form:"location"`
	Wallet                 primitive.ObjectID   `bson:"wallet" form:"wallet" default:"null"`
	Account                BankAccount          `bson:"account" form:"account" default:"{}"`
	RefreshToken           string               `bson:"refresh_token,omitempty" form:"refresh_token,omitempty"`
	EmailVerified          bool                 `bson:"email_confirmed" form:"email_confirmed" default:"false"`
	EmailVerificationToken string               `bson:"email_verification_token,omitempty" form:"email_verification_token,omitempty"`
	PasswordResetToken     string               `bson:"password_reset_token,omitempty" form:"password_reset_token,omitempty"`
	KYCStatus              KYCStatus            `bson:"kyc_status,omitempty" form:"kyc_status,omitempty" default:"unverified"`
	Role                   Roles                `bson:"role" form:"role" default:"user"`
	CreatedAt              primitive.DateTime   `bson:"created_at" form:"created_at" default:"Now()"`
	UpdatedAt              primitive.DateTime   `bson:"updated_at" form:"updated_at" default:"Now()"`
}

type Avatar struct {
	Alt string `json:"alt,omitempty" bson:"alt,omitempty"`
	URL string `json:"url,omitempty" bson:"url,omitempty"`
}

type SocialNetwork struct {
	Network string `json:"network,omitempty" bson:"network,omitempty"`
	Link    string `json:"link,omitempty" bson:"link,omitempty"`
}

type UserResponse struct {
	ID                     primitive.ObjectID   `bson:"_id" json:"_id"`
	FirstName              string               `json:"first_name"`
	LastName               string               `json:"last_name"`
	Email                  string               `bson:"email" json:"email"`
	Username               string               `bson:"username" json:"username"`
	Avatar                 Avatar               `bson:"avatar" json:"avatar"`
	Social                 SocialNetwork        `bson:"social" json:"social"`
	Friends                []primitive.ObjectID `bson:"friends" json:"friends"`
	Location               string               `bson:"location" json:"location"`
	Wallet                 primitive.ObjectID   `bson:"wallet" json:"wallet"`
	Account                BankAccount          `bson:"account" json:"account"`
	RefreshToken           string               `bson:"refresh_token" json:"refresh_token"`
	EmailVerified          bool                 `bson:"email_confirmed" json:"email_confirmed"`
	EmailVerificationToken string               `bson:"email_verification_token,omitempty" json:"email_verification_token,omitempty"`
	PasswordResetToken     string               `bson:"password_reset_token,omitempty" json:"password_reset_token,omitempty"`
	KYCStatus              KYCStatus            `bson:"kyc_status" json:"kyc_status"`
	Role                   Roles                `bson:"role" json:"role"`
	CreatedAt              primitive.DateTime   `bson:"created_at" json:"created_at"`
	UpdatedAt              primitive.DateTime   `bson:"updated_at" json:"updated_at"`
}

type UserUpdate struct {
	Email    string        `bson:"email" json:"email"`
	Username string        `bson:"username" json:"username"`
	Avatar   Avatar        `bson:"avatar" json:"avatar"`
	Social   SocialNetwork `bson:"social" json:"social"`
	Location string        `bson:"location" json:"location"`
}

type CreatUser struct {
	Fullname      string             `json:"fullname"`
	Username      string             `json:"username"`
	Avatar        interface{}        `json:"avatar"`
	Email         string             `json:"email"`
	Password      string             `json:"password"`
	Social        interface{}        `json:"social"`
	Role          Roles              `json:"role"`
	RefreshToken  string             `json:"refreshToken,omitempty"`
	EmailVerified bool               `json:"emailConfirmed,omitempty"`
	CreatedAt     primitive.DateTime `json:"created_at"`
	UpdatedAt     primitive.DateTime `json:"updated_at"`
}

type GetUserById struct {
	ID primitive.ObjectID `json:"id"`
}

type GetUserByEmailStruct struct {
	Email string `json:"email"`
}

type UpdateUserAvatar struct {
	ID        primitive.ObjectID `json:"id"`
	Avatar    string             `json:"avatar"`
	CreatedAt primitive.DateTime `json:"created_at"`
	UpdatedAt primitive.DateTime `json:"updated_at"`
}

/*
Get user data by:
- ID
- Email
- Username
- From token
*/
func GetUser(ctx context.Context, filter bson.M) (UserResponse, error) {
	var user UserResponse
	funcName := ut.GetFunctionName()

	opts := options.FindOne().SetProjection(bson.M{
		"password":     0,
		"refreshToken": 0})

	if err := usersCollection.FindOne(ctx, filter, opts).Decode(&user); err != nil {
		SetError(err, "error", funcName)
		return UserResponse{}, err
	}

	return user, nil
}

func GetUserAllInfo(ctx context.Context, filter bson.M) (UserResponse, error) {
	var user UserResponse
	funcName := ut.GetFunctionName()

	if err := usersCollection.FindOne(ctx, filter).Decode(&user); err != nil {
		SetError(err, "error", funcName)
		return UserResponse{}, err
	}

	return user, nil
}

// Get user by ID
func GetUserByID(ctx context.Context, userID primitive.ObjectID) UserResponse {
	filter := bson.M{"_id": userID}

	user, err := GetUser(ctx, filter)
	if err != nil {
		return UserResponse{}
	}

	return user
}

// Get user by Email
func GetUserByEmail(ctx context.Context, email string) UserResponse {
	filter := bson.M{"email": email}

	funcName := ut.GetFunctionName()

	user, err := GetUser(ctx, filter)
	if err != nil {
		SetError(err, "error", funcName)
		return UserResponse{}
	}

	return user
}

// Get user by Username
func GetUserByUsername(ctx context.Context, username string) UserResponse {

	filter := bson.M{"username": username}

	user, err := GetUser(ctx, filter)
	if err != nil {
		return UserResponse{}
	}

	return user
}

// Get user from token
func GetUserFromToken(c *gin.Context) (UserResponse, error) {
	check, ok := c.Get("user") // Check if user is logged in
	if !ok {
		return UserResponse{}, errors.New("Unauthorized")
	}

	user := check.(UserResponse)

	return user, nil
}

func GetUsers(ctx context.Context, filter bson.M) ([]UserResponse, error) {
	var users []UserResponse
	funcName := ut.GetFunctionName()

	opts := options.Find().SetProjection(bson.M{
		"password":     0,
		"refreshToken": 0})

	cursor, err := usersCollection.Find(ctx, filter, opts)
	if err != nil {
		SetError(err, "error", funcName)
		return nil, err
	}

	if err = cursor.All(ctx, &users); err != nil {
		SetError(err, "error", funcName)
		return nil, err
	}

	return users, nil
}

// Check if Email is verified
func CheckIfEmailVerificationExists(ctx context.Context, email string) (bool, error) {
	var user UserResponse

	filter := bson.M{"email": email}
	user, err := GetUser(ctx, filter)
	if err != nil {
		return false, err
	}
	return user.EmailVerified, nil
}

// Email Verification
// Generate random string as token for email verification
func GenerateEmailVerificationToken(email string) string {
	return ut.RandomString(6, email)
}

// Password Reset
// Genereate random 4 char string as token for password reset
func GeneratePasswordResetToken(email string) string {
	return ut.RandomString(4, email)
}

// Send email verification email
// Update email verification token in database
// Send email
// Return error
// Accepts context and email
// TODO: Add email template
func SendVerificationEmail(ctx context.Context, email string) error {
	// Get user data after updating email verification token
	filter := bson.M{"email": email}
	update := bson.M{"$set": bson.M{"email_verification_token": GenerateEmailVerificationToken(email)}}
	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)

	var user UserResponse
	if err := usersCollection.FindOneAndUpdate(ctx, filter, update, opts).Decode(&user); err != nil {
		return err
	}

	// Send email
	r := em.NewRequest([]string{email}, "Email Verification", user.EmailVerificationToken)

	template := struct {
		Title string
		Body  string
	}{
		Title: "Email Verification",
		Body:  user.EmailVerificationToken,
	}

	err := r.ParseTemplate("email-verification.html", template)
	if err != nil {
		return err
	}

	ok, err := r.SendEmail()
	if err != nil {
		return err
	}

	if !ok {
		return errors.New("email not sent")
	}

	return nil
}

// Verify email with token
func VerifyEmail(ctx context.Context, email string, token string) error {
	// get user data
	user := GetUserByEmail(ctx, email)

	// check if token is valid
	if user.EmailVerificationToken != token {
		return errors.New("invalid verification token")
	}

	filter := bson.M{"email": email}
	update := bson.M{"$set": bson.M{"email_confirmed": true, "email_verification_token": ""}}

	_, err := usersCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	return nil
}

func UpdateUser(ctx context.Context, filter bson.M, update bson.M) error {
	funcName := ut.GetFunctionName()

	_, err := usersCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		SetDebug(err.Error(), funcName)
		return err
	}

	return nil
}

func RemoveVerificationCode(ctx context.Context, email string) error {
	filter := bson.M{"email": email}
	update := bson.M{"$set": bson.M{"email_verification_token": ""}}

	err := UpdateUser(ctx, filter, update)
	if err != nil {
		return err
	}

	return nil
}

// SendPasswordResetEmail sends password reset email
// It sends a token to the user's email
// It updates the user's password reset token in the database
// It returns error
func SendPasswordResetEmail(ctx context.Context, email string) error {
	// Get user data after updating password reset token
	filter := bson.M{"email": email}
	update := bson.M{"$set": bson.M{"password_reset_token": GeneratePasswordResetToken(email)}}
	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)

	var user UserResponse
	if err := usersCollection.FindOneAndUpdate(ctx, filter, update, opts).Decode(&user); err != nil {
		return err
	}

	// Send email
	r := em.NewRequest([]string{email}, "Password Reset", user.PasswordResetToken)

	template := struct {
		Title string
		Body  string
	}{
		Title: "Password Reset",
		Body:  user.PasswordResetToken,
	}

	err := r.ParseTemplate("password-reset.html", template)
	if err != nil {
		return err
	}

	ok, err := r.SendEmail()
	if err != nil {
		return err
	}

	if !ok {
		return errors.New("email not sent")
	}

	return nil
}
