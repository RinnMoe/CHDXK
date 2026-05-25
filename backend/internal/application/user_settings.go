package application

type UserSettingsDTO struct {
	CurrentSemester string `json:"current_semester"`
}

type UpdateUserSettingsCommand struct {
	CurrentSemester string `json:"current_semester"`
}
