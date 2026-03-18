package ds

type MissionOrbitItem struct {
	MissionID uint `gorm:"primaryKey;autoIncrement:false"`
	OrbitID   uint `gorm:"primaryKey;autoIncrement:false"`

	Payload float64

	Velocity float64 // Скорость км/с
	Period   float64 // Период мин

	Orbit OrbitType `gorm:"foreignKey:OrbitID;constraint:OnDelete:RESTRICT"`
}