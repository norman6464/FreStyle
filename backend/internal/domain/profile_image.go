package domain

type ProfileImageUploadURL struct {
	UploadURL string `json:"uploadUrl"`
	ImageURL  string `json:"imageUrl"`
	Key       string `json:"key"`
	ExpiresIn int    `json:"expiresIn"`
}
