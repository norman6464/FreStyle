package domain

// ProfileImageUploadURL は profile アイコン用の S3 直接アップロード URL を表す。
// フロントの ProfileRepository.getImagePresignedUrl が期待する形式
// （uploadUrl: PUT 対象、imageUrl: アップロード後に表示する URL）に合わせる。
type ProfileImageUploadURL struct {
	UploadURL string `json:"uploadUrl"`
	ImageURL  string `json:"imageUrl"`
	Key       string `json:"key"`
	ExpiresIn int    `json:"expiresIn"`
}
