package application

type SendRegisterCodeCommand struct {
	Email string `json:"email"`
}

type RegisterCommand struct {
	Email    string `json:"email"`
	Code     string `json:"code"`
	Password string `json:"password"`
}

type LoginCommand struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type AccountDTO struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Role     string `json:"role"`
}

type SendResetCodeCommand struct {
	Email string `json:"email"`
}

type ResetPasswordCommand struct {
	Email       string `json:"email"`
	Code        string `json:"code"`
	NewPassword string `json:"new_password"`
}
