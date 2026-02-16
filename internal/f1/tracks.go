package f1

type location struct {
	Lat, Lon float64
}

var circuitLocations = map[string]location{
	"Bahrain":     {26.03, 50.51},
	"Jeddah":      {21.63, 39.10},
	"Melbourne":   {-37.85, 144.97},
	"Suzuka":      {34.84, 136.54},
	"Shanghai":    {31.34, 121.22},
	"Miami":       {25.96, -80.24},
	"Imola":       {44.34, 11.71},
	"Monaco":      {43.73, 7.42},
	"Barcelona":   {41.57, 2.26},
	"Montreal":    {45.50, -73.52},
	"Spielberg":   {47.22, 14.76},
	"Silverstone": {52.07, -1.02},
	"Spa":         {50.44, 5.97},
	"Zandvoort":   {52.39, 4.54},
	"Monza":       {45.62, 9.29},
	"Baku":        {40.37, 49.85},
	"Singapore":   {1.29, 103.86},
	"Austin":      {30.13, -97.64},
	"Mexico City": {19.40, -99.09},
	"São Paulo":   {-23.70, -46.70},
	"Las Vegas":   {36.11, -115.17},
	"Lusail":      {25.49, 51.45},
	"Abu Dhabi":   {24.47, 54.60},
}

// CircuitCoords returns latitude and longitude for an F1 circuit
// identified by its Location field from the meetings API.
func CircuitCoords(loc string) (lat, lon float64, ok bool) {
	l, ok := circuitLocations[loc]
	if !ok {
		return 0, 0, false
	}
	return l.Lat, l.Lon, true
}
