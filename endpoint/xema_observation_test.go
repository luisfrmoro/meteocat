package endpoint

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/luisfrmoro/meteocat/model"
)

const (
	testErrorExpectedNilObservations       = "expected nil observations, got %v"
	testErrorExpectedNilVariables          = "expected nil variables, got %v"
	testErrorExpectedPath                  = "expected path %s, got %s"
	testErrorExpectedStationObservationPtr = "expected *model.StationObservationList, got %T"
	testErrorExpectedVariablePtr           = "expected *model.VariableList, got %T"
)

// TestObservations_Success verifies that Observations
// parses a valid response with station observations.
func TestObservations_Success(t *testing.T) {
	testDate := time.Date(2020, time.June, 16, 0, 0, 0, 0, time.UTC)
	expectedObservations := model.StationObservationList{
		{
			Code: "CC",
			Variables: []model.VariableObservation{
				{
					Code: 1,
					Readings: []model.Reading{
						{
							Data: model.MeteocatTime{Time: time.Date(2020, 6, 16, 0, 0, 0, 0, time.UTC)},
							DataExtrem: &model.MeteocatTime{
								Time: time.Date(2020, 6, 16, 0, 5, 0, 0, time.UTC),
							},
							Value:    947.3,
							Status:   "V",
							TimeBase: "SH",
						},
					},
				},
				{
					Code: 30,
					Readings: []model.Reading{
						{
							Data:     model.MeteocatTime{Time: time.Date(2020, 6, 16, 0, 0, 0, 0, time.UTC)},
							Value:    0.6,
							Status:   "V",
							TimeBase: "SH",
						},
						{
							Data:     model.MeteocatTime{Time: time.Date(2020, 6, 16, 0, 30, 0, 0, time.UTC)},
							Value:    0.6,
							Status:   "V",
							TimeBase: "SH",
						},
					},
				},
			},
		},
	}

	expectedPath := "/xema/v1/estacions/mesurades/CC/2020/06/16"

	mockDo := func(ctx context.Context, method, path string, out any) *model.APIError {
		if method != "GET" {
			t.Errorf(testErrorMethodExpected, method)
		}
		if path != expectedPath {
			t.Errorf(testErrorExpectedPath, expectedPath, path)
		}

		listPtr, ok := out.(*model.StationObservationList)
		if !ok {
			t.Fatalf(testErrorExpectedStationObservationPtr, out)
		}

		data, _ := json.Marshal(expectedObservations)
		json.Unmarshal(data, listPtr)

		return nil
	}

	ctx := context.Background()
	observations, apiErr := Observations(ctx, mockDo, "CC", testDate)
	if apiErr != nil {
		t.Fatalf(testErrorNoError, apiErr)
	}

	validateObservationsResponse(t, observations)
}

// validateObservationsResponse validates the structure and content of observations response.
func validateObservationsResponse(t *testing.T, observations model.StationObservationList) {
	t.Helper()

	if len(observations) != 1 {
		t.Fatalf("expected 1 station observation, got %d", len(observations))
	}

	if observations[0].Code != "CC" {
		t.Errorf("expected Code CC, got %s", observations[0].Code)
	}

	if len(observations[0].Variables) != 2 {
		t.Fatalf("expected 2 variables, got %d", len(observations[0].Variables))
	}

	validateFirstVariable(t, observations[0].Variables[0])
	validateSecondVariable(t, observations[0].Variables[1])
}

// validateFirstVariable validates the first variable in observations.
func validateFirstVariable(t *testing.T, varObs model.VariableObservation) {
	t.Helper()

	if varObs.Code != 1 {
		t.Errorf("expected variable code 1, got %d", varObs.Code)
	}
	if len(varObs.Readings) != 1 {
		t.Fatalf("expected 1 reading for variable 1, got %d", len(varObs.Readings))
	}
	if varObs.Readings[0].Value != 947.3 {
		t.Errorf("expected value 947.3, got %f", varObs.Readings[0].Value)
	}
	if varObs.Readings[0].Status != "V" {
		t.Errorf("expected status V, got %s", varObs.Readings[0].Status)
	}
	if varObs.Readings[0].DataExtrem == nil {
		t.Error("expected DataExtrem to be set")
	}
}

// validateSecondVariable validates the second variable in observations.
func validateSecondVariable(t *testing.T, varObs model.VariableObservation) {
	t.Helper()

	if varObs.Code != 30 {
		t.Errorf("expected variable code 30, got %d", varObs.Code)
	}
	if len(varObs.Readings) != 2 {
		t.Fatalf("expected 2 readings for variable 30, got %d", len(varObs.Readings))
	}
}

