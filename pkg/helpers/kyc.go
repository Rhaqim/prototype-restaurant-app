package helpers

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type KYC struct {
	Phone          string             `bson:"phone" json:"phone" binding:"required"`
	DOB            time.Time          `bson:"dob" json:"dob" binding:"required"`
	Address        string             `bson:"address" json:"address" binding:"required"`
	City           primitive.ObjectID `bson:"city" json:"city" binding:"required"`
	State          primitive.ObjectID `bson:"state" json:"state" binding:"required"`
	Zip            string             `bson:"zip" json:"zip" binding:"required"`
	Country        primitive.ObjectID `bson:"country" json:"country" binding:"required"`
	IdentityType   string             `bson:"identityType" json:"identityType" binding:"required"`
	IdentityNumber string             `bson:"identityNumber" json:"identityNumber" binding:"required"`
	IdentityPhoto  interface{}        `bson:"identityPhoto" json:"identityPhoto" binding:"required"`
	UpdatedAt      primitive.DateTime `bson:"updatedAt" json:"updatedAt" default:"Now()"`
}

// ValidateKYC DOB
func ValidateKYCDOB(dob time.Time) bool {
	// Check if year is less than 18
	return dob.Year() < time.Now().Year()-18
}
