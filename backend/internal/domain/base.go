// Package domain は FreStyle の純粋なドメイン構造体・列挙・定数を集める層。
//
// クリーンアーキテクチャ的観点:
//   - domain は他のどの層にも依存しない (純粋なロジックのみ)
//   - usecase / repository / handler が domain を参照する
//   - GORM tag は本来 infrastructure 関心ごとだが、本プロジェクトでは
//     pragmatic な妥協として domain 構造体に直書きする運用にしている
//     （DTO ↔ DB Entity の二重定義を避け、移行コストを抑えるため）
//   - DB スキーマは Go domain を「正」として AutoMigrate で構築する
//     （詳細: docs/16-go-schema-design.md @ frestyle-infrastructure）
package domain

import "time"

// BaseEntity は全ての永続化ドメイン共通のフィールドをまとめる埋め込み用構造体。
// 各 entity は無名フィールドとして埋め込んで使う:
//
//	type User struct {
//	    BaseEntity
//	    ...
//	}
//
// 主キー / タイムスタンプの規約をプロジェクト全体で統一することで、
// 「テーブルによって id 型が違う」「updated_at がないテーブルがある」のような不揃いを防ぐ。
type BaseEntity struct {
	ID        uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime" json:"createdAt"`
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updatedAt"`
}