// TestObservations_DateFormatting verifies that the date is correctly formatted in the URL.
func TestObservations_DateFormatting(t *testing.T) {
	tests := []struct {
		name         string
		stationCode  string
		date         time.Time
		expectedPath string
	}{
		{
			name:         "single digit month and day",
			stationCode:  "AB",
			date:         time.Date(2020, 1, 5, 0, 0, 0, 0, time.UTC),
			expectedPath: "/xema/v1/estacions/mesurades/AB/2020/01/05",
		},
		{
			name:         "double digit month and day",
			stationCode:  "XY",
			date:         time.Date(2020, 12, 31, 0, 0, 0, 0, time.UTC),
			expectedPath: "/xema/v1/estacions/mesurades/XY/2020/12/31",
		},
		{
			name:         "leap year february",
			stationCode:  "CD",
			date:         time.Date(2020, 2, 29, 0, 0, 0, 0, time.UTC),
			expectedPath: "/xema/v1/estacions/mesurades/CD/2020/02/29",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDo := func(ctx context.Context, method, path string, out any) *model.APIError {
				if path != tt.expectedPath {
					t.Errorf(testErrorExpectedPath, tt.expectedPath, path)
				}
				listPtr, ok := out.(*model.StationObservationList)
				if !ok {
					t.Fatalf(testErrorExpectedStationObservationPtr, out)
				}
				*listPtr = model.StationObservationList{}
				return nil
			}

			ctx := context.Background()
			_, apiErr := Observations(ctx, mockDo, tt.stationCode, tt.date)
			if apiErr != nil {
				t.Fatalf(testErrorNoError, apiErr)
			}
		})
	}
}

// TestObservations_APIError verifies that API errors are properly propagated.
func TestObservations_APIError(t *testing.T) {
	expectedError := &model.APIError{
		Code:    404,
		Message: "Station not found",
	}

	mockDo := func(ctx context.Context, method, path string, out any) *model.APIError {
		return expectedError
	}

	ctx := context.Background()
	testDate := time.Date(2020, 6, 16, 0, 0, 0, 0, time.UTC)
	observations, apiErr := Observations(ctx, mockDo, "INVALID", testDate)

	if apiErr == nil {
		t.Fatal(testErrorExpectedErrorNil)
	}
	if apiErr.Code != 404 {
		t.Errorf("expected error code 404, got %d", apiErr.Code)
	}
	if apiErr.Message != "Station not found" {
		t.Errorf("expected message 'Station not found', got %s", apiErr.Message)
	}
	if observations != nil {
		t.Errorf(testErrorExpectedNilObservations, observations)
	}
}

// TestObservations_EmptyResult verifies that an empty result is handled correctly.
func TestObservations_EmptyResult(t *testing.T) {
	mockDo := func(ctx context.Context, method, path string, out any) *model.APIError {
		listPtr, ok := out.(*model.StationObservationList)
		if !ok {
			t.Fatalf(testErrorExpectedStationObservationPtr, out)
		}
		*listPtr = model.StationObservationList{}
		return nil
	}

	ctx := context.Background()
	testDate := time.Date(2020, 6, 16, 0, 0, 0, 0, time.UTC)
	observations, apiErr := Observations(ctx, mockDo, "CC", testDate)

	if apiErr != nil {
		t.Fatalf(testErrorNoError, apiErr)
	}
	if observations == nil {
		t.Fatal("expected empty list, got nil")
	}
	if len(observations) != 0 {
		t.Errorf("expected 0 observations, got %d", len(observations))
	}
}

// TestVariables_Success verifies that Variables
// parses a valid response with variable metadata.
func TestVariables_Success(t *testing.T) {
	expectedVariables := model.VariableList{
		{
			Code:     1,
			Name:     "Pressió atmosfèrica màxima",
			Unit:     "hPa",
			Acronym:  "Px",
			Type:     "DAT",
			Decimals: 1,
		},
		{
			Code:     30,
			Name:     "Temperatura",
			Unit:     "°C",
			Acronym:  "T",
			Type:     "DAT",
			Decimals: 1,
		},
		{
			Code:     97,
			Name:     "Temperatura superficial del mar",
			Unit:     "°C",
			Acronym:  "TMAR",
			Type:     "DAT",
			Decimals: 1,
		},
	}

	mockDo := func(ctx context.Context, method, path string, out any) *model.APIError {
		if method != "GET" {
			t.Errorf(testErrorMethodExpected, method)
		}
		if path != variablesMetadataPath {
			t.Errorf(testErrorExpectedPath, variablesMetadataPath, path)
		}

		listPtr, ok := out.(*model.VariableList)
		if !ok {
			t.Fatalf(testErrorExpectedVariablePtr, out)
		}

		data, _ := json.Marshal(expectedVariables)
		json.Unmarshal(data, listPtr)

		return nil
	}

	ctx := context.Background()
	variables, apiErr := Variables(ctx, mockDo)
	if apiErr != nil {
		t.Fatalf(testErrorNoError, apiErr)
	}

	validateVariablesResponse(t, variables)
}

