package helpers

import "go.mongodb.org/mongo-driver/bson/primitive"

type KYC struct {
	Phone          string             `bson:"phone" json:"phone"`
	DOB            string             `bson:"dob" json:"dob"`
	Address        string             `bson:"address" json:"address"`
	City           primitive.ObjectID `bson:"city" json:"city"`
	State          primitive.ObjectID `bson:"state" json:"state"`
	Zip            string             `bson:"zip" json:"zip"`
	Country        primitive.ObjectID `bson:"country" json:"country"`
	IdentityType   string             `bson:"identityType" json:"identityType"`
	IdentityNumber string             `bson:"identityNumber" json:"identityNumber"`
	IdentityPhoto  interface{}        `bson:"identityPhoto" json:"identityPhoto"`
	UpdatedAt      primitive.DateTime `bson:"updatedAt" json:"updatedAt" default:"Now()"`
}
