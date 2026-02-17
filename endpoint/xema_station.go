package endpoint

import (
	"context"
	"net/url"
	"time"

	"github.com/luisfrmoro/meteocat/model"
)

const stationMetadataPath = "/xema/v1/estacions/metadades"

// StationMetadataFilter holds optional filter values for station metadata requests.
type StationMetadataFilter struct {
	Status *model.StationStatus
	Date   *time.Time
}

// StationMetadataOption configures optional filters for station metadata requests.
type StationMetadataOption func(*StationMetadataFilter)

// WithStationStatus filters station metadata by operational status.
// Note: The METEOCAT API requires that if you use WithStationStatus, you must also use WithStationDate.
func WithStationStatus(status model.StationStatus) StationMetadataOption {
	return func(filter *StationMetadataFilter) {
		filter.Status = &status
	}
}

// WithStationDate filters station metadata by a specific date.
// The API expects the date formatted as yyyy-MM-DDZ.
// Note: The METEOCAT API requires that if you use WithStationDate, you must also use WithStationStatus.
func WithStationDate(date time.Time) StationMetadataOption {
	return func(filter *StationMetadataFilter) {
		filter.Date = &date
	}
}

// Stations fetches the list of XEMA station metadata from the METEOCAT API.
// The endpoint can optionally filter results by operational status and date.
//
// IMPORTANT: The METEOCAT API requires that if one optional filter is provided,
// both Status and Date must be provided together. They are interdependent.
// To request all stations without filtering, provide no filters.
// To filter by status on a specific date, provide both WithStationStatus and WithStationDate.
//
// Parameters:
//   - ctx: context for cancellation and timeouts
//   - do: function to perform the actual HTTP request (typically client.do or a mock)
//   - opts: optional filters for status and date (both required if filtering)
//
// Returns:
//   - model.StationList: list of station metadata
//   - *model.APIError: error if the request fails or data cannot be parsed
func Stations(ctx context.Context, do DoFunc, opts ...StationMetadataOption) (model.StationList, *model.APIError) {
	resource := stationMetadataPath
	filter := StationMetadataFilter{}
	for _, opt := range opts {
		if opt != nil {
			opt(&filter)
		}
	}

	query := url.Values{}
	if filter.Status != nil {
		query.Set("estat", string(*filter.Status))
	}
	if filter.Date != nil {
		query.Set("data", filter.Date.UTC().Format("2006-01-02Z"))
	}
	if encoded := query.Encode(); encoded != "" {
		resource = resource + "?" + encoded
	}

	var list model.StationList
	if err := do(ctx, "GET", resource, &list); err != nil {
		return nil, err
	}
	return list, nil
}
