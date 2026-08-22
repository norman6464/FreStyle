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
	cases := []struct {
		name     string
		isPublic bool
		viewer   uint64
		want     bool
	}{
		{"公開×所有者", true, owner, true},
		{"公開×非所有者", true, 99, true},
		{"公開×未認証(viewerID=0)", true, 0, true},
		{"非公開×所有者", false, owner, true},
		{"非公開×非所有者", false, 99, false},
		{"非公開×未認証(viewerID=0)", false, 0, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			d := &RichDocument{OwnerID: owner, IsPublic: tc.isPublic}
			if got := d.CanBeReadBy(tc.viewer); got != tc.want {
				t.Fatalf("CanBeReadBy(%d) = %v, want %v", tc.viewer, got, tc.want)
			}
		})
	}
}
