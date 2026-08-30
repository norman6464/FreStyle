package domain

import "testing"

func Test_DocumentKind_Valid(t *testing.T) {
	cases := []struct {
		name string
		k    DocumentKind
		want bool
	}{
		{"note", DocumentKindNote, true},
		{"course-chapter", DocumentKindCourseChapter, true},
		{"空文字", DocumentKind(""), false},
		{"未知kind", DocumentKind("weird"), false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.k.Valid(); got != tc.want {
				t.Fatalf("Valid() = %v, want %v", got, tc.want)
			}
		})
	}
}

func Test_RichDocument_CanBeReadBy(t *testing.T) {
	const owner uint64 = 7
	wsA := "0198a000-0000-7000-8000-0000000000c1"
	wsB := "0198a000-0000-7000-8000-0000000000c2"
	cases := []struct {
		name            string
		isPublic        bool
		docWorkspace    *string
		viewer          uint64
		viewerWorkspace WorkspaceRef
		want            bool
	}{
		{"公開×所有者×同一ワークスペース", true, &wsA, owner, WorkspaceRefOf(wsA), true},
		{"公開×所有者×文書は別ワークスペース(移管後の古い値)", true, &wsB, owner, WorkspaceRefOf(wsA), true},
		{"公開×所有者×ワークスペース不明(NULL)", true, nil, owner, WorkspaceRefOf(wsA), true},
		{"公開×所有者×閲覧者は未所属", true, nil, owner, NoWorkspace(), true},
		{"公開×非所有者×同一ワークスペース", true, &wsA, 99, WorkspaceRefOf(wsA), true},
		{"公開×非所有者×別ワークスペース", true, &wsA, 99, WorkspaceRefOf(wsB), false},
		{"公開×非所有者×ワークスペース不明(NULL)", true, nil, 99, WorkspaceRefOf(wsA), false},
		{"公開×非所有者×閲覧者は未所属", true, &wsA, 99, NoWorkspace(), false},
		{"公開×未認証(viewerID=0)", true, &wsA, 0, NoWorkspace(), false},
		{"非公開×所有者", false, &wsA, owner, WorkspaceRefOf(wsA), true},
		{"非公開×非所有者×同一ワークスペース", false, &wsA, 99, WorkspaceRefOf(wsA), false},
		{"非公開×未認証(viewerID=0)", false, &wsA, 0, NoWorkspace(), false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			d := &RichDocument{OwnerID: owner, IsPublic: tc.isPublic, WorkspaceID: tc.docWorkspace}
			if got := d.CanBeReadBy(tc.viewer, tc.viewerWorkspace); got != tc.want {
				t.Fatalf("CanBeReadBy(%d, %+v) = %v, want %v", tc.viewer, tc.viewerWorkspace, got, tc.want)
			}
		})
	}
}

func Test_RichDocument_CompanyRef(t *testing.T) {
	if _, affiliated := (&RichDocument{}).CompanyRef().CompanyID(); affiliated {
		t.Fatal("company_id が NULL の文書は未所属を返すべき")
	}
	id := uint64(3)
	got, affiliated := (&RichDocument{CompanyID: &id}).CompanyRef().CompanyID()
	if !affiliated || got != id {
		t.Fatalf("CompanyRef() = (%d, %v), want (3, true)", got, affiliated)
	}
}

func Test_RichDocument_WorkspaceRef(t *testing.T) {
	if _, affiliated := (&RichDocument{}).WorkspaceRef().WorkspaceID(); affiliated {
		t.Fatal("workspace_id が NULL の文書は未所属を返すべき")
	}
	wid := "0198a000-0000-7000-8000-0000000000c1"
	got, affiliated := (&RichDocument{WorkspaceID: &wid}).WorkspaceRef().WorkspaceID()
	if !affiliated || got != wid {
		t.Fatalf("WorkspaceRef() = (%q, %v), want (%q, true)", got, affiliated, wid)
	}
}
