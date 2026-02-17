package model

// Region represents a regional administrative division with its unique identifier and name.
// This data structure is used by the METEOCAT API to provide regional reference information.
// Regions are used as geographic divisions and are referenced by other endpoints
// such as municipalities to indicate their administrative zone.
type Region struct {
	// Code is the unique numeric identifier for the region
	Code int `json:"codi"`

	// Name is the official name of the region
	Name string `json:"nom"`
}

// RegionList represents a collection of regions returned by the METEOCAT API
type RegionList []Region

// Coordinates represents geographic coordinates in decimal format.
// Latitude and Longitude are expressed in decimal degrees as per the
// World Geodetic System (WGS84).
type Coordinates struct {
	// Latitude is the latitude coordinate in decimal degrees (range: -90 to 90)
	Latitude float64 `json:"latitud"`

	// Longitude is the longitude coordinate in decimal degrees (range: -180 to 180)
	Longitude float64 `json:"longitud"`
}

// Municipality represents a municipality with its geographic and administrative information.
// This data structure is used by the METEOCAT API to provide specific municipality information
// including coordinates and the region to which it belongs.
type Municipality struct {
	// Code is the unique identifier for the municipality (typically a 6-digit string)
	Code string `json:"codi"`

	// Name is the official name of the municipality
	Name string `json:"nom"`

	// Coordinates contains the geographic location of the municipality center
	Coordinates Coordinates `json:"coordenades"`

	// Region contains reference information about the region to which this municipality belongs
	Region *Region `json:"comarca,omitempty"`
}

// MunicipalityList represents a collection of municipalities returned by the METEOCAT API
type MunicipalityList []Municipality

// SymbolValue represents an individual meteorological symbol value within a symbol category.
// Each value corresponds to a specific meteorological condition or state within its category,
// with associated icons for day and night representation.
type SymbolValue struct {
	// Code is the unique identifier for this symbol value within its category
	Code string `json:"codi"`

	// Name is the human-readable name of the symbol value (e.g., "Cel serè" for clear sky)
	Name string `json:"nom"`

	// Description provides additional context about the symbol value
	Description string `json:"descripcio"`

	// Category indicates the meteorological category this symbol belongs to (e.g., "nuvolositat" for cloudiness)
	Category string `json:"categoria"`

	// IconURL is the HTTP URL to the day-time icon representation of this symbol (SVG or PNG format)
	IconURL string `json:"icona"`

	// IconURLNight is the HTTP URL to the night-time icon representation of this symbol.
	// May be empty if no distinct night icon is available.
	IconURLNight string `json:"icona_nit"`
}

// Symbol represents a meteorological symbol category with its possible values and descriptions.
// This data structure groups related weather condition symbols together, such as sky state,
// precipitation types, or snow accumulation, each with their specific codes and representations.
//
// The Symbol structure is used by the METEOCAT API to provide a reference catalog of all
// possible meteorological symbols that may appear in forecast and observation data endpoints.
type Symbol struct {
	// Name is the identifier name of the symbol category (e.g., "cel" for sky state)
	Name string `json:"nom"`

	// Description provides a human-readable description of what this symbol category represents
	Description string `json:"descripcio"`

	// Values is a slice of all possible symbol values within this category
	Values []SymbolValue `json:"valors"`
}

// SymbolList represents a collection of meteorological symbol categories returned by the METEOCAT API
type SymbolList []Symbol
