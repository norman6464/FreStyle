package persistence

import (
	"context"
	"database/sql"
	"time"

	"github.com/norman6464/FreStyle/backend/internal/adapter/persistence/sqlcgen"
	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
)

// companyApplicationRepository は [repository.CompanyApplicationRepository] の実装。
// クエリは sqlc 生成コード（生 SQL）で、接続プール（*sql.DB）をそのまま受け取る。
type companyApplicationRepository struct{ db *sql.DB }

func NewCompanyApplicationRepository(db *sql.DB) repository.CompanyApplicationRepository {
	return &companyApplicationRepository{db: db}
}

func (r *companyApplicationRepository) Create(ctx context.Context, app *domain.CompanyApplication) error {
	now := time.Now()
	createdAt := app.CreatedAt
	if createdAt.IsZero() {
		createdAt = now // GORM autoCreateTime 相当（ゼロのときだけ now）
	}
	updatedAt := app.UpdatedAt
	if updatedAt.IsZero() {
		updatedAt = now // GORM autoUpdateTime 相当（ゼロのときだけ now）
	}
	row, err := sqlcgen.New(r.db).InsertCompanyApplication(ctx, sqlcgen.InsertCompanyApplicationParams{
		CompanyName:   app.CompanyName,
		ApplicantName: app.ApplicantName,
		Email:         app.Email,
		Message:       app.Message,
		Status:        app.Status,
		CreatedAt:     createdAt,
		UpdatedAt:     updatedAt,
	})
	if err != nil {
		return err
	}
	app.ID = uint64(row.ID)
	app.CreatedAt = row.CreatedAt
	app.UpdatedAt = row.UpdatedAt
	return nil
}

func (r *companyApplicationRepository) ListAll(ctx context.Context) ([]domain.CompanyApplication, error) {
	rows, err := sqlcgen.New(r.db).ListCompanyApplications(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]domain.CompanyApplication, 0, len(rows))
	for _, row := range rows {
		out = append(out, domain.CompanyApplication{
			ID:            uint64(row.ID),
			CompanyName:   row.CompanyName,
			ApplicantName: row.ApplicantName,
			Email:         row.Email,
			Message:       row.Message,
			Status:        row.Status,
			CreatedAt:     row.CreatedAt,
			UpdatedAt:     row.UpdatedAt,
		})
	}
	return out, nil
}

// UpdateStatus は申請の status を更新する。対象行が無ければ domain.ErrNotFound を返す。
//
// 0 行更新を成功にしてはいけない理由:
//
//	UPDATE は 1 行も一致しなくても SQL としては成功する。ここで nil を返すと handler は
//	204 を返し、管理者の画面では承認/却下が済んだように見えるのに DB には何も書かれていない。
//	この経路には上位での存在確認が無いため、他の管理者が先に消した申請・存在しない id を
//	叩いた場合が実際に 0 行になる。行が無いことは「更新できた」ではなく「対象が無い」なので
//	404 として返す。
func (r *companyApplicationRepository) UpdateStatus(ctx context.Context, id uint64, status string) error {
	id64, ok := toInt64ID(id)
	if !ok {
		return domain.ErrNotFound // 存在し得ない id = 対象なし
	}
	// :execrows なので実際に書き換わった行数が返る（:exec だと 0 行でも成功と区別が付かない）。
	affected, err := sqlcgen.New(r.db).UpdateCompanyApplicationStatus(ctx, sqlcgen.UpdateCompanyApplicationStatusParams{
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
