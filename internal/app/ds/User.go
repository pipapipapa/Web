package ds

type User struct {
	ID       uint   `gorm:"primaryKey" json:"id"`
	Login    string `gorm:"type:varchar(50);unique;not null" json:"login"`
	Password string `gorm:"type:varchar(255);not null" json:"-"`
	Role     string `gorm:"type:varchar(20);not null" json:"role"` // 'CLIENT', 'MODERATOR'
	FullName string `gorm:"type:varchar(100)" json:"full_name"`
}