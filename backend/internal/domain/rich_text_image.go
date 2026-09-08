package domain

// RichTextImageUploadURL はオブジェクトストレージへの直接アップロード用に発行する署名付き URL を表す。
type RichTextImageUploadURL struct {
	URL       string `json:"url"`
	Key       string `json:"key"`
	PublicURL string `json:"publicUrl"`
	ExpiresIn int    `json:"expiresIn"`
}
