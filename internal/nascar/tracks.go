package nascar

// TrackLocation holds geographic coordinates for a NASCAR track.
type TrackLocation struct {
	Lat, Lon float64
}

// trackLocations maps NASCAR track_id to geographic coordinates.
var trackLocations = map[int]TrackLocation{
	4:   {34.30, -80.11},  // Darlington Raceway
	14:  {36.52, -82.26},  // Bristol Motor Speedway
	22:  {36.63, -79.85},  // Martinsville Speedway
	26:  {37.59, -77.42},  // Richmond Raceway
	40:  {25.45, -80.41},  // Homestead-Miami Speedway
	41:  {39.12, -94.83},  // Kansas Speedway
	42:  {36.27, -115.01}, // Las Vegas Motor Speedway
	43:  {33.04, -97.28},  // Texas Motor Speedway
	45:  {38.63, -90.15},  // World Wide Technology Raceway
	52:  {36.09, -86.39},  // Nashville Superspeedway
	75:  {19.40, -99.09},  // Autódromo Hermanos Rodríguez
	82:  {33.57, -86.06},  // Talladega Superspeedway
	84:  {33.37, -112.31}, // Phoenix Raceway
	99:  {38.16, -122.45}, // Sonoma Raceway
	103: {39.19, -75.53},  // Dover Motor Speedway
	105: {29.19, -81.07},  // Daytona International Speedway
	111: {33.39, -84.32},  // Atlanta Motor Speedway
	123: {39.79, -86.23},  // Indianapolis Motor Speedway
	133: {42.07, -84.24},  // Michigan International Speedway
	138: {43.36, -71.46},  // New Hampshire Motor Speedway
	157: {42.34, -76.93},  // Watkins Glen International
	159: {36.10, -80.25},  // Bowman Gray Stadium
	162: {35.35, -80.68},  // Charlotte Motor Speedway
	177: {36.22, -81.10},  // North Wilkesboro Speedway
	198: {41.05, -75.51},  // Pocono Raceway
	206: {41.68, -92.22},  // Iowa Speedway
	210: {35.35, -80.68},  // Charlotte Motor Speedway Road Course
	214: {30.13, -97.64},  // Circuit of The Americas
	218: {41.86, -87.62},  // Chicago Street Race
}

// TrackCoords returns the latitude and longitude for a given track ID.
func TrackCoords(trackID int) (lat, lon float64, ok bool) {
	loc, ok := trackLocations[trackID]
	if !ok {
		return 0, 0, false
	}
	return loc.Lat, loc.Lon, true
}
