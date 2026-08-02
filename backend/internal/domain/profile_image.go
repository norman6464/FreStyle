package domain

// ProfileImageUploadURL は profile アイコン用の S3 直接アップロード URL を表す。
// UploadURL は PUT 対象、ImageURL はアップロード後に表示するパス。
// ImageURL に配信ドメインは含めない（FRESTYLE-234。ドメイン変更で保存済みデータが壊れないように）。
type ProfileImageUploadURL struct {
	UploadURL string `json:"uploadUrl"`
	ImageURL  string `json:"imageUrl"`
	Key       string `json:"key"`
	ExpiresIn int    `json:"expiresIn"`
}
