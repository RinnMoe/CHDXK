package application

import "time"

type AccountDTO struct {
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
	HasPointsAccount bool      `json:"has_points_account,omitempty"`
	SessionID        string    `json:"session_id,omitempty"`
	ExpiresAt        time.Time `json:"expires_at,omitempty"`
	ImpersonatedBy   string    `json:"impersonated_by,omitempty"`
}
