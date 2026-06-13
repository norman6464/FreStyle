package ses

import (
	"strings"
	"testing"
)

func Test_招待メール構築_マジックリンクと表示名を含む(t *testing.T) {
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
func Test_招待メール構築_HTMLインジェクションをエスケープ(t *testing.T) {
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

func Test_招待メール構築_scope空は省略(t *testing.T) {
	_, htmlBody, textBody := BuildInvitationMail("https://x", "", "", "")
	if strings.Contains(htmlBody, "招待元の会社") || strings.Contains(textBody, "招待元の会社") {
		t.Errorf("scope sections should be omitted when companyName is empty")
	}
}

// role の生値（company_admin / trainee 等）はメール本文に出さない。
// 一般受信者には社内ロール体系の名称が伝わらないため、本文では会社名と表示名のみ示し、
// 権限はログイン後の機能体験で暗黙的に理解してもらう設計。
func Test_招待メール構築_生のroleを露出しない(t *testing.T) {
	for _, role := range []string{"company_admin", "trainee", "super_admin"} {
		_, htmlBody, textBody := BuildInvitationMail("https://x", "", "ACME", role)
		if strings.Contains(htmlBody, role) {
			t.Errorf("htmlBody must not contain raw role %q, got: %s", role, htmlBody)
		}
		if strings.Contains(textBody, role) {
			t.Errorf("textBody must not contain raw role %q, got: %s", role, textBody)
		}
		if strings.Contains(htmlBody, "付与される権限") || strings.Contains(textBody, "付与される権限") {
			t.Errorf("body must not contain 付与される権限 label (role=%s)", role)
		}
	}
}

func Test_マジックリンクURL_末尾スラッシュを除去(t *testing.T) {
	got := MagicLinkURL("https://normanblog.com/", "abc")
	want := "https://normanblog.com/invitations/accept?token=abc"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func Test_マジックリンクURL_tokenをエンコード(t *testing.T) {
	// UUID v4 はエンコード不要だが、万一 token にメタ文字が含まれた場合の安全性を確認。
	got := MagicLinkURL("https://example.com", "abc def&x=1")
	if !strings.Contains(got, "token=abc+def%26x%3D1") {
		t.Errorf("token must be query-escaped, got: %s", got)
	}
}
