// Package persistence は usecase 層が定義した port の永続化実装
// （GORM / DynamoDB / S3 presigner 等）を集約する。wiring は router.go で行う。
package persistence

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/norman6464/FreStyle/backend/internal/adapter/persistence/sqlcgen"
	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
	"gorm.io/gorm"
)

// userRepository は [repository.UserRepository] の実装。
// 読み取りは sqlc 生成コード（生 SQL 直書き）、書き込みは GORM（autoTime / 採番）を使う。
// 接続は GORM の *sql.DB を sqlc に共有する（別 pool を持たない）。
type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) repository.UserRepository {
	return &userRepository{db: db}
}

// userRow は user 系クエリが返す共通の行形（users 全列 + JOIN で解決した role_name）。
// 各クエリの生成 Row 型はフィールド構成が同一なので、この型へ変換して 1 つの詰め替えに集約する。
type userRow = sqlcgen.GetUserByIDRow

// toDomainUser は sqlc 生成モデル → domain への詰め替え。
// Role はロールマスタ（roles）を JOIN して解決した role_name。
// id 系は DB が bigint(int64) で domain が uint64。値は採番シーケンス由来で常に非負・int64 範囲内のため
// 変換は安全（gosec G115 は persistence の id 境界として .golangci.yml で除外）。
func toDomainUser(row userRow) *domain.User {
	u := &domain.User{
		ID:        uint64(row.ID),
		Email:     row.Email,
		Name:      row.Name,
		Role:      domain.RoleName(row.RoleName),
		IsActive:  row.IsActive,
		CreatedAt: row.CreatedAt,
		UpdatedAt: row.UpdatedAt,
	}
	if row.RoleID.Valid {
		u.RoleID = uint16(row.RoleID.Int16)
	}
	if row.CompanyID.Valid {
		cid := uint64(row.CompanyID.Int64)
		u.CompanyID = &cid
	}
	if row.AiChatEnabled.Valid {
		v := row.AiChatEnabled.Bool
		u.AiChatEnabled = &v
	}
	if row.DeletedAt.Valid {
		t := row.DeletedAt.Time
		u.DeletedAt = &t
	}
	return u
}

func (r *userRepository) FindByCognitoSub(ctx context.Context, sub string) (*domain.User, error) {
	sqlDB, err := r.db.DB()
	if err != nil {
		return nil, err
	}
	row, err := sqlcgen.New(sqlDB).GetUserByCognitoSub(ctx, sub)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return toDomainUser(userRow(row)), nil
}

func (r *userRepository) FindActiveByEmail(ctx context.Context, email string) (*domain.User, error) {
	sqlDB, err := r.db.DB()
	if err != nil {
		return nil, err
	}
	rows, err := sqlcgen.New(sqlDB).ListActiveUsersByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, nil
	}
	// uq_users_email_active があれば最大 1 行。既存データの重複で index 未作成のまま起動して
	// いる環境では複数行になり得るが、その状態でのログインは別人アカウントへの解決になり得る
	// ため拒否する（起動時 WARNING の重複解消を促す）。
	if len(rows) > 1 {
		return nil, fmt.Errorf("email %q のアクティブユーザーが %d 件あります（uq_users_email_active 未作成。重複を解消してください）", email, len(rows))
	}
	row := rows[0]
	u := toDomainUser(userRow{
		ID: row.ID, Email: row.Email, Name: row.Name,
		CompanyID: row.CompanyID, RoleID: row.RoleID,
		AiChatEnabled: row.AiChatEnabled, IsActive: row.IsActive,
		CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt, DeletedAt: row.DeletedAt,
		RoleName: row.RoleName,
	})
	if row.PasswordHash.Valid {
		v := row.PasswordHash.String
		u.PasswordHash = &v
	}
	return u, nil
}

