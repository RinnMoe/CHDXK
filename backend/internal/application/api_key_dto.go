package application

import (
	"time"

	"jcourse/internal/domain/auth"
)

type ApiKeyDTO struct {
	ID         int        `json:"id"`
	Name       string     `json:"name"`
	Key        string     `json:"key,omitempty"`
	KeyMasked  string     `json:"key_masked"`
	Role       string     `json:"role"`
	UserID     int        `json:"user_id"`
	CreatedAt  time.Time  `json:"created_at"`
	LastUsedAt *time.Time `json:"last_used_at"`
}

func newApiKeyDTO(k auth.ApiKey, includeKey bool) ApiKeyDTO {
	dto := ApiKeyDTO{
		ID:         k.ID,
		Name:       k.Name,
		KeyMasked:  auth.MaskApiKey(k.Key),
		Role:       k.Role,
		UserID:     k.UserID,
		CreatedAt:  k.CreatedAt,
		LastUsedAt: k.LastUsedAt,
	}
	if includeKey {
		dto.Key = k.Key
	}
	return dto
}

type CreateApiKeyCommand struct {
	Name string `json:"name"`
}
