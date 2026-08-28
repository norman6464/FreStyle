package usecase_test

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

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
	got, err := uc.Execute(context.Background(), usecase.ResolvePageRefTitlesInput{
		WorkspaceID: kbRefWS, UserID: 7, Doc: kbRefDoc(visible, hidden),
	})
	assert.NoError(t, err)

	var doc map[string]any
	assert.NoError(t, json.Unmarshal([]byte(got), &doc))
	inline := doc["content"].([]any)[0].(map[string]any)["content"].([]any)
	first := inline[0].(map[string]any)["attrs"].(map[string]any)
	second := inline[1].(map[string]any)["attrs"].(map[string]any)
	assert.Equal(t, "設計メモ v2", first["title"], "閲覧できる参照は現在の題名になる")
	// deny のある参照には題名を入れない。保存されていた title も読み出し時に剥がす
	// （剥がす前に保存された doc から、権限を失った読み手へ古い題名が返らないように）。
	assert.Nil(t, second["title"])
	repo.AssertExpectations(t)
}

func Test_ページ参照の題名解決_参照が無ければ問い合わせず原文のまま(t *testing.T) {
	repo := &mockKBPermissionRepo{}
	uc := usecase.NewResolvePageRefTitlesUseCase(repo)
	doc := `{"type":"doc","content":[{"type":"paragraph","content":[{"type":"text","text":"参照なし"}]}]}`

	got, err := uc.Execute(context.Background(), usecase.ResolvePageRefTitlesInput{
		WorkspaceID: kbRefWS, UserID: 7, Doc: doc,
	})

	assert.NoError(t, err)
	assert.Equal(t, doc, got)
	repo.AssertNotCalled(t, "ListWorkspacePageViewFactsByIDs")
}

func Test_ページ参照の題名解決_壊れたdocや取得失敗では原文を返す(t *testing.T) {
	repo := &mockKBPermissionRepo{}
	uc := usecase.NewResolvePageRefTitlesUseCase(repo)

	broken := `{"type":"doc","content":[`
	got, err := uc.Execute(context.Background(), usecase.ResolvePageRefTitlesInput{
		WorkspaceID: kbRefWS, UserID: 7, Doc: broken,
	})
	assert.NoError(t, err, "読めない doc はエラーではない（保存経路の検証が弾く領分）")
	assert.Equal(t, broken, got, "読めない doc はそのまま返す（本文が開けることが題名より重い）")

	id := "00000000-0000-7000-8000-000000000001"
	repo.On("ListWorkspacePageViewFactsByIDs", mock.Anything, kbRefWS, uint64(7), []string{id}).
		Return(nil, assert.AnError)
	got2, err2 := uc.Execute(context.Background(), usecase.ResolvePageRefTitlesInput{
		WorkspaceID: kbRefWS, UserID: 7, Doc: kbRefDoc(id),
	})
	assert.ErrorIs(t, err2, assert.AnError, "事実の取得失敗は握り潰さず返す（気づけるように）")
	// 失敗でも本文は返すが、保存されていた title は剥がして返す（可視判定を通っていない
	// 題名を、失敗経路からも出さない）。参照そのものは残る。
	assert.Contains(t, got2, id)
	assert.NotContains(t, got2, `"無題"`)
}

func Test_ページ参照の題名解決_同じ参照は1回だけ数える(t *testing.T) {
	repo := &mockKBPermissionRepo{}
	dup := "00000000-0000-7000-8000-000000000001"
	repo.On("ListWorkspacePageViewFactsByIDs", mock.Anything, kbRefWS, uint64(7), []string{dup}).
		Return([]repository.PageWithViewFacts{kbViewableFacts(dup, "本題")}, nil)
	uc := usecase.NewResolvePageRefTitlesUseCase(repo)

	got, err := uc.Execute(context.Background(), usecase.ResolvePageRefTitlesInput{
		WorkspaceID: kbRefWS, UserID: 7, Doc: kbRefDoc(dup, dup),
	})
	assert.NoError(t, err)

	// 同じ ID は 1 回で問い合わせ、両方の出現が書き換わる。
	assert.Contains(t, got, `"本題"`)
	assert.NotContains(t, got, `"無題"`)
	repo.AssertNumberOfCalls(t, "ListWorkspacePageViewFactsByIDs", 1)
}

