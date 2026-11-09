package dao

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"todo/server/internal/domain"
	"todo/server/internal/errs"
)

type NoteDAO struct {
	db *gorm.DB
}

func NewNoteDAO(db *gorm.DB) *NoteDAO {
	return &NoteDAO{db: db}
}

func (d *NoteDAO) ListByUser(ctx context.Context, userID uint) ([]domain.Note, error) {
	var notes []domain.Note
	err := d.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC, id DESC").
		Find(&notes).Error
	return notes, err
}

func (d *NoteDAO) FindByID(ctx context.Context, id, userID uint) (*domain.Note, error) {
	var note domain.Note
	if err := d.db.WithContext(ctx).Where("id = ? AND user_id = ?", id, userID).First(&note).Error; err != nil {
		return nil, wrapNoteError(err)
	}
	return &note, nil
}

func (d *NoteDAO) Create(ctx context.Context, note *domain.Note) error {
	return d.db.WithContext(ctx).Create(note).Error
}

func (d *NoteDAO) Update(ctx context.Context, id, userID uint, updates map[string]any) error {
	return d.db.WithContext(ctx).
		Model(&domain.Note{}).
		Where("id = ? AND user_id = ?", id, userID).
		Updates(updates).Error
}

func (d *NoteDAO) Delete(ctx context.Context, id, userID uint) error {
	res := d.db.WithContext(ctx).Where("id = ? AND user_id = ?", id, userID).Delete(&domain.Note{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return errs.NotFound("便签不存在")
	}
	return nil
}

func (d *NoteDAO) DeleteByUser(ctx context.Context, userID uint) error {
	return d.db.WithContext(ctx).Where("user_id = ?", userID).Delete(&domain.Note{}).Error
}

func wrapNoteError(err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return errs.NotFound("便签不存在")
	}
	return err
}
