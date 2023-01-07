package main

import (
	"context"

	"github.com/Rhaqim/thedutchapp/pkg/config"
	"github.com/Rhaqim/thedutchapp/pkg/helpers"
	"go.mongodb.org/mongo-driver/bson"

	// "github.com/Rhaqim/thedutchapp/pkg/handlers"
	ut "github.com/Rhaqim/thedutchapp/pkg/utils"
)

func main() {
	// run := handlers.GinRouter()
	// port := ut.GetEnv("PORT")

	// run.Run(port)
	var ctx = context.Background()

	var filter = bson.M{"role": helpers.User}
	// err := helpers.SetUserIDsCache(ctx, filter, config.UserRole)
	// if err != nil {
	// 	panic(err)
	// }

	users, err := helpers.GetUserIDsFromCache(ctx, filter, config.UserRole)
	if err != nil {
		panic(err)
	}

	helpers.SetDebug(ut.ToJSON(users), ut.GetFunctionName())
}
