package helpers

import (
	"context"
	"math"

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
		distance = ls.GetDistance(lat, lon, restaurant.Latitude, restaurant.Longitude)
		if distance < minDistance || minDistance == 0 {
			minDistance = distance
			closestRestaurant = restaurant
		}
	}

	return closestRestaurant, nil
}

/* make api call to google maps api given an address and return the lat and lon */
func GetLatLon(address string) (float64, float64, error) {
	var funcName = ut.GetFunctionName()
	var lat float64
	var lon float64

	SetInfo("Getting lat and lon for address: "+address, funcName)

	// make api call to google maps api
	// return lat and lon

	return lat, lon, nil
}
