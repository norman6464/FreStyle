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

// OidcProviderCognito は user_oidc_identities.provider に入れる鍵。
//
// **値が "cognito" のままなのは歴史的な理由で、いま使っている発行者を指してはいない。**
// これは DB に保存済みの値なので、変えると既存行の書き換え（データ移行）になる。
// 発行者を切り替えたときに一緒に変えなかったのは、移行のタイミングと
// 本番データの扱いが別の判断だから。名前を先に変えて値だけ残すと、
// 「どちらが正か」が読めなくなるので、両方まとめて直すまでこのままにする。
const OidcProviderCognito = "cognito"
