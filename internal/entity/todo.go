package entity

import "time"

type Todo struct {
	ID        int64     `json:"id" gorm:"primaryKey"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	DueDate   time.Time `json:"due_date"`
	Completed bool      `json:"completed"`
	UserID    int64     `json:"user_id"`
}
