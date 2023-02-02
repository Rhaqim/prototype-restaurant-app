package helpers

import (
	"context"
	"encoding/json"
	"errors"
	"math"

	"github.com/Rhaqim/thedutchapp/pkg/config"
	ut "github.com/Rhaqim/thedutchapp/pkg/utils"
	"go.mongodb.org/mongo-driver/bson"
)

// location services
// Services for user and restaurant locations
// recommender system based on location
//

type LocationService interface {
	GetDistance(lat1, lon1, lat2, lon2 float64) float64
	GetClosestRestaurant(lat, lon float64) (Restaurant, error)
}

type locationService struct{}

func NewLocationService() LocationService {
	return &locationService{}
}

func (ls *locationService) GetDistance(lat1, lon1, lat2, lon2 float64) float64 {
	// Convert degrees to radians
	lat1 = lat1 * math.Pi / 180.0
	lon1 = lon1 * math.Pi / 180.0

	lat2 = lat2 * math.Pi / 180.0
	lon2 = lon2 * math.Pi / 180.0

	// Haversine formula
	dlon := lon2 - lon1
	dlat := lat2 - lat1
	a := math.Pow(math.Sin(dlat/2), 2) + math.Cos(lat1)*math.Cos(lat2)*math.Pow(math.Sin(dlon/2), 2)
	c := 2 * math.Asin(math.Sqrt(a))
	r := 6371

	// calculate the result
	return c * float64(r)
}

func (ls *locationService) GetClosestRestaurant(lat, lon float64) (Restaurant, error) {
	var funcName = ut.GetFunctionName()
	var restaurants []Restaurant
	var restaurant Restaurant

	filter := bson.M{}

	var c = context.Background()

	restaurants, err := GetRestaurants(c, filter)
	if err != nil {
		SetDebug(err.Error(), funcName)
		return restaurant, err
	}

	var minDistance float64
	var distance float64
	var closestRestaurant Restaurant

	for _, restaurant := range restaurants {
		distance = ls.GetDistance(lat, lon, restaurant.MapInfo.Lat, restaurant.MapInfo.Long)
		if distance < minDistance || minDistance == 0 {
			minDistance = distance
			closestRestaurant = restaurant
		}
	}

	return closestRestaurant, nil
}

/* make api call to google maps api given an address and return the lat and lon */
func GetLatLong(address Address) (float64, float64, string, error) {
	var funcName = ut.GetFunctionName()
	var lat float64
	var lon float64
	var placeID string

	// convert address to string
	addressString := address.HouseNumber + "+" + address.Street + "+" + address.City + "+" + address.Zipcode

	var outputFormat = "json"
	var parameters = "address=" + addressString + "&key=" + config.GoogleMapsAPIKey
	var googleMapsAPIURL = "https://maps.googleapis.com/maps/api/geocode/" + outputFormat + "?" + parameters

	// make api call to google maps api
	status, body := ut.BaseAPICaller(googleMapsAPIURL, "GET", nil)

	if status != 200 {
		SetDebug("Error getting lat and lon for address: "+addressString, funcName)
		return lat, lon, placeID, errors.New("Error getting lat and lon for address: " + addressString)
	}

	// parse json response
	var data map[string]interface{}
	err := json.Unmarshal([]byte(body), &data)
	if err != nil {
		SetDebug("Error parsing json response for address: "+addressString, funcName)
		return lat, lon, placeID, err
	}

	// get lat and lon
	lat = data["results"].([]interface{})[0].(map[string]interface{})["geometry"].(map[string]interface{})["location"].(map[string]interface{})["lat"].(float64)
	lon = data["results"].([]interface{})[0].(map[string]interface{})["geometry"].(map[string]interface{})["location"].(map[string]interface{})["lng"].(float64)
	placeID = data["results"].([]interface{})[0].(map[string]interface{})["place_id"].(string)

	return lat, lon, placeID, nil
}