// validateVariablesResponse validates the structure and content of variables response.
func validateVariablesResponse(t *testing.T, variables model.VariableList) {
	t.Helper()

	if len(variables) != 3 {
		t.Fatalf("expected 3 variables, got %d", len(variables))
	}

	validateFirstVariableMetadata(t, variables[0])

	// Verify second variable
	if variables[1].Code != 30 {
		t.Errorf("expected Code 30, got %d", variables[1].Code)
	}

	validateThirdVariableMetadata(t, variables[2])
}

// validateFirstVariableMetadata validates the first variable metadata.
func validateFirstVariableMetadata(t *testing.T, variable model.Variable) {
	t.Helper()

	if variable.Code != 1 {
		t.Errorf("expected Code 1, got %d", variable.Code)
	}
	if variable.Name != "Pressió atmosfèrica màxima" {
		t.Errorf("expected name 'Pressió atmosfèrica màxima', got %s", variable.Name)
	}
	if variable.Unit != "hPa" {
		t.Errorf("expected unit 'hPa', got %s", variable.Unit)
	}
	if variable.Acronym != "Px" {
		t.Errorf("expected acronym 'Px', got %s", variable.Acronym)
	}
	if variable.Type != "DAT" {
		t.Errorf("expected type 'DAT', got %s", variable.Type)
	}
	if variable.Decimals != 1 {
		t.Errorf("expected decimals 1, got %d", variable.Decimals)
	}
}

// validateThirdVariableMetadata validates the third variable metadata.
func validateThirdVariableMetadata(t *testing.T, variable model.Variable) {
	t.Helper()

	if variable.Code != 97 {
		t.Errorf("expected Code 97, got %d", variable.Code)
	}
	if variable.Acronym != "TMAR" {
		t.Errorf("expected acronym 'TMAR', got %s", variable.Acronym)
	}
}

// TestVariables_APIError verifies that API errors are properly propagated.
func TestVariables_APIError(t *testing.T) {
	expectedError := &model.APIError{
		Code:    401,
		Message: testErrorInvalidAPIKey,
	}

	mockDo := func(ctx context.Context, method, path string, out any) *model.APIError {
		return expectedError
	}

	ctx := context.Background()
	variables, apiErr := Variables(ctx, mockDo)

	if apiErr == nil {
		t.Fatal(testErrorExpectedErrorNil)
	}
	if apiErr.Code != 401 {
		t.Errorf("expected error code 401, got %d", apiErr.Code)
	}
	if apiErr.Message != testErrorInvalidAPIKey {
		t.Errorf("expected message '%s', got %s", testErrorInvalidAPIKey, apiErr.Message)
	}
	if variables != nil {
		t.Errorf(testErrorExpectedNilVariables, variables)
	}
}

// TestVariables_EmptyResult verifies that an empty result is handled correctly.
func TestVariables_EmptyResult(t *testing.T) {
	mockDo := func(ctx context.Context, method, path string, out any) *model.APIError {
		listPtr, ok := out.(*model.VariableList)
		if !ok {
			t.Fatalf("expected *model.VariableList, got %T", out)
		}
		*listPtr = model.VariableList{}
		return nil
	}

	ctx := context.Background()
	variables, apiErr := Variables(ctx, mockDo)

	if apiErr != nil {
		t.Fatalf(testErrorNoError, apiErr)
	}
	if variables == nil {
		t.Fatal("expected empty list, got nil")
	}
	if len(variables) != 0 {
		t.Errorf("expected 0 variables, got %d", len(variables))
	}
}

// TestObservations_ContextCancellation verifies context cancellation handling.
func TestObservations_ContextCancellation(t *testing.T) {
	mockDo := func(ctx context.Context, method, path string, out any) *model.APIError {
		return &model.APIError{
			Message: "context canceled",
		}
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	testDate := time.Date(2020, 6, 16, 0, 0, 0, 0, time.UTC)
	observations, apiErr := Observations(ctx, mockDo, "CC", testDate)

	if apiErr == nil {
		t.Fatal(testErrorExpectedErrorNil)
	}
	if observations != nil {
		t.Errorf(testErrorExpectedNilObservations, observations)
	}
}

// TestVariables_ContextCancellation verifies context cancellation handling.
func TestVariables_ContextCancellation(t *testing.T) {
	mockDo := func(ctx context.Context, method, path string, out any) *model.APIError {
		return &model.APIError{
			Message: "context canceled",
		}
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	variables, apiErr := Variables(ctx, mockDo)

	if apiErr == nil {
		t.Fatal(testErrorExpectedErrorNil)
	}
	if variables != nil {
		t.Errorf(testErrorExpectedNilVariables, variables)
	}
}
