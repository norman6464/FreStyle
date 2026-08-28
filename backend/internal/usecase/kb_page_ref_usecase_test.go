package usecase_test

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
)

const kbRefWS = "00000000-0000-7000-8000-00000000aaaa"

func kbRefDoc(refs ...string) string {
	content := ""
	for i, id := range refs {
		if i > 0 {
			content += ","
		}
		content += fmt.Sprintf(`{"type":"pageRef","attrs":{"pageId":%q,"title":"無題"}}`, id)
	}
	return `{"type":"doc","content":[{"type":"paragraph","content":[` + content + `]}]}`
}

func kbViewableFacts(id, title string) repository.PageWithViewFacts {
	return repository.PageWithViewFacts{
		Page:  domain.Page{ID: id, Title: title},
		Facts: domain.PageViewFacts{Role: rolePtr(domain.GrantRoleViewer)},
	}
}

func rolePtr(r domain.GrantRole) *domain.GrantRole { return &r }

func Test_ページ参照の題名解決_閲覧できる参照だけを現在の題名にする(t *testing.T) {
	repo := &mockKBPermissionRepo{}
	visible := "00000000-0000-7000-8000-000000000001"
	hidden := "00000000-0000-7000-8000-000000000002"
	// hidden は経路上に deny がある（事実だけ返し、判定は usecase 側の ResolvePageView）。
	deniedFacts := repository.PageWithViewFacts{
		Page: domain.Page{ID: hidden, Title: "隠しページの新題名"},
		Facts: domain.PageViewFacts{
			Role: rolePtr(domain.GrantRoleViewer),
			View: &domain.RestrictionFacts{DeniedAnywhere: true},
		},
	}
	repo.On("ListWorkspacePageViewFactsByIDs", mock.Anything, kbRefWS, uint64(7), []string{visible, hidden}).
		Return([]repository.PageWithViewFacts{kbViewableFacts(visible, "設計メモ v2"), deniedFacts}, nil)

	uc := usecase.NewResolvePageRefTitlesUseCase(repo)
	got := uc.Execute(context.Background(), usecase.ResolvePageRefTitlesInput{
		WorkspaceID: kbRefWS, UserID: 7, Doc: kbRefDoc(visible, hidden),
	})

	var doc map[string]any
	assert.NoError(t, json.Unmarshal([]byte(got), &doc))
	inline := doc["content"].([]any)[0].(map[string]any)["content"].([]any)
	first := inline[0].(map[string]any)["attrs"].(map[string]any)
	second := inline[1].(map[string]any)["attrs"].(map[string]any)
	assert.Equal(t, "設計メモ v2", first["title"], "閲覧できる参照は現在の題名になる")
	// deny のある参照は差し替えない。保存されていた文字（著者が書けた情報）のまま。
	assert.Equal(t, "無題", second["title"])
	repo.AssertExpectations(t)
}

func Test_ページ参照の題名解決_参照が無ければ問い合わせず原文のまま(t *testing.T) {
	repo := &mockKBPermissionRepo{}
	uc := usecase.NewResolvePageRefTitlesUseCase(repo)
	doc := `{"type":"doc","content":[{"type":"paragraph","content":[{"type":"text","text":"参照なし"}]}]}`

	got := uc.Execute(context.Background(), usecase.ResolvePageRefTitlesInput{
		WorkspaceID: kbRefWS, UserID: 7, Doc: doc,
	})

	assert.Equal(t, doc, got)
	repo.AssertNotCalled(t, "ListWorkspacePageViewFactsByIDs")
}

func Test_ページ参照の題名解決_壊れたdocや取得失敗では原文を返す(t *testing.T) {
	repo := &mockKBPermissionRepo{}
	uc := usecase.NewResolvePageRefTitlesUseCase(repo)

	broken := `{"type":"doc","content":[`
	assert.Equal(t, broken, uc.Execute(context.Background(), usecase.ResolvePageRefTitlesInput{
		WorkspaceID: kbRefWS, UserID: 7, Doc: broken,
	}), "読めない doc はそのまま返す（本文が開けることが題名より重い）")

	id := "00000000-0000-7000-8000-000000000001"
	repo.On("ListWorkspacePageViewFactsByIDs", mock.Anything, kbRefWS, uint64(7), []string{id}).
		Return(nil, assert.AnError)
	doc := kbRefDoc(id)
	assert.Equal(t, doc, uc.Execute(context.Background(), usecase.ResolvePageRefTitlesInput{
		WorkspaceID: kbRefWS, UserID: 7, Doc: doc,
	}), "事実の取得失敗でも原文を返す")
}

func Test_ページ参照の題名解決_同じ参照は1回だけ数え解決数に天井がある(t *testing.T) {
	repo := &mockKBPermissionRepo{}
	dup := "00000000-0000-7000-8000-000000000001"
	repo.On("ListWorkspacePageViewFactsByIDs", mock.Anything, kbRefWS, uint64(7), []string{dup}).
		Return([]repository.PageWithViewFacts{kbViewableFacts(dup, "本題")}, nil)
	uc := usecase.NewResolvePageRefTitlesUseCase(repo)

	got := uc.Execute(context.Background(), usecase.ResolvePageRefTitlesInput{
		WorkspaceID: kbRefWS, UserID: 7, Doc: kbRefDoc(dup, dup),
	})

	// 同じ ID は 1 回で問い合わせ、両方の出現が書き換わる。
	assert.Contains(t, got, `"本題"`)
	assert.NotContains(t, got, `"無題"`)
	repo.AssertNumberOfCalls(t, "ListWorkspacePageViewFactsByIDs", 1)

	// 天井（100）を超えた分は解決しない: 101 個の参照 → 問い合わせは先頭 100 個。
	repo2 := &mockKBPermissionRepo{}
	ids := make([]string, 101)
	for i := range ids {
		ids[i] = fmt.Sprintf("00000000-0000-7000-8000-%012d", i+1)
	}
	repo2.On("ListWorkspacePageViewFactsByIDs", mock.Anything, kbRefWS, uint64(7),
		mock.MatchedBy(func(got []string) bool { return len(got) == 100 })).
		Return([]repository.PageWithViewFacts{}, nil)
	uc2 := usecase.NewResolvePageRefTitlesUseCase(repo2)
	uc2.Execute(context.Background(), usecase.ResolvePageRefTitlesInput{
		WorkspaceID: kbRefWS, UserID: 7, Doc: kbRefDoc(ids...),
	})
	repo2.AssertExpectations(t)
}
