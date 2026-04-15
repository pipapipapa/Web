package ds

type MissionOrbitItem struct {
	MissionID 	uint 		`gorm:"primaryKey;autoIncrement:false" json:"-"`
	OrbitID   	uint 		`gorm:"primaryKey;autoIncrement:false" json:"orbit_id"`

	Payload 	*float64 	`json:"payload"`

	Velocity 	*float64 	`json:"velocity"`
	Period   	*float64 	`json:"period"`

	Orbit		OrbitType 	`gorm:"foreignKey:OrbitID;constraint:OnDelete:RESTRICT" json:"orbit"`
}