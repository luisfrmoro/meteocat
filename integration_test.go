//go:build integration

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

// TestIntegrationRegions verifies that the Regions endpoint returns
// valid reference data for regional divisions (comarques).
func TestIntegrationRegions(t *testing.T) {
	client, ctx, cancel := setupIntegrationClient(t)
	defer cancel()

	regions, apiErr := client.Regions(ctx)
	if apiErr != nil {
		t.Fatalf("regions request: %v", apiErr)
	}

	if len(regions) == 0 {
		t.Fatal("expected at least one region")
	}

	t.Logf("Retrieved %d regions", len(regions))

	// Validate each region has required fields
	for i, region := range regions {
		if region.Code == 0 {
			t.Fatalf("region %d: expected Code to be set", i)
		}
		if strings.TrimSpace(region.Name) == "" {
			t.Fatalf("region %d (Code=%d): expected Name to be set", i, region.Code)
		}
		t.Logf("  Region %d: Code=%d, Name=%q", i, region.Code, region.Name)
	}

	t.Log("✓ Regions endpoint validation completed successfully")
}

// validateMunicipalityFields checks that a municipality has all required fields.
func validateMunicipalityFields(t *testing.T, i int, mun *model.Municipality) {
	t.Helper()

	if strings.TrimSpace(mun.Code) == "" {
		t.Fatalf("municipality %d: expected Code to be set", i)
	}
	if strings.TrimSpace(mun.Name) == "" {
		t.Fatalf("municipality %d (Code=%s): expected Name to be set", i, mun.Code)
	}
}

// TestIntegrationMunicipalities verifies that the Municipalities endpoint returns
// valid reference data including coordinates and region associations.
func TestIntegrationMunicipalities(t *testing.T) {
	client, ctx, cancel := setupIntegrationClient(t)
	defer cancel()

	municipalities, apiErr := client.Municipalities(ctx)
	if apiErr != nil {
		t.Fatalf("municipalities request: %v", apiErr)
	}

	if len(municipalities) == 0 {
		t.Fatal("expected at least one municipality")
	}

	t.Logf("Retrieved %d municipalities", len(municipalities))

	// Validate each municipality has required fields
	hasCoordinates := false
	hasRegionRef := false

	for i, mun := range municipalities {
		validateMunicipalityFields(t, i, &mun)

		// Track data availability
		if mun.Coordinates.Latitude != 0 || mun.Coordinates.Longitude != 0 {
			hasCoordinates = true
		}
		if mun.Region != nil {
			hasRegionRef = true
		}

		if i < 3 {
			t.Logf("  Municipality %d: Code=%s, Name=%q, Lat=%.4f, Lon=%.4f", i, mun.Code, mun.Name, mun.Coordinates.Latitude, mun.Coordinates.Longitude)
		}
	}

	// Verify that coordinate data is present
	if !hasCoordinates {
		t.Error("expected at least one municipality with coordinate data")
	}

	// Verify that region associations are present
	if !hasRegionRef {
		t.Error("expected at least one municipality with region reference")
	}

	t.Log("✓ Municipalities endpoint validation completed successfully")
}

// validateSymbolValue checks that a symbol value has all required fields.
func validateSymbolValue(t *testing.T, categoryIdx, valueIdx int, value *model.SymbolValue) {
	t.Helper()

	if strings.TrimSpace(value.Code) == "" {
		t.Fatalf("symbol[%d].value[%d]: expected Code to be set", categoryIdx, valueIdx)
	}
	if strings.TrimSpace(value.Name) == "" {
		t.Fatalf("symbol[%d].value[%d] (Code=%s): expected Name to be set", categoryIdx, valueIdx, value.Code)
	}
}

// validateSymbolCategory checks that a symbol category is valid.
func validateSymbolCategory(t *testing.T, i int, symbol *model.Symbol) (hasValues bool) {
	t.Helper()

	if strings.TrimSpace(symbol.Name) == "" {
		t.Fatalf("symbol %d: expected Name to be set", i)
	}

	if len(symbol.Values) == 0 {
		t.Logf("  WARNING: Symbol category %d (%q) has no values", i, symbol.Name)
		return false
	}

	// Validate each symbol value
	for j, value := range symbol.Values {
		validateSymbolValue(t, i, j, &value)
	}

	if i < 3 {
		t.Logf("  Symbol category %d: Name=%q, Values=%d", i, symbol.Name, len(symbol.Values))
	}

	return true
}

