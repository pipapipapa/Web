package ds

import "time"

type Mission struct {
	ID        	uint      `gorm:"primaryKey" json:"id"`
	Status    	string    `gorm:"type:varchar(20);not null;default:'DRAFT'" json:"status"` // DRAFT, DELETED, FORMED, COMPLETED, REJECTED
	CreatedAt 	time.Time `gorm:"not null" json:"created_at"`
	AuthorID  	uint      `gorm:"not null" json:"author_id"`

	FormedAt    *time.Time 	`json:"formed_at"`
	CompletedAt *time.Time 	`json:"completed_at"`
	ModeratorID *uint		`json:"moderator_id"`

	SatelliteMassKg *int 	`gorm:"default:0" json:"satellite_mass_kg"`

	Author     	User               `gorm:"foreignKey:AuthorID;constraint:OnDelete:RESTRICT" json:"author"`
	Moderator  	*User              `gorm:"foreignKey:ModeratorID;constraint:OnDelete:RESTRICT" json:"moderator"`
	OrbitItems 	[]MissionOrbitItem `gorm:"foreignKey:MissionID;constraint:OnDelete:RESTRICT" json:"orbit_items"`
}