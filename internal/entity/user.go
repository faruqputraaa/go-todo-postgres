package entity

type User struct {
	ID       int64  `json:"id" gorm:"primaryKey"`
	Username string `json:"username"`
	Password string `json:"-"`
	Role     string `json:"role"`
	FullName string `json:"full_name"`
}