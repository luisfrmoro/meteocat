package meteocat

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"net/http"
	"reflect"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/luisfrmoro/meteocat/endpoint"
	"github.com/luisfrmoro/meteocat/model"
)

const (
	baseURL           = "https://api.meteo.cat"
	userAgent         = "meteocat-go/0.1.0"
	contentTypeHeader = "Content-Type"
)

// Client manages requests to the METEOCAT HTTP API.
// It handles the METEOCAT API workflow: performs a single request to fetch data from the API endpoint,
// which returns the data directly in the response body.
//
// The APIKey must be loaded from secure storage and must not be logged or serialized.
// Do not modify any fields after construction; changing them concurrently may introduce race conditions.
type Client struct {
	baseURL         string
	httpClient      *http.Client
	userAgent       string
	maxResponseBody int64
	apiKey          string `json:"-"`
}

// String implements fmt.Stringer but intentionally omits the API key.
func (c *Client) String() string {
	return fmt.Sprintf("meteocat.Client{BaseURL:%s, APIKeySet:%t}", c.baseURL, c.apiKey != "")
}

// NewClient constructs a new *Client using the provided API key.
// If httpClient is nil, a sensible default with a 10s timeout is used.
// The apiKey must be a valid METEOCAT API key; it will be used in the Authorization header for all requests.
func NewClient(apiKey string, httpClient *http.Client) (*Client, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("api key is required")
	}

	if httpClient == nil {
		httpClient = &http.Client{Timeout: 10 * time.Second}
	}

	return &Client{
		baseURL:         baseURL,
		httpClient:      httpClient,
		userAgent:       userAgent,
		maxResponseBody: 10 << 20, // 10 MB
		apiKey:          apiKey,
	}, nil
}

// isJSONContent returns true if the content type indicates JSON or a JSON-based media type.
func isJSONContent(contentType string) bool {
	ct := strings.ToLower(contentType)
	return strings.Contains(ct, "application/json") || strings.Contains(ct, "+json")
}

// validateHTTPOut checks that 'out' is a non-nil pointer to a valid unmarshal target.
func validateHTTPOut(out any) *model.APIError {
	if out == nil {
		return &model.APIError{Message: "no output provided"}
	}
	outValue := reflect.ValueOf(out)
	if outValue.Kind() != reflect.Ptr || outValue.IsNil() {
		return &model.APIError{Message: "out must be a non-nil pointer"}
	}
	return nil
}

