package endpoint

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/luisfrmoro/meteocat/model"
)

const (
	testErrorExpectedNilForecast      = "expected nil forecast, got %v"
	testErrorExpectedForecastPtr      = "expected *model.MunicipalityHourlyForecast, got %T"
	testErrorExpectedForecastPath     = "expected path %s, got %s"
	testErrorExpectedDays             = "expected %d days, got %d"
	testErrorExpectedVariables        = "expected variable %s to be present"
	testErrorExpectedReadings         = "expected %d readings, got %d"
	testErrorExpectedMunicipalityCode = "expected municipality code %s, got %s"
)

// TestMunicipalHourlyForecast_Success verifies that MunicipalHourlyForecast
// parses a valid response with hourly forecast data.
func TestMunicipalHourlyForecast_Success(t *testing.T) {
	// Create test data matching the METEOCAT API response format
	expectedForecast := model.MunicipalityHourlyForecast{
		MunicipalityCode: "250019",
		Days: []model.ForecastDay{
			{
				Date: "2020-08-20Z",
				Variables: &model.ForecastVariables{
					Temperature: &model.Temperature{
						Unit: "°C",
						Values: []model.HourlyValue{
							{
								Value: "16.9",
								Time:  model.MeteocatTime{Time: parseTime(t, "2020-08-20T00:00Z")},
							},
							{
								Value: "17.6",
								Time:  model.MeteocatTime{Time: parseTime(t, "2020-08-20T01:00Z")},
							},
						},
					},
					Precipitation: &model.Precipitation{
						Unit: "mm",
						Values: []model.HourlyValue{
							{
								Value: "0.0",
								Time:  model.MeteocatTime{Time: parseTime(t, "2020-08-20T00:00Z")},
							},
							{
								Value: "0.0",
								Time:  model.MeteocatTime{Time: parseTime(t, "2020-08-20T01:00Z")},
							},
						},
					},
				},
			},
			{
				Date: "2020-08-21Z",
				Variables: &model.ForecastVariables{
					Temperature: &model.Temperature{
						Unit: "°C",
						Values: []model.HourlyValue{
							{
								Value: "18.2",
								Time:  model.MeteocatTime{Time: parseTime(t, "2020-08-21T00:00Z")},
							},
						},
					},
				},
			},
		},
	}

	expectedPath := "/pronostic/v1/municipalHoraria/250019"

	mockDo := func(ctx context.Context, method, path string, out any) *model.APIError {
		// Verify correct HTTP method
		if method != "GET" {
			t.Errorf(testErrorMethodExpected, method)
		}

		// Verify correct API path
		if path != expectedPath {
			t.Errorf(testErrorExpectedForecastPath, expectedPath, path)
		}

		// Verify output parameter is a pointer
		forecastPtr, ok := out.(*model.MunicipalityHourlyForecast)
		if !ok {
			t.Fatalf(testErrorExpectedForecastPtr, out)
		}

		// Simulate unmarshaling the response
		data, _ := json.Marshal(expectedForecast)
		json.Unmarshal(data, forecastPtr)

		return nil
	}

	ctx := context.Background()
	forecast, apiErr := MunicipalHourlyForecast(ctx, mockDo, "250019")

	// Verify no error occurred
	if apiErr != nil {
		t.Fatalf(testErrorNoError, apiErr)
	}

	// Validate the forecast response
	validateForecastResponse(t, forecast, expectedForecast)
}

// TestMunicipalHourlyForecast_APIError verifies that API errors are properly propagated
func TestMunicipalHourlyForecast_APIError(t *testing.T) {
	expectedError := &model.APIError{
		Code:    404,
		Message: "Municipality not found",
	}

	mockDo := func(ctx context.Context, method, path string, out any) *model.APIError {
		return expectedError
	}

	ctx := context.Background()
	forecast, apiErr := MunicipalHourlyForecast(ctx, mockDo, "999999")

	// Verify error is returned
	if apiErr == nil {
		t.Fatal(testErrorExpectedErrorNil)
	}

	// Verify error details
	if apiErr.Code != 404 {
		t.Errorf("expected error code 404, got %d", apiErr.Code)
	}
	if apiErr.Message != "Municipality not found" {
		t.Errorf("expected error message 'Municipality not found', got %s", apiErr.Message)
	}

	// Verify forecast is nil/zero
	if forecast.MunicipalityCode != "" {
		t.Errorf(testErrorExpectedNilForecast, forecast)
	}
}

// TestMunicipalHourlyForecast_PathConstruction verifies that the resource path
// is constructed correctly with the municipality code
func TestMunicipalHourlyForecast_PathConstruction(t *testing.T) {
	testCases := []struct {
		municipalityCode string
		expectedPath     string
	}{
		{"250019", "/pronostic/v1/municipalHoraria/250019"},
		{"080193", "/pronostic/v1/municipalHoraria/080193"},
		{"170121", "/pronostic/v1/municipalHoraria/170121"},
	}

	for _, tc := range testCases {
		paths := []string{}

		mockDo := func(ctx context.Context, method, path string, out any) *model.APIError {
			paths = append(paths, path)

			// Return empty forecast
			forecastPtr := out.(*model.MunicipalityHourlyForecast)
			*forecastPtr = model.MunicipalityHourlyForecast{}

			return nil
		}

		ctx := context.Background()
		MunicipalHourlyForecast(ctx, mockDo, tc.municipalityCode)

		if len(paths) != 1 {
			t.Fatalf("expected 1 path, got %d", len(paths))
		}
		if paths[0] != tc.expectedPath {
			t.Errorf("expected path %s, got %s", tc.expectedPath, paths[0])
		}
	}
}