func (r *userRepository) CognitoSubjectByUserID(ctx context.Context, userID uint64) (string, error) {
	id64, ok := toInt64ID(userID)
	if !ok {
		return "", nil
	}
	sqlDB, err := r.db.DB()
	if err != nil {
		return "", err
	}
	subject, err := sqlcgen.New(sqlDB).GetCognitoSubjectByUserID(ctx, id64)
	if errors.Is(err, sql.ErrNoRows) {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	return subject, nil
}

func (r *userRepository) FindByID(ctx context.Context, id uint64) (*domain.User, error) {
	id64, ok := toInt64ID(id)
	if !ok {
		return nil, nil // int64 範囲外 = 存在し得ない id
	}
	sqlDB, err := r.db.DB()
	if err != nil {
		return nil, err
	}
	row, err := sqlcgen.New(sqlDB).GetUserByID(ctx, id64)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return toDomainUser(row), nil
}

func (r *userRepository) ListByRole(ctx context.Context, role domain.RoleName) ([]domain.User, error) {
	sqlDB, err := r.db.DB()
	if err != nil {
		return nil, err
	}
	rows, err := sqlcgen.New(sqlDB).ListUsersByRole(ctx, string(role))
	if err != nil {
		return nil, err
	}
	users := make([]domain.User, 0, len(rows))
	for _, row := range rows {
		users = append(users, *toDomainUser(userRow(row)))
	}
	return users, nil
}

func (r *userRepository) ListByCompanyID(ctx context.Context, companyID uint64) ([]domain.User, error) {
	cid, ok := toInt64ID(companyID)
	if !ok {
		// 範囲外の ID は該当なしと同じ扱い。nil を返すと JSON が null になり
		// フロントの map が落ちるため空スライスにする（FRESTYLE-77）。
		return make([]domain.User, 0), nil
	}
	sqlDB, err := r.db.DB()
	if err != nil {
		return nil, err
	}
	rows, err := sqlcgen.New(sqlDB).ListUsersByCompanyID(ctx, sql.NullInt64{Int64: cid, Valid: true})
	if err != nil {
		return nil, err
	}
	users := make([]domain.User, 0, len(rows))
	for _, row := range rows {
		users = append(users, *toDomainUser(userRow(row)))
	}
	return users, nil
}

// resolveRoleID はロール名を roles.id に解決する。未知の名前はエラー（黙って別ロールにしない）。
func (r *userRepository) resolveRoleID(ctx context.Context, roleName domain.RoleName) (uint16, error) {
	var role domain.Role
	if err := r.db.WithContext(ctx).Where("name = ?", string(roleName)).Take(&role).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, fmt.Errorf("unknown role %q", roleName)
		}
		return 0, err
	}
	return role.ID, nil
}

// CreateWithOidcIdentity は users 行と OIDC identity を単一トランザクションで作成する。
// 正規化後は識別子（identity）を持たないユーザーは存在し得ないため、両者を不可分にする。
// identity 側が (provider, subject) 競合などで失敗するとトランザクションごと巻き戻り、
// users 行だけが残る（＝ログイン不能な孤児）状態を作らない。
func (r *userRepository) CreateWithOidcIdentity(ctx context.Context, user *domain.User, provider, subject string) error {
	roleID, err := r.resolveRoleID(ctx, user.Role)
	if err != nil {
		return err
	}
	user.RoleID = roleID
	// ローリングデプロイ中に旧タスクが読む旧カラム cognito_sub へ subject を併記する
	// （撤去は FRESTYLE-311 の後続 PR）。identity（正）とは同一トランザクションで対に作る。
	if provider == domain.OidcProviderCognito {
		user.CognitoSub = subject
	}
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(user).Error; err != nil {
			return err
		}
		return ensureOidcIdentityTx(tx, user.ID, provider, subject)
	})
}

// EnsureOidcIdentity は (provider, subject) の identity を無ければ作る（冪等）。
// 既存ユーザーへの provider 追加・張り直し（セルフヒール）に使う。
// subject が別ユーザーに紐付いている場合は黙って成功にせずエラーを返す
// （無音で放置するとサイレントなログイン不能を作るため）。
func (r *userRepository) EnsureOidcIdentity(ctx context.Context, userID uint64, provider, subject string) error {
	return ensureOidcIdentityTx(r.db.WithContext(ctx), userID, provider, subject)
}

