// Package persistence は usecase 層が定義した port の永続化実装
// （sqlc 生成コード / DynamoDB / S3 presigner 等）を集約する。wiring は router.go で行う。
package persistence

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/norman6464/FreStyle/backend/internal/adapter/persistence/sqlcgen"
	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
)

// userRepository は [repository.UserRepository] の実装。
// クエリは sqlc 生成コード（生 SQL）で、接続プール（*sql.DB）をそのまま受け取る。
type userRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) repository.UserRepository {
	return &userRepository{db: db}
}

// queries は保持している接続プールで sqlc の Queries を作る（別 pool を持たない）。
func (r *userRepository) queries() *sqlcgen.Queries {
	return sqlcgen.New(r.db)
}

// withTx は 1 つのトランザクションを開き、その中でだけ有効な Queries を fn に渡す。
// fn がエラーを返せば（あるいは Commit に失敗すれば）書き込みはすべて巻き戻る。
func (r *userRepository) withTx(ctx context.Context, fn func(qtx *sqlcgen.Queries) error) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }() // Commit 済みなら no-op
	if err := fn(sqlcgen.New(r.db).WithTx(tx)); err != nil {
		return err
	}
	return tx.Commit()
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
	u.RoleID = row.RoleID
	if row.WorkspaceID.Valid {
		wid := row.WorkspaceID.UUID.String()
		u.WorkspaceID = &wid
	}
	if row.DeletedAt.Valid {
		t := row.DeletedAt.Time
		u.DeletedAt = &t
	}
	return u
}

func (r *userRepository) FindByCognitoSub(ctx context.Context, sub string) (*domain.User, error) {
	q := r.queries()
	row, err := q.GetUserByCognitoSub(ctx, sub)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return toDomainUser(userRow(row)), nil
}

func (r *userRepository) FindActiveByEmail(ctx context.Context, email string) (*domain.User, error) {
	q := r.queries()
	rows, err := q.ListActiveUsersByEmail(ctx, email)
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
		WorkspaceID: row.WorkspaceID, RoleID: row.RoleID,
		IsActive:  row.IsActive,
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
	q := r.queries()
	subject, err := q.GetCognitoSubjectByUserID(ctx, id64)
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
	q := r.queries()
	row, err := q.GetUserByID(ctx, id64)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return toDomainUser(row), nil
}

func (r *userRepository) ListByRole(ctx context.Context, role domain.RoleName) ([]domain.User, error) {
	q := r.queries()
	rows, err := q.ListUsersByRole(ctx, string(role))
	if err != nil {
		return nil, err
	}
	users := make([]domain.User, 0, len(rows))
	for _, row := range rows {
		users = append(users, *toDomainUser(userRow(row)))
	}
	return users, nil
}

