package f1

// Meeting represents an F1 race weekend from the OpenF1 API.
type Meeting struct {
	MeetingKey       int    `json:"meeting_key"`
	MeetingName      string `json:"meeting_name"`
	OfficialName     string `json:"meeting_official_name"`
	Location         string `json:"location"`
	CountryName      string `json:"country_name"`
	CircuitShortName string `json:"circuit_short_name"`
	DateStart        string `json:"date_start"`
	DateEnd          string `json:"date_end"`
	Year             int    `json:"year"`
}

// Session represents an F1 session (practice, qualifying, race).
type Session struct {
	SessionKey       int    `json:"session_key"`
	SessionType      string `json:"session_type"`
	SessionName      string `json:"session_name"`
	DateStart        string `json:"date_start"`
	DateEnd          string `json:"date_end"`
	CircuitShortName string `json:"circuit_short_name"`
	CountryName      string `json:"country_name"`
	Location         string `json:"location"`
	MeetingKey       int    `json:"meeting_key"`
}

// Position represents a driver's position at a point in time.
type Position struct {
	DriverNumber int    `json:"driver_number"`
	Position     int    `json:"position"`
	Date         string `json:"date"`
}

// RaceControlMessage represents a race control event (flags, etc).
type RaceControlMessage struct {
	Category  string `json:"category"`
	Flag      string `json:"flag"`
	Message   string `json:"message"`
	LapNumber int    `json:"lap_number"`
}

// Stint represents a driver's tire stint in a session.
type Stint struct {
	DriverNumber     int    `json:"driver_number"`
	StintNumber      int    `json:"stint_number"`
	Compound         string `json:"compound"`
	LapStart         int    `json:"lap_start"`
	LapEnd           int    `json:"lap_end"`
	TyreAgeAtFitting int    `json:"tyre_age_at_fitting"`
}

// DriverInfo represents driver metadata for a session.
type DriverInfo struct {
	DriverNumber int    `json:"driver_number"`
	FullName     string `json:"full_name"`
	FirstName    string `json:"first_name"`
	LastName     string `json:"last_name"`
	TeamName     string `json:"team_name"`
	NameAcronym  string `json:"name_acronym"`
}