// ensureOidcIdentityTx は identity を冪等に挿入する（db は base 接続でもトランザクションでも良い）。
// 既存 (provider, subject) が別ユーザー所有ならエラー、自分の所有なら成功にする。
func ensureOidcIdentityTx(db *gorm.DB, userID uint64, provider, subject string) error {
	res := db.Exec(
		`INSERT INTO user_oidc_identities (user_id, provider, subject, created_at, updated_at)
		 VALUES (?, ?, ?, NOW(), NOW())
		 ON CONFLICT (provider, subject) DO NOTHING`, userID, provider, subject,
	)
	if res.Error != nil {
		// (user_id, provider) の一意制約違反（同一ユーザーが別 subject を保持）はここでエラーになる。
		return res.Error
	}
	if res.RowsAffected == 1 {
		return nil
	}
	// 挿入されなかった = (provider, subject) が既に存在する。所有者が自分なら冪等成功。
	var identity domain.UserOidcIdentity
	if err := db.
		Where("provider = ? AND subject = ?", provider, subject).
		Take(&identity).Error; err != nil {
		return err
	}
	if identity.UserID != userID {
		return fmt.Errorf(
			"oidc identity conflict: provider=%s の subject は既に user %d に紐付いています（要求 user %d）",
			provider, identity.UserID, userID,
		)
	}
	return nil
}

// UpdateAiChatEnabled は AI チャットの個別上書きを更新する。enabled=nil で NULL（会社設定に従う）に戻す。
func (r *userRepository) UpdateAiChatEnabled(ctx context.Context, userID uint64, enabled *bool) error {
	var value any = gorm.Expr("NULL")
	if enabled != nil {
		value = *enabled
	}
	return r.db.WithContext(ctx).
		Model(&domain.User{}).
		Where("id = ?", userID).
		Update("ai_chat_enabled", value).Error
}

// UpdateActive はユーザーアカウントの有効/無効を更新する（false で無効化 → ログイン/利用不可）。
// 対象が存在しなければ gorm.ErrRecordNotFound を返す（handler が 404 にマップ）。
func (r *userRepository) UpdateActive(ctx context.Context, userID uint64, active bool) error {
	res := r.db.WithContext(ctx).
		Model(&domain.User{}).
		Where("id = ?", userID).
		Update("is_active", active)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// SoftDelete はユーザーを論理削除する（deleted_at = NOW()）。以後 FindByCognitoSub 等で除外され、
// 認証時にも弾かれる。既に削除済み / 存在しない場合は gorm.ErrRecordNotFound を返す。
// OIDC identity も削除して subject の占有を解く（同じ OIDC アカウントの再招待を可能にする。
// ここで消し損ねても起動時バックフィルの掃除が自己修復する）。
func (r *userRepository) SoftDelete(ctx context.Context, userID uint64) error {
	res := r.db.WithContext(ctx).
		Model(&domain.User{}).
		Where("id = ? AND deleted_at IS NULL", userID).
		Update("deleted_at", gorm.Expr("NOW()"))
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Delete(&domain.UserOidcIdentity{}).Error
}

func (r *userRepository) UpdateName(ctx context.Context, userID uint64, name string) error {
	return r.db.WithContext(ctx).
		Model(&domain.User{}).
		Where("id = ?", userID).
		Update("name", name).Error
}

func (r *userRepository) UpdateRole(ctx context.Context, userID uint64, role domain.RoleName) error {
	roleID, err := r.resolveRoleID(ctx, role)
	if err != nil {
		return err
	}
	// 正は role_id。ローリングデプロイ中に旧タスクが読む旧 role カラムへも併記する
	// （撤去は FRESTYLE-311 の後続 PR）。
	return r.db.WithContext(ctx).
		Model(&domain.User{}).
		Where("id = ?", userID).
		Updates(map[string]any{"role": string(role), "role_id": roleID}).Error
}

func (r *userRepository) UpdateCompanyID(ctx context.Context, userID uint64, companyID uint64) error {
	return r.db.WithContext(ctx).
		Model(&domain.User{}).
		Where("id = ?", userID).
		Update("company_id", companyID).Error
}
