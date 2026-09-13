package domain

import "time"

// UserOidcIdentity は OIDC の発行者が付けたユーザー識別子。
// 認証の都合を users 本体から分離するための正規化テーブル。
// 1 ユーザーは発行者ごとに 1 identity（uq_user_oidc_user_provider）、
// 同一発行者内で subject は一意（uq_user_oidc_provider_subject）。
type UserOidcIdentity struct {
	ID     uint64 `json:"id"`
	UserID uint64 `json:"userId"`
	// Provider は発行者を区別する鍵。複数の発行者を並べられるよう列で持つ。
	Provider string `json:"provider"`
	// Subject は発行者が付ける不変のユーザー識別子（トークンの sub）。
	Subject   string    `json:"subject"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// OidcProviderDefault は user_oidc_identities.provider に入れる既定の鍵。
//
// **特定の発行者を指す値ではない。** 本番は GCIP、ローカルは Dex と、稼働する発行者は
// デプロイ先ごとに 1 つに決まる（infra/oidc.Verifier が発行者非依存に作られている理由と同じ）。
// アプリはどの発行者かを問わず、この 1 つの鍵で識別 subject を突き合わせる。
const OidcProviderDefault = "oidc"
