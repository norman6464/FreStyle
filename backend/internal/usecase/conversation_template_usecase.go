package usecase

import (
	"context"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/repository"
)

type ListConversationTemplatesUseCase struct {
	repo repository.ConversationTemplateRepository
}

func NewListConversationTemplatesUseCase(r repository.ConversationTemplateRepository) *ListConversationTemplatesUseCase {
	return &ListConversationTemplatesUseCase{repo: r}
}

func (u *ListConversationTemplatesUseCase) Execute(ctx context.Context) ([]domain.ConversationTemplate, error) {
	return u.repo.ListActive(ctx)
}
