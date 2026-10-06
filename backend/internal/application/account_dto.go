package application

import "time"

type AccountDTO struct {
	Nickname           string     `json:"nickname"`
	LocalRole          string     `json:"local_role"`
	Permissions        []string   `json:"permissions"`
	MustChangePassword bool       `json:"must_change_password"`
	HasPassword        bool       `json:"has_password"`
	SuspendedAt        *time.Time `json:"suspended_at"`
	SuspendTill        *time.Time `json:"suspend_till"`
	MutedUntil         *time.Time `json:"muted_until"`
	ModerationReason   string     `json:"moderation_reason"`

	IsBootstrap      bool      `json:"is_bootstrap"`
	BoundNickname    string    `json:"bound_nickname,omitempty"`
	ID               int       `json:"id"`
	Username         string    `json:"username"`
	Email            string    `json:"email"`
	Role             string    `json:"role"`
	Subject          string    `json:"subject,omitempty"`
	IdentityRole     string    `json:"identity_role,omitempty"`
	IdentityType     string    `json:"identity_type,omitempty"`
	PermissionMode   string    `json:"permission_mode,omitempty"`
	ViolationCount   int       `json:"violation_count,omitempty"`
	InitialPassword  bool      `json:"initial_password,omitempty"`
	HasPointsAccount bool      `json:"has_points_account"`
	SessionID        string    `json:"session_id,omitempty"`
	ExpiresAt        time.Time `json:"expires_at,omitempty"`
	ImpersonatedBy   string    `json:"impersonated_by,omitempty"`
}
