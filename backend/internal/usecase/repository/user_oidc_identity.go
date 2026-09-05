package repository

import "context"

// UserOidcIdentityRepository は user_oidc_identities テーブルへのアクセスを提供する。
// User 本体（UserRepository）とは独立した domain で、既存ユーザーへの provider
// 追加・張り直し（セルフヒール）でも単独で呼ばれる。新規ユーザー作成と不可分に
// 行いたい場合は、呼び出し側（usecase）が TxManager.DoInTx の中で
// UserRepository.Create と併せて呼ぶ。
type UserOidcIdentityRepository interface {
	// EnsureIdentity は (provider, subject) の identity を無ければ作る（冪等）。
	// 既存 (provider, subject) が別ユーザー所有の場合はエラーを返す
	// （無音で放置するとサイレントなログイン不能を作るため）。
	EnsureIdentity(ctx context.Context, userID uint64, provider, subject string) error
}