func Test_ページ参照の題名解決_解決数の天井は文書順の先頭100件(t *testing.T) {
	// 「長さが 100」だけでは、末尾や任意の 100 件を選ぶ実装でも通ってしまう。
	// 契約は**文書順の先頭 100 件**なので、選ばれた ID の並びまで固定する。
	repo := &mockKBPermissionRepo{}
	ids := make([]string, 101)
	for i := range ids {
		ids[i] = fmt.Sprintf("00000000-0000-7000-8000-%012d", i+1)
	}
	repo.On("ListWorkspacePageViewFactsByIDs", mock.Anything, kbRefWS, uint64(7), ids[:100]).
		Return([]repository.PageWithViewFacts{}, nil)
	uc := usecase.NewResolvePageRefTitlesUseCase(repo)

	_, err := uc.Execute(context.Background(), usecase.ResolvePageRefTitlesInput{
		WorkspaceID: kbRefWS, UserID: 7, Doc: kbRefDoc(ids...),
	})

	assert.NoError(t, err)
	repo.AssertExpectations(t)
}

func Test_ページ参照の題名解決_pageIdの表記ゆれは正規形へ寄せて照合する(t *testing.T) {
	// repository は ID を小文字・ハイフン区切りへ正規化して返す。保存された参照が
	// 大文字で書かれていても、突き合わせが外れて題名だけ差し替わらない、を防ぐ。
	repo := &mockKBPermissionRepo{}
	canonical := "00000000-0000-7000-8000-0000000000ab"
	repo.On("ListWorkspacePageViewFactsByIDs", mock.Anything, kbRefWS, uint64(7), []string{canonical}).
		Return([]repository.PageWithViewFacts{kbViewableFacts(canonical, "正規形の題名")}, nil)
	uc := usecase.NewResolvePageRefTitlesUseCase(repo)

	upper := "00000000-0000-7000-8000-0000000000AB"
	got, err := uc.Execute(context.Background(), usecase.ResolvePageRefTitlesInput{
		WorkspaceID: kbRefWS, UserID: 7, Doc: kbRefDoc(upper),
	})

	assert.NoError(t, err)
	assert.Contains(t, got, "正規形の題名")
	repo.AssertExpectations(t)
}

func Test_ページ参照の題名は保存時に剥がされる(t *testing.T) {
	// title は読み手ごとの派生値。保存されると、閲覧できる編集者の画面で解決された
	// 現在の題名が本文へ焼き込まれ、閲覧できない読み手にも返ってしまう。
	id := "00000000-0000-7000-8000-000000000001"
	doc := kbRefDoc(id)

	got := usecase.StripPageRefTitles(doc)

	assert.NotContains(t, got, `"無題"`)
	assert.Contains(t, got, id, "参照そのもの（pageId）は残る")

	// 参照が無い doc は触らない（同じ文字列のまま）。
	plain := `{"type":"doc","content":[{"type":"paragraph","content":[{"type":"text","text":"本文"}]}]}`
	assert.Equal(t, plain, usecase.StripPageRefTitles(plain))
}

func Test_ページ参照の題名解決_アーカイブ済みの参照は題名に採らない(t *testing.T) {
	// 隠したページの現在の題名を本文へ映さない（検索が現役だけを対象にするのと同じ線引き）。
	repo := &mockKBPermissionRepo{}
	id := "00000000-0000-7000-8000-000000000001"
	archivedAt := time.Now()
	archived := repository.PageWithViewFacts{
		Page:  domain.Page{ID: id, Title: "隠した題名", ArchivedAt: &archivedAt},
		Facts: domain.PageViewFacts{Role: rolePtr(domain.GrantRoleViewer)},
	}
	repo.On("ListWorkspacePageViewFactsByIDs", mock.Anything, kbRefWS, uint64(7), []string{id}).
		Return([]repository.PageWithViewFacts{archived}, nil)
	uc := usecase.NewResolvePageRefTitlesUseCase(repo)

	got, err := uc.Execute(context.Background(), usecase.ResolvePageRefTitlesInput{
		WorkspaceID: kbRefWS, UserID: 7, Doc: kbRefDoc(id),
	})

	assert.NoError(t, err)
	assert.NotContains(t, got, "隠した題名", "現在の題名は入れない")
	// 原文をそのまま返しても通る検証にしない: 保存されていた「無題」も剥がれている
	//（＝読み出し側の strip と解決の両方が実際に走った）ことまで確かめる。
	assert.NotContains(t, got, `"無題"`)
	repo.AssertExpectations(t)
}

