// Package smtp は SES を使わない環境(staging)向けのメール送信実装。
// 宛先は box 上のメールキャッチャー(同一 Docker ネットワーク内の SMTP サーバー)で、
// 認証・TLS なしの内部ネットワーク前提。実メールは外部へ送信されない。
package smtp

import (
	"context"
	"fmt"
	"mime"
	"net/mail"
	netsmtp "net/smtp"
	"strings"
)

// Sender は usecase.MagicLinkSender を SMTP で満たす送信実装。
type Sender struct {
	addr string // host:port
	from string // 送信元アドレス(ベアアドレス。エンベロープと From ヘッダの両方に使う)
}

// NewSender は host:port のメールキャッチャーへ from 名義で送る Sender を返す。
func NewSender(host, port, from string) *Sender {
	return &Sender{addr: host + ":" + port, from: from}
}

// SendInvitationEmail は multipart/alternative(text + html)の 1 通を SMTP で送信する。
func (s *Sender) SendInvitationEmail(_ context.Context, to, subject, htmlBody, textBody string) error {
	// ヘッダインジェクション対策: to はリクエスト由来のため、アドレスとして検証してから
	// パース結果のアドレス部のみをヘッダとエンベロープに使う(改行や追加ヘッダを排除)。
	parsed, err := mail.ParseAddress(to)
	if err != nil {
		return fmt.Errorf("smtp: invalid to address %q: %w", to, err)
	}
	to = parsed.Address
	// 件名も改行を除去してからエンコードする(本文は DATA 部のヘッダ外なのでそのままでよい)。
	subject = strings.NewReplacer("\r", " ", "\n", " ").Replace(subject)

	const boundary = "frestyle-invitation-boundary"
	var b strings.Builder
	b.WriteString("From: " + s.from + "\r\n")
	b.WriteString("To: " + to + "\r\n")
	// 日本語件名は RFC 2047 の Q エンコードで包む。
	b.WriteString("Subject: " + mime.QEncoding.Encode("utf-8", subject) + "\r\n")
	b.WriteString("MIME-Version: 1.0\r\n")
	b.WriteString("Content-Type: multipart/alternative; boundary=" + boundary + "\r\n")
	b.WriteString("\r\n")
	for _, part := range []struct{ contentType, body string }{
		{"text/plain; charset=utf-8", textBody},
		{"text/html; charset=utf-8", htmlBody},
	} {
		b.WriteString("--" + boundary + "\r\n")
		b.WriteString("Content-Type: " + part.contentType + "\r\n\r\n")
		b.WriteString(part.body + "\r\n")
	}
	b.WriteString("--" + boundary + "--\r\n")

	if err := netsmtp.SendMail(s.addr, nil, s.from, []string{to}, []byte(b.String())); err != nil {
		return fmt.Errorf("smtp send to %s via %s: %w", to, s.addr, err)
	}
	return nil
}
