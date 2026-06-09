package persistence

import (
	"context"
	"database/sql"
	"errors"

	"github.com/norman6464/FreStyle/backend/internal/adapter/persistence/sqlcgen"
	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
	"gorm.io/gorm"
)

// companyRepository は [repository.CompanyRepository] の実装。
// 読み取りは sqlc 生成コード（生 SQL）、書き込み（UpdateAiChatEnabled）は生 SQL の Exec。
type companyRepository struct{ db *gorm.DB }

func NewCompanyRepository(db *gorm.DB) repository.CompanyRepository {
	return &companyRepository{db: db}
}

func toDomainCompany(row sqlcgen.Company) domain.Company {
	return domain.Company{
		ID:                       uint64(row.ID),
		Name:                     row.Name,
		AiChatEnabledForTrainees: row.AiChatEnabledForTrainees,
		CreatedAt:                row.CreatedAt,
		UpdatedAt:                row.UpdatedAt,
	}
}

func (r *companyRepository) ListAll(ctx context.Context) ([]domain.Company, error) {
	sqlDB, err := r.db.DB()
	if err != nil {
		return nil, err
	}
	rows, err := sqlcgen.New(sqlDB).ListCompanies(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]domain.Company, 0, len(rows))
	for _, row := range rows {
		out = append(out, toDomainCompany(row))
	}
	return out, nil
}

func (r *companyRepository) FindByID(ctx context.Context, id uint64) (*domain.Company, error) {
	id64, ok := toInt64ID(id)
	if !ok {
		return nil, gorm.ErrRecordNotFound // 存在し得ない id = not found
	}
	sqlDB, err := r.db.DB()
	if err != nil {
		return nil, err
	}
	row, err := sqlcgen.New(sqlDB).GetCompanyByID(ctx, id64)
	if errors.Is(err, sql.ErrNoRows) {
		// AiChatEnabledForUserUseCase が ErrRecordNotFound を見て「会社行なし = 既定 true」にする契約を維持。
		return nil, gorm.ErrRecordNotFound
	}
	if err != nil {
		return nil, err
	}
	c := toDomainCompany(row)
	return &c, nil
}

// UpdateAiChatEnabled は ai_chat_enabled_for_trainees を更新する（生 SQL 直書き / updated_at も更新）。
func (r *companyRepository) UpdateAiChatEnabled(ctx context.Context, companyID uint64, enabled bool) error {
	const q = `UPDATE companies SET ai_chat_enabled_for_trainees = ?, updated_at = NOW() WHERE id = ?`
	return r.db.WithContext(ctx).Exec(q, enabled, companyID).Error
}
