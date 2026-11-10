package domain

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

type Subtask struct {
	ID    string `json:"id"`
	Title string `json:"title"`
	Done  bool   `json:"done"`
}

type Subtasks []Subtask

func (s Subtasks) Value() (driver.Value, error) {
	if s == nil {
		return "[]", nil
	}
	buf, err := json.Marshal(s)
	if err != nil {
		return nil, err
	}
	return string(buf), nil
}

func (s *Subtasks) Scan(value any) error {
	switch v := value.(type) {
	case nil:
		*s = Subtasks{}
		return nil
	case []byte:
		return json.Unmarshal(v, s)
	case string:
		return json.Unmarshal([]byte(v), s)
	default:
		return fmt.Errorf("不支持的子任务类型: %T", value)
	}
}

type Task struct {
	ID          uint       `gorm:"primaryKey" json:"id"`
	UserID      uint       `gorm:"index;not null" json:"-"`
	Title       string     `gorm:"size:200;not null" json:"title"`
	Description string     `gorm:"type:text" json:"description"`
	Notes       string     `gorm:"type:mediumtext" json:"notes"`
	Subtasks    Subtasks   `gorm:"type:json" json:"subtasks"`
	Status      string     `gorm:"size:16;index;not null;default:todo" json:"status"`
	Priority    string     `gorm:"size:16;not null;default:medium" json:"priority"`
	Pinned      bool       `gorm:"default:false" json:"pinned"`
	Position    int        `gorm:"default:0" json:"position"`
	CompletedAt *time.Time `gorm:"index" json:"completed_at"`
	ArchivedAt  *time.Time `gorm:"index" json:"archived_at"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

var taskStatuses = map[string]bool{"todo": true, "doing": true, "done": true}

var taskPriorities = map[string]bool{"low": true, "medium": true, "high": true}

func ValidStatus(status string) bool { return taskStatuses[status] }

func ValidPriority(priority string) bool { return taskPriorities[priority] }
