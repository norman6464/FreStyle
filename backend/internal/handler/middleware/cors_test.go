package middleware

import "testing"

func Test_許可オリジン判定(t *testing.T) {
	cases := []struct {
		origin string
		want   bool
	}{
		{"https://frestyle.jp", true},
		// 旧ドメインは撤去済みのため許可しない(FRESTYLE-226)。http/https 双方を確認する。
		{"https://normanblog.com", false},
		{"http://normanblog.com", false},
		{"http://localhost:5173", true},
		{"https://evil.example.com", false},
		{"", false},
	}
	for _, tc := range cases {
		t.Run(tc.origin, func(t *testing.T) {
			if got := IsAllowedOrigin(tc.origin); got != tc.want {
				t.Fatalf("IsAllowedOrigin(%q) = %v, want %v", tc.origin, got, tc.want)
			}
		})
	}
}
