package web

import (
	"errors"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"todo/server/internal/errs"
)

func abortWithError(c *gin.Context, err error) {
	status := http.StatusInternalServerError
	message := "服务器内部错误"
	switch {
	case errors.Is(err, errs.ErrInvalidInput):
		status, message = http.StatusBadRequest, err.Error()
	case errors.Is(err, errs.ErrInvalidCredentials), errors.Is(err, errs.ErrUnauthorized):
		status, message = http.StatusUnauthorized, err.Error()
	case errors.Is(err, errs.ErrForbidden):
		status, message = http.StatusForbidden, err.Error()
	case errors.Is(err, errs.ErrNotFound):
		status, message = http.StatusNotFound, err.Error()
	case errors.Is(err, errs.ErrConflict):
		status, message = http.StatusConflict, err.Error()
	default:
		log.Printf("internal error: %v", err)
	}
	c.AbortWithStatusJSON(status, gin.H{"error": message})
}
