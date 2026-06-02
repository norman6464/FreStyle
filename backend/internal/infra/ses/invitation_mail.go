package ses

import (
	"fmt"
	"html"
	"net/url"
	"strings"
)

// BuildInvitationMail は招待メールの subject / HTML / text を返す。
// HTML 非対応クライアント向けに text 本文も用意する。role は本文には出さない（互換のため受け取るだけ）。
// displayName / companyName が空なら該当行を省略する。
func BuildInvitationMail(magicLink, displayName, companyName, _ string) (subject, htmlBody, textBody string) {
	subject = "FreStyle へようこそ — 管理者からの招待"

	greeting := "FreStyle にご招待いただきありがとうございます。"
	if displayName != "" {
		greeting = fmt.Sprintf("%s さん、FreStyle にご招待いただきありがとうございます。", displayName)
	}

	scopeText := ""
	if companyName != "" {
		scopeText = fmt.Sprintf("\n\n- 招待元の会社: %s", companyName)
	}

	textBody = fmt.Sprintf(`%s

下記のリンクをクリックして、ログイン手続きを完了してください。

%s%s

このリンクの有効期限は 7 日間です。
心当たりがない場合はこのメールを破棄してください。

— FreStyle 運営事務局
`, greeting, magicLink, scopeText)

	htmlBody = buildInvitationHTML(magicLink, greeting, companyName)
	return subject, htmlBody, textBody
}

// buildInvitationHTML は HTML 本文を組み立てる。埋め込む値は html.EscapeString で
// エスケープして HTML インジェクションを防ぐ。
func buildInvitationHTML(magicLink, greeting, companyName string) string {
	var sb strings.Builder
	sb.WriteString(`<!DOCTYPE html><html><body style="font-family: -apple-system, BlinkMacSystemFont, 'Helvetica Neue', sans-serif; color: #1f2937; line-height: 1.7;">`)
	sb.WriteString(`<div style="max-width: 560px; margin: 0 auto; padding: 24px;">`)
	sb.WriteString(`<h2 style="color: #111827;">FreStyle へようこそ</h2>`)
	sb.WriteString(`<p>` + html.EscapeString(greeting) + `</p>`)

	if companyName != "" {
		sb.WriteString(`<ul style="background: #f9fafb; padding: 16px 24px; border-radius: 8px;">`)
		sb.WriteString(`<li>招待元の会社: <strong>` + html.EscapeString(companyName) + `</strong></li>`)
		sb.WriteString(`</ul>`)
	}

	sb.WriteString(`<p style="margin-top: 24px;">下記のボタンからログイン手続きにお進みください。</p>`)
	sb.WriteString(`<p style="margin: 24px 0;"><a href="` + html.EscapeString(magicLink) + `" style="background: #4f46e5; color: #ffffff; padding: 12px 24px; border-radius: 6px; text-decoration: none; display: inline-block;">ログイン手続きへ進む</a></p>`)
	sb.WriteString(`<p style="font-size: 13px; color: #6b7280;">ボタンが押せない場合は、下記の URL をブラウザに貼り付けてください。<br>`)
	sb.WriteString(`<span style="word-break: break-all;">` + html.EscapeString(magicLink) + `</span></p>`)
	sb.WriteString(`<hr style="border: none; border-top: 1px solid #e5e7eb; margin: 24px 0;">`)
	sb.WriteString(`<p style="font-size: 12px; color: #6b7280;">このリンクの有効期限は 7 日間です。<br>心当たりがない場合は、このメールを破棄してください。</p>`)
	sb.WriteString(`<p style="font-size: 12px; color: #6b7280;">— FreStyle 運営事務局</p>`)
	sb.WriteString(`</div></body></html>`)
	return sb.String()
}

// MagicLinkURL は token からフロント受諾画面の絶対 URL を組み立てる（baseURL は APP_BASE_URL）。
func MagicLinkURL(baseURL, token string) string {
	base := strings.TrimRight(baseURL, "/")
	return base + "/invitations/accept?token=" + url.QueryEscape(token)
}