// TestIntegrationSymbols verifies that the Symbols endpoint returns
// valid reference data for meteorological symbols organized by category.
func TestIntegrationSymbols(t *testing.T) {
	client, ctx, cancel := setupIntegrationClient(t)
	defer cancel()

	symbols, apiErr := client.Symbols(ctx)
	if apiErr != nil {
		t.Fatalf("symbols request: %v", apiErr)
	}

	if len(symbols) == 0 {
		t.Fatal("expected at least one symbol category")
	}

	t.Logf("Retrieved %d symbol categories", len(symbols))

	// Validate each symbol category has required fields
	totalSymbolValues := 0

	for i, symbol := range symbols {
		if validateSymbolCategory(t, i, &symbol) {
			totalSymbolValues += len(symbol.Values)
		}
	}

	if totalSymbolValues == 0 {
		t.Error("expected at least one symbol value across all categories")
	}

	t.Logf("✓ Symbols endpoint validation completed successfully (total values: %d)", totalSymbolValues)
}

// TestIntegrationRegionsAndMunicipalitiesConsistency verifies that municipality
// region references are consistent with the regions reference data.
func TestIntegrationRegionsAndMunicipalitiesConsistency(t *testing.T) {
	client, ctx, cancel := setupIntegrationClient(t)
	defer cancel()

	// Fetch reference data
	regions, apiErr := client.Regions(ctx)
	if apiErr != nil {
		t.Fatalf("regions request: %v", apiErr)
	}

	municipalities, apiErr := client.Municipalities(ctx)
	if apiErr != nil {
		t.Fatalf("municipalities request: %v", apiErr)
	}

	// Create a map of region codes for fast lookup
	regionMap := make(map[int]*model.Region)
	for i := range regions {
		regionMap[regions[i].Code] = &regions[i]
	}

	t.Logf("Checking %d municipalities against %d regions", len(municipalities), len(regions))

	// Validate that all municipality region references exist in regions data
	invalidRegionReferences := 0
	for i, mun := range municipalities {
		if mun.Region == nil {
			// Region reference is optional (omitempty)
			continue
		}

		if _, exists := regionMap[mun.Region.Code]; !exists {
			t.Logf("municipality[%d] (%s): region Code %d not found in regions reference data", i, mun.Name, mun.Region.Code)
			invalidRegionReferences++
		}
	}

	if invalidRegionReferences > 0 {
		t.Fatalf("found %d municipalities with invalid region references", invalidRegionReferences)
	}

	t.Log("✓ Regions and Municipalities consistency validation completed successfully")
}

// TestIntegrationDataStability verifies that multiple calls to the endpoints
// return consistent data (no changes between calls within a short timespan).
func TestIntegrationDataStability(t *testing.T) {
	client, ctx, cancel := setupIntegrationClient(t)
	defer cancel()

	// Fetch regions twice
	regions1, apiErr := client.Regions(ctx)
	if apiErr != nil {
		t.Fatalf("first regions request: %v", apiErr)
	}

	regions2, apiErr := client.Regions(ctx)
	if apiErr != nil {
		t.Fatalf("second regions request: %v", apiErr)
	}

	// Verify count matches
	if len(regions1) != len(regions2) {
		t.Fatalf("regions count changed between calls: %d vs %d", len(regions1), len(regions2))
	}

	// Verify first few entries match
	for i := 0; i < len(regions1) && i < 5; i++ {
		if regions1[i].Code != regions2[i].Code || regions1[i].Name != regions2[i].Name {
			t.Fatalf("region[%d] changed between calls: %+v vs %+v", i, regions1[i], regions2[i])
		}
	}

	// Fetch municipalities twice
	municipalities1, apiErr := client.Municipalities(ctx)
	if apiErr != nil {
		t.Fatalf("first municipalities request: %v", apiErr)
	}

	municipalities2, apiErr := client.Municipalities(ctx)
	if apiErr != nil {
		t.Fatalf("second municipalities request: %v", apiErr)
	}

	// Verify count matches
	if len(municipalities1) != len(municipalities2) {
		t.Fatalf("municipalities count changed between calls: %d vs %d", len(municipalities1), len(municipalities2))
	}

	// Fetch symbols twice
	symbols1, apiErr := client.Symbols(ctx)
	if apiErr != nil {
		t.Fatalf("first symbols request: %v", apiErr)
	}

	symbols2, apiErr := client.Symbols(ctx)
	if apiErr != nil {
		t.Fatalf("second symbols request: %v", apiErr)
	}

	// Verify count matches
	if len(symbols1) != len(symbols2) {
		t.Fatalf("symbols count changed between calls: %d vs %d", len(symbols1), len(symbols2))
	}

	t.Logf("✓ Data stability validation completed successfully")
	t.Logf("  Regions: %d (stable across 2 calls)", len(regions1))
	t.Logf("  Municipalities: %d (stable across 2 calls)", len(municipalities1))
	t.Logf("  Symbol categories: %d (stable across 2 calls)", len(symbols1))
}
