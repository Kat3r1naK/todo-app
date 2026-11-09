package service

import (
	"context"
	"strings"
	"unicode/utf8"

	"todo/server/internal/domain"
	"todo/server/internal/errs"
	"todo/server/internal/repository"
)

type CreateTaskInput struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Status      string `json:"status"`
	Priority    string `json:"priority"`
}

type UpdateTaskPatch struct {
	Title       *string `json:"title"`
	Description *string `json:"description"`
	Status      *string `json:"status"`
	Priority    *string `json:"priority"`
	Pinned      *bool   `json:"pinned"`
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
	task := &domain.Task{
		UserID:      userID,
		Title:       title,
		Description: strings.TrimSpace(input.Description),
		Status:      status,
		Priority:    priority,
		Position:    position,
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
	status := task.Status
	if patch.Status != nil {
		if !domain.ValidStatus(*patch.Status) {
			return nil, errs.BadRequest("无效的状态")
		}
		status = *patch.Status
		updates["status"] = status
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
