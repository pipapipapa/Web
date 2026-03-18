package ds

type OrbitType struct {
	ID          uint    `gorm:"primaryKey" json:"id"`
	Name        string  `gorm:"type:varchar(100);not null" json:"name"`
	Description string  `gorm:"type:text;not null" json:"description"`
	Status      string  `gorm:"type:varchar(20);not null;default:'ACTIVE'" json:"-"` 
	ImageKey    *string `gorm:"type:varchar(255)" json:"image_key"`
	VideoKey	*string `gorm:"type:varchar(255)" json:"video_key"`
	
	AltitudeKm  int     `gorm:"not null" json:"altitude_km"` 
}