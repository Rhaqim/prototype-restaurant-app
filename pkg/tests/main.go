package tests

// import (
// 	"bytes"
// 	"context"
// 	"net/http"
// 	"net/http/httptest"

// 	"github.com/gin-gonic/gin"
// 	"github.com/stretchr/testify/suite"
// 	"go.mongodb.org/mongo-driver/bson"
// 	"go.mongodb.org/mongo-driver/mongo"
// 	"go.mongodb.org/mongo-driver/mongo/options"
// )

// type TestSuite struct {
// 	suite.Suite
// 	router *gin.Engine
// 	client *mongo.Client
// 	db     *mongo.Database
// }

// func (suite *TestSuite) SetupSuite() {
// 	// Create a new gin router
// 	suite.router = gin.New()

// 	// Connect to the MongoDB test database
// 	client, err := mongo.NewClient(options.Client().ApplyURI("mongodb://localhost:27017"))
// 	if err != nil {
// 		suite.T().Fatalf("Failed to connect to MongoDB: %v", err)
// 	}
// 	suite.client = client
// 	suite.db = client.Database("test")

// 	// Add routes to the router
// 	suite.router.GET("/items", func(c *gin.Context) {
// 		c.JSON(http.StatusOK, gin.H{"message": "List of items"})
// 	})
// 	suite.router.POST("/items", func(c *gin.Context) {
// 		c.JSON(http.StatusOK, gin.H{"message": "Item created"})
// 	})
// }

// func (suite *TestSuite) TearDownSuite() {
// 	// Drop the test database after all tests have run
// 	suite.db.Drop(context.TODO())
// 	suite.client.Disconnect(context.TODO())
// }

// func (suite *TestSuite) TestListItems() {
// 	// Create a request to send to the router
// 	req, _ := http.NewRequest("GET", "/items", nil)
// 	rr := httptest.NewRecorder()

// 	// Send the request to the router and record the response
// 	suite.router.ServeHTTP(rr, req)

// 	// Check that the response has a 200 status code
// 	suite.Equal(http.StatusOK, rr.Code)

// 	// Check that the response body is as expected
// 	suite.Equal("List of items", rr.Body.String())
// }

// func (suite *TestSuite) TestCreateItem() {
// 	// Create a request to send to the router
// 	data := []byte(`{"name":"Test item"}`)
// 	req, _ := http.NewRequest("POST", "/items", bytes.NewBuffer(data))
// 	rr := httptest.NewRecorder()

// 	// Send the request to the router and record the response
// 	suite.router.ServeHTTP(rr, req)

// 	// Check that the response has a 200 status code
// 	suite.Equal(http.StatusOK, rr.Code)

// 	// check if item was added to the database
// 	var item Item
// 	err := suite.db.Collection("items").FindOne(context.TODO(), bson.M{"name": "Test item"}).Decode(&item)
// 	suite.NoError(err)
// 	suite.Equal("Test item", item.Name)

// 	// Check that the response body is as expected
// 	suite.Equal("Item created", rr.Body.String())
// }