func (r *userRepository) ListByWorkspaceID(ctx context.Context, workspaceID string) ([]domain.User, error) {
	wid, ok := toNullUUID(workspaceID)
	if !ok {
		return make([]domain.User, 0), nil
	}
	q := r.queries()
	rows, err := q.ListUsersByWorkspaceID(ctx, wid)
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
func resolveRoleID(ctx context.Context, q *sqlcgen.Queries, roleName domain.RoleName) (int32, error) {
	id, err := q.GetRoleIDByName(ctx, string(roleName))
	if errors.Is(err, sql.ErrNoRows) {
		return 0, fmt.Errorf("unknown role %q", roleName)
	}
	if err != nil {
		return 0, err
	}
	return id, nil
}

// insertUserTx は users 行を 1 件作り、採番結果を user へ書き戻す。
// created_at / updated_at に DB 既定値は無いのでここで値を決める（ゼロのときだけ now）。
func insertUserTx(ctx context.Context, q *sqlcgen.Queries, user *domain.User) error {
	now := time.Now()
	createdAt := user.CreatedAt
	if createdAt.IsZero() {
		createdAt = now
	}
	updatedAt := user.UpdatedAt
	if updatedAt.IsZero() {
		updatedAt = now
	}
	params := sqlcgen.InsertUserParams{
		Email:     user.Email,
		Name:      user.Name,
		RoleID:    user.RoleID,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}
	if user.PasswordHash != nil {
		params.PasswordHash = sql.NullString{String: *user.PasswordHash, Valid: true}
	}
	wid, ok := nullWorkspaceID(user.WorkspaceID)
	if !ok {
		return fmt.Errorf("workspace_id が不正な形式です: %q", *user.WorkspaceID)
	}
	params.WorkspaceID = wid
	if user.DeletedAt != nil {
		params.DeletedAt = sql.NullTime{Time: *user.DeletedAt, Valid: true}
	}
	var (
		newID        int64
		newCreatedAt time.Time
		newUpdatedAt time.Time
	)
	if user.ID == 0 {
		row, err := q.InsertUser(ctx, params)
		if err != nil {
			if isUniqueViolation(err) {
				return repository.ErrEmailTaken
			}
			return err
		}
		newID, newCreatedAt, newUpdatedAt = row.ID, row.CreatedAt, row.UpdatedAt
	} else {
		// 呼び出し側が id を決めた場合はそれを使う（採番シーケンスは進めない）。
		fixedID, ok := toInt64ID(user.ID)
		if !ok {
			return fmt.Errorf("user id %d が int64 の範囲外です", user.ID)
		}
		row, err := q.InsertUserWithID(ctx, sqlcgen.InsertUserWithIDParams{
			ID:           fixedID,
			Email:        params.Email,
			PasswordHash: params.PasswordHash,
			Name:         params.Name,
			WorkspaceID:  params.WorkspaceID,
			RoleID:       params.RoleID,
			CreatedAt:    params.CreatedAt,
			UpdatedAt:    params.UpdatedAt,
			DeletedAt:    params.DeletedAt,
		})
		if err != nil {
			if isUniqueViolation(err) {
				return repository.ErrEmailTaken
			}
			return err
		}
		newID, newCreatedAt, newUpdatedAt = row.ID, row.CreatedAt, row.UpdatedAt
	}
	user.ID = uint64(newID)
	user.CreatedAt = newCreatedAt
	user.UpdatedAt = newUpdatedAt
	// is_active は常に true で作る（作成直後のアカウントは有効。停止は UpdateActive の仕事）。
	user.IsActive = true
	return nil
}

// CreateWithOidcIdentity は users 行と OIDC identity を単一トランザクションで作成する。
// 正規化後は識別子（identity）を持たないユーザーは存在し得ないため、両者を不可分にする。
// identity 側が (provider, subject) 競合などで失敗するとトランザクションごと巻き戻り、
// users 行だけが残る（＝ログイン不能な孤児）状態を作らない。
func (r *userRepository) CreateWithOidcIdentity(ctx context.Context, user *domain.User, provider, subject string) error {
	return r.withTx(ctx, func(qtx *sqlcgen.Queries) error {
		return createWithOidcIdentity(ctx, qtx, user, provider, subject)
	})
}

// createWithOidcIdentity は渡された接続またはトランザクション上で
// users行とOIDC identityを作成する。
func createWithOidcIdentity(
	ctx context.Context,
	q *sqlcgen.Queries,
	user *domain.User,
	provider string,
	subject string,
) error {
	roleID, err := resolveRoleID(ctx, q, user.Role)
	if err != nil {
		return err
	}
	user.RoleID = roleID
	if err := insertUserTx(ctx, q, user); err != nil {
		return err
	}
	return ensureOidcIdentityTx(ctx, q, user.ID, provider, subject)
}

// bootstrapSuperAdminLockKey は「最初の運営管理者を作る」経路を直列化する advisory lock のキー。
// 起動時マイグレーション（database.migrateAdvisoryLockKey）とは別の値にする。
const bootstrapSuperAdminLockKey int64 = 7_419_063

// CreateFirstSuperAdminWithOidcIdentity は super_admin が 1 人も居ないときに限りユーザーを作る。
//
// 「0 人であること」を確かめてから作るまでのあいだに別のトランザクションが 1 人目を作れば、
// 招待を経ないこの経路が 2 人目・3 人目にも開いたままになる。判定を呼び出し側に置くと
// READ COMMITTED では互いの未コミット行が見えず、同時に来た 2 本がどちらも「0 人」を見る。
// そこで判定と INSERT を同じトランザクションに入れ、さらに advisory lock で経路自体を
// 直列化する（ロックはコミットで解放され、次の 1 本は必ず確定済みの 1 人目を見る）。
// email の一意索引に頼らないのは、索引が守るのは「同じアドレス」であって
// 「super_admin が 1 人」ではないため（別アドレスなら索引は素通りする）。
func (r *userRepository) CreateFirstSuperAdminWithOidcIdentity(
	ctx context.Context, user *domain.User, provider, subject string,
) (bool, error) {
	if user.Role != domain.RoleSuperAdmin {
		return false, fmt.Errorf("最初の運営管理者の作成に role %q が渡されました（super_admin 専用の経路です）", user.Role)
	}
	q := r.queries()
	roleID, err := resolveRoleID(ctx, q, user.Role)
	if err != nil {
		return false, err
	}
	user.RoleID = roleID

	created := false
	err = r.withTx(ctx, func(qtx *sqlcgen.Queries) error {
		// ロックは必ずこのトランザクション（qtx）で取る。pg_advisory_xact_lock は
		// トランザクションスコープなので、別の接続で発行すると取った直後に解放され、
		// 「0 人か確かめて作る」のあいだを誰も守らなくなる。
		// pgbouncer（transaction pooler）前提のため、セッションロックは使えない。
		if err := qtx.AcquireBootstrapSuperAdminLock(ctx, bootstrapSuperAdminLockKey); err != nil {
			return err
		}
		existing, err := qtx.CountActiveSuperAdmins(ctx, string(domain.RoleSuperAdmin))
		if err != nil {
			return err
		}
		if existing > 0 {
			return nil // created = false のまま。免除は既に閉じている。
		}
		if err := insertUserTx(ctx, qtx, user); err != nil {
			return err
		}
		if err := ensureOidcIdentityTx(ctx, qtx, user.ID, provider, subject); err != nil {
			return err
		}
		created = true
		return nil
	})
	if err != nil {
		return false, err
	}
	return created, nil
}

// EnsureOidcIdentity は (provider, subject) の identity を無ければ作る（冪等）。
// 既存ユーザーへの provider 追加・張り直し（セルフヒール）に使う。
// subject が別ユーザーに紐付いている場合は黙って成功にせずエラーを返す
// （無音で放置するとサイレントなログイン不能を作るため）。
func (r *userRepository) EnsureOidcIdentity(ctx context.Context, userID uint64, provider, subject string) error {
	q := r.queries()
	return ensureOidcIdentityTx(ctx, q, userID, provider, subject)
}

// ensureOidcIdentityTx は identity を冪等に挿入する（q は base 接続でもトランザクションでも良い）。
// 既存 (provider, subject) が別ユーザー所有ならエラー、自分の所有なら成功にする。
func ensureOidcIdentityTx(ctx context.Context, q *sqlcgen.Queries, userID uint64, provider, subject string) error {
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

// UpdateActive はユーザーアカウントの有効/無効を更新する（false で無効化 → ログイン/利用不可）。
// 対象が存在しなければ domain.ErrNotFound を返す（handler が 404 にマップ）。
func (r *userRepository) UpdateActive(ctx context.Context, userID uint64, active bool) error {
	id64, ok := toInt64ID(userID)
	if !ok {
		return domain.ErrNotFound // 存在し得ない id = not found
	}
	q := r.queries()
	affected, err := q.UpdateUserActive(ctx, sqlcgen.UpdateUserActiveParams{ID: id64, IsActive: active})
	if err != nil {
		return err
	}
	if affected == 0 {
		return domain.ErrNotFound
	}
	return nil
}

// SoftDelete はユーザーを論理削除する（deleted_at = now()）。以後 FindByCognitoSub 等で除外され、
// 認証時にも弾かれる。既に削除済み / 存在しない場合は domain.ErrNotFound を返す。
// OIDC identity も削除して subject の占有を解く（同じ OIDC アカウントの再招待を可能にする。
// ここで消し損ねても起動時バックフィルの掃除が自己修復する）。
//
// 2 文を 1 トランザクションにまとめないのは、無効化を必ず残すため。identity の掃除が失敗した
// ときに巻き戻すと、消したはずの利用者が有効なまま戻ってしまう（掃除漏れはバックフィルが直す）。
func (r *userRepository) SoftDelete(ctx context.Context, userID uint64) error {
	id64, ok := toInt64ID(userID)
	if !ok {
		return domain.ErrNotFound // 存在し得ない id = not found
	}
	q := r.queries()
	affected, err := q.SoftDeleteUser(ctx, id64)
	if err != nil {
		return err
	}
	if affected == 0 {
		return domain.ErrNotFound
	}
	return q.DeleteOidcIdentitiesByUserID(ctx, id64)
}

// UpdateName は氏名だけを更新する。対象が存在しなければ domain.ErrNotFound を返す。
func (r *userRepository) UpdateName(ctx context.Context, userID uint64, name string) error {
	id64, ok := toInt64ID(userID)
	if !ok {
		return domain.ErrNotFound // 存在し得ない id = not found
	}
	q := r.queries()
	affected, err := q.UpdateUserName(ctx, sqlcgen.UpdateUserNameParams{ID: id64, Name: name})
	if err != nil {
		return err
	}
	if affected == 0 {
		return domain.ErrNotFound
	}
	return nil
}

// UpdateRole は役割だけを更新する。対象が存在しなければ domain.ErrNotFound を返す。
// 昇格が 1 行も当たっていないのに成功を返すと、権限が上がったつもりの利用者が生まれるため、
// ここは特に 0 件を握り潰してはいけない。
func (r *userRepository) UpdateRole(ctx context.Context, userID uint64, role domain.RoleName) error {
	id64, ok := toInt64ID(userID)
	if !ok {
		return domain.ErrNotFound // 存在し得ない id = not found
	}
	q := r.queries()
	roleID, err := resolveRoleID(ctx, q, role)
	if err != nil {
		return err
	}
	affected, err := q.UpdateUserRoleID(ctx, sqlcgen.UpdateUserRoleIDParams{ID: id64, RoleID: roleID})
	if err != nil {
		return err
	}
	if affected == 0 {
		return domain.ErrNotFound
	}
	return nil
}

// UpdateWorkspaceID は所属ワークスペースを付け替える（招待の受諾で呼ばれる）。
// workspaceID は呼び出し側が既に解決した値をそのまま渡す（サブクエリで引き直さない）。
func (r *userRepository) UpdateWorkspaceID(ctx context.Context, userID uint64, workspaceID *string) error {
	id64, ok := toInt64ID(userID)
	if !ok {
		return domain.ErrNotFound // 存在し得ない id = not found
	}
	wid, ok := nullWorkspaceID(workspaceID)
	if !ok {
		return fmt.Errorf("workspace_id が不正な形式です: %q", *workspaceID)
	}
	q := r.queries()
	affected, err := q.UpdateUserWorkspaceID(ctx, sqlcgen.UpdateUserWorkspaceIDParams{
		ID:          id64,
		WorkspaceID: wid,
	})
	if err != nil {
		return err
	}
	if affected == 0 {
		return domain.ErrNotFound
	}
	return nil
}
