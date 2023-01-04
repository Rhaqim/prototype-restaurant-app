package helpers

import (
	"errors"
	"strings"
	"time"
)

type OpenHours struct {
	Day   string `json:"day" bson:"day"`
	Open  string `json:"open" bson:"open"`
	Close string `json:"close" bson:"close"`
}

// Validate Open hours Days, Open and Close times
func (h OpenHours) Validate() error {
	if !h.DayIsValid() {
		return errors.New("invalid day")
	}
	if !h.TimeIsValid(h.Open) {
		return errors.New("invalid open time")
	}
	if !h.TimeIsValid(h.Close) {
		return errors.New("invalid close time")
	}
	if !h.OpenBeforeClose() {
		return errors.New("open time must be before close time")
	}
	return nil
}

// Check if day is valid
func (h OpenHours) DayIsValid() bool {
	days := []string{"monday", "tuesday", "wednesday", "thursday", "friday", "saturday", "sunday"}
	for _, day := range days {
		// convert to lowercase
		if strings.ToLower(h.Day) == day {
			return true
		}
	}
	return false
}

// Check if time is valid
func (h OpenHours) TimeIsValid(t string) bool {
	_, err := time.Parse("15:04", t)
	return err == nil
}

// Check if open time is before close time
func (h OpenHours) OpenBeforeClose() bool {
	open, _ := time.Parse("15:04", h.Open)
	close, _ := time.Parse("15:04", h.Close)
	return open.Before(close)
}

// Available times
func (h OpenHours) AvailableTimes() []string {
	var times []string
	open, _ := time.Parse("15:04", h.Open)
	close, _ := time.Parse("15:04", h.Close)
	for open.Before(close) {
		times = append(times, open.Format("15:04"))
		open = open.Add(time.Minute * 30)
	}
	return times
}

// Custom date format '2020-01-01' for marshalling and unmarshalling requests
type CustomDate struct {
	time.Time
}

func (c *CustomDate) UnmarshalJSON(b []byte) error {
	s := strings.Trim(string(b), "\"")
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return err
	}
	c.Time = t
	return nil
}

func (c CustomDate) MarshalJSON() ([]byte, error) {
	return []byte(c.Time.Format("\"2006-01-02\"")), nil
}
