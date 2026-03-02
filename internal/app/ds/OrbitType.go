package ds

type OrbitType struct {
	ID          uint    `gorm:"primaryKey"`
	Name        string  `gorm:"type:varchar(100);not null"`
	Description string  `gorm:"type:text;not null"`
	Status      string  `gorm:"type:varchar(20);not null;default:'ACTIVE'"` 
	ImageKey    *string `gorm:"type:varchar(255)"`
	VideoKey	*string `gorm:"type:varchar(255)"`
	
	AltitudeKm  int     `gorm:"not null"` 
}