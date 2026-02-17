package endpoint

import (
	"context"

	"github.com/luisfrmoro/meteocat/model"
)

// DoFunc abstracts the HTTP client call, allowing endpoint functions to be testable
// and decoupled from the specific HTTP client implementation.
// It mirrors the signature of Client.do() for dependency injection during testing.
type DoFunc func(ctx context.Context, method, resource string, out any) *model.APIError

// Regions fetches the list of all regional administrative divisions from the METEOCAT API.
// This endpoint returns metadata about the geographic divisions (regions) of the service area,
// including their unique codes and names. Regions are used as administrative groupings
// and are referenced by other METEOCAT services such as municipality data.
//
// The region codes returned by this endpoint are useful for filtering or organizing
// other geographic data by administrative zone.
//
// Parameters:
//   - ctx: context for cancellation and timeouts
//   - do: function to perform the actual HTTP request (typically client.do or a mock)
//
// Returns:
//   - model.RegionList: slice of regions with their unique identifiers and names
//   - *model.APIError: error if the request fails or data cannot be parsed
//
// Example:
//
//	client, _ := meteocat.NewClient("your-api-key", nil)
//	regions, err := client.Regions(context.Background())
//	if err != nil {
//		log.Fatal(err)
//	}
//	for _, r := range regions {
//		fmt.Printf("%d: %s\n", r.Code, r.Name)
//	}
func Regions(ctx context.Context, do DoFunc) (model.RegionList, *model.APIError) {
	var list model.RegionList
	if err := do(ctx, "GET", "/referencia/v1/comarques", &list); err != nil {
		return nil, err
	}
	return list, nil
}

// Municipalities fetches the list of all municipalities from the METEOCAT API.
// This endpoint returns complete municipality data including geographic coordinates,
// administrative information, and region references. Municipalities are the finest
// geographic units available in the METEOCAT API.
//
// The municipality codes returned by this endpoint may be required for other METEOCAT
// services such as detailed weather forecasts or local observation data.
//
// Each municipality includes:
// - Unique identifier (typically a 6-digit code)
// - Official name
// - Geographic coordinates (latitude and longitude in decimal degrees)
// - Reference to the region (administrative division) to which it belongs
//
// Parameters:
//   - ctx: context for cancellation and timeouts
//   - do: function to perform the actual HTTP request (typically client.do or a mock)
//
// Returns:
//   - model.MunicipalityList: slice of municipalities with their complete information
//   - *model.APIError: error if the request fails or data cannot be parsed
//
// Example:
//
//	client, _ := meteocat.NewClient("your-api-key", nil)
//	municipalities, err := client.Municipalities(context.Background())
//	if err != nil {
//		log.Fatal(err)
//	}
//	for _, m := range municipalities {
//		if m.Region != nil {
//			fmt.Printf("%s (region: %s)\n", m.Name, m.Region.Name)
//		}
//		fmt.Printf("  Coordinates: %.4f°N, %.4f°E\n", m.Coordinates.Latitude, m.Coordinates.Longitude)
//	}
func Municipalities(ctx context.Context, do DoFunc) (model.MunicipalityList, *model.APIError) {
	var list model.MunicipalityList
	if err := do(ctx, "GET", "/referencia/v1/municipis", &list); err != nil {
		return nil, err
	}
	return list, nil
}

// Symbols fetches the complete catalog of meteorological symbols from the METEOCAT API.
// This endpoint returns reference information for all symbol categories and their possible values
// used across METEOCAT's forecast and observational data endpoints.
//
// The symbol catalog includes:
// - Multiple symbol categories (e.g., sky state, precipitation types, snow accumulation)
// - Detailed descriptions and categorizations
// - Visual representations (icons) for both day and night conditions
// - Unique codes for each symbol value within its category
//
// This reference data is essential for correctly interpreting meteorological symbols
// that appear in forecast and observation responses. Each symbol value includes:
// - Unique code within its category
// - Human-readable name
// - Categorical classification
// - HTTP URLs to SVG/PNG icon representations (separate day and night icons)
//
// Parameters:
//   - ctx: context for cancellation and timeouts
//   - do: function to perform the actual HTTP request (typically client.do or a mock)
//
// Returns:
//   - model.SymbolList: slice of meteorological symbol categories with all values
//   - *model.APIError: error if the request fails or data cannot be parsed
//
// Example:
//
//	client, _ := meteocat.NewClient("your-api-key", nil)
//	symbols, err := client.Symbols(context.Background())
//	if err != nil {
//		log.Fatal(err)
//	}
//	for _, category := range symbols {
//		fmt.Printf("Category: %s\n", category.Name)
//		for _, value := range category.Values {
//			fmt.Printf("  - %s (code: %s)\n", value.Name, value.Code)
//			if value.IconURL != "" {
//				fmt.Printf("    Day icon: %s\n", value.IconURL)
//			}
//		}
//	}
func Symbols(ctx context.Context, do DoFunc) (model.SymbolList, *model.APIError) {
	var list model.SymbolList
	if err := do(ctx, "GET", "/referencia/v1/simbols", &list); err != nil {
		return nil, err
	}
	return list, nil
}
