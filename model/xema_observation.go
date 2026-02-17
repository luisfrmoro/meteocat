package model

// Variable represents the metadata of a single XEMA variable.
// Variables are the fundamental units used to record observations from stations,
// such as atmospheric pressure, temperature, humidity, wind speed, etc.
// Not all stations measure the same variables; the set of available sensors
// may be different at each station (e.g., snow depth is not measured where it doesn't snow).
type Variable struct {
	// Code is the unique numeric identifier of the variable
	Code int `json:"codi"`

	// Name is the official descriptive name of the variable in Catalan
	Name string `json:"nom"`

	// Unit is the measurement unit used to express variable values (e.g., "hPa", "°C")
	Unit string `json:"unitat"`

	// Acronym is a short abbreviation for the variable name (e.g., "Px" for pressure, "TMAR" for sea temperature)
	Acronym string `json:"acronim"`

	// Type is the variable classification provided by the API (e.g., "DAT" for data)
	Type string `json:"tipus"`

	// Decimals indicates the number of decimal places used to express values of this variable
	Decimals int `json:"decimals"`
}

// VariableList represents a collection of variable metadata returned by the METEOCAT API.
type VariableList []Variable

// Reading represents a single observation measurement at a specific point in time.
// Each reading includes a measured value, timestamp, validation status, and time base.
type Reading struct {
	// Data is the timestamp of the observation in UTC indicating the measurement period
	Data MeteocatTime `json:"data"`

	// DataExtrem is an optional timestamp indicating the exact moment when an extreme value
	// (maximum or minimum) was recorded during the measurement period.
	// This field is only present for extreme value measurements.
	DataExtrem *MeteocatTime `json:"dataExtrem,omitempty"`

	// Value is the numeric measurement result
	Value float64 `json:"valor"`

	// Status indicates the result of the quality control validation process.
	// Possible values:
	//   - "" (blank): validation process has not started
	//   - "T": validation process started but result pending
	//   - "V": data is considered valid
	//   - "N": data is considered invalid
	Status string `json:"estat"`

	// TimeBase describes the measurement interval (base temporal).
	// Possible values:
	//   - "HO": Hourly (Horària)
	//   - "SH": Semi-hourly (Semi-horària)
	//   - "DM": 10-minute interval (10 minutal)
	//   - "MI": Minutal (Minutal)
	TimeBase string `json:"baseHoraria"`
}

// VariableObservation groups all readings for a single variable measured at a station.
type VariableObservation struct {
	// Code is the unique numeric identifier of the variable
	Code int `json:"codi"`

	// Readings is the list of observation values for this variable
	Readings []Reading `json:"lectures"`
}

// StationObservation represents all observations of all variables recorded at a station for a specific day.
type StationObservation struct {
	// Code is the unique identifier of the station (e.g., "CC")
	Code string `json:"codi"`

	// Variables is the list of all variables measured at the station on the specified day
	Variables []VariableObservation `json:"variables"`
}

// StationObservationList represents a collection of observations returned by the METEOCAT API.
type StationObservationList []StationObservation
