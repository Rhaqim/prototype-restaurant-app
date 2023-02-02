package helpers

import (
	"context"
	"encoding/json"

	"github.com/Rhaqim/thedutchapp/pkg/config"
	ut "github.com/Rhaqim/thedutchapp/pkg/utils"
)

type Country struct {
	Alpha2              string   `json:"alpha2"`
	Alpha3              string   `json:"alpha3"`
	CountryCallingCodes []string `json:"countryCallingCodes"`
	Currencies          []string `json:"currencies"`
	Ioc                 string   `json:"ioc"`
	Languages           []string `json:"languages"`
	Name                string   `json:"name"`
	Status              string   `json:"status"`
}

func LoadCountriesDB() error {
	// Load countries from the json file
	var countries []Country

	// Open the json file
	data := ut.LoadJsonFile("countries.json")

	// Unmarshal the json file
	err := json.Unmarshal(data, &countries)
	if err != nil {
		return err
	}

	var errChan = make(chan error, 1)

	// Insert the countries into the database
	go func() {
		for _, country := range countries {
			_, err := config.CountryCollection.InsertOne(context.Background(), country)

			if err != nil {
				errChan <- err
				return
			}
		}
	}()

	// Check if there is an error
	select {
	case err := <-errChan:
		return err
	default:
	}

	return nil
}