// TestMunicipalHourlyForecast_EmptyVariables verifies handling of empty or missing variables
func TestMunicipalHourlyForecast_EmptyVariables(t *testing.T) {
	forecastWithEmptyVariables := model.MunicipalityHourlyForecast{
		MunicipalityCode: "250019",
		Days: []model.ForecastDay{
			{
				Date:      "2020-08-20Z",
				Variables: &model.ForecastVariables{},
			},
		},
	}

	mockDo := func(ctx context.Context, method, path string, out any) *model.APIError {
		forecastPtr := out.(*model.MunicipalityHourlyForecast)
		data, _ := json.Marshal(forecastWithEmptyVariables)
		json.Unmarshal(data, forecastPtr)
		return nil
	}

	ctx := context.Background()
	forecast, apiErr := MunicipalHourlyForecast(ctx, mockDo, "250019")

	if apiErr != nil {
		t.Fatalf(testErrorNoError, apiErr)
	}

	if forecast.MunicipalityCode != "250019" {
		t.Errorf(testErrorExpectedMunicipalityCode, "250019", forecast.MunicipalityCode)
	}

	if len(forecast.Days) != 1 {
		t.Errorf(testErrorExpectedDays, 1, len(forecast.Days))
	}

	// Verify temperature is nil for empty variables
	if forecast.Days[0].Variables.Temperature != nil {
		t.Errorf("expected nil temperature, got %v", forecast.Days[0].Variables.Temperature)
	}
}

// TestMunicipalHourlyForecast_TypedVariables verifies type-safe access to meteorological variables
func TestMunicipalHourlyForecast_TypedVariables(t *testing.T) {
	forecastVariables := &model.ForecastVariables{
		Temperature: &model.Temperature{
			Unit: "°C",
			Values: []model.HourlyValue{
				{
					Value: "20.5",
					Time:  model.MeteocatTime{Time: parseTime(t, "2020-08-20T12:00Z")},
				},
			},
		},
	}

	// Test existing temperature variable
	temp := forecastVariables.Temperature
	if temp == nil {
		t.Fatal("expected temperature variable to exist")
	}
	if temp.Unit != "°C" {
		t.Errorf("expected unit °C, got %s", temp.Unit)
	}
	if len(temp.Values) != 1 {
		t.Errorf("expected 1 reading, got %d", len(temp.Values))
	}
	if temp.Values[0].Value != model.StringOrFloat64("20.5") {
		t.Errorf("expected value 20.5, got %s", temp.Values[0].Value)
	}

	// Test non-existent precipitation variable
	precip := forecastVariables.Precipitation
	if precip != nil {
		t.Errorf("expected nil for non-existent precipitation, got %v", precip)
	}

	// Test with empty ForecastVariables
	emptyVars := &model.ForecastVariables{}
	if emptyVars.Temperature != nil {
		t.Errorf("expected nil temperature in empty variables")
	}
}

// validateForecastResponse validates the structure and content of forecast response
func validateForecastResponse(t *testing.T, forecast, expected model.MunicipalityHourlyForecast) {
	t.Helper()

	// Verify municipality code
	if forecast.MunicipalityCode != expected.MunicipalityCode {
		t.Errorf(testErrorExpectedMunicipalityCode, expected.MunicipalityCode, forecast.MunicipalityCode)
	}

	// Verify number of days
	if len(forecast.Days) != len(expected.Days) {
		t.Errorf(testErrorExpectedDays, len(expected.Days), len(forecast.Days))
	}

	// Validate each day
	for i, day := range forecast.Days {
		validateForecastDay(t, i, day, expected.Days[i])
	}
}

// validateForecastDay validates a single day in the forecast
func validateForecastDay(t *testing.T, dayIndex int, day, expectedDay model.ForecastDay) {
	t.Helper()

	if day.Date != expectedDay.Date {
		t.Errorf("day %d: expected date %s, got %s", dayIndex, expectedDay.Date, day.Date)
	}

	if day.Variables == nil {
		t.Errorf("day %d: expected variables to be present, got nil", dayIndex)
		return
	}

	// Verify variables for first day
	if dayIndex == 0 {
		validateFirstDayVariables(t, day)
	}
}

// validateFirstDayVariables validates temperature and precipitation variables in the first day
func validateFirstDayVariables(t *testing.T, day model.ForecastDay) {
	t.Helper()

	// Verify temperature variable
	if day.Variables.Temperature == nil {
		t.Errorf(testErrorExpectedVariables, "temp")
	} else if len(day.Variables.Temperature.Values) < 2 {
		t.Errorf(testErrorExpectedReadings, 2, len(day.Variables.Temperature.Values))
	}

	// Verify precipitation variable
	if day.Variables.Precipitation == nil {
		t.Errorf(testErrorExpectedVariables, "precipitacio")
	} else if len(day.Variables.Precipitation.Values) < 2 {
		t.Errorf(testErrorExpectedReadings, 2, len(day.Variables.Precipitation.Values))
	}
}

// parseTime is a helper function to parse RFC3339 timestamps for tests
func parseTime(t *testing.T, timeStr string) time.Time {
	layouts := []string{
		"2006-01-02T15:04:05Z",
		"2006-01-02T15:04Z",
		"2006-01-02Z",
	}

	for _, layout := range layouts {
		parsed, err := time.Parse(layout, timeStr)
		if err == nil {
			return parsed
		}
	}

	t.Fatalf("unable to parse time: %s", timeStr)
	return time.Time{}
}
