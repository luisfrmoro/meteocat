package endpoint

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/luisfrmoro/meteocat/model"
)

// Test message constants
const (
	testErrorMethodExpected     = "expected method GET, got %s"
	testErrorNoError            = "expected no error, got %v"
	testErrorExpectedErrorNil   = "expected error, got nil"
	testErrorInvalidAPIKey      = "Invalid API key"
	testErrorExpectedNilRegions = "expected nil regions, got %v"
	testErrorExpectedNilMun     = "expected nil municipalities, got %v"
	testErrorExpectedNilSymbols = "expected nil symbols, got %v"
)

// TestRegions_Success verifies that the Regions function correctly
// parses a valid response with multiple regions
func TestRegions_Success(t *testing.T) {
	// Create test data matching the METEOCAT API response format
	expectedRegions := model.RegionList{
		{
			Code: 5,
			Name: "Region 5",
		},
		{
			Code: 1,
			Name: "Region 1",
		},
		{
			Code: 41,
			Name: "Region 41",
		},
	}

	// Mock DoFunc that simulates successful API response
	mockDo := func(ctx context.Context, method, path string, out any) *model.APIError {
		// Verify correct HTTP method
		if method != "GET" {
			t.Errorf(testErrorMethodExpected, method)
		}

		// Verify correct API path
		if path != "/referencia/v1/comarques" {
			t.Errorf("expected path /referencia/v1/comarques, got %s", path)
		}

		// Simulate unmarshaling the response into out parameter
		listPtr, ok := out.(*model.RegionList)
		if !ok {
			t.Fatalf("expected *model.RegionList, got %T", out)
		}

		// Marshal and unmarshal to simulate real JSON processing
		data, _ := json.Marshal(expectedRegions)
		json.Unmarshal(data, listPtr)

		return nil
	}

	ctx := context.Background()
	regions, apiErr := Regions(ctx, mockDo)

	// Verify no error occurred
	if apiErr != nil {
		t.Fatalf(testErrorNoError, apiErr)
	}

	// Verify correct number of regions
	if len(regions) != 3 {
		t.Errorf("expected 3 regions, got %d", len(regions))
	}

	// Verify first region data
	if regions[0].Code != 5 {
		t.Errorf("expected Code 5, got %d", regions[0].Code)
	}
	if regions[0].Name != "Region 5" {
		t.Errorf("expected name 'Region 5', got %s", regions[0].Name)
	}

	// Verify second region
	if regions[1].Name != "Region 1" {
		t.Errorf("expected name 'Region 1', got %s", regions[1].Name)
	}

	// Verify third region
	if regions[2].Code != 41 {
		t.Errorf("expected Code 41, got %d", regions[2].Code)
	}
}

// TestRegions_APIError verifies that API errors are properly propagated
func TestRegions_APIError(t *testing.T) {
	expectedError := &model.APIError{
		Code:    401,
		Message: testErrorInvalidAPIKey,
	}

	// Mock DoFunc that simulates an API error
	mockDo := func(ctx context.Context, method, path string, out any) *model.APIError {
		return expectedError
	}

	ctx := context.Background()
	regions, apiErr := Regions(ctx, mockDo)

	// Verify error is returned
	if apiErr == nil {
		t.Fatal(testErrorExpectedErrorNil)
	}

	// Verify error details
	if apiErr.Code != 401 {
		t.Errorf("expected error code 401, got %d", apiErr.Code)
	}
	if apiErr.Message != testErrorInvalidAPIKey {
		t.Errorf("expected error message '%s', got %s", testErrorInvalidAPIKey, apiErr.Message)
	}

	// Verify no data is returned
	if regions != nil {
		t.Errorf(testErrorExpectedNilRegions, regions)
	}
}

