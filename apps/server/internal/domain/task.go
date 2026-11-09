package domain

import "time"

type Task struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	UserID      uint      `gorm:"index;not null" json:"-"`
	Title       string    `gorm:"size:200;not null" json:"title"`
	Description string    `gorm:"type:text" json:"description"`
	Status      string    `gorm:"size:16;index;not null;default:todo" json:"status"`
	Priority    string    `gorm:"size:16;not null;default:medium" json:"priority"`
	Pinned      bool      `gorm:"default:false" json:"pinned"`
	Position    int       `gorm:"default:0" json:"position"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

var taskStatuses = map[string]bool{"todo": true, "doing": true, "done": true}

var taskPriorities = map[string]bool{"low": true, "medium": true, "high": true}

func ValidStatus(status string) bool { return taskStatuses[status] }

func ValidPriority(priority string) bool { return taskPriorities[priority] }
