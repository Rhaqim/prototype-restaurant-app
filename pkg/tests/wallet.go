package tests

// Write tests for the wallet controller
// func TestCreateWallet(t *testing.T) {
// 	// Create a test user
// 	user := User{
// 		ID:       primitive.NewObjectID(),
// 		Email:    "
// 		Password: "password",
// 		Role:     "user",
// 	}
// 	user.Password, _ = auth.HashPassword(user.Password)
// 	user.CreatedAt, user.UpdatedAt = hp.CreatedAtUpdatedAt()

// 	// Create a test wallet
// 	wallet := Wallet{
// 		ID:       primitive.NewObjectID(),
// 		UserID:   user.ID,
// 		TxnPin:   "password",
// 		Balance:  0,
// 		Currency: "NGN",
// 	}
// 	wallet.TxnPin, _ = auth.HashPassword(wallet.TxnPin)
// 	wallet.CreatedAt, wallet.UpdatedAt = hp.CreatedAtUpdatedAt()

// 	// Create a test request
// 	request := Wallet{
// 		TxnPin:   "password",
// 		Balance:  0,
// 		Currency: "NGN",
// 	}

// 	// Create a test response
// 	response := Wallet{
// 		ID:       wallet.ID,
// 		UserID:   user.ID,
// 		TxnPin:   wallet.TxnPin,
// 		Balance:  wallet.Balance,
// 		Currency: wallet.Currency,
// 		CreatedAt: wallet.CreatedAt,
// 		UpdatedAt: wallet.UpdatedAt,
// 	}

// 	// Create a test context
// 	ctx := context.Background()

// 	// Create a test router
// 	router := gin.Default()

// 	// Create a test controller
// 	controller := WalletController{
// 		Router: router,
// 		DB:     db,
// 	}

// 	// Create a test route
// 	controller.CreateWallet()

// 	// Create a test server
// 	server := httptest.NewServer(router)

// 	// Create a test client
// 	client := server.Client()

// 	// Create a test request
// 	req, err := http.NewRequest("POST", server.URL+"/wallets", bytes.NewBuffer([]byte(`{"txn_pin": "password", "balance": 0, "currency": "NGN"}`)))
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	// Create a test token
// 	token, err := auth.CreateToken(user.ID, user.Email, user.Role)
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	// Add the token to the request
// 	req.Header.Add("Authorization", "Bearer "+token)

// 	// Send the request
// 	res, err := client.Do(req)
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	// Close the response body
// 	defer res.Body.Close()

// 	// Read the response body
// 	body, err := ioutil.ReadAll(res.Body)
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	// Unmarshal the response body
// 	var walletResponse Wallet
// 	err = json.Unmarshal(body, &walletResponse)
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	// Assert that the response status code is 201
// 	assert.Equal(t, http.StatusCreated, res.StatusCode)

// 	// Assert that the response body is equal to the expected response
// 	assert.Equal(t, response, walletResponse)

// 	// Assert that the wallet was created in the database
// 	var walletDB Wallet
// 	err = db.Collection("wallets").FindOne(ctx, bson.M{"_id": wallet.ID}).Decode(&walletDB)
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	// Assert that the wallet in the database is equal to the expected wallet
// 	assert.Equal(t, wallet, walletDB)
// }
