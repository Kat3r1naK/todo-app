package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"todo/server/internal/domain"
	"todo/server/internal/errs"
	"todo/server/internal/repository"
)

type CreateTaskInput struct {
	Title       string           `json:"title"`
	Description string           `json:"description"`
	Notes       string           `json:"notes"`
	Subtasks    []domain.Subtask `json:"subtasks"`
	Status      string           `json:"status"`
	Priority    string           `json:"priority"`
}

type UpdateTaskPatch struct {
	Title       *string           `json:"title"`
	Description *string           `json:"description"`
	Notes       *string           `json:"notes"`
	Subtasks    *[]domain.Subtask `json:"subtasks"`
	Status      *string           `json:"status"`
	Priority    *string           `json:"priority"`
	Pinned      *bool             `json:"pinned"`
	Archived    *bool             `json:"archived"`
}

const (
	maxNotesRunes   = 20000
	maxSubtasks     = 100
	maxSubtaskRunes = 200
)

func normalizeSubtasks(items []domain.Subtask) ([]domain.Subtask, error) {
	if len(items) > maxSubtasks {
		return nil, errs.BadRequest("子任务过多")
	}
	normalized := make([]domain.Subtask, 0, len(items))
	seen := make(map[string]bool, len(items))
	for _, item := range items {
		title := strings.TrimSpace(item.Title)
		if title == "" {
			continue
		}
		if utf8.RuneCountInString(title) > maxSubtaskRunes {
			return nil, errs.BadRequest("子任务标题过长")
		}
		id := strings.TrimSpace(item.ID)
		if id == "" || seen[id] {
			id = newSubtaskID()
		}
		seen[id] = true
		normalized = append(normalized, domain.Subtask{ID: id, Title: title, Done: item.Done})
	}
	return normalized, nil
}

func newSubtaskID() string {
	buf := make([]byte, 8)
	if _, err := rand.Read(buf); err != nil {
		return strconv.FormatInt(time.Now().UnixNano(), 36)
	}
	return hex.EncodeToString(buf)
}

type TaskService struct {
	tasks repository.TaskRepository
}

func NewTaskService(tasks repository.TaskRepository) *TaskService {
	return &TaskService{tasks: tasks}
}

func (s *TaskService) List(ctx context.Context, userID uint) ([]domain.Task, error) {
	return s.tasks.ListByUser(ctx, userID)
}

func (s *TaskService) Create(ctx context.Context, userID uint, input CreateTaskInput) (*domain.Task, error) {
	title := strings.TrimSpace(input.Title)
	if title == "" {
		return nil, errs.BadRequest("任务标题不能为空")
	}
	if utf8.RuneCountInString(title) > 200 {
		return nil, errs.BadRequest("标题过长")
	}
	status := input.Status
	if status == "" {
		status = "todo"
	}
	if !domain.ValidStatus(status) {
		return nil, errs.BadRequest("无效的状态")
	}
	priority := input.Priority
	if priority == "" {
		priority = "medium"
	}
	if !domain.ValidPriority(priority) {
		return nil, errs.BadRequest("无效的优先级")
	}
	max, found, err := s.tasks.MaxPosition(ctx, userID, status)
	if err != nil {
		return nil, err
	}
	position := 0
	if found {
		position = max + 1
	}
	notes := strings.TrimSpace(input.Notes)
	if utf8.RuneCountInString(notes) > maxNotesRunes {
		return nil, errs.BadRequest("备注过长")
	}
	subtasks, err := normalizeSubtasks(input.Subtasks)
	if err != nil {
		return nil, err
	}
	task := &domain.Task{
		UserID:      userID,
		Title:       title,
		Description: strings.TrimSpace(input.Description),
		Notes:       notes,
		Subtasks:    domain.Subtasks(subtasks),
		Status:      status,
		Priority:    priority,
		Position:    position,
	}
	if status == "done" {
		now := time.Now()
		task.CompletedAt = &now
	}
	if err := s.tasks.Create(ctx, task); err != nil {
		return nil, err
	}
	return task, nil
}

func (s *TaskService) Update(ctx context.Context, userID, taskID uint, patch UpdateTaskPatch) (*domain.Task, error) {
	task, err := s.tasks.FindByID(ctx, taskID, userID)
	if err != nil {
		return nil, err
	}

	updates := map[string]any{}
	if patch.Title != nil {
		title := strings.TrimSpace(*patch.Title)
		if title == "" {
			return nil, errs.BadRequest("任务标题不能为空")
		}
		if utf8.RuneCountInString(title) > 200 {
			return nil, errs.BadRequest("标题过长")
		}
		updates["title"] = title
	}
	if patch.Description != nil {
		updates["description"] = strings.TrimSpace(*patch.Description)
	}
	if patch.Notes != nil {
		notes := strings.TrimSpace(*patch.Notes)
		if utf8.RuneCountInString(notes) > maxNotesRunes {
			return nil, errs.BadRequest("备注过长")
		}
		updates["notes"] = notes
	}
	if patch.Subtasks != nil {
		subtasks, err := normalizeSubtasks(*patch.Subtasks)
		if err != nil {
			return nil, err
		}
		updates["subtasks"] = domain.Subtasks(subtasks)
	}
	status := task.Status
	if patch.Status != nil {
		if !domain.ValidStatus(*patch.Status) {
			return nil, errs.BadRequest("无效的状态")
		}
		status = *patch.Status
		updates["status"] = status
		if status == "done" && task.Status != "done" {
			updates["completed_at"] = time.Now()
		} else if status != "done" && task.Status == "done" {
			updates["completed_at"] = nil
		}
	}
	if patch.Priority != nil {
		if !domain.ValidPriority(*patch.Priority) {
			return nil, errs.BadRequest("无效的优先级")
		}
		updates["priority"] = *patch.Priority
	}
	if patch.Pinned != nil {
		updates["pinned"] = *patch.Pinned
		if *patch.Pinned && !task.Pinned {
			min, found, err := s.tasks.MinPosition(ctx, userID, status)
			if err != nil {
				return nil, err
			}
			position := 0
			if found {
				position = min - 1
			}
			updates["position"] = position
		}
	}
	if patch.Archived != nil {
		if *patch.Archived {
			if status != "done" {
				return nil, errs.BadRequest("只有已完成的任务可以归档")
			}
			if task.ArchivedAt == nil {
				updates["archived_at"] = time.Now()
			}
		} else {
			updates["archived_at"] = nil
		}
	}

	if len(updates) > 0 {
		if err := s.tasks.Update(ctx, taskID, userID, updates); err != nil {
			return nil, err
		}
	}
	return s.tasks.FindByID(ctx, taskID, userID)
}

func (s *TaskService) Delete(ctx context.Context, userID, taskID uint) error {
	return s.tasks.Delete(ctx, taskID, userID)
}

func (s *TaskService) Reorder(ctx context.Context, userID uint, columns map[string][]uint) error {
	if len(columns) == 0 {
		return errs.BadRequest("排序数据不能为空")
	}
	for status := range columns {
		if !domain.ValidStatus(status) {
			return errs.BadRequest("无效的状态")
		}
	}
	return s.tasks.Reorder(ctx, userID, columns)
}
