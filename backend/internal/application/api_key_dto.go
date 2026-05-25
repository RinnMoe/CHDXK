package application

import (
	"strconv"
	"time"

	"jcourse/internal/domain/auth"
)

type ApiKeyDTO struct {
	ID         string     `json:"id"`
	Name       string     `json:"name"`
	Key        string     `json:"key"`
	Role       string     `json:"role"`
	UserID     int        `json:"user_id"`
	CreatedAt  time.Time  `json:"created_at"`
	LastUsedAt *time.Time `json:"last_used_at"`
}

func newApiKeyDTO(k auth.ApiKey, key string) ApiKeyDTO {
	return ApiKeyDTO{
		ID:         strconv.FormatInt(k.ID, 10),
		Name:       k.Name,
		Key:        key,
		Role:       k.Role,
		UserID:     k.UserID,
		CreatedAt:  k.CreatedAt,
		LastUsedAt: k.LastUsedAt,
	}
}

type CreateApiKeyCommand struct {
	Name string `json:"name"`
}
