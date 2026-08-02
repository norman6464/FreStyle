package domain

// NoteImageUploadURL は S3 への直接アップロード用に発行する署名付き URL を表す。
type NoteImageUploadURL struct {
	URL string `json:"url"`
	Key string `json:"key"`
	// PublicURL はアップロード後に img / Markdown から参照する表示用パス。
	// 配信ドメインは含めない（FRESTYLE-234。ドメイン変更で保存済みデータが壊れないように）。
	PublicURL string `json:"publicUrl"`
	ExpiresIn int    `json:"expiresIn"`
}
