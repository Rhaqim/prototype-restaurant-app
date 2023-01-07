package admin

import (
	"context"
	"sync"
	"time"

	"github.com/Rhaqim/thedutchapp/pkg/auth"
	"github.com/Rhaqim/thedutchapp/pkg/config"
	hp "github.com/Rhaqim/thedutchapp/pkg/helpers"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var adminCollection = config.AdminCollection
var authCollection = config.UserCollection

// AdminModel struct for admin model
type AdminModel struct {
	ID        primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	FirstName string             `json:"first_name" bson:"first_name" binding:"required"`
	LastName  string             `json:"last_name" bson:"last_name" binding:"required"`
	Username  string             `json:"username" bson:"username" binding:"required"`
	Email     string             `json:"email" bson:"email" binding:"required"`
	Password  string             `json:"password" bson:"password" binding:"required"`
	Role      hp.Roles           `json:"role" bson:"role"`
	CreatedAt primitive.DateTime `json:"created_at" bson:"created_at"`
	UpdatedAt primitive.DateTime `json:"updated_at" bson:"updated_at"`
}

// AdminInterface interface for admin model
// to implement methods
type AdminInterface interface {
	CreateAdmin(ctx context.Context) error
	AdminSignIn(ctx context.Context) (string, error)
}

// NewAdminModel creates a new admin model
// and returns a pointer to it
func NewAdminModel(firstName, lastName, username, email, password string) *AdminModel {
	hashedPassword, _ := auth.HashPassword(password)
	var new_username = "admin_" + username

	return &AdminModel{
		ID:        primitive.NewObjectID(),
		FirstName: firstName,
		LastName:  lastName,
		Username:  new_username,
		Email:     email,
		Password:  hashedPassword,
		Role:      hp.Admin,
		CreatedAt: primitive.NewDateTimeFromTime(time.Now()),
		UpdatedAt: primitive.NewDateTimeFromTime(time.Now()),
	}
}

// CreateAdmin creates a new admin
// and inserts into admin collection and auth collection
// returns error if any
func (a *AdminModel) CreateAdmin(ctx context.Context) error {
	hashedPassword, _ := auth.HashPassword(a.Password)
	var new_username = "admin_" + a.Username

	a.ID = primitive.NewObjectID()
	a.Username = new_username
	a.Password = hashedPassword
	a.Role = hp.Admin
	a.CreatedAt = primitive.NewDateTimeFromTime(time.Now())
	a.UpdatedAt = primitive.NewDateTimeFromTime(time.Now())

	// Insert into admin collection and auth collection goroutine
	var wg sync.WaitGroup
	wg.Add(2)

	errChan := make(chan error, 2)

	go func() {
		defer wg.Done()
		_, err := adminCollection.InsertOne(ctx, a)
		if err != nil {
			errChan <- err
		}
	}()

	go func() {
		defer wg.Done()
		_, err := authCollection.InsertOne(ctx, a)
		if err != nil {
			errChan <- err
		}
	}()

	wg.Wait()
	close(errChan)

	for err := range errChan {
		if err != nil {
			return err
		}
	}

	return nil
}