// TestMunicipalities_Success verifies that the Municipalities function correctly
// parses a valid response with multiple municipalities
func TestMunicipalities_Success(t *testing.T) {
	// Create test data matching the METEOCAT API response format
	expectedMunicipalities := model.MunicipalityList{
		{
			Code: "250019",
			Name: "Municipality A",
			Coordinates: &model.Coordinates{
				Latitude:  42.16239244076299,
				Longitude: 1.0928929183862726,
			},
			Region: &model.Region{
				Code: 25,
				Name: "Region 25",
			},
		},
		{
			Code: "430521",
			Name: "Municipality B",
			Coordinates: &model.Coordinates{
				Latitude:  40.90796337891453,
				Longitude: 0.49184755773298194,
			},
			Region: &model.Region{
				Code: 9,
				Name: "Region 9",
			},
		},
	}

	// Mock DoFunc that simulates successful API response
	mockDo := func(ctx context.Context, method, path string, out any) *model.APIError {
		// Verify correct HTTP method
		if method != "GET" {
			t.Errorf(testErrorMethodExpected, method)
		}

		// Verify correct API path
		if path != "/referencia/v1/municipis" {
			t.Errorf("expected path /referencia/v1/municipis, got %s", path)
		}

		// Simulate unmarshaling the response into out parameter
		listPtr, ok := out.(*model.MunicipalityList)
		if !ok {
			t.Fatalf("expected *model.MunicipalityList, got %T", out)
		}

		// Marshal and unmarshal to simulate real JSON processing
		data, _ := json.Marshal(expectedMunicipalities)
		json.Unmarshal(data, listPtr)

		return nil
	}

	ctx := context.Background()
	municipalities, apiErr := Municipalities(ctx, mockDo)

	// Verify no error occurred
	if apiErr != nil {
		t.Fatalf(testErrorNoError, apiErr)
	}

	// Verify correct number of municipalities
	if len(municipalities) != 2 {
		t.Errorf("expected 2 municipalities, got %d", len(municipalities))
	}

	// Verify first municipality data
	if municipalities[0].Code != "250019" {
		t.Errorf("expected Code '250019', got %s", municipalities[0].Code)
	}
	if municipalities[0].Name != "Municipality A" {
		t.Errorf("expected name 'Municipality A', got %s", municipalities[0].Name)
	}

	// Verify coordinates
	if municipalities[0].Coordinates == nil {
		t.Fatal("expected coordinates, got nil")
	}
	if municipalities[0].Coordinates.Latitude != 42.16239244076299 {
		t.Errorf("expected latitude 42.16239244076299, got %f", municipalities[0].Coordinates.Latitude)
	}
	if municipalities[0].Coordinates.Longitude != 1.0928929183862726 {
		t.Errorf("expected longitude 1.0928929183862726, got %f", municipalities[0].Coordinates.Longitude)
	}

	// Verify region reference
	if municipalities[0].Region == nil {
		t.Fatal("expected region reference, got nil")
	}
	if municipalities[0].Region.Name != "Region 25" {
		t.Errorf("expected region 'Region 25', got %s", municipalities[0].Region.Name)
	}
}

// TestMunicipalities_APIError verifies that API errors are properly propagated
func TestMunicipalities_APIError(t *testing.T) {
	expectedError := &model.APIError{
		Code:    401,
		Message: testErrorInvalidAPIKey,
	}

	// Mock DoFunc that simulates an API error
	mockDo := func(ctx context.Context, method, path string, out any) *model.APIError {
		return expectedError
	}

	ctx := context.Background()
	municipalities, apiErr := Municipalities(ctx, mockDo)

	// Verify error is returned
	if apiErr == nil {
		t.Fatal(testErrorExpectedErrorNil)
	}

	// Verify no data is returned
	if municipalities != nil {
		t.Errorf(testErrorExpectedNilMun, municipalities)
	}
}

