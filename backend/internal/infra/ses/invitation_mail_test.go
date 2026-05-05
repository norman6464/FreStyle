package ses

import (
	"strings"
	"testing"
)

func TestBuildInvitationMail_IncludesMagicLinkAndDisplayName(t *testing.T) {
	link := "https://normanblog.com/invitations/accept?token=abc-123"
	subject, htmlBody, textBody := BuildInvitationMail(link, "山田太郎", "株式会社FreStyle", "company_admin")

	if subject == "" {
		t.Fatal("subject must not be empty")
	}
	if !strings.Contains(textBody, link) {
		t.Errorf("textBody must contain magic link, got: %s", textBody)
	}
	if !strings.Contains(htmlBody, link) {
		t.Errorf("htmlBody must contain magic link")
	}
	if !strings.Contains(textBody, "山田太郎") {
		t.Errorf("textBody must contain displayName")
	}
	if !strings.Contains(htmlBody, "株式会社FreStyle") {
		t.Errorf("htmlBody must contain companyName")
	}
}

// 招待者がメールアドレス・displayName・会社名にスクリプトを混入できる箇所がないことを確認する。
// HTML body には displayName / companyName / role がそのまま埋め込まれるため、< / > が
// エスケープされ、HTML タグとして解釈されないことを検証する。
// （`onerror=alert(1)` のような文字列自体は残るが、タグが閉じていなければ XSS は成立しない）。
func TestBuildInvitationMail_EscapesHTMLInjection(t *testing.T) {
	_, htmlBody, _ := BuildInvitationMail(
		"https://example.com/x?token=t",
		`<script>alert(1)</script>`,
		`<img src=x onerror=alert(1)>`,
		`trainee`,
	)
	if strings.Contains(htmlBody, "<script>alert(1)</script>") {
		t.Errorf("htmlBody must escape <script> tags as plain text")
	}
	if strings.Contains(htmlBody, "<img src=x") {
		t.Errorf("htmlBody must escape <img> tags as plain text")
	}
	// 安全側で `&lt;` / `&gt;` への置換が行われていることを確認
	if !strings.Contains(htmlBody, "&lt;script&gt;") {
		t.Errorf("htmlBody must contain HTML-escaped form of <script>, got: %s", htmlBody)
	}
}

func TestBuildInvitationMail_OmitsScopeWhenEmpty(t *testing.T) {
	_, htmlBody, textBody := BuildInvitationMail("https://x", "", "", "")
	if strings.Contains(htmlBody, "招待元の会社") || strings.Contains(textBody, "招待元の会社") {
		t.Errorf("scope sections should be omitted when companyName/role are empty")
	}
}

func TestMagicLinkURL_TrimsTrailingSlash(t *testing.T) {
	got := MagicLinkURL("https://normanblog.com/", "abc")
	want := "https://normanblog.com/invitations/accept?token=abc"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestMagicLinkURL_EncodesToken(t *testing.T) {
	// UUID v4 はエンコード不要だが、万一 token にメタ文字が含まれた場合の安全性を確認。
	got := MagicLinkURL("https://example.com", "abc def&x=1")
	if !strings.Contains(got, "token=abc+def%26x%3D1") {
		t.Errorf("token must be query-escaped, got: %s", got)
	}
}
