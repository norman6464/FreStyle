package persistence

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/norman6464/FreStyle/backend/internal/adapter/persistence/sqlcgen"
	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
)

// userOidcIdentityRepository は [repository.UserOidcIdentityRepository] の実装。
type userOidcIdentityRepository struct {
	baseRepository
}

// NewUserOidcIdentityRepository は user_oidc_identities の repository を組み立てる。
func NewUserOidcIdentityRepository(db *sql.DB) repository.UserOidcIdentityRepository {
	return &userOidcIdentityRepository{baseRepository{db: db}}
}

// EnsureIdentity は (provider, subject) の identity を無ければ作る（冪等）。
// 既存ユーザーへの provider 追加・張り直し（セルフヒール）にも、新規ユーザー作成の
// 一部として TxManager.DoInTx の中から呼ぶ場合にも使う。
// subject が別ユーザーに紐付いている場合は黙って成功にせずエラーを返す
// （無音で放置するとサイレントなログイン不能を作るため）。
func (r *userOidcIdentityRepository) EnsureIdentity(ctx context.Context, userID uint64, provider, subject string) error {
	q := sqlcgen.New(r.dbtx(ctx))
	id64, ok := toInt64ID(userID)
	if !ok {
		return fmt.Errorf("user id %d が int64 の範囲外です", userID)
	}
	inserted, err := q.InsertOidcIdentityIfAbsent(ctx, sqlcgen.InsertOidcIdentityIfAbsentParams{
		UserID:   id64,
		Provider: provider,
		Subject:  subject,
	})
	if err != nil {
		// (user_id, provider) の一意制約違反（同一ユーザーが別 subject を保持）はここでエラーになる。
		return err
	}
	if inserted == 1 {
		return nil
	}
	// 挿入されなかった = (provider, subject) が既に存在する。所有者が自分なら冪等成功。
	ownerID, err := q.GetOidcIdentityOwner(ctx, sqlcgen.GetOidcIdentityOwnerParams{
		Provider: provider,
		Subject:  subject,
	})
	if errors.Is(err, sql.ErrNoRows) {
		return domain.ErrNotFound // 直後に消えた（従来の Take と同じシグナル）
	}
	if err != nil {
		return err
	}
	if ownerID != id64 {
		return fmt.Errorf(
			"oidc identity conflict: provider=%s の subject は既に user %d に紐付いています（要求 user %d）",
			provider, ownerID, userID,
		)
	}
	return nil
}
