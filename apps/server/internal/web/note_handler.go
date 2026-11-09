package web

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"todo/server/internal/errs"
	"todo/server/internal/service"
)

type NoteHandler struct {
	notes *service.NoteService
}

func NewNoteHandler(notes *service.NoteService) *NoteHandler {
	return &NoteHandler{notes: notes}
}

func (h *NoteHandler) List(c *gin.Context) {
	notes, err := h.notes.List(c.Request.Context(), currentUser(c).ID)
	if err != nil {
		abortWithError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"notes": notes})
}

func (h *NoteHandler) Create(c *gin.Context) {
	var req service.CreateNoteInput
	if err := c.ShouldBindJSON(&req); err != nil {
		abortWithError(c, errs.BadRequest("请求格式错误"))
		return
	}
	note, err := h.notes.Create(c.Request.Context(), currentUser(c).ID, req)
	if err != nil {
		abortWithError(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"note": note})
}

func (h *NoteHandler) Update(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var req service.UpdateNotePatch
	if err := c.ShouldBindJSON(&req); err != nil {
		abortWithError(c, errs.BadRequest("请求格式错误"))
		return
	}
	note, err := h.notes.Update(c.Request.Context(), currentUser(c).ID, id, req)
	if err != nil {
		abortWithError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"note": note})
}

func (h *NoteHandler) Delete(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	if err := h.notes.Delete(c.Request.Context(), currentUser(c).ID, id); err != nil {
		abortWithError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}
