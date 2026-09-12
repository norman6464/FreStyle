package profile

import (
	"context"
	"errors"
	"time"

	"github.com/norman6464/frestyle/backend/internal/domain"
	"github.com/norman6464/frestyle/backend/internal/usecase/repository"
)

// GetProfileUseCase は指定 user のプロフィールを返す。
type GetProfileUseCase struct {
	profiles repository.ProfileRepository
}

func NewGetProfileUseCase(p repository.ProfileRepository) *GetProfileUseCase {
	return &GetProfileUseCase{profiles: p}
}

func (u *GetProfileUseCase) Execute(ctx context.Context, userID uint64) (*domain.Profile, error) {
	if userID == 0 {
		return nil, errors.New("userID is required")
	}
	return u.profiles.FindByUserID(ctx, userID)
}

// UpdateProfileUseCase はプロフィールの任意フィールドを upsert する。
type UpdateProfileUseCase struct {
	profiles repository.ProfileRepository
}

func NewUpdateProfileUseCase(p repository.ProfileRepository) *UpdateProfileUseCase {
	return &UpdateProfileUseCase{profiles: p}
}

type UpdateProfileInput struct {
	UserID     uint64
	Bio        string
	AvatarURL  string
	StatusText string
}

func (u *UpdateProfileUseCase) Execute(ctx context.Context, in UpdateProfileInput) (*domain.Profile, error) {
	if in.UserID == 0 {
		return nil, errors.New("userID is required")
	}
	p := &domain.Profile{
		UserID:     in.UserID,
		Bio:        in.Bio,
		AvatarURL:  in.AvatarURL,
		StatusText: in.StatusText,
	}
	if err := u.profiles.Upsert(ctx, p); err != nil {
		return nil, err
	}
	return p, nil
}

// UpdateStatusInput は PUT /me/status の入力（段 14）。絵文字・テキスト・失効時刻だけを
// 扱う。bio / avatarURL には触れない（UpdateProfileUseCase の専管。互いの担当を混ぜない）。
type UpdateStatusInput struct {
	UserID    uint64
	Emoji     string
	Text      string
	ExpiresAt *time.Time
}

// UpdateStatusUseCase は一言ステータス（絵文字・テキスト・失効時刻）だけを upsert する。
type UpdateStatusUseCase struct {
	profiles repository.ProfileRepository
}

func NewUpdateStatusUseCase(p repository.ProfileRepository) *UpdateStatusUseCase {
	return &UpdateStatusUseCase{profiles: p}
}

func (u *UpdateStatusUseCase) Execute(ctx context.Context, in UpdateStatusInput) (*domain.Profile, error) {
	if in.UserID == 0 {
		return nil, errors.New("userID is required")
	}
	return u.profiles.UpdateStatus(ctx, in.UserID, in.Emoji, in.Text, in.ExpiresAt)
}

// ListMyIdentitiesUseCase は本人の認証方法一覧を返す（段 14。表示専用）。
type ListMyIdentitiesUseCase struct {
	identities repository.UserOidcIdentityRepository
}

func NewListMyIdentitiesUseCase(r repository.UserOidcIdentityRepository) *ListMyIdentitiesUseCase {
	return &ListMyIdentitiesUseCase{identities: r}
}

func (u *ListMyIdentitiesUseCase) Execute(ctx context.Context, userID uint64) ([]domain.UserIdentity, error) {
	if userID == 0 {
		return nil, errors.New("userID is required")
	}
	return u.identities.ListByUserID(ctx, userID)
}

// IssueProfileImageUploadURLUseCase は profile アイコン用 PUT 署名付き URL を発行する。
type IssueProfileImageUploadURLUseCase struct {
	presigner repository.ProfileImagePresigner
}

func NewIssueProfileImageUploadURLUseCase(p repository.ProfileImagePresigner) *IssueProfileImageUploadURLUseCase {
	return &IssueProfileImageUploadURLUseCase{presigner: p}
}

func (u *IssueProfileImageUploadURLUseCase) Execute(ctx context.Context, userID uint64, contentType string, size int64) (*domain.ProfileImageUploadURL, error) {
	if userID == 0 {
		return nil, errors.New("userID is required")
	}
	return u.presigner.Generate(ctx, userID, contentType, size)
}
