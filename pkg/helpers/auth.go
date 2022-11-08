package helpers

type SignIn struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type SignOut struct {
	Username string `json:"username"`
}

type RefreshToken struct {
	RefreshToken string `json:"refresh_token"`
}

type ForgotPassword struct {
	Email string `json:"email"`
}

type ResetPassword struct {
	RefreshToken string `json:"refresh_token"`
	OldPassword  string `json:"old_password"`
	NewPassword  string `json:"new_password"`
}
