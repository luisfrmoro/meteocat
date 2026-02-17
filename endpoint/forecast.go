package endpoint

import (
	"context"
	"fmt"

	"github.com/luisfrmoro/meteocat/model"
)

const (
	municipalHourlyForecastPath = "/pronostic/v1/municipalHoraria"
)

// MunicipalHourlyForecast fetches a 72-hour hourly weather forecast for a specific municipality.
// This endpoint provides detailed meteorological predictions updated twice daily (approximately
// at 5 AM and 5 PM), with an hourly temporal resolution covering a 3-day forecast window.
//
// The forecast includes seven meteorological variables for each hour:
// - Temperatura (Temperature in °C)
// - Precipitació (Precipitation in mm)
// - Humitat relativa (Relative humidity in %)
// - Velocitat del vent (Wind speed in km/h)
// - Direcció del vent (Wind direction in degrees)
// - Temperatura de xafogor (Apparent temperature in °C)
// - Estat del cel (Sky conditions via symbol code)
//
// The municipality code must be obtained from the municipalities metadata endpoint.
// Symbol codes can be resolved using the symbols metadata endpoint to get visual representations.
//
// Parameters:
//   - ctx: context for cancellation and timeouts
//   - do: function to perform the actual HTTP request (typically client.do or a mock)
//   - municipalityCode: the unique 6-digit identifier of the municipality (e.g., "250019")
//
// Returns:
//   - model.MunicipalityHourlyForecast: forecast containing 3 days of hourly predictions
//   - *model.APIError: error if the request fails, municipality code is invalid, or data cannot be parsed
//
// Example:
//
//	client, _ := meteocat.NewClient("your-api-key", nil)
//	forecast, err := client.MunicipalHourlyForecast(context.Background(), "250019")
//	if err != nil {
//		log.Fatal(err)
//	}
//	fmt.Printf("Forecast for municipality %s\n", forecast.MunicipalityCode)
//	for _, day := range forecast.Days {
//		fmt.Printf("Date: %s\n", day.Date)
//		// Access individual variables using day.Variables.GetVariable("temp")
//		// or iterate over all available variables
//		if tempVar := day.Variables.GetVariable("temp"); tempVar != nil {
//			fmt.Printf("  Temperature unit: %s\n", tempVar.Unit)
//			fmt.Printf("  Number of hourly readings: %d\n", len(tempVar.Values))
//		}
//	}
func MunicipalHourlyForecast(ctx context.Context, do DoFunc, municipalityCode string) (model.MunicipalityHourlyForecast, *model.APIError) {
	resource := fmt.Sprintf("%s/%s", municipalHourlyForecastPath, municipalityCode)

	var forecast model.MunicipalityHourlyForecast
	if err := do(ctx, "GET", resource, &forecast); err != nil {
		return model.MunicipalityHourlyForecast{}, err
	}
	return forecast, nil
}
