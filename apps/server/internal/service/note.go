package service

import (
	"context"
	"strings"
	"unicode/utf8"

	"todo/server/internal/domain"
	"todo/server/internal/errs"
	"todo/server/internal/repository"
)

type CreateNoteInput struct {
	Content string `json:"content"`
	Color   string `json:"color"`
}

type UpdateNotePatch struct {
	Content *string `json:"content"`
	Color   *string `json:"color"`
}

type NoteService struct {
	notes repository.NoteRepository
}

func NewNoteService(notes repository.NoteRepository) *NoteService {
	return &NoteService{notes: notes}
}

func (s *NoteService) List(ctx context.Context, userID uint) ([]domain.Note, error) {
	return s.notes.ListByUser(ctx, userID)
}

func (s *NoteService) Create(ctx context.Context, userID uint, input CreateNoteInput) (*domain.Note, error) {
	content := strings.TrimSpace(input.Content)
	if content == "" {
		return nil, errs.BadRequest("便签内容不能为空")
	}
	if utf8.RuneCountInString(content) > 2000 {
		return nil, errs.BadRequest("便签内容过长")
	}
	color := input.Color
	if color == "" {
		color = "yellow"
	}
	if !domain.ValidNoteColor(color) {
		return nil, errs.BadRequest("无效的便签颜色")
	}
	note := &domain.Note{UserID: userID, Content: content, Color: color}
	if err := s.notes.Create(ctx, note); err != nil {
		return nil, err
	}
	return note, nil
}

func (s *NoteService) Update(ctx context.Context, userID, noteID uint, patch UpdateNotePatch) (*domain.Note, error) {
	if _, err := s.notes.FindByID(ctx, noteID, userID); err != nil {
		return nil, err
	}
	updates := map[string]any{}
	if patch.Content != nil {
		content := strings.TrimSpace(*patch.Content)
		if content == "" {
			return nil, errs.BadRequest("便签内容不能为空")
		}
		if utf8.RuneCountInString(content) > 2000 {
			return nil, errs.BadRequest("便签内容过长")
		}
		updates["content"] = content
	}
	if patch.Color != nil {
		if !domain.ValidNoteColor(*patch.Color) {
			return nil, errs.BadRequest("无效的便签颜色")
		}
		updates["color"] = *patch.Color
	}
	if len(updates) > 0 {
		if err := s.notes.Update(ctx, noteID, userID, updates); err != nil {
			return nil, err
		}
	}
	return s.notes.FindByID(ctx, noteID, userID)
}

func (s *NoteService) Delete(ctx context.Context, userID, noteID uint) error {
	return s.notes.Delete(ctx, noteID, userID)
}
