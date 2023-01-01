package helpers

import (
	"context"
	"encoding/json"

	"github.com/Rhaqim/thedutchapp/pkg/config"
	ut "github.com/Rhaqim/thedutchapp/pkg/utils"
)

type Country struct {
	Alpha2              string
	Alpha3              string
	CountryCallingCodes []string
	Currencies          []string
	Ioc                 string
	Languages           []string
	Name                string
	Status              string
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

	// Insert the countries into the database
	for _, country := range countries {
		_, err := config.CountryCollection.InsertOne(context.Background(), country)

		if err != nil {
			return err
		}

	}
	return nil
}
