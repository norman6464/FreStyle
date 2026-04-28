package handler

import "testing"

func TestUserStatsResolveUserID_MeKeyword(t *testing.T) {
	h := &UserStatsHandler{}
	uid, err := h.resolveUserID(makeCtx(7, "me"))
	if err != nil || uid != 7 {
		t.Fatalf("'me' should resolve to current user; got uid=%d err=%v", uid, err)
	}
}

func TestUserStatsResolveUserID_MismatchNumericIsForbidden(t *testing.T) {
	h := &UserStatsHandler{}
	if _, err := h.resolveUserID(makeCtx(7, "99")); err != errUserStatsForbidden {
		t.Fatalf("mismatch numeric should be forbidden; got %v", err)
	}
}

func TestUserStatsResolveUserID_NoCurrentUserIsUnauthorized(t *testing.T) {
	h := &UserStatsHandler{}
	if _, err := h.resolveUserID(makeCtx(0, "me")); err != errUserStatsUnauthorized {
		t.Fatalf("no current user should be unauthorized; got %v", err)
	}
}
