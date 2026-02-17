//go:build xema_integration

package meteocat

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/luisfrmoro/meteocat/model"
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

// validateStationFields checks that a station has required fields.
func validateStationFields(t *testing.T, i int, station *model.Station) {
	t.Helper()

	if strings.TrimSpace(station.Code) == "" {
		t.Fatalf("station %d: expected Code to be set", i)
	}
	if strings.TrimSpace(station.Name) == "" {
		t.Fatalf("station %d (Code=%s): expected Name to be set", i, station.Code)
	}
	if station.Coordinates.Latitude == 0 && station.Coordinates.Longitude == 0 {
		t.Fatalf("station %d (Code=%s): expected Coordinates to be set", i, station.Code)
	}
	if strings.TrimSpace(station.Municipality.Code) == "" {
		t.Fatalf("station %d (Code=%s): expected Municipality.Code to be set", i, station.Code)
	}
	if strings.TrimSpace(station.Municipality.Name) == "" {
		t.Fatalf("station %d (Code=%s): expected Municipality.Name to be set", i, station.Code)
	}
}

// TestIntegrationStations verifies that the XEMA station metadata endpoint returns
// valid station data.
func TestIntegrationStations(t *testing.T) {
	client, ctx, cancel := setupIntegrationClient(t)
	defer cancel()

	stations, apiErr := client.Stations(ctx)
	if apiErr != nil {
		t.Fatalf("stations metadata request: %v", apiErr)
	}

	if len(stations) == 0 {
		t.Fatal("expected at least one station")
	}

	t.Logf("Retrieved %d stations", len(stations))

	for i, station := range stations {
		validateStationFields(t, i, &station)

		if len(station.States) == 0 {
			t.Fatalf("station %d (Code=%s): expected at least one state", i, station.Code)
		}
		if station.States[0].StartDate.IsZero() {
			t.Fatalf("station %d (Code=%s): expected state StartDate to be set", i, station.Code)
		}

		if i < 3 {
			t.Logf("  Station %d: Code=%s, Name=%q, Lat=%.4f, Lon=%.4f", i, station.Code, station.Name, station.Coordinates.Latitude, station.Coordinates.Longitude)
		}
	}

	t.Log("✓ Stations metadata endpoint validation completed successfully")
}

// validateVariableFields checks that a variable has required fields.
func validateVariableFields(t *testing.T, i int, variable *model.Variable) {
	t.Helper()

	if variable.Code == 0 {
		t.Fatalf("variable %d: expected Code to be set", i)
	}
	if strings.TrimSpace(variable.Name) == "" {
		t.Fatalf("variable %d (Code=%d): expected Name to be set", i, variable.Code)
	}
	if strings.TrimSpace(variable.Unit) == "" {
		t.Fatalf("variable %d (Code=%d): expected Unit to be set", i, variable.Code)
	}
	if strings.TrimSpace(variable.Acronym) == "" {
		t.Fatalf("variable %d (Code=%d): expected Acronym to be set", i, variable.Code)
	}
	if strings.TrimSpace(variable.Type) == "" {
		t.Fatalf("variable %d (Code=%d): expected Type to be set", i, variable.Code)
	}
	if variable.Decimals < 0 {
		t.Fatalf("variable %d (Code=%d): Decimals should not be negative", i, variable.Code)
	}
}

// validateObservationFields checks that a station observation has required fields.
func validateObservationFields(t *testing.T, i int, obs *model.StationObservation) {
	t.Helper()

	if strings.TrimSpace(obs.Code) == "" {
		t.Fatalf("observation %d: expected Code to be set", i)
	}
	if len(obs.Variables) == 0 {
		t.Fatalf("observation %d (Code=%s): expected at least one variable", i, obs.Code)
	}

	for j, varObs := range obs.Variables {
		if varObs.Code == 0 {
			t.Fatalf("observation %d, variable %d: expected Code to be set", i, j)
		}
		if len(varObs.Readings) == 0 {
			t.Fatalf("observation %d (Code=%s), variable %d (Code=%d): expected at least one reading", i, obs.Code, j, varObs.Code)
		}

		for k, reading := range varObs.Readings {
			if reading.Data.IsZero() {
				t.Fatalf("observation %d, variable %d, reading %d: expected Data timestamp to be set", i, j, k)
			}
			// DataExtrem is optional (only present for extreme values)
			// Status can be empty (blank) when validation has not started yet
			// Valid Status values: "" (not validated), "T" (pending), "V" (valid), "N" (invalid)
			if strings.TrimSpace(reading.TimeBase) == "" {
				t.Fatalf("observation %d, variable %d, reading %d: expected TimeBase to be set", i, j, k)
			}
		}
	}
}

