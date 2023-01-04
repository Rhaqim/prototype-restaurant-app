package helpers

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type KYCPhoto struct {
	Front interface{} `json:"front" bson:"front"`
	Back  interface{} `json:"back" bson:"back"`
}

type KYCStatus string

const (
	Unverified KYCStatus = "unverified"
	Verified   KYCStatus = "verified"
	Rejected   KYCStatus = "rejected"
)

type KYC struct {
	Phone          string             `bson:"phone" json:"phone" binding:"required,e164"`
	DOB            time.Time          `bson:"dob" json:"dob" binding:"required" time_format:"2006-01-02"`
	Address        string             `bson:"address" json:"address" binding:"required"`
	City           primitive.ObjectID `bson:"city" json:"city" binding:"required"`
	State          primitive.ObjectID `bson:"state" json:"state" binding:"required"`
	Zip            string             `bson:"zip" json:"zip" binding:"required"`
	CountryCode    string             `bson:"countryCode" json:"countryCode" binding:"required,iso3166_1_alpha2"`
	IdentityType   IdentityType       `bson:"identityType" json:"identityType" binding:"required"`
	IdentityNumber string             `bson:"identityNumber" json:"identityNumber" binding:"required"`
	IdentityPhoto  KYCPhoto           `bson:"identityPhoto" json:"identityPhoto" binding:"required"`
	KYCStatus      KYCStatus          `bson:"kycStatus" json:"kycStatus" default:"unverified"`
	UpdatedAt      primitive.DateTime `bson:"updated_at" json:"updated_at" default:"Now()"`
}

// ValidateKYC DOB
func ValidateKYCDOB(dob time.Time) bool {
	// Check if year is less than 18
	return dob.Year() < time.Now().Year()-18
}

type IdentityType string

const (
	// Identity Types
	Passport IdentityType = "passport"
	National IdentityType = "national id card"
	License  IdentityType = "driver's license"
)

func (IT IdentityType) String() string {
	switch IT {
	case Passport:
		return "passport"
	case National:
		return "national id card"
	case License:
		return "driver's license"
	default:
		return "passport"
	}
}

// Check if KYC is complete
func CheckKYCStatus(user UserResponse) bool {
	return user.KYCStatus == Verified
}

// For Nigerians BVN is the Identity Number
// For other countries Passport Number is the Identity Number
type BVN struct {
	BVN         uint `bson:"bvn" json:"bvn" binding:"required,number"`
	BVNVerified bool `bson:"bvnVerified" json:"bvnVerified" default:"false"`
}
