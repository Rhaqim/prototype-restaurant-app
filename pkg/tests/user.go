package tests

// Write tests for the user controller and return the user object for use in other tests
// func TestCreateUser(t *testing.T) {
// 	// Create a test request
// 	request := User{
// 		Email:    "example@gmail.com",
// 		Password: "password",
// 		Role:     "user",
// 	}
//
// 	// Create a test response
// 	response := User{
// 		ID:       primitive.NewObjectID(),
// 		Email:    request.Email,
// 		Password: request.Password,
// 		Role:     request.Role,
// 		CreatedAt: hp.CreatedAtUpdatedAt(),
// 		UpdatedAt: hp.CreatedAtUpdatedAt(),
// 	}
//
// 	// Create a test context
// 	ctx := context.Background()
//
// 	// Create a test router
// 	router := gin.Default()
//
// 	// Create a test controller
// 	controller := UserController{
// 		Router: router,
// 		DB:     db,
// 	}
//
// 	// Create a test route
// 	controller.CreateUser()
//
// 	// Create a test server
// 	server := httptest.NewServer(router)
//
// 	// Create a test client
// 	client := server.Client()
//
// 	// Create a test request
// 	req, err := http.NewRequestWithContext(ctx, "POST", server.URL+"/users", bytes.NewBuffer([]byte(`{"email":"`+request.Email+`","password":"`+request.Password+`","role":"`+request.Role+`"}`)))
// 	if err != nil {
// 		t.Fatal(err)
// 	}
//
// 	// Create a test response recorder
// 	recorder := httptest.NewRecorder()
//
// 	// Create a test response
// 	res, err := client.Do(req)
// 	if err != nil {
// 		t.Fatal(err)
// 	}
//
// 	// Create a test response body
// 	body, err := ioutil.ReadAll(res.Body)
// 	if err != nil {
// 		t.Fatal(err)
// 	}
//
// 	// Create a test response
// 	var user User
// 	err = json.Unmarshal(body, &user)
// 	if err != nil {
// 		t.Fatal(err)
// 	}
//
// 	// Assert that the response is equal to the expected response
// 	assert.Equal(t, response, user)
//
// 	// Return the user object
// 	return user
// }
