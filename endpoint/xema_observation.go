package endpoint

import (
	"context"
	"fmt"
	"time"

	"github.com/luisfrmoro/meteocat/model"
)

const (
	stationObservationsPath = "/xema/v1/estacions/mesurades"
	variablesMetadataPath   = "/xema/v1/variables/mesurades/metadades"
)

// Observations fetches all observations of all variables recorded at a station for a specific day.
// The endpoint returns observation measurements grouped by variable, with each variable containing
// a list of readings taken throughout the day.
//
// Parameters:
//   - ctx: context for cancellation and timeouts
//   - do: function to perform the actual HTTP request (typically client.do or a mock)
//   - stationCode: the unique identifier of the station (e.g., "CC")
//   - date: the specific date for which observations are requested
//
// Returns:
//   - model.StationObservationList: list of observations with all variables and readings
//   - *model.APIError: error if the request fails or data cannot be parsed
func Observations(ctx context.Context, do DoFunc, stationCode string, date time.Time) (model.StationObservationList, *model.APIError) {
	year := date.UTC().Year()
	month := date.UTC().Month()
	day := date.UTC().Day()

	resource := fmt.Sprintf("%s/%s/%04d/%02d/%02d", stationObservationsPath, stationCode, year, month, day)

	var list model.StationObservationList
	if err := do(ctx, "GET", resource, &list); err != nil {
		return nil, err
	}
	return list, nil
}

// Variables fetches the metadata of all XEMA variables.
// The endpoint returns information about all variables independently from the stations where they are measured.
// This reference data is essential for understanding variable codes, units, decimal precision, and other properties.
//
// Parameters:
//   - ctx: context for cancellation and timeouts
//   - do: function to perform the actual HTTP request (typically client.do or a mock)
//
// Returns:
//   - model.VariableList: list of variable metadata
//   - *model.APIError: error if the request fails or data cannot be parsed
func Variables(ctx context.Context, do DoFunc) (model.VariableList, *model.APIError) {
	var list model.VariableList
	if err := do(ctx, "GET", variablesMetadataPath, &list); err != nil {
		return nil, err
	}
	return list, nil
}
