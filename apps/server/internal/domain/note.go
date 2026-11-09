package domain

import "time"

type Note struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"index;not null" json:"-"`
	Content   string    `gorm:"type:text;not null" json:"content"`
	Color     string    `gorm:"size:16;default:yellow" json:"color"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

var noteColors = map[string]bool{"yellow": true, "green": true, "blue": true, "pink": true, "purple": true}

func ValidNoteColor(color string) bool { return noteColors[color] }
