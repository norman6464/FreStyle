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

// adminInvitationRepository は [repository.AdminInvitationRepository] の実装。
// クエリは sqlc 生成コード（生 SQL）で、接続プール（*sql.DB）をそのまま受け取る。
type adminInvitationRepository struct{ db *sql.DB }

func NewAdminInvitationRepository(db *sql.DB) repository.AdminInvitationRepository {
	return &adminInvitationRepository{db: db}
}

// invitationRow は招待系の読み取りクエリが返す共通の行形。全クエリの SELECT に workspace_id
// を加えた結果、列リストが invitations の全列と一致し、sqlc は個別の Row 型ではなくテーブル型
// （sqlcgen.Invitation）をそのまま返すようになった。
type invitationRow = sqlcgen.Invitation

func toDomainAdminInvitation(row invitationRow) domain.AdminInvitation {
	inv := domain.AdminInvitation{
		ID:        uint64(row.ID),
		Email:     row.Email,
		Role:      domain.RoleName(row.Role),
		Name:      row.Name,
		Status:    row.Status,
		ExpiresAt: row.ExpiresAt,
		CreatedAt: row.CreatedAt,
	}
	if row.Token.Valid {
		// token は *string（未設定を NULL にして UNIQUE を避ける）。NULL は nil のまま。
		tok := row.Token.String
		inv.Token = &tok
	}
	if row.WorkspaceID.Valid {
		wid := row.WorkspaceID.UUID.String()
		inv.WorkspaceID = &wid
	}
	return inv
}

// pending 以外（accepted / canceled）は除外する（行は物理削除せず status のみ更新）。
// created_at は一意でないため、以下の一覧・単一取得はいずれも id をタイブレークに置いて順序を固定する
// （特に FindPendingByEmail は 1 件しか返さず、順序が揺れると「どの招待が受理されるか」が変わる）。
func (r *adminInvitationRepository) ListAll(ctx context.Context) ([]domain.AdminInvitation, error) {
	rows, err := sqlcgen.New(r.db).ListPendingInvitations(ctx, domain.InvitationStatusPending)
	if err != nil {
		return nil, err
	}
	return toDomainAdminInvitations(rows), nil
}

func (r *adminInvitationRepository) ListByWorkspaceID(ctx context.Context, workspaceID string) ([]domain.AdminInvitation, error) {
	wid, ok := toNullUUID(workspaceID)
	if !ok {
		return []domain.AdminInvitation{}, nil // 不正 / 空の ID は該当なしと同じ扱い
	}
	rows, err := sqlcgen.New(r.db).ListPendingInvitationsByWorkspace(ctx, sqlcgen.ListPendingInvitationsByWorkspaceParams{
		WorkspaceID: wid,
		Status:      domain.InvitationStatusPending,
	})
	if err != nil {
		return nil, err
	}
	return toDomainAdminInvitations(rows), nil
}

func toDomainAdminInvitations(rows []invitationRow) []domain.AdminInvitation {
	out := make([]domain.AdminInvitation, 0, len(rows))
	for _, row := range rows {
		out = append(out, toDomainAdminInvitation(row))
	}
	return out
}

