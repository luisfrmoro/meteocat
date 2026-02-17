package model

import (
	"encoding/json"
	"fmt"
)

// HourlyValue represents a single meteorological measurement at a specific hour.
// Each hourly value includes the actual measurement and its corresponding timestamp.
type HourlyValue struct {
	// Value is the numeric measurement result for this hour, parsed from JSON
	// The API may return this as either a string or a number
	Value StringOrFloat64 `json:"valor"`

	// Time is the timestamp (in UTC) of the measurement
	Time MeteocatTime `json:"data"`
}

// StringOrFloat64 handles JSON values that may be either strings or numbers
type StringOrFloat64 string

// UnmarshalJSON unmarshals a string or number into StringOrFloat64
func (s *StringOrFloat64) UnmarshalJSON(data []byte) error {
	if len(data) == 0 {
		*s = "0"
		return nil
	}

	// Try to unmarshal as string first (with quotes)
	var str string
	if err := json.Unmarshal(data, &str); err == nil {
		*s = StringOrFloat64(str)
		return nil
	}

	// If that fails, try to unmarshal as number
	var num float64
	if err := json.Unmarshal(data, &num); err == nil {
		*s = StringOrFloat64(fmt.Sprintf("%v", num))
		return nil
	}

	// If both fail, store the raw string representation
	*s = StringOrFloat64(string(data))
	return nil
}

// MarshalJSON marshals StringOrFloat64 to JSON as a quoted string
func (s StringOrFloat64) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(s))
}

// Temperature represents hourly temperature forecasts in Celsius
type Temperature struct {
	Unit   string        `json:"unitat"`
	Values []HourlyValue `json:"valors"`
}

// ApparentTemperature represents hourly apparent temperature (xafogor) forecasts in Celsius
type ApparentTemperature struct {
	Unit   string        `json:"unitat"`
	Values []HourlyValue `json:"valors"`
}

// Humidity represents hourly relative humidity forecasts in percentage
type Humidity struct {
	Unit   string        `json:"unitat"`
	Values []HourlyValue `json:"valors"`
}

// Precipitation represents hourly precipitation forecasts in millimeters
type Precipitation struct {
	Unit   string        `json:"unitat"`
	Values []HourlyValue `json:"valor"`
}

// WindSpeed represents hourly wind speed forecasts in km/h
type WindSpeed struct {
	Unit   string        `json:"unitat"`
	Values []HourlyValue `json:"valors"`
}

// WindDirection represents hourly wind direction forecasts in degrees
type WindDirection struct {
	Unit   string        `json:"unitat,omitempty"`
	Values []HourlyValue `json:"valors"`
}

// SkyConditions represents hourly sky condition/weather symbol forecasts
type SkyConditions struct {
	Unit   string        `json:"unitat,omitempty"`
	Values []HourlyValue `json:"valors"`
}

// ForecastVariables holds all meteorological variables available in a forecast for a day.
// Each variable contains an hourly time series of measurements with their units.
type ForecastVariables struct {
	// Temperature is the temperature in degrees Celsius
	Temperature *Temperature `json:"temp"`

	// ApparentTemperature is the apparent temperature (feels like) in degrees Celsius
	ApparentTemperature *ApparentTemperature `json:"tempXafogor"`

	// Humidity is the relative humidity in percentage
	Humidity *Humidity `json:"humitat"`

	// Precipitation is the precipitation in millimeters
	Precipitation *Precipitation `json:"precipitacio"`

	// WindSpeed is the wind speed in km/h
	WindSpeed *WindSpeed `json:"velVent"`

	// WindDirection is the wind direction in degrees
	WindDirection *WindDirection `json:"dirVent"`

	// SkyConditions is the sky state/weather conditions as symbol codes
	SkyConditions *SkyConditions `json:"estatCel"`
}

// ForecastDay represents all forecast data for a single day.
// It contains the date and all meteorological variables available for that day.
type ForecastDay struct {
	// Date is the date of the forecast in format "YYYY-MM-DDTZ" (e.g., "2020-08-20Z")
	Date string `json:"data"`

	// Variables holds all meteorological variables available for this forecast day
	Variables *ForecastVariables `json:"variables"`
}

// MunicipalityHourlyForecast represents a complete 72-hour hourly forecast for a single municipality.
// The forecast is updated twice daily (approximately at 5 AM and 5 PM) and provides
// hourly predictions for temperature, precipitation, humidity, wind speed/direction,
// apparent temperature, and sky conditions.
type MunicipalityHourlyForecast struct {
	// MunicipalityCode is the unique 6-digit identifier for the municipality (e.g., "250019")
	// This code can be obtained from the municipalities metadata endpoint
	MunicipalityCode string `json:"codiMunicipi"`

	// Days contains the forecast data for each day in the 72-hour window
	// Typically contains 3 days of data with hourly resolution
	Days []ForecastDay `json:"dies"`
}
