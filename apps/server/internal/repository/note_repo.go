package repository

import (
	"todo/server/internal/repository/dao"
)

type noteRepository struct {
	*dao.NoteDAO
}

func NewNoteRepository(noteDAO *dao.NoteDAO) NoteRepository {
	return &noteRepository{NoteDAO: noteDAO}
}
