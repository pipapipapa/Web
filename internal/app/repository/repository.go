package repository

type OrbitType struct {
	ID          int
	Name        string
	AltitudeKm  int
	Description string 
	ImageKey    string 
	VideoKey    string 
	Type        string 
	Weight      string 
	Velocity    float64 
	Period      float64 
}

type MissionProject struct {
	ID              int
	SatelliteName   string
	SatelliteMassKg int
	CalculatedPeriod float64
	CalculatedVelocity float64
	CalculatedAltitude int
	
	SelectedOrbits  []MissionOrbitItem 
}

type MissionOrbitItem struct {
	Orbit    OrbitType
	Priority int
	Comment  string 
}

const MinioBaseURL = "http://127.0.0.1:9000/orbits/"