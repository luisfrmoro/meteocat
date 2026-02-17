# meteocat

Small Go wrapper for the METEOCAT API. Early-stage project: it may contain errors and is likely incomplete.

## What it does

This library provides a type-safe Go client for METEOCAT (Servei Meteorològic de Catalunya), the official meteorological service of Catalonia. It handles the API workflow automatically: call an endpoint and receive decoded data directly.

Currently implements reference data, XEMA observation endpoints, and weather forecast endpoints:
- **Regions (Comarques)**: Administrative divisions of Catalonia
- **Municipalities (Municipis)**: All Catalan municipalities with coordinates
- **Symbols**: Meteorological symbol catalog with icons
- **XEMA stations metadata**: Station catalog with location, network, and status history
- **XEMA observations**: Daily observations from weather stations
- **XEMA variables metadata**: Measurement variable definitions and properties
- **Municipal Hourly Forecasts**: 72-hour hourly weather predictions for any municipality

## Status

- **Experimental and evolving**: Core reference data and XEMA observation endpoints implemented; 72-hour hourly forecasts added
- **Planned features**: Additional forecast types (daily, weekly), alerts, radar imagery, and historical data endpoints
- Well-tested with comprehensive unit and integration tests for implemented features
- Type-safe models for all data structures
- Subject to breaking changes as the API matures

## Features

- **Typed models** for all data structures (regions, municipalities, symbols, stations)
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

    // Get all municipalities with coordinates (if provided by the API)
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
        if muni.Coordinates != nil {
            fmt.Printf("Location: lat=%.4f, lon=%.4f\n", 
                muni.Coordinates.Latitude, muni.Coordinates.Longitude)
        }
        if muni.Region != nil {
            fmt.Printf("Region: %s\n", muni.Region.Name)
        }
    }

    // Get XEMA stations metadata
    stations, apiErr := client.Stations(ctx)
    if apiErr != nil {
        log.Fatal(apiErr)
    }
    fmt.Printf("Stations: %d\n", len(stations))

    // Get XEMA variables metadata
    variables, apiErr := client.Variables(ctx)
    if apiErr != nil {
        log.Fatal(apiErr)
    }
    fmt.Printf("Variables: %d\n", len(variables))

    // Get observations from a station for a specific date
    if len(stations) > 0 {
        stationCode := stations[0].Code
        date := time.Date(2026, 2, 16, 0, 0, 0, 0, time.UTC)
        observations, apiErr := client.Observations(ctx, stationCode, date)
        if apiErr != nil {
            log.Fatal(apiErr)
        }
        fmt.Printf("Observations for station %s: %d\n", stationCode, len(observations))
        if len(observations) > 0 && len(observations[0].Variables) > 0 {
            fmt.Printf("Variables measured: %d\n", len(observations[0].Variables))
        }
    }

    // Get 72-hour hourly forecast for a municipality
    municipalityCode := "080193" // Barcelona
    forecast, apiErr := client.MunicipalHourlyForecast(ctx, municipalityCode)
    if apiErr != nil {
        log.Fatal(apiErr)
    }
    fmt.Printf("Forecast for municipality %s: %d days\n", forecast.MunicipalityCode, len(forecast.Days))
    
    if len(forecast.Days) > 0 {
        firstDay := forecast.Days[0]
        fmt.Printf("Date: %s\n", firstDay.Date)
        
        // Access temperature forecasts
        if tempVar := firstDay.Variables.GetVariable("temp"); tempVar != nil {
            fmt.Printf("Temperature unit: %s\n", tempVar.Unit)
            fmt.Printf("Hourly readings available: %d\n", len(tempVar.Values))
            if len(tempVar.Values) > 0 {
                firstReading := tempVar.Values[0]
                fmt.Printf("First reading: %s°C at %s\n", 
                    firstReading.Value, firstReading.Time.Format("2006-01-02 15:04"))
            }
        }
        
        // Access precipitation forecasts
        if precipVar := firstDay.Variables.GetVariable("precipitacio"); precipVar != nil {
            fmt.Printf("Precipitation: %s readings available\n", precipVar.Unit)
        }
    }
```

## How it works

METEOCAT's API uses a simple single-request workflow: call a method and receive decoded data directly.

## Available endpoints

### Reference Data Endpoints

| Method | Endpoint | Returns |
|--------|----------|---------|
| `Regions(ctx)` | `/referencia/v1/comarques` | Administrative divisions of Catalonia |
| `Municipalities(ctx)` | `/referencia/v1/municipis` | Municipalities with WGS84 coordinates |
| `Symbols(ctx)` | `/referencia/v1/simbols` | Weather symbols with day/night icons |

### XEMA (Automatic Weather Stations Network) Endpoints

| Method | Endpoint | Returns |
|--------|----------|---------|
| `Stations(ctx, ...opts)` | `/xema/v1/estacions/metadades` | Station metadata with location and status (filters: status+date required together) |
| `Observations(ctx, stationCode, date)` | `/xema/v1/estacions/mesurades/{code}/{YYYY}/{MM}/{DD}` | Daily observations for all variables at a specific station |
| `Variables(ctx)` | `/xema/v1/variables/mesurades/metadades` | Metadata for all measurement variables (codes, units, decimals) |

### Weather Forecast Endpoints

| Method | Endpoint | Returns |
|--------|----------|---------|
| `MunicipalHourlyForecast(ctx, municipalityCode)` | `/pronostic/v1/municipalHoraria/{municipalityCode}` | 72-hour hourly forecast with 7 meteorological variables |

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

# Run reference integration tests (requires METEOCAT_API_KEY environment variable)
go test -tags=reference_integration ./...

# Run XEMA integration tests (requires METEOCAT_API_KEY environment variable)
go test -tags=xema_integration ./...

# Run forecast integration tests (requires METEOCAT_API_KEY environment variable)
go test -tags=forecast_integration ./...
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
4. Include integration test in the relevant integration file (e.g. [reference_integration_test.go](reference_integration_test.go) or [xema_integration_test.go](xema_integration_test.go))
5. Update this README with the new endpoint

**Code style**:
- Consistent with Go conventions
- Test coverage for new code
- Clear docstrings for public APIs
- Handle errors explicitly (no `panic()`)

**Planned endpoints to contribute**:
- Weather forecasts (daily, weekly)
- Multivariable observations (specific time ranges and variables)
- Alerts and warnings
- Radar and satellite imagery
- Lightning detection data
- Historical aggregated data

---

## Data origin and reuse

Weather data is provided by METEOCAT (Servei Meteorològic de Catalunya), the official meteorological service of Catalonia. Please cite METEOCAT as the data source and review their terms of service for data reuse policies.

This library is not affiliated with, sponsored by, or endorsed by METEOCAT or the Generalitat de Catalunya.

For complete API documentation, visit: https://apidocs.meteocat.gencat.cat/