// FindPendingByEmail は保留中の招待を email で引く。
// 突き合わせは domain.NormalizeEmail と同じ正規形どうしで行う。引数は Go 側で畳み、
// 列は users の一意索引と同じ SQL 式 lower(btrim(email, E'\t\n\x0B\f\r ')) で畳む
// （btrim の文字集合は domain.EmailTrimCutset と同じもの）。
// ユーザー側の email は正規形で保存・照会するため、招待作成時に大文字混じり・前後空白付きで
// 入った行を byte 一致で探すと「招待したのに招待が見つからない」状態になる
// （同じアドレスの解釈が 2 つある状態を作らない）。
func (r *adminInvitationRepository) FindPendingByEmail(ctx context.Context, email string) (*domain.AdminInvitation, error) {
	row, err := sqlcgen.New(r.db).FindPendingInvitationByEmail(ctx, sqlcgen.FindPendingInvitationByEmailParams{
		EmailNormal: domain.NormalizeEmail(email),
		Status:      domain.InvitationStatusPending,
	})
	if errors.Is(err, sql.ErrNoRows) {
		// 該当なしは招待ユーザーでない通常サインアップなので nil, nil を返す。
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	inv := toDomainAdminInvitation(row)
	return &inv, nil
}

func (r *adminInvitationRepository) FindByID(ctx context.Context, id uint64) (*domain.AdminInvitation, error) {
	if id == 0 {
		return nil, nil
	}
	id64, ok := toInt64ID(id)
	if !ok {
		return nil, nil // 存在し得ない id = 該当なし
	}
	row, err := sqlcgen.New(r.db).FindInvitationByID(ctx, id64)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	inv := toDomainAdminInvitation(row)
	return &inv, nil
}

func (r *adminInvitationRepository) FindPendingByToken(ctx context.Context, token string) (*domain.AdminInvitation, error) {
	if token == "" {
		return nil, nil
	}
	// 期限切れは DB 側で弾く。比較は DB 関数でなく Go の UTC 現在時刻をバインドする
	// （DB エンジン非依存 / ローカル TZ 設定に左右されない）。
	row, err := sqlcgen.New(r.db).FindPendingInvitationByToken(ctx, sqlcgen.FindPendingInvitationByTokenParams{
		Token:  sql.NullString{String: token, Valid: true},
		Status: domain.InvitationStatusPending,
		Now:    time.Now().UTC(),
	})
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	inv := toDomainAdminInvitation(row)
	return &inv, nil
}

// Create は招待行を保存する。email は正規形に畳んでから書く。
// 保存側だけ生のままだと、FindPendingByEmail を正規形の式に揃えても既存行と新規行で
// 「どちらの表現で入っているか」がばらつき、同じアドレスの解釈が 2 つある状態が続く。
// 呼び元（usecase）も畳んだ値を渡すが、この経路を通る限り必ず正規形になることをここで保証する。
func (r *adminInvitationRepository) Create(ctx context.Context, inv *domain.AdminInvitation) error {
	wid, ok := nullWorkspaceID(inv.WorkspaceID)
	if !ok {
		// 1 行も書けていないので nil を返さない（呼び出し側が作成できたと誤認する）。
		return fmt.Errorf("workspace_id が不正な形式です: %q", *inv.WorkspaceID)
	}
	inv.Email = domain.NormalizeEmail(inv.Email)
	createdAt := inv.CreatedAt
	if createdAt.IsZero() {
		createdAt = time.Now() // GORM autoCreateTime 相当（ゼロのときだけ now）
	}
	var token sql.NullString
	if inv.Token != nil {
		token = sql.NullString{String: *inv.Token, Valid: true}
	}
	row, err := sqlcgen.New(r.db).InsertInvitation(ctx, sqlcgen.InsertInvitationParams{
		WorkspaceID: wid,
		Email:       inv.Email,
		Role:        string(inv.Role),
		Name:        inv.Name,
		Status:      inv.Status,
		Token:       token,
		ExpiresAt:   inv.ExpiresAt,
		CreatedAt:   createdAt,
	})
	if err != nil {
		return err
	}
	inv.ID = uint64(row.ID)
	inv.CreatedAt = row.CreatedAt
	return nil
}

// UpdateStatus は招待の status を更新する。対象行が無ければ domain.ErrNotFound を返す。
//
// 0 行更新を成功にしてはいけない理由:
//
//	この更新は「招待を取り消した」「招待を消費した」という事実そのものを表す。
//	0 行のまま成功を返すと、取り消したはずの招待が pending のまま残っているのに
//	画面上は取り消し済みに見える、という食い違いが起きる。
//	呼び出し側（Cancel / ログイン時の受諾）はどちらも直前に招待を読み出しているので、
//	ここに落ちるのは「読み出しと更新のあいだに招待が消えた」競合のときだけ。
func (r *adminInvitationRepository) UpdateStatus(ctx context.Context, id uint64, status string) error {
	return updateInvitationStatus(ctx, sqlcgen.New(r.db), id, status)
}

func updateInvitationStatus(
	ctx context.Context,
	q *sqlcgen.Queries,
	id uint64,
	status string,
) error {
	id64, ok := toInt64ID(id)
	if !ok {
		return domain.ErrNotFound // 存在し得ない id = 対象なし
	}
	// :execrows なので実際に書き換わった行数が返る（:exec だと 0 行でも成功と区別が付かない）。
	affected, err := q.UpdateInvitationStatus(ctx, sqlcgen.UpdateInvitationStatusParams{
		ID:     id64,
		Status: status,
	})
	if err != nil {
		return err
	}
	if affected == 0 {
		return domain.ErrNotFound
	}
	return nil
}
