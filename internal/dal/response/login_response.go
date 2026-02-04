package response

type LoginResponse struct {
	Id           int64  `json:"id"`
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
	Username     string `json:"username"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refreshToken" binding:"required"`
}
