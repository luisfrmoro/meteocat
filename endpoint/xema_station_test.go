package endpoint

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/luisfrmoro/meteocat/model"
)

const (
	testErrorExpectedNilStations = "expected nil stations, got %v"
)

// TestStations_SuccessNoFilters verifies that Stations
// parses a valid response without query filters.
func TestStations_SuccessNoFilters(t *testing.T) {
	startDate := time.Date(1995, 11, 15, 10, 0, 0, 0, time.UTC)
	expectedStations := model.StationList{
		{
			Code: "CC",
			Name: "Oris",
			Type: "A",
			Coordinates: model.Coordinates{
				Latitude:  42.075052799,
				Longitude: 2.20980884646,
			},
			Location: "Abocador comarcal",
			Altitude: 626,
			Municipality: model.Municipality{
				Code: "081509",
				Name: "Oris",
			},
			County: model.Region{
				Code: 24,
				Name: "Osona",
			},
			Province: model.StationProvince{
				Code: 8,
				Name: "Barcelona",
			},
			Network: model.StationNetwork{
				Code: 1,
				Name: "XEMA",
			},
			States: []model.StationState{
				{
					Code:      2,
					StartDate: model.MeteocatTime{Time: startDate},
					EndDate:   nil,
				},
			},
		},
	}

	mockDo := func(ctx context.Context, method, path string, out any) *model.APIError {
		if method != "GET" {
			t.Errorf(testErrorMethodExpected, method)
		}
		if path != stationMetadataPath {
			t.Errorf("expected path %s, got %s", stationMetadataPath, path)
		}

		listPtr, ok := out.(*model.StationList)
		if !ok {
			t.Fatalf("expected *model.StationList, got %T", out)
		}

		data, _ := json.Marshal(expectedStations)
		json.Unmarshal(data, listPtr)

		return nil
	}

	ctx := context.Background()
	stations, apiErr := Stations(ctx, mockDo)
	if apiErr != nil {
		t.Fatalf(testErrorNoError, apiErr)
	}

	if len(stations) != 1 {
		t.Fatalf("expected 1 station, got %d", len(stations))
	}

	if stations[0].Code != "CC" {
		t.Errorf("expected Code CC, got %s", stations[0].Code)
	}
	if stations[0].Municipality.Code != "081509" {
		t.Errorf("expected municipality code 081509, got %s", stations[0].Municipality.Code)
	}
	if len(stations[0].States) != 1 {
		t.Fatalf("expected 1 state, got %d", len(stations[0].States))
	}
	if stations[0].States[0].EndDate != nil {
		t.Fatalf("expected nil end date, got %v", stations[0].States[0].EndDate)
	}
}

// TestStations_SuccessWithFilters verifies that Stations
// appends query parameters when filters are provided.
func TestStations_SuccessWithFilters(t *testing.T) {
	filterDate := time.Date(2026, 2, 17, 0, 0, 0, 0, time.UTC)
	expectedPath := stationMetadataPath + "?data=2026-02-17Z&estat=ope"

	mockDo := func(ctx context.Context, method, path string, out any) *model.APIError {
		if method != "GET" {
			t.Errorf(testErrorMethodExpected, method)
		}
		if path != expectedPath {
			t.Errorf("expected path %s, got %s", expectedPath, path)
		}

		listPtr, ok := out.(*model.StationList)
		if !ok {
			t.Fatalf("expected *model.StationList, got %T", out)
		}
		*listPtr = model.StationList{}

		return nil
	}

	ctx := context.Background()
	stations, apiErr := Stations(
		ctx,
		mockDo,
		WithStationStatus(model.StationStatusOperational),
		WithStationDate(filterDate),
	)
	if apiErr != nil {
		t.Fatalf(testErrorNoError, apiErr)
	}
	if stations == nil {
		t.Fatalf("expected empty list, got nil")
	}
}

// TestStations_APIError verifies that API errors are properly propagated.
func TestStations_APIError(t *testing.T) {
	expectedError := &model.APIError{
		Code:    401,
		Message: testErrorInvalidAPIKey,
	}

	mockDo := func(ctx context.Context, method, path string, out any) *model.APIError {
		return expectedError
	}

	ctx := context.Background()
	stations, apiErr := Stations(ctx, mockDo)
	if apiErr == nil {
		t.Fatal(testErrorExpectedErrorNil)
	}
	if apiErr.Code != 401 {
		t.Errorf("expected error code 401, got %d", apiErr.Code)
	}
	if stations != nil {
		t.Errorf(testErrorExpectedNilStations, stations)
	}
}
