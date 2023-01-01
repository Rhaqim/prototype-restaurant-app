package utils

// // Update the Product with the new order
// product_id, err := primitive.ObjectIDFromHex(request.Product.ID.Hex())
// if err != nil {
// 	response := hp.SetError(err, "Error converting id to object id", funcName)
// 	c.AbortWithStatusJSON(http.StatusInternalServerError, response)
// 	return
// }

// product_filter := bson.M{"_id": product_id}
// product_update := bson.M{
// 	// decrement stock by quantity
// 	"$inc": bson.M{
// 		"stock": -request.Quantity,
// 	},
// }

// _, err = productCollection.UpdateOne(ctx, product_filter, product_update)
// if err != nil {
// 	response := hp.SetError(err, "Error updating hosted event", funcName)
// 	c.AbortWithStatusJSON(http.StatusInternalServerError, response)
// 	return
// }

// // Update the event with the new order
// id, err := primitive.ObjectIDFromHex(request.Event.ID.Hex())
// if err != nil {
// 	response := hp.SetError(err, "Error converting id to object id", funcName)
// 	c.AbortWithStatusJSON(http.StatusInternalServerError, response)
// 	return
// }

// filter := bson.M{"_id": id}
// update := bson.M{
// 	"$push": bson.M{
// 		"orders": insertResult.InsertedID,
// 	},
// }

// updateResult, err := eventCollection.UpdateOne(ctx, filter, update)
// if err != nil {
// 	response := hp.SetError(err, "Error updating hosted event", funcName)
// 	c.AbortWithStatusJSON(http.StatusInternalServerError, response)
// 	return
// }

// response := hp.SetSuccess(" order created", updateResult, funcName)
// c.JSON(http.StatusOK, response)
// }

/*
FOR GETTING THE SOCIALS
*/
// func GetSocial(c *gin.Context) {
// 	collection := config.MI.DB.Collection("socials")
// 	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
// 	defer cancel()
// 	var socials []models.Social
// 	cur, err := collection.Find(ctx, bson.M{})
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, helpers.SetError(err, "Error while getting socials", "GetSocial"))
// 		return
// 	}
// 	defer cur.Close(ctx)
// 	for cur.Next(ctx) {
// 		var social models.Social
// 		err := cur.Decode(&social)
// 		if err != nil {
// 			c.JSON(http.StatusInternalServerError, helpers.SetError(err, "Error while decoding socials", "GetSocial"))
// 			return
// 		}
// 		socials = append(socials, social)
// 	}
// 	if err := cur.Err(); err != nil {
// 		c.JSON(http.StatusInternalServerError, helpers.SetError(err, "Error while getting socials", "GetSocial"))
// 		return
// 	}
// 	c.JSON(http.StatusOK, helpers.SetSuccess("Socials found", socials, "GetSocial"))
// }

/*
FOR UPDATING THE BILL OF THE EVENT
*/
///////
// for _, product := range request.Products {
// 	product_id, err := primitive.ObjectIDFromHex(product.ProductID.Hex())
// 	if err != nil {
// 		response := hp.SetError(err, "Error converting id to object id", funcName)
// 		c.AbortWithStatusJSON(http.StatusInternalServerError, response)
// 		return
// 	}

// 	// fetch product
// 	product_fetched, err := hp.GetProductbyID(ctx, product_id)
// 	if err != nil {
// 		response := hp.SetError(err, "Error finding product", funcName)
// 		c.AbortWithStatusJSON(http.StatusInternalServerError, response)
// 		return
// 	}

// 	event_filter := bson.M{"_id": event_id}
// 	event_update := bson.M{
// 		"$push": bson.M{
// 			"orders": insertResult.InsertedID,
// 		},
// 		// update bill with new order
// 		"$inc": bson.M{
// 			"bill": float64(product_fetched.Price * float64(product.Quantity)),
// 		},
// 	}

// 	_, err = eventCollection.UpdateOne(ctx, event_filter, event_update)
// 	if err != nil {
// 		response := hp.SetError(err, "Error updating hosted event", funcName)
// 		c.AbortWithStatusJSON(http.StatusInternalServerError, response)
// 		return
// 	}

// }

/*
For Updating the Product Stock
*/

// for _, product := range request.Products {
// 	product_id, err := primitive.ObjectIDFromHex(product.ProductID.Hex())
// 	if err != nil {
// 		response := hp.SetError(err, "Error converting id to object id", funcName)
// 		c.AbortWithStatusJSON(http.StatusInternalServerError, response)
// 		return
// 	}

// 	product_filter := bson.M{"_id": product_id}
// 	product_update := bson.M{
// 		// decrement stock by quantity
// 		"$inc": bson.M{
// 			"stock": -product.Quantity,
// 		},
// 	}

// 	_, err = productCollection.UpdateOne(ctx, product_filter, product_update)
// 	if err != nil {
// 		response := hp.SetError(err, "Error updating product", funcName)
// 		c.AbortWithStatusJSON(http.StatusInternalServerError, response)
// 		return
// 	}

// }