func Test_パンくず_閲覧できる祖先だけがclosureの順で返る(t *testing.T) {
	pages := &mockKnowledgeBaseRepo{}
	perms := &mockKBPermissionRepo{}
	// closure の順は「ん」→「あ」（根が「ん」）。facts は題名順（「あ」が先）で返す —
	// 題名順をそのまま使う退行をこの並びで捕まえる。
	root := "00000000-0000-7000-8000-00000000000a"
	child := "00000000-0000-7000-8000-00000000000b"
	hidden := "00000000-0000-7000-8000-00000000000c"
	pages.On("ListAncestorPageIDs", mock.Anything, kbRefWS, "page-x").
		Return([]string{root, child, hidden}, nil)
	perms.On("ListWorkspacePageViewFactsByIDs", mock.Anything, kbRefWS, uint64(7),
		[]string{root, child, hidden}).
		Return([]repository.PageWithViewFacts{
			// 題名順: 「あ」(child) が先、「ん」(root) が後。hidden は deny の事実つき。
			kbViewableFacts(child, "あ"),
			kbViewableFacts(root, "ん"),
			{
				Page: domain.Page{ID: hidden, Title: "見えない段"},
				Facts: domain.PageViewFacts{
					Role: rolePtr(domain.GrantRoleViewer),
					View: &domain.RestrictionFacts{DeniedAnywhere: true},
				},
			},
		}, nil)
	uc := usecase.NewListViewableAncestorsUseCase(pages, perms)

	got, err := uc.Execute(context.Background(), usecase.ListViewableAncestorsInput{
		WorkspaceID: kbRefWS, UserID: 7, PageID: "page-x",
	})

	assert.NoError(t, err)
	// 並びは closure（根から）。facts の題名順に引きずられない。
	// 見えない段は行ごと消える（題名どころか実在も知らせない）。
	assert.Equal(t, []usecase.AncestorRef{
		{ID: root, Title: "ん"},
		{ID: child, Title: "あ"},
	}, got)
}

func Test_パンくず_アーカイブ済みの祖先も閲覧できる限り含める(t *testing.T) {
	// アーカイブ済みのページは /p で開ける。経路から抜くと「その段が無い」かのように
	// 場所を偽るので、可視である限り出す（題名解決と除外の線引きが違う）。
	pages := &mockKnowledgeBaseRepo{}
	perms := &mockKBPermissionRepo{}
	arch := "00000000-0000-7000-8000-00000000000d"
	archivedAt := time.Now()
	pages.On("ListAncestorPageIDs", mock.Anything, kbRefWS, "page-y").
		Return([]string{arch}, nil)
	perms.On("ListWorkspacePageViewFactsByIDs", mock.Anything, kbRefWS, uint64(7), []string{arch}).
		Return([]repository.PageWithViewFacts{{
			Page:  domain.Page{ID: arch, Title: "片付けた親", ArchivedAt: &archivedAt},
			Facts: domain.PageViewFacts{Role: rolePtr(domain.GrantRoleViewer)},
		}}, nil)
	uc := usecase.NewListViewableAncestorsUseCase(pages, perms)

	got, err := uc.Execute(context.Background(), usecase.ListViewableAncestorsInput{
		WorkspaceID: kbRefWS, UserID: 7, PageID: "page-y",
	})

	assert.NoError(t, err)
	assert.Equal(t, []usecase.AncestorRef{{ID: arch, Title: "片付けた親"}}, got)
}

func Test_パンくず_祖先が無ければ空のsliceを返す(t *testing.T) {
	// nil を返すと JSON で null になり、フロントの ancestors.map が落ちる。
	pages := &mockKnowledgeBaseRepo{}
	perms := &mockKBPermissionRepo{}
	pages.On("ListAncestorPageIDs", mock.Anything, kbRefWS, "root-page").
		Return([]string{}, nil)
	uc := usecase.NewListViewableAncestorsUseCase(pages, perms)

	got, err := uc.Execute(context.Background(), usecase.ListViewableAncestorsInput{
		WorkspaceID: kbRefWS, UserID: 7, PageID: "root-page",
	})

	assert.NoError(t, err)
	assert.NotNil(t, got)
	assert.Empty(t, got)
	perms.AssertNotCalled(t, "ListWorkspacePageViewFactsByIDs")
}
