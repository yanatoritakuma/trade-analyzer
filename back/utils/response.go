package utils

import (
	"errors"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/yanatoritakuma/trade-analyzer/back/domain"
)

// ErrorResponse は統一エラーレスポンス形式。
type ErrorResponse struct {
	Error string `json:"error"`
}

// HandleError はドメインエラーを適切なHTTPステータスへマッピングして返す。
func HandleError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, domain.ErrNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	case errors.Is(err, domain.ErrUnauthorized):
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
	case errors.Is(err, domain.ErrForbidden):
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
	case errors.Is(err, domain.ErrAccountDisabled):
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
	case errors.Is(err, domain.ErrInvalidInput),
		errors.Is(err, domain.ErrAlreadyExists),
		errors.Is(err, domain.ErrInvalidCode),
		errors.Is(err, domain.ErrExpiredCode),
		errors.Is(err, domain.ErrUsedCode):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	default:
		log.Printf("unexpected error: %+v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "サーバーエラーが発生しました"})
	}
}
