package ses

import (
	"fmt"
	"html"
	"net/url"
	"strings"
)

// BuildInvitationMail は招待メールの subject / HTML / text を返す。
// magicLink には ?token=<UUID> 付きの URL をそのまま渡す（呼び出し側で組み立て）。
//
// HTML / text 双方を組み立てる理由:
//   - SES の SendEmail で両方提供しておくと、HTML を許容しないクライアントでもテキスト本文が読める
//   - 受信側のフィッシング誤判定を避けるため、URL は本文中に普通の文字列として明示的に提示する
//
// displayName / companyName / role が空の場合は本文から該当行を省略する（過度に固定文言にしない）。
func BuildInvitationMail(magicLink, displayName, companyName, role string) (subject, htmlBody, textBody string) {
	subject = "FreStyle へようこそ — 管理者からの招待"

	greeting := "FreStyle にご招待いただきありがとうございます。"
	if displayName != "" {
		greeting = fmt.Sprintf("%s さん、FreStyle にご招待いただきありがとうございます。", displayName)
	}

	scopeLines := []string{}
	if companyName != "" {
		scopeLines = append(scopeLines, fmt.Sprintf("- 招待元の会社: %s", companyName))
	}
	if role != "" {
		scopeLines = append(scopeLines, fmt.Sprintf("- 付与される権限: %s", role))
	}
	scopeText := ""
	if len(scopeLines) > 0 {
		scopeText = "\n\n" + strings.Join(scopeLines, "\n")
	}

	textBody = fmt.Sprintf(`%s

下記のリンクをクリックして、ログイン手続きを完了してください。

%s%s

このリンクの有効期限は 7 日間です。
心当たりがない場合はこのメールを破棄してください。

— FreStyle 運営事務局
`, greeting, magicLink, scopeText)

	htmlBody = buildInvitationHTML(magicLink, greeting, displayName, companyName, role)
	return subject, htmlBody, textBody
}

// buildInvitationHTML は HTML 本文を組み立てる。受信者が入力した値はすべて html.EscapeString で
// エスケープしてから埋め込み、HTML インジェクションを防ぐ。リンク URL は url.QueryEscape ではなく
// HTML 属性向けに html.EscapeString で十分（URL 自体の組み立てはこの関数の責務外）。
func buildInvitationHTML(magicLink, greeting, displayName, companyName, role string) string {
	var sb strings.Builder
	sb.WriteString(`<!DOCTYPE html><html><body style="font-family: -apple-system, BlinkMacSystemFont, 'Helvetica Neue', sans-serif; color: #1f2937; line-height: 1.7;">`)
	sb.WriteString(`<div style="max-width: 560px; margin: 0 auto; padding: 24px;">`)
	sb.WriteString(`<h2 style="color: #111827;">FreStyle へようこそ</h2>`)
	sb.WriteString(`<p>` + html.EscapeString(greeting) + `</p>`)

	if companyName != "" || role != "" {
		sb.WriteString(`<ul style="background: #f9fafb; padding: 16px 24px; border-radius: 8px;">`)
		if companyName != "" {
			sb.WriteString(`<li>招待元の会社: <strong>` + html.EscapeString(companyName) + `</strong></li>`)
		}
		if role != "" {
			sb.WriteString(`<li>付与される権限: <strong>` + html.EscapeString(role) + `</strong></li>`)
		}
		sb.WriteString(`</ul>`)
	}
	_ = displayName // displayName は greeting で既に使用済

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

// MagicLinkURL は invitations の token から「フロント受諾画面」の絶対 URL を組み立てる。
// baseURL は env APP_BASE_URL（例: https://normanblog.com）を渡す想定。
// token は UUID なので URL エンコード不要だが、安全のため QueryEscape を通す。
func MagicLinkURL(baseURL, token string) string {
	base := strings.TrimRight(baseURL, "/")
	return base + "/invitations/accept?token=" + url.QueryEscape(token)
}
