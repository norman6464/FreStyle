package domain

// NoteImageUploadURL は S3 への直接アップロード用に発行する署名付き URL を表す。
type NoteImageUploadURL struct {
	URL string `json:"url"`
	Key string `json:"key"`
	// PublicURL はアップロード後に img / Markdown から参照する配信用 URL（CDN 経由）。
	PublicURL string `json:"publicUrl"`
	ExpiresIn int    `json:"expiresIn"`
}
