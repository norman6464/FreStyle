// Package domain は他層に依存しない純粋なドメイン構造体・列挙・定数を集める。
// GORM tag は pragmatic な妥協として domain に直書きし、DB スキーマは AutoMigrate で構築する。
package domain

import "time"

// BaseEntity は永続化ドメイン共通の主キー / タイムスタンプを埋め込みで提供する。
type BaseEntity struct {
	ID        uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime" json:"createdAt"`
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updatedAt"`
}
