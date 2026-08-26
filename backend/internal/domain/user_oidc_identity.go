package domain

import "time"

// UserOidcIdentity は OIDC プロバイダ由来のユーザー識別子（Cognito の sub 等）。
// 認証プロバイダの都合を users 本体から分離するための正規化テーブル（FRESTYLE-311）。
// 1 ユーザーはプロバイダごとに 1 identity（uq_user_oidc_user_provider）、
// 同一プロバイダ内で subject は一意（uq_user_oidc_provider_subject）。
type UserOidcIdentity struct {
	ID     uint64 `json:"id"`
	UserID uint64 `json:"userId"`
	// Provider は発行元（現状 'cognito' のみ）。将来の複数 IdP を見越して列で持つ。
	Provider string `json:"provider"`
	// Subject はプロバイダが発行する不変のユーザー識別子（Cognito の sub）。
	Subject   string    `json:"subject"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// OidcProviderCognito は現在唯一の OIDC プロバイダ名。
const OidcProviderCognito = "cognito"
