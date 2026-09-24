package models

import "time"

// User is created on first OAuth login and identified by (Provider, ProviderUserID).
type User struct {
	ID             uint      `json:"id"`
	Provider       string    `json:"provider" gorm:"uniqueIndex:idx_user_provider_uid"`
	ProviderUserID string    `json:"-" gorm:"uniqueIndex:idx_user_provider_uid"`
	Email          string    `json:"email"`
	Name           string    `json:"name"`
	AvatarURL      string    `json:"avatarUrl"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
}
