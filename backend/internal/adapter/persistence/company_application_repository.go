package persistence

import (
	"context"
	"time"

	"github.com/norman6464/FreStyle/backend/internal/adapter/persistence/sqlcgen"
	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
	"gorm.io/gorm"
)

// companyApplicationRepository は [repository.CompanyApplicationRepository] の実装。
// クエリは sqlc 生成コード（生 SQL）で、GORM からは接続プール（*sql.DB）だけを借りる。
type companyApplicationRepository struct{ db *gorm.DB }

func NewCompanyApplicationRepository(db *gorm.DB) repository.CompanyApplicationRepository {
	return &companyApplicationRepository{db: db}
}

func (r *companyApplicationRepository) Create(ctx context.Context, app *domain.CompanyApplication) error {
	sqlDB, err := r.db.DB()
	if err != nil {
		return err
	}
	now := time.Now()
	createdAt := app.CreatedAt
	if createdAt.IsZero() {
		createdAt = now // GORM autoCreateTime 相当（ゼロのときだけ now）
	}
	updatedAt := app.UpdatedAt
	if updatedAt.IsZero() {
		updatedAt = now // GORM autoUpdateTime 相当（ゼロのときだけ now）
	}
	row, err := sqlcgen.New(sqlDB).InsertCompanyApplication(ctx, sqlcgen.InsertCompanyApplicationParams{
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
	sqlDB, err := r.db.DB()
	if err != nil {
		return nil, err
	}
	rows, err := sqlcgen.New(sqlDB).ListCompanyApplications(ctx)
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

func (r *companyApplicationRepository) UpdateStatus(ctx context.Context, id uint64, status string) error {
	id64, ok := toInt64ID(id)
	if !ok {
		return nil // 存在し得ない id = 対象なし
	}
	sqlDB, err := r.db.DB()
	if err != nil {
		return err
	}
	return sqlcgen.New(sqlDB).UpdateCompanyApplicationStatus(ctx, sqlcgen.UpdateCompanyApplicationStatusParams{
		ID:     id64,
		Status: status,
	})
}