// TestSymbols_Success verifies that the Symbols function correctly
// parses a valid response with multiple symbol categories
func TestSymbols_Success(t *testing.T) {
	// Create test data matching the METEOCAT API response format
	expectedSymbols := model.SymbolList{
		{
			Name:        "sky",
			Description: "sky state",
			Values: []model.SymbolValue{
				{
					Code:         "1",
					Name:         "Clear Sky",
					Description:  "",
					Category:     "cloudiness",
					IconURL:      "https://example.com/sky/1.svg",
					IconURLNight: "https://example.com/sky/1n.svg",
				},
				{
					Code:         "2",
					Name:         "Cloudy",
					Description:  "",
					Category:     "cloudiness",
					IconURL:      "https://example.com/sky/2.svg",
					IconURLNight: "https://example.com/sky/2n.svg",
				},
			},
		},
		{
			Name:        "snow",
			Description: "snow accumulation",
			Values: []model.SymbolValue{
				{
					Code:         "1",
					Name:         "Negligible",
					Description:  "less than 2 cm in 24 hours",
					Category:     "snow_accumulation",
					IconURL:      "https://example.com/snow/1.png",
					IconURLNight: "",
				},
			},
		},
	}

	// Mock DoFunc that simulates successful API response
	mockDo := func(ctx context.Context, method, path string, out any) *model.APIError {
		// Verify correct HTTP method
		if method != "GET" {
			t.Errorf(testErrorMethodExpected, method)
		}

		// Verify correct API path
		if path != "/referencia/v1/simbols" {
			t.Errorf("expected path /referencia/v1/simbols, got %s", path)
		}

		// Simulate unmarshaling the response into out parameter
		listPtr, ok := out.(*model.SymbolList)
		if !ok {
			t.Fatalf("expected *model.SymbolList, got %T", out)
		}

		// Marshal and unmarshal to simulate real JSON processing
		data, _ := json.Marshal(expectedSymbols)
		json.Unmarshal(data, listPtr)

		return nil
	}

	ctx := context.Background()
	symbols, apiErr := Symbols(ctx, mockDo)

	// Verify no error occurred
	if apiErr != nil {
		t.Fatalf(testErrorNoError, apiErr)
	}

	// Verify correct number of symbol categories
	if len(symbols) != 2 {
		t.Errorf("expected 2 symbol categories, got %d", len(symbols))
	}

	// Verify first symbol category
	if symbols[0].Name != "sky" {
		t.Errorf("expected name 'sky', got %s", symbols[0].Name)
	}

	// Verify values in first category
	if len(symbols[0].Values) != 2 {
		t.Errorf("expected 2 values in first category, got %d", len(symbols[0].Values))
	}

	// Verify first symbol value
	if symbols[0].Values[0].Code != "1" {
		t.Errorf("expected Code '1', got %s", symbols[0].Values[0].Code)
	}
	if symbols[0].Values[0].Name != "Clear Sky" {
		t.Errorf("expected name 'Clear Sky', got %s", symbols[0].Values[0].Name)
	}
	if symbols[0].Values[0].IconURL != "https://example.com/sky/1.svg" {
		t.Errorf("expected IconURL, got %s", symbols[0].Values[0].IconURL)
	}

	// Verify second symbol category
	if symbols[1].Name != "snow" {
		t.Errorf("expected name 'snow', got %s", symbols[1].Name)
	}
}

// TestSymbols_APIError verifies that API errors are properly propagated
func TestSymbols_APIError(t *testing.T) {
	expectedError := &model.APIError{
		Code:    401,
		Message: testErrorInvalidAPIKey,
	}

	// Mock DoFunc that simulates an API error
	mockDo := func(ctx context.Context, method, path string, out any) *model.APIError {
		return expectedError
	}

	ctx := context.Background()
	symbols, apiErr := Symbols(ctx, mockDo)

	// Verify error is returned
	if apiErr == nil {
		t.Fatal(testErrorExpectedErrorNil)
	}

	// Verify no data is returned
	if symbols != nil {
		t.Errorf(testErrorExpectedNilSymbols, symbols)
	}
}

// TestSymbols_EmptyResult verifies handling of empty symbol list
func TestSymbols_EmptyResult(t *testing.T) {
	// Mock DoFunc that returns empty list
	mockDo := func(ctx context.Context, method, path string, out any) *model.APIError {
		listPtr, ok := out.(*model.SymbolList)
		if !ok {
			t.Fatalf("expected *model.SymbolList, got %T", out)
		}

		// Return empty list
		*listPtr = model.SymbolList{}
		return nil
	}

	ctx := context.Background()
	symbols, apiErr := Symbols(ctx, mockDo)

	// Verify no error occurred
	if apiErr != nil {
		t.Fatalf(testErrorNoError, apiErr)
	}

	// Verify empty list
	if len(symbols) != 0 {
		t.Errorf("expected 0 symbols, got %d", len(symbols))
	}
}

// TestRegions_ContextCancellation verifies context cancellation is respected
func TestRegions_ContextCancellation(t *testing.T) {
	// Mock DoFunc that checks for cancelled context
	mockDo := func(ctx context.Context, method, path string, out any) *model.APIError {
		// Simulate checking context cancellation
		if ctx.Err() != nil {
			return &model.APIError{
				Code:    0,
				Message: "context cancelled",
			}
		}
		return nil
	}

	// Create cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	regions, apiErr := Regions(ctx, mockDo)

	// Verify error is returned
	if apiErr == nil {
		t.Fatal("expected error for cancelled context, got nil")
	}

	// Verify no data is returned
	if regions != nil {
		t.Errorf(testErrorExpectedNilRegions, regions)
	}
}
