package smtp

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"strings"
	"testing"
	"time"
)

// fakeSMTPServer は 1 接続だけ受けて DATA の中身を received へ流す最小 SMTP サーバー。
func fakeSMTPServer(t *testing.T) (addr string, received <-chan string) {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	t.Cleanup(func() { _ = ln.Close() })

	ch := make(chan string, 1)
	go func() {
		conn, err := ln.Accept()
		if err != nil {
			return
		}
		defer conn.Close()
		br := bufio.NewReader(conn)
		fmt.Fprintf(conn, "220 fake ready\r\n")
		var data strings.Builder
		inData := false
		for {
			line, err := br.ReadString('\n')
			if err != nil {
				return
			}
			if inData {
				if strings.TrimRight(line, "\r\n") == "." {
					inData = false
					fmt.Fprintf(conn, "250 ok\r\n")
					ch <- data.String()
					continue
				}
				data.WriteString(line)
				continue
			}
			switch cmd := strings.ToUpper(strings.TrimSpace(line)); {
			case strings.HasPrefix(cmd, "DATA"):
				inData = true
				fmt.Fprintf(conn, "354 go ahead\r\n")
			case strings.HasPrefix(cmd, "QUIT"):
				fmt.Fprintf(conn, "221 bye\r\n")
				return
			default: // EHLO / HELO / MAIL FROM / RCPT TO
				fmt.Fprintf(conn, "250 ok\r\n")
			}
		}
	}()
	return ln.Addr().String(), ch
}

func Test_SMTP送信_宛先と件名と両形式の本文が届く(t *testing.T) {
	addr, received := fakeSMTPServer(t)
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		t.Fatalf("split addr: %v", err)
	}

	s := NewSender(host, port, "noreply@staging.example.jp")
	if err := s.SendInvitationEmail(
		context.Background(),
		"invitee@example.com", "FreStyle への招待", "<p>magic-link</p>", "magic-link-text",
	); err != nil {
		t.Fatalf("send: %v", err)
	}

	var got string
	select {
	case got = <-received:
	case <-time.After(3 * time.Second):
		t.Fatal("fake サーバーに DATA が届かなかった")
	}

	for _, want := range []string{
		"From: noreply@staging.example.jp",
		"To: invitee@example.com",
		"multipart/alternative",
		"text/plain; charset=utf-8",
		"text/html; charset=utf-8",
		"<p>magic-link</p>",
		"magic-link-text",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("送信データに %q が含まれない。got:\n%s", want, got)
		}
	}
	// 日本語件名は RFC 2047 でエンコードされる(生の UTF-8 が Subject 行に出ない)。
	if !strings.Contains(got, "Subject: =?utf-8?q?") {
		t.Errorf("Subject が Q エンコードされていない。got:\n%s", got)
	}
}

func Test_SMTP送信_宛先の改行注入は拒否する(t *testing.T) {
	// ヘッダインジェクション(CRLF で Bcc 等を注入)はアドレス検証で弾く。
	s := NewSender("127.0.0.1", "1", "noreply@staging.example.jp")
	err := s.SendInvitationEmail(
		context.Background(),
		"a@example.com\r\nBcc: victim@example.com", "s", "h", "t",
	)
	if err == nil {
		t.Fatal("改行入りの宛先が拒否されなかった")
	}
}

func Test_SMTP送信_件名の改行は除去される(t *testing.T) {
	addr, received := fakeSMTPServer(t)
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		t.Fatalf("split addr: %v", err)
	}

	s := NewSender(host, port, "noreply@staging.example.jp")
	if err := s.SendInvitationEmail(
		context.Background(),
		"invitee@example.com", "subject\r\nBcc: victim@example.com", "h", "t",
	); err != nil {
		t.Fatalf("send: %v", err)
	}

	var got string
	select {
	case got = <-received:
	case <-time.After(3 * time.Second):
		t.Fatal("fake サーバーに DATA が届かなかった")
	}
	// 改行が除去されていれば "Bcc:" は Subject 行の中の無害な文字列として残るだけで、
	// 独立したヘッダ行(行頭の Bcc:)にはならない。
	for _, line := range strings.Split(got, "\n") {
		if strings.HasPrefix(strings.TrimRight(line, "\r"), "Bcc:") {
			t.Errorf("件名経由でヘッダ行が注入された。got:\n%s", got)
		}
	}
}

func Test_SMTP送信_接続失敗はエラーを返す(t *testing.T) {
	// 未使用ポートに向けて送ると接続エラーになる。
	s := NewSender("127.0.0.1", "1", "noreply@staging.example.jp")
	if err := s.SendInvitationEmail(context.Background(), "a@example.com", "s", "h", "t"); err == nil {
		t.Fatal("接続できないのにエラーが返らなかった")
	}
}
