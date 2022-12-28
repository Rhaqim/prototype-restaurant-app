package controllers

/*

Creating a new Order with Gorooutines, concurrency, WaitGroups and channels

*/

// func CreateNewOrder(c *gin.Context) {
// 	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
// 	defer cancel()

// 	var funcName = ut.GetFunctionName()

// 	var request hp.Order

// 	if err := c.ShouldBindJSON(&request); err != nil {
// 		response := hp.SetError(err, "Error binding json", funcName)
// 		c.AbortWithStatusJSON(http.StatusBadRequest, response)
// 		return
// 	}

// 	request.ID = primitive.NewObjectID()

// 	defer database.ConnectMongoDB().Disconnect(context.TODO())
// }
