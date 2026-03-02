package ds

type User struct {
	ID       uint   `gorm:"primaryKey"`
	Login    string `gorm:"type:varchar(50);unique;not null"`
	Password string `gorm:"type:varchar(255);not null"`
	Role     string `gorm:"type:varchar(20);not null"` // 'CLIENT', 'MODERATOR'
	FullName string `gorm:"type:varchar(100)"`
}