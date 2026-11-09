package repository

import (
	"todo/server/internal/repository/dao"
)

type taskRepository struct {
	*dao.TaskDAO
}

func NewTaskRepository(taskDAO *dao.TaskDAO) TaskRepository {
	return &taskRepository{TaskDAO: taskDAO}
}
