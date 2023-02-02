package helpers

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type KYC struct {
	Phone              string             `bson:"phone" json:"phone" binding:"required,e164"`
	DOB                CustomDate         `bson:"dob" json:"dob" binding:"required" time_format:"2006-01-02"`
	Address            Address            `bson:"address" json:"address" binding:"required"`
	MapInfo            MapInfo            `bson:"map_info" json:"map_info" binding:"required"`
	IdentityType       IdentityType       `bson:"identity_type" json:"identity_type" binding:"required"`
	IdentityNumber     string             `bson:"identity_number" json:"identity_number" binding:"required,min=5,max=20"`
	IdentityExpiration CustomDate         `bson:"identity_expiration" json:"identity_expiration" binding:"required" time_format:"2006-01-02"`
	IdentityPhoto      KYCPhoto           `bson:"identity_photo" json:"identity_photo" binding:"required"`
	SelfieImage        Avatar             `bson:"selfie_image" json:"selfie_image" binding:"required"`
	KYCStatus          KYCStatus          `bson:"kyc_status" json:"kyc_status" default:"unverified"`
	UpdatedAt          primitive.DateTime `bson:"updated_at" json:"updated_at" default:"Now()"`
}

type KYCPhoto struct {
	Front Avatar `json:"front" bson:"front"`
	Back  Avatar `json:"back" bson:"back"`
}

type KYCStatus string

const (
	Unverified KYCStatus = "unverified"
	Verified   KYCStatus = "verified"
	Rejected   KYCStatus = "rejected"
)

func (KS KYCStatus) String() string {
	switch KS {
	case Unverified:
		return "unverified"
	case Verified:
		return "verified"
	case Rejected:
		return "rejected"
	default:
		return "unverified"
	}
}

// ValidateKYC DOB
func ValidateKYCDOB(dob CustomDate) bool {
	// Check if year is less than 18
	return dob.Year() < time.Now().Year()-18
}

type IdentityType string

const (
	// Identity Types
	Passport IdentityType = "passport"
	National IdentityType = "national_id"
	License  IdentityType = "drivers_license"
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
