package handler

import (
	"errors"
	"testing"
)

func Test_ユーザー統計_ユーザーID解決_meキーワード(t *testing.T) {
	h := &UserStatsHandler{}
	uid, err := h.resolveUserID(makeCtx(7, "me"))
	if err != nil || uid != 7 {
		t.Fatalf("'me' should resolve to current user; got uid=%d err=%v", uid, err)
	}
}

func Test_ユーザー統計_ユーザーID解決_不一致の数値は禁止(t *testing.T) {
	h := &UserStatsHandler{}
	if _, err := h.resolveUserID(makeCtx(7, "99")); !errors.Is(err, errUserStatsForbidden) {
		t.Fatalf("mismatch numeric should be forbidden; got %v", err)
	}
}

func Test_ユーザー統計_ユーザーID解決_カレントユーザーなしは未認証(t *testing.T) {
	h := &UserStatsHandler{}
	if _, err := h.resolveUserID(makeCtx(0, "me")); !errors.Is(err, errUserStatsUnauthorized) {
		t.Fatalf("no current user should be unauthorized; got %v", err)
	}
}
