package model

// StationStatus defines the operational status filter values supported by the API.
// During its lifetime, a station can have different operational states.
type StationStatus string

const (
	// StationStatusOperational indicates the station is currently functioning and recording data.
	StationStatusOperational StationStatus = "ope"

	// StationStatusDismantled indicates the station is no longer installed and not recording data.
	StationStatusDismantled StationStatus = "des"

	// StationStatusTemporaryShutdown indicates the station is temporarily stopped (e.g., for repairs).
	StationStatusTemporaryShutdown StationStatus = "bte"
)

// Station represents the metadata of an XEMA station as returned by the METEOCAT API.
// It includes identification details, location, administrative references, network,
// and operational status history.
type Station struct {
	// Code is the unique identifier for the station (e.g., "CC")
	Code string `json:"codi"`

	// Name is the official name of the station
	Name string `json:"nom"`

	// Type is the station type identifier provided by the API
	Type string `json:"tipus"`

	// Coordinates is the geographic position of the station
	Coordinates Coordinates `json:"coordenades"`

	// Location describes the physical placement of the station
	Location string `json:"emplacament"`

	// Altitude is the station elevation in meters
	Altitude float64 `json:"altitud"`

	// Municipality is the municipality where the station is located
	Municipality Municipality `json:"municipi"`

	// County is the county (comarca) where the station is located
	County Region `json:"comarca"`

	// Province is the province where the station is located
	Province StationProvince `json:"provincia"`

	// Network is the network to which the station belongs
	Network StationNetwork `json:"xarxa"`

	// States lists the station operational states over time
	States []StationState `json:"estats"`
}

// StationList represents a collection of XEMA stations returned by the METEOCAT API.
type StationList []Station

// StationProvince represents the province reference associated with a station.
type StationProvince struct {
	// Code is the numeric identifier of the province
	Code int `json:"codi"`

	// Name is the official name of the province
	Name string `json:"nom"`
}

// StationNetwork represents the network reference associated with a station.
type StationNetwork struct {
	// Code is the numeric identifier of the network
	Code int `json:"codi"`

	// Name is the official name of the network
	Name string `json:"nom"`
}

// StationState represents an operational state of a station with its time window.
type StationState struct {
	// Code is the numeric identifier for the station state
	Code int `json:"codi"`

	// StartDate is the start date/time of the state in RFC3339 format
	StartDate MeteocatTime `json:"dataInici"`

	// EndDate is the end date/time of the state in RFC3339 format, or nil if ongoing
	EndDate *MeteocatTime `json:"dataFi"`
}
