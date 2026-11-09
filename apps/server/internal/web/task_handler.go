package web

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"todo/server/internal/errs"
	"todo/server/internal/service"
)

type TaskHandler struct {
	tasks *service.TaskService
}

func NewTaskHandler(tasks *service.TaskService) *TaskHandler {
	return &TaskHandler{tasks: tasks}
}

func (h *TaskHandler) List(c *gin.Context) {
	tasks, err := h.tasks.List(c.Request.Context(), currentUser(c).ID)
	if err != nil {
		abortWithError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"tasks": tasks})
}

func (h *TaskHandler) Create(c *gin.Context) {
	var req service.CreateTaskInput
	if err := c.ShouldBindJSON(&req); err != nil {
		abortWithError(c, errs.BadRequest("请求格式错误"))
		return
	}
	task, err := h.tasks.Create(c.Request.Context(), currentUser(c).ID, req)
	if err != nil {
		abortWithError(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"task": task})
}

func (h *TaskHandler) Update(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var req service.UpdateTaskPatch
	if err := c.ShouldBindJSON(&req); err != nil {
		abortWithError(c, errs.BadRequest("请求格式错误"))
		return
	}
	task, err := h.tasks.Update(c.Request.Context(), currentUser(c).ID, id, req)
	if err != nil {
		abortWithError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"task": task})
}

func (h *TaskHandler) Delete(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	if err := h.tasks.Delete(c.Request.Context(), currentUser(c).ID, id); err != nil {
		abortWithError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (h *TaskHandler) Reorder(c *gin.Context) {
	var req struct {
		Columns map[string][]uint `json:"columns"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		abortWithError(c, errs.BadRequest("请求格式错误"))
		return
	}
	if err := h.tasks.Reorder(c.Request.Context(), currentUser(c).ID, req.Columns); err != nil {
		abortWithError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}