// TestIntegrationVariables verifies that the XEMA variables metadata endpoint returns
// valid variable metadata for all measurement variables.
func TestIntegrationVariables(t *testing.T) {
	client, ctx, cancel := setupIntegrationClient(t)
	defer cancel()

	variables, apiErr := client.Variables(ctx)
	if apiErr != nil {
		t.Fatalf("variables metadata request: %v", apiErr)
	}

	if len(variables) == 0 {
		t.Fatal("expected at least one variable")
	}

	t.Logf("Retrieved %d variables", len(variables))

	for i, variable := range variables {
		validateVariableFields(t, i, &variable)

		if i < 5 {
			t.Logf("  Variable %d: Code=%d, Name=%q, Unit=%s, Decimals=%d", i, variable.Code, variable.Name, variable.Unit, variable.Decimals)
		}
	}

	t.Log("✓ Variables metadata endpoint validation completed successfully")
}

// TestIntegrationObservations verifies that the XEMA observations endpoint returns
// valid observation data for a station on a specific date.
func TestIntegrationObservations(t *testing.T) {
	const dateFormat = "2006-01-02"

	client, ctx, cancel := setupIntegrationClient(t)
	defer cancel()

	// First get a list of operational stations to find a valid station code
	stationsDate := time.Date(2026, time.February, 16, 0, 0, 0, 0, time.UTC)
	stations, apiErr := client.Stations(ctx, WithStationStatus(model.StationStatusOperational), WithStationDate(stationsDate))
	if apiErr != nil {
		t.Fatalf("stations metadata request: %v", apiErr)
	}

	if len(stations) == 0 {
		t.Fatal("expected at least one operational station")
	}

	// Use the first operational station
	stationCode := stations[0].Code
	t.Logf("Using station %q for observations test", stationCode)

	// Use a date known to have data from the API documentation example
	// The example shows data from 2026-02-16
	testDate := time.Date(2026, time.February, 16, 0, 0, 0, 0, time.UTC)

	observations, apiErr := client.Observations(ctx, stationCode, testDate)
	if apiErr != nil {
		t.Fatalf("observations request for station %s on %s: %v", stationCode, testDate.Format(dateFormat), apiErr)
	}

	// Note: Not all stations may have observations on all dates.
	// If the response is empty, this is expected behavior, not an error.
	if len(observations) == 0 {
		t.Logf("No observations found for station %s on %s (this is expected if the station has no data for this date)", stationCode, testDate.Format(dateFormat))
		return
	}

	t.Logf("Retrieved %d station observations for date %s", len(observations), testDate.Format(dateFormat))

	for i, obs := range observations {
		validateObservationFields(t, i, &obs)

		totalReadings := 0
		for _, varObs := range obs.Variables {
			totalReadings += len(varObs.Readings)
		}

		t.Logf("  Station %s: %d variables, %d total readings", obs.Code, len(obs.Variables), totalReadings)

		// Log details of first variable if it exists
		if len(obs.Variables) > 0 {
			firstVar := obs.Variables[0]
			if len(firstVar.Readings) > 0 {
				firstReading := firstVar.Readings[0]
				t.Logf("    First variable (Code=%d): %d readings, first value=%.2f at %s", firstVar.Code, len(firstVar.Readings), firstReading.Value, firstReading.Data.Format("2006-01-02T15:04Z"))
			}
		}
	}

	t.Log("✓ Observations endpoint validation completed successfully")
}
