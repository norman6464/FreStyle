package domain

// ProfileImageUploadURL は profile アイコン用の S3 直接アップロード URL を表す。
// UploadURL は PUT 対象、ImageURL はアップロード後に表示する URL。
type ProfileImageUploadURL struct {
	UploadURL string `json:"uploadUrl"`
	ImageURL  string `json:"imageUrl"`
	Key       string `json:"key"`
	ExpiresIn int    `json:"expiresIn"`
}
