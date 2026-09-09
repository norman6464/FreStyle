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
//
// 以前は撤去済みの旧発行者を指す名前・値のまま残っていた（発行者を切り替えたときに
// 値まで変えると既存行の書き換え＝データ移行になり、名前だけ先に変えると
// 「どちらが正か」が読めなくなるため、両方まとめて直すまで意図的に据え置いていた）。
// 2026-09-09、本番データを全削除し user_oidc_identities が 0 件になったタイミングで
// 名前と値を一緒に直した。
const OidcProviderDefault = "oidc"
