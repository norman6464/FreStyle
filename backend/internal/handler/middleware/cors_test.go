package middleware

import "testing"

func TestIsAllowedOrigin(t *testing.T) {
	cases := []struct {
		origin string
		want   bool
	}{
		{"https://normanblog.com", true},
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
