# meteocat

Small Go wrapper for the METEOCAT API. Early-stage project: it may contain errors and is likely incomplete.

## What it does

This library provides a type-safe Go client for METEOCAT (Servei Meteorològic de Catalunya), the official meteorological service of Catalonia. It handles the API workflow automatically: call an endpoint and receive decoded data directly.

Currently implements reference data endpoints:
- **Regions (Comarques)**: Administrative divisions of Catalonia
- **Municipalities (Municipis)**: All Catalan municipalities with coordinates
- **Symbols**: Meteorological symbol catalog with icons

## Status

- **Experimental and evolving**: Limited subset of METEOCAT endpoints implemented (reference data only)
- **Planned features**: Weather forecasts, observations, and real-time data endpoints
- Well-tested with comprehensive unit and integration tests for implemented features
- Type-safe models for all data structures
- Subject to breaking changes as the API matures

## Features

- **Typed models** for all data structures (regions, municipalities, symbols)
- **Safe HTTP handling**: Size limits (10 MB), charset normalization, JSON validation
- **Clear error handling**: Structured APIError type with HTTP and METEOCAT error codes
- **Secure**: API key never serialized to JSON or logs
- **Flexible context**: All methods accept `context.Context` for cancellation and timeouts
- **Custom HTTP client**: Use your own `*http.Client` or default with 10s timeout
- **Charset support**: Handles Catalan characters with ISO-8859-1, Windows-1252, and UTF-8 normalization

## Installation

```bash
go get github.com/luisfrmoro/meteocat
```

Requires Go 1.18 or later. The library has zero external dependencies - only Go standard library.

## Quick start

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/luisfrmoro/meteocat"
)

func main() {
    client, err := meteocat.NewClient("YOUR_API_KEY", nil)
    if err != nil {
        log.Fatal(err)
    }
    ctx := context.Background()

    // Get all Catalan regions (comarques)
    regions, apiErr := client.Regions(ctx)
    if apiErr != nil {
        log.Fatal(apiErr)
    }
    fmt.Printf("Regions: %d\n", len(regions))

    // Get all municipalities with coordinates
    municipalities, apiErr := client.Municipalities(ctx)
    if apiErr != nil {
        log.Fatal(apiErr)
    }
    fmt.Printf("Municipalities: %d\n", len(municipalities))

    // Get meteorological symbols catalog
    symbols, apiErr := client.Symbols(ctx)
    if apiErr != nil {
        log.Fatal(apiErr)
    }
    fmt.Printf("Symbol categories: %d\n", len(symbols))

    // Access specific data
    if len(municipalities) > 0 {
        muni := municipalities[0]
        fmt.Printf("Municipality: %s (Code: %s)\n", muni.Name, muni.Code)
        fmt.Printf("Location: lat=%.4f, lon=%.4f\n", 
            muni.Coordinates.Latitude, muni.Coordinates.Longitude)
        if muni.Region != nil {
            fmt.Printf("Region: %s\n", muni.Region.Name)
        }
    }
}
```

## How it works

METEOCAT's API uses a simple single-request workflow: call a method and receive decoded data directly.

## Available endpoints

| Method | Endpoint | Returns |
|--------|----------|---------|
| `Regions(ctx)` | `/referencia/v1/comarques` | Administrative divisions of Catalonia |
| `Municipalities(ctx)` | `/referencia/v1/municipis` | Municipalities with WGS84 coordinates |
| `Symbols(ctx)` | `/referencia/v1/simbols` | Weather symbols with day/night icons |

---

## Getting started

### API key

You need a METEOCAT API key. Keep it out of logs and source control. The client passes it via the `x-api-key` header. **The API key is never serialized to JSON** - safe for logging.

Get your API key at: https://apidocs.meteocat.gencat.cat/section/informacio-general/plans-i-registre/

### Client initialization

```go
// Use default HTTP client (10s timeout)
client, err := meteocat.NewClient("YOUR_API_KEY", nil)

// Or provide your own HTTP client for full control
customClient := &http.Client{
    Timeout: 30 * time.Second,
}
client, err := meteocat.NewClient("YOUR_API_KEY", customClient)
```

---

## Security & Reliability

- **API key protection**: Never serialized or logged
- **Network safety**: 10 MB response limit, 10s timeout, charset normalization for Catalan characters
- **Error handling**: Structured `APIError` type with clear messages

---

## Testing

```bash
# Run all tests
go test ./...

# Run integration tests (requires METEOCAT_API_KEY environment variable)
go test -tags=integration ./...
```

Comprehensive unit and integration tests included. Integration tests require a valid API key set in `METEOCAT_API_KEY`.

---

## Technical details

- **Charset normalization**: Automatic handling of ISO-8859-1, Windows-1252, and UTF-8 for Catalan characters
- **Geographic coordinates**: WGS84 system (compatible with GPS and web mapping)
- **Geographic scope**: Catalonia-specific data and future endpoints

---

## Contributing

Issues and PRs are welcome.

**If you add new endpoints**:
1. Add handler in `endpoint/` with tests
2. Add types in `model/` with tests
3. Add public method to `Client` type in [client.go](client.go)
4. Include integration test in [integration_test.go](integration_test.go)
5. Update this README with the new endpoint

**Code style**:
- Consistent with Go conventions
- Test coverage for new code
- Clear docstrings for public APIs
- Handle errors explicitly (no `panic()`)

**Planned endpoints to contribute**:
- Weather forecasts (hourly, daily, weekly)
- Real-time observations from weather stations
- Alerts and warnings
- Radar and satellite imagery
- Historical data

---

## Data origin and reuse

Weather data is provided by METEOCAT (Servei Meteorològic de Catalunya), the official meteorological service of Catalonia. Please cite METEOCAT as the data source and review their terms of service for data reuse policies.

This library is not affiliated with, sponsored by, or endorsed by METEOCAT or the Generalitat de Catalunya.

For complete API documentation, visit: https://apidocs.meteocat.gencat.cat/