// prepareRequest creates a new HTTP request with the given context, method, and URL,
// applying standard headers (Accept, User-Agent, Authorization).
func (c *Client) prepareRequest(ctx context.Context, method, url string) (*http.Request, *model.APIError) {
	req, err := http.NewRequestWithContext(ctx, method, url, nil)
	if err != nil {
		return nil, &model.APIError{Message: fmt.Sprintf("create request: %v", err)}
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", c.userAgent)
	req.Header.Set("x-api-key", c.apiKey)

	return req, nil
}

// readResponseBody reads the response body with a size limit to prevent OOM attacks.
func (c *Client) readResponseBody(resp *http.Response) ([]byte, *model.APIError) {
	limitedReader := io.LimitReader(resp.Body, c.maxResponseBody+1)
	respBytes, err := io.ReadAll(limitedReader)
	if err != nil {
		return nil, &model.APIError{Code: resp.StatusCode, Message: fmt.Sprintf("read response: %v", err)}
	}

	if int64(len(respBytes)) > c.maxResponseBody {
		return nil, &model.APIError{Code: resp.StatusCode, Message: fmt.Sprintf("response body too large: limit %d bytes", c.maxResponseBody)}
	}

	return respBytes, nil
}

// handleSuccessResponse unmarshals a successful JSON response into 'out'.
func (c *Client) handleSuccessResponse(resp *http.Response, respBytes []byte, out any) *model.APIError {
	contentType := resp.Header.Get(contentTypeHeader)
	isJSON := isJSONContent(contentType)

	if len(respBytes) == 0 {
		if resp.StatusCode == http.StatusNoContent {
			return nil
		}
		if !isJSON {
			return &model.APIError{Code: resp.StatusCode, Message: fmt.Sprintf("empty response body for content-type %q", contentType)}
		}
		return &model.APIError{Code: resp.StatusCode, Message: "empty response body"}
	}

	if !isJSON {
		return &model.APIError{Code: resp.StatusCode, Message: fmt.Sprintf("unexpected content-type %q", contentType)}
	}

	if err := json.Unmarshal(respBytes, out); err != nil {
		return &model.APIError{Code: resp.StatusCode, Message: fmt.Sprintf("unmarshal response: %v", err)}
	}

	return nil
}

// handleErrorResponse parses an API error response, attempting to extract structured error information.
func (c *Client) handleErrorResponse(resp *http.Response, respBytes []byte) *model.APIError {
	var apiErr model.APIError
	if len(respBytes) == 0 {
		return &model.APIError{Code: resp.StatusCode, Message: http.StatusText(resp.StatusCode)}
	}

	if err := json.Unmarshal(respBytes, &apiErr); err != nil || (apiErr.Message == "" && apiErr.Code == 0) {
		return &model.APIError{Code: resp.StatusCode, Message: string(respBytes)}
	}

	if apiErr.Code == 0 {
		apiErr.Code = resp.StatusCode
	}

	return &apiErr
}

func (c *Client) readAndNormalizeJSON(resp *http.Response) ([]byte, *model.APIError) {
	respBytes, apiErr := c.readResponseBody(resp)
	if apiErr != nil {
		return nil, apiErr
	}

	return normalizeJSONBytes(resp.Header.Get(contentTypeHeader), respBytes)
}

// normalizeJSONBytes ensures JSON payloads are decoded as UTF-8 before unmarshalling.
// It converts from common legacy encodings (ISO-8859-1, Windows-1252) when detected
// via Content-Type or when the payload contains invalid UTF-8.
func normalizeJSONBytes(contentType string, respBytes []byte) ([]byte, *model.APIError) {
	if len(respBytes) == 0 {
		return respBytes, nil
	}

	charset := ""
	if contentType != "" {
		_, params, err := mime.ParseMediaType(contentType)
		if err == nil {
			charset = strings.ToLower(strings.TrimSpace(params["charset"]))
		}
	}

	switch charset {
	case "", "utf-8", "utf8":
		if utf8.Valid(respBytes) {
			return respBytes, nil
		}
		return latin1ToUTF8(respBytes), nil
	case "iso-8859-1", "iso8859-1", "latin1":
		return latin1ToUTF8(respBytes), nil
	case "iso-8859-15", "iso8859-15", "latin9":
		return latin9ToUTF8(respBytes), nil
	case "windows-1252", "cp1252":
		return windows1252ToUTF8(respBytes), nil
	default:
		if utf8.Valid(respBytes) {
			return respBytes, nil
		}
		return nil, &model.APIError{Message: fmt.Sprintf("unsupported charset %q in content-type %q", charset, contentType)}
	}
}

func latin1ToUTF8(input []byte) []byte {
	output := make([]byte, 0, len(input)*2)
	for _, b := range input {
		if b < 0x80 {
			output = append(output, b)
			continue
		}
		output = append(output, 0xC0|(b>>6), 0x80|(b&0x3F))
	}
	return output
}

func latin9ToUTF8(input []byte) []byte {
	// ISO-8859-15 differs from ISO-8859-1 at a few positions (e.g., Euro sign).
	latin9 := map[byte]rune{
		0xA4: 0x20AC, // Euro sign
		0xA6: 0x0160,
		0xA8: 0x0161,
		0xB4: 0x017D,
		0xB8: 0x017E,
		0xBC: 0x0152,
		0xBD: 0x0153,
		0xBE: 0x0178,
	}

	output := make([]byte, 0, len(input)*2)
	for _, b := range input {
		if r, ok := latin9[b]; ok {
			output = append(output, string(r)...)
			continue
		}
		if b < 0x80 {
			output = append(output, b)
			continue
		}
		output = append(output, 0xC0|(b>>6), 0x80|(b&0x3F))
	}

	return output
}

func windows1252ToUTF8(input []byte) []byte {
	// Map 0x80-0x9F to their Unicode equivalents.
	win1252 := [32]rune{
		0x20AC, 0xFFFD, 0x201A, 0x0192,
		0x201E, 0x2026, 0x2020, 0x2021,
		0x02C6, 0x2030, 0x0160, 0x2039,
		0x0152, 0xFFFD, 0x017D, 0xFFFD,
		0xFFFD, 0x2018, 0x2019, 0x201C,
		0x201D, 0x2022, 0x2013, 0x2014,
		0x02DC, 0x2122, 0x0161, 0x203A,
		0x0153, 0xFFFD, 0x017E, 0x0178,
	}

	output := make([]byte, 0, len(input)*2)
	for _, b := range input {
		switch {
		case b < 0x80:
			output = append(output, b)
		case b >= 0xA0:
			output = append(output, 0xC0|(b>>6), 0x80|(b&0x3F))
		default:
			r := win1252[b-0x80]
			output = append(output, string(r)...)
		}
	}

	return output
}

// do performs an HTTP request to fetch data from the METEOCAT API.
// It makes a single request to the API endpoint and unmarshals the response data directly into `out`.
//
// Parameters:
//   - ctx: context for cancellation and timeouts
//   - method: HTTP method (typically "GET")
//   - resource: API endpoint relative to baseURL (e.g., "/api/forecasts/50441")
//   - out: non-nil pointer where the data will be unmarshaled
//
// Returns *APIError on any failure (HTTP errors, parsing errors, network errors, etc.)
func (c *Client) do(ctx context.Context, method, resource string, out any) *model.APIError {
	if err := validateHTTPOut(out); err != nil {
		return err
	}

	// Request to METEOCAT API endpoint
	url := c.baseURL + "/" + strings.TrimLeft(resource, "/")
	req, apiErr := c.prepareRequest(ctx, method, url)
	if apiErr != nil {
		return apiErr
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return &model.APIError{Message: fmt.Sprintf("request to METEOCAT API: %v", err)}
	}
	defer func() {
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
	}()

	respBytes, apiErr := c.readAndNormalizeJSON(resp)
	if apiErr != nil {
		return apiErr
	}

	// Handle response status
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return c.handleErrorResponse(resp, respBytes)
	}

	// Unmarshal response directly into out
	if apiErr := c.handleSuccessResponse(resp, respBytes, out); apiErr != nil {
		return apiErr
	}

	return nil
}

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
func (c *Client) Regions(ctx context.Context) (model.RegionList, *model.APIError) {
	return endpoint.Regions(ctx, c.do)
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
func (c *Client) Municipalities(ctx context.Context) (model.MunicipalityList, *model.APIError) {
	return endpoint.Municipalities(ctx, c.do)
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
func (c *Client) Symbols(ctx context.Context) (model.SymbolList, *model.APIError) {
	return endpoint.Symbols(ctx, c.do)
}

// StationMetadataOption configures optional filters for station metadata requests.
type StationMetadataOption = endpoint.StationMetadataOption

// WithStationStatus filters station metadata by operational status.
func WithStationStatus(status model.StationStatus) StationMetadataOption {
	return endpoint.WithStationStatus(status)
}

// WithStationDate filters station metadata by a specific date.
// The API expects the date formatted as yyyy-MM-DDZ.
func WithStationDate(date time.Time) StationMetadataOption {
	return endpoint.WithStationDate(date)
}

// Stations fetches the list of XEMA station metadata from the METEOCAT API.
// The endpoint can optionally filter results by operational status and date.
//
// IMPORTANT: The METEOCAT API requires that if one optional filter is provided,
// both Status and Date must be provided together. They are interdependent.
// To request all stations without filtering, call Stations with no options.
// To filter by status on a specific date, you must provide both
// WithStationStatus and WithStationDate options.
//
// Parameters:
//   - ctx: context for cancellation and timeouts
//   - opts: optional filters for status and date (must provide both if filtering)
//
// Returns:
//   - model.StationList: list of station metadata
//   - *model.APIError: error if the request fails or data cannot be parsed
//
// Example without filters (all stations):
//
//	client, _ := meteocat.NewClient("your-api-key", nil)
//	stations, err := client.Stations(context.Background())
//	if err != nil {
//		log.Fatal(err)
//	}
//	for _, s := range stations {
//		fmt.Printf("%s: %s\n", s.Code, s.Name)
//	}
//
// Example with filters (status and date required together):
//
//	client, _ := meteocat.NewClient("your-api-key", nil)
//	date := time.Date(2026, 2, 17, 0, 0, 0, 0, time.UTC)
//	stations, err := client.Stations(
//		context.Background(),
//		meteocat.WithStationStatus(model.StationStatusOperational),
//		meteocat.WithStationDate(date),
//	)
//	if err != nil {
//		log.Fatal(err)
//	}
//	for _, s := range stations {
//		fmt.Printf("%s: %s\n", s.Code, s.Name)
//	}
func (c *Client) Stations(ctx context.Context, opts ...StationMetadataOption) (model.StationList, *model.APIError) {
	return endpoint.Stations(ctx, c.do, opts...)
}

// Variable type alias for metadata of a single XEMA variable.
type Variable = model.Variable

// VariableList type alias for a collection of variable metadata.
type VariableList = model.VariableList

// Reading type alias for a single observation measurement at a specific point in time.
type Reading = model.Reading

// VariableObservation type alias grouping all readings for a single variable.
type VariableObservation = model.VariableObservation

// StationObservation type alias for all observations recorded at a station for a specific day.
type StationObservation = model.StationObservation

// StationObservationList type alias for a collection of observations.
type StationObservationList = model.StationObservationList

// Observations fetches all observations of all variables recorded at a station for a specific day.
// The endpoint returns observation measurements grouped by variable, with each variable containing
// a list of readings taken throughout the day.
//
// Parameters:
//   - ctx: context for cancellation and timeouts
//   - stationCode: the unique identifier of the station (e.g., "CC")
//   - date: the specific date for which observations are requested
//
// Returns:
//   - StationObservationList: list of observations with all variables and readings
//   - *APIError: error if the request fails or data cannot be parsed
//
// Example:
//
//	client, _ := meteocat.NewClient("your-api-key", nil)
//	date := time.Date(2020, time.June, 16, 0, 0, 0, 0, time.UTC)
//	obs, err := client.Observations(context.Background(), "CC", date)
//	if err != nil {
//		log.Fatal(err)
//	}
//	for _, stationObs := range obs {
//		fmt.Printf("Station %s:\n", stationObs.Code)
//		for _, varObs := range stationObs.Variables {
//			fmt.Printf("  Variable %d: %d readings\n", varObs.Code, len(varObs.Readings))
//		}
//	}
func (c *Client) Observations(ctx context.Context, stationCode string, date time.Time) (StationObservationList, *model.APIError) {
	return endpoint.Observations(ctx, c.do, stationCode, date)
}

// Variables fetches the metadata of all XEMA variables.
// The endpoint returns information about all variables independently from the stations where they are measured.
// This reference data is essential for understanding variable codes, units, decimal precision, and other properties
// used across all XEMA observation endpoints.
//
// Parameters:
//   - ctx: context for cancellation and timeouts
//
// Returns:
//   - VariableList: list of variable metadata
//   - *APIError: error if the request fails or data cannot be parsed
//
// Example:
//
//	client, _ := meteocat.NewClient("your-api-key", nil)
//	vars, err := client.Variables(context.Background())
//	if err != nil {
//		log.Fatal(err)
//	}
//	for _, v := range vars {
//		fmt.Printf("%d: %s (%s) - %d decimals\n", v.Code, v.Name, v.Unit, v.Decimals)
//	}
func (c *Client) Variables(ctx context.Context) (VariableList, *model.APIError) {
	return endpoint.Variables(ctx, c.do)
}
