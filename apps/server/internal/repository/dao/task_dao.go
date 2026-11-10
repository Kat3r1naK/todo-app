package dao

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"gorm.io/gorm"

	"todo/server/internal/domain"
	"todo/server/internal/errs"
)

type TaskDAO struct {
	db *gorm.DB
}

func NewTaskDAO(db *gorm.DB) *TaskDAO {
	return &TaskDAO{db: db}
}

func (d *TaskDAO) ListByUser(ctx context.Context, userID uint) ([]domain.Task, error) {
	var tasks []domain.Task
	err := d.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("position ASC, id ASC").
		Find(&tasks).Error
	return tasks, err
}

func (d *TaskDAO) FindByID(ctx context.Context, id, userID uint) (*domain.Task, error) {
	var task domain.Task
	if err := d.db.WithContext(ctx).Where("id = ? AND user_id = ?", id, userID).First(&task).Error; err != nil {
		return nil, wrapTaskError(err)
	}
	return &task, nil
}

func (d *TaskDAO) Create(ctx context.Context, task *domain.Task) error {
	return d.db.WithContext(ctx).Create(task).Error
}

func (d *TaskDAO) Update(ctx context.Context, id, userID uint, updates map[string]any) error {
	return d.db.WithContext(ctx).
		Model(&domain.Task{}).
		Where("id = ? AND user_id = ?", id, userID).
		Updates(updates).Error
}

func (d *TaskDAO) Delete(ctx context.Context, id, userID uint) error {
	res := d.db.WithContext(ctx).Where("id = ? AND user_id = ?", id, userID).Delete(&domain.Task{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return errs.NotFound("任务不存在")
	}
	return nil
}

func (d *TaskDAO) DeleteByUser(ctx context.Context, userID uint) error {
	return d.db.WithContext(ctx).Where("user_id = ?", userID).Delete(&domain.Task{}).Error
}

func (d *TaskDAO) MaxPosition(ctx context.Context, userID uint, status string) (int, bool, error) {
	var value sql.NullInt64
	err := d.db.WithContext(ctx).
		Model(&domain.Task{}).
		Where("user_id = ? AND status = ?", userID, status).
		Select("MAX(position)").
		Scan(&value).Error
	if err != nil {
		return 0, false, err
	}
	return int(value.Int64), value.Valid, nil
}

func (d *TaskDAO) MinPosition(ctx context.Context, userID uint, status string) (int, bool, error) {
	var value sql.NullInt64
	err := d.db.WithContext(ctx).
		Model(&domain.Task{}).
		Where("user_id = ? AND status = ?", userID, status).
		Select("MIN(position)").
		Scan(&value).Error
	if err != nil {
		return 0, false, err
	}
	return int(value.Int64), value.Valid, nil
}

func (d *TaskDAO) Reorder(ctx context.Context, userID uint, columns map[string][]uint) error {
	tx := d.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return tx.Error
	}
	for status, ids := range columns {
		for position, id := range ids {
			updates := map[string]any{"status": status, "position": position}
			if status == "done" {
				updates["completed_at"] = gorm.Expr("COALESCE(completed_at, ?)", time.Now())
			} else {
				updates["completed_at"] = nil
			}
			if err := tx.Model(&domain.Task{}).
				Where("id = ? AND user_id = ?", id, userID).
				Updates(updates).Error; err != nil {
				tx.Rollback()
				return err
			}
		}
	}
	return tx.Commit().Error
}

func (d *TaskDAO) CountByUser(ctx context.Context) (map[uint]int64, error) {
	type countRow struct {
		UserID uint
		Total  int64
	}
	var rows []countRow
	if err := d.db.WithContext(ctx).
		Model(&domain.Task{}).
		Select("user_id, COUNT(*) AS total").
		Group("user_id").
		Scan(&rows).Error; err != nil {
		return nil, err
	}
	counts := make(map[uint]int64, len(rows))
	for _, row := range rows {
		counts[row.UserID] = row.Total
	}
	return counts, nil
}

func wrapTaskError(err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return errs.NotFound("任务不存在")
	}
	return err
}
