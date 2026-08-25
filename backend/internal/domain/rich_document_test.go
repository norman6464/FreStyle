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
	companyA := uint64(1)
	companyB := uint64(2)
	cases := []struct {
		name          string
		isPublic      bool
		docCompany    *uint64
		viewer        uint64
		viewerCompany CompanyRef
		want          bool
	}{
		{"公開×所有者×同一会社", true, &companyA, owner, CompanyRefOf(companyA), true},
		{"公開×所有者×文書は別会社(移管後の古い値)", true, &companyB, owner, CompanyRefOf(companyA), true},
		{"公開×所有者×会社不明(NULL)", true, nil, owner, CompanyRefOf(companyA), true},
		{"公開×所有者×閲覧者は未所属", true, nil, owner, NoCompany(), true},
		{"公開×非所有者×同一会社", true, &companyA, 99, CompanyRefOf(companyA), true},
		{"公開×非所有者×別会社", true, &companyA, 99, CompanyRefOf(companyB), false},
		{"公開×非所有者×会社不明(NULL)", true, nil, 99, CompanyRefOf(companyA), false},
		{"公開×非所有者×閲覧者は未所属", true, &companyA, 99, NoCompany(), false},
		{"公開×未認証(viewerID=0)", true, &companyA, 0, NoCompany(), false},
		{"非公開×所有者", false, &companyA, owner, CompanyRefOf(companyA), true},
		{"非公開×非所有者×同一会社", false, &companyA, 99, CompanyRefOf(companyA), false},
		{"非公開×未認証(viewerID=0)", false, &companyA, 0, NoCompany(), false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			d := &RichDocument{OwnerID: owner, IsPublic: tc.isPublic, CompanyID: tc.docCompany}
			if got := d.CanBeReadBy(tc.viewer, tc.viewerCompany); got != tc.want {
				t.Fatalf("CanBeReadBy(%d, %+v) = %v, want %v", tc.viewer, tc.viewerCompany, got, tc.want)
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
