package repository

import (
	"context"

	"github.com/norman6464/frestyle/backend/internal/domain"
)

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
	// ListByUserID は本人の認証方法一覧を返す（段 14。表示専用。subject を返すのは
	// 呼び出し元が /me 経路にだけ配線する前提 — domain.UserIdentity の doc 参照）。
	ListByUserID(ctx context.Context, userID uint64) ([]domain.UserIdentity, error)
}
