package domain

// NoteImageUploadURL は S3 への直接アップロード用に発行する署名付き URL を表す。
type NoteImageUploadURL struct {
	URL       string `json:"url"`
	Key       string `json:"key"`
	ExpiresIn int    `json:"expiresIn"`
}
