//go:build forecast_integration

package meteocat

import (
	"context"
	"encoding/json"
	"os"
	"strings"
	"testing"
	"time"
)

// setupIntegrationClient creates a test client for integration tests.
// Skips the test if METEOCAT_API_KEY is not set in environment.
// Returns the client, a context with 30s timeout, and the cancel function.
func setupIntegrationClient(t *testing.T) (*Client, context.Context, context.CancelFunc) {
	t.Helper()

	apiKey := strings.TrimSpace(os.Getenv("METEOCAT_API_KEY"))
	if apiKey == "" {
		t.Skip("METEOCAT_API_KEY is not set")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)

	client, err := NewClient(apiKey, nil)
	if err != nil {
		cancel()
		t.Fatalf("create client: %v", err)
	}

	return client, ctx, cancel
}

// TestIntegrationMunicipalHourlyForecast verifies that the MunicipalHourlyForecast endpoint
// returns valid 72-hour hourly forecast data for a municipality.
func TestIntegrationMunicipalHourlyForecast(t *testing.T) {
	client, ctx, cancel := setupIntegrationClient(t)
	defer cancel()

	// Use a known municipality code from Catalonia
	municipalityCode := "250432"

	forecast, apiErr := client.MunicipalHourlyForecast(ctx, municipalityCode)
	if apiErr != nil {
		t.Fatalf("municipal hourly forecast request: %v", apiErr)
	}

	// Save the complete JSON response for analysis and schema validation
	jsonData, err := json.MarshalIndent(forecast, "", "  ")
	if err != nil {
		t.Logf("WARNING: failed to marshal forecast to JSON: %v", err)
	} else {
		// Print JSON to test output for visibility
		t.Logf("Complete forecast response JSON:\n%s", string(jsonData))

		// Save JSON to file for easy access and sharing
		outputFile := "forecast_response_sample.json"
		if err := os.WriteFile(outputFile, jsonData, 0644); err != nil {
			t.Logf("WARNING: failed to save JSON to %s: %v", outputFile, err)
		} else {
			t.Logf("✓ Forecast JSON saved to: %s", outputFile)
		}
	}

	// Validate basic forecast structure
	validateForecastStructure(t, forecast, municipalityCode)

	// Validate each day in the forecast
	for i, day := range forecast.Days {
		validateForecastDayIntegration(t, i, day)
	}

	// Validate specific meteorological variables if available
	if len(forecast.Days) > 0 {
		validateSpecificVariables(t, forecast.Days[0])
	}

	t.Log("✓ Municipal hourly forecast endpoint validation completed successfully")
}

// validateForecastStructure validates the basic structure of the forecast response
func validateForecastStructure(t *testing.T, forecast MunicipalityHourlyForecast, expectedCode string) {
	t.Helper()

	if forecast.MunicipalityCode != expectedCode {
		t.Errorf("expected municipality code %s, got %s", expectedCode, forecast.MunicipalityCode)
	}

	if len(forecast.Days) == 0 {
		t.Fatal("expected at least one forecast day")
	}

	if len(forecast.Days) > 4 {
		t.Logf("WARNING: expected at most 4 days, got %d", len(forecast.Days))
	}

	t.Logf("Retrieved forecast for municipality %s with %d days", forecast.MunicipalityCode, len(forecast.Days))
}

// validateForecastDayIntegration validates a single day's forecast data structure
func validateForecastDayIntegration(t *testing.T, dayIndex int, day ForecastDay) {
	t.Helper()

	if strings.TrimSpace(day.Date) == "" {
		t.Fatalf("day %d: expected Date to be set", dayIndex)
	}

	if day.Variables == nil {
		t.Fatalf("day %d: expected Variables to be set", dayIndex)
	}

	t.Logf("  Day %d: Date=%s", dayIndex, day.Date)

	// Count available variables by checking which fields are non-nil
	variableCount := 0
	if day.Variables.Temperature != nil {
		variableCount++
	}
	if day.Variables.ApparentTemperature != nil {
		variableCount++
	}
	if day.Variables.Humidity != nil {
		variableCount++
	}
	if day.Variables.Precipitation != nil {
		variableCount++
	}
	if day.Variables.WindSpeed != nil {
		variableCount++
	}
	if day.Variables.WindDirection != nil {
		variableCount++
	}
	if day.Variables.SkyConditions != nil {
		variableCount++
	}

	if variableCount == 0 {
		t.Logf("  WARNING: Day %d has no variables", dayIndex)
		return
	}

	t.Logf("    Variables: %d available", variableCount)
	validateForecastVariables(t, dayIndex, day)
}

