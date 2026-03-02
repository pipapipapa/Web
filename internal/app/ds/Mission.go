package ds

import "time"

type Mission struct {
	ID        uint      `gorm:"primaryKey"`
	Status    string    `gorm:"type:varchar(20);not null;default:'DRAFT'"` // DRAFT, DELETED, FORMED, COMPLETED, REJECTED
	CreatedAt time.Time `gorm:"not null"`
	AuthorID  uint      `gorm:"not null"`

	FormedAt    *time.Time
	CompletedAt *time.Time
	ModeratorID *uint

	SatelliteMassKg int 	`gorm:"default:0"`

	Author     User               `gorm:"foreignKey:AuthorID;constraint:OnDelete:RESTRICT"`
	Moderator  *User              `gorm:"foreignKey:ModeratorID;constraint:OnDelete:RESTRICT"`
	OrbitItems []MissionOrbitItem `gorm:"foreignKey:MissionID;constraint:OnDelete:RESTRICT"`
}