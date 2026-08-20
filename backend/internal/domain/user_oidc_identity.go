package domain

import "time"

// UserOidcIdentity は OIDC プロバイダ由来のユーザー識別子（Cognito の sub 等）。
// 認証プロバイダの都合を users 本体から分離するための正規化テーブル（FRESTYLE-311）。
// 1 ユーザーはプロバイダごとに 1 identity（uq_user_oidc_user_provider）、
// 同一プロバイダ内で subject は一意（uq_user_oidc_provider_subject）。
type UserOidcIdentity struct {
	ID     uint64 `gorm:"primaryKey" json:"id"`
	UserID uint64 `gorm:"column:user_id;not null;uniqueIndex:uq_user_oidc_user_provider" json:"userId"`
	// Provider は発行元（現状 'cognito' のみ）。将来の複数 IdP を見越して列で持つ。
	Provider string `gorm:"column:provider;not null;default:cognito;uniqueIndex:uq_user_oidc_user_provider;uniqueIndex:uq_user_oidc_provider_subject" json:"provider"`
	// Subject はプロバイダが発行する不変のユーザー識別子（Cognito の sub）。
	Subject   string    `gorm:"column:subject;not null;uniqueIndex:uq_user_oidc_provider_subject" json:"subject"`
	CreatedAt time.Time `gorm:"column:created_at" json:"createdAt"`
	UpdatedAt time.Time `gorm:"column:updated_at" json:"updatedAt"`
}

func (UserOidcIdentity) TableName() string { return "user_oidc_identities" }

// OidcProviderCognito は現在唯一の OIDC プロバイダ名。
const OidcProviderCognito = "cognito"