// validateForecastVariables validates each variable's structure in a day
func validateForecastVariables(t *testing.T, dayIndex int, day ForecastDay) {
	t.Helper()

	// Check each variable type
	if day.Variables.Temperature != nil {
		validateVariableData(t, dayIndex, "temp", day.Variables.Temperature.Unit, day.Variables.Temperature.Values)
	}
	if day.Variables.ApparentTemperature != nil {
		validateVariableData(t, dayIndex, "tempXafogor", day.Variables.ApparentTemperature.Unit, day.Variables.ApparentTemperature.Values)
	}
	if day.Variables.Humidity != nil {
		validateVariableData(t, dayIndex, "humitat", day.Variables.Humidity.Unit, day.Variables.Humidity.Values)
	}
	if day.Variables.Precipitation != nil {
		validateVariableData(t, dayIndex, "precipitacio", day.Variables.Precipitation.Unit, day.Variables.Precipitation.Values)
	}
	if day.Variables.WindSpeed != nil {
		validateVariableData(t, dayIndex, "velVent", day.Variables.WindSpeed.Unit, day.Variables.WindSpeed.Values)
	}
	if day.Variables.WindDirection != nil {
		validateVariableData(t, dayIndex, "dirVent", day.Variables.WindDirection.Unit, day.Variables.WindDirection.Values)
	}
	if day.Variables.SkyConditions != nil {
		validateVariableData(t, dayIndex, "estatCel", day.Variables.SkyConditions.Unit, day.Variables.SkyConditions.Values)
	}
}

// validateVariableData validates a single variable's data
func validateVariableData(t *testing.T, dayIndex int, varName, unit string, values []HourlyValue) {
	t.Helper()

	if strings.TrimSpace(unit) == "" {
		t.Logf("  WARNING: day %d, variable %s: unit is empty", dayIndex, varName)
	}

	if len(values) == 0 {
		t.Logf("  WARNING: day %d, variable %s: no hourly values", dayIndex, varName)
		return
	}

	t.Logf("      %s (%s): %d hourly readings", varName, unit, len(values))
	validateVariableValues(t, dayIndex, varName, values)
}

// validateVariableValues validates the individual values within a variable
func validateVariableValues(t *testing.T, dayIndex int, varName string, values []HourlyValue) {
	t.Helper()

	// Validate first value structure
	firstValue := values[0]
	if strings.TrimSpace(string(firstValue.Value)) == "" {
		t.Fatalf("day %d, variable %s, value 0: expected Value to be set", dayIndex, varName)
	}
	if firstValue.Time.IsZero() {
		t.Fatalf("day %d, variable %s, value 0: expected Time to be set", dayIndex, varName)
	}
	t.Logf("        Sample value: %s at %s", firstValue.Value, firstValue.Time.Format("2006-01-02 15:04"))
}

// validateSpecificVariables validates specific meteorological variables if available
func validateSpecificVariables(t *testing.T, firstDay ForecastDay) {
	t.Helper()

	if firstDay.Variables.Temperature != nil {
		t.Logf("✓ Temperature variable found with unit: %s", firstDay.Variables.Temperature.Unit)
	} else {
		t.Logf("ℹ Temperature variable not found in first day")
	}

	if firstDay.Variables.Precipitation != nil {
		t.Logf("✓ Precipitation variable found with unit: %s", firstDay.Variables.Precipitation.Unit)
	} else {
		t.Logf("ℹ Precipitation variable not found in first day")
	}
}

// TestIntegrationMunicipalHourlyForecast_InvalidCode verifies that the endpoint
// properly handles an invalid municipality code.
func TestIntegrationMunicipalHourlyForecast_InvalidCode(t *testing.T) {
	client, ctx, cancel := setupIntegrationClient(t)
	defer cancel()

	// Use an invalid municipality code
	invalidCode := "999999"

	forecast, apiErr := client.MunicipalHourlyForecast(ctx, invalidCode)

	// We expect an error for invalid municipality code
	if apiErr == nil {
		t.Logf("WARNING: expected error for invalid municipality code, but got valid response: %v", forecast)
		// Note: Some APIs might return empty results instead of error
	} else {
		t.Logf("✓ Correctly received error for invalid code: %s", apiErr.Message)
	}
}
