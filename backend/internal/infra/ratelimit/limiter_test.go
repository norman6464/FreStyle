package ratelimit

import (
	"testing"
	"time"
)

func Test_レートリミッタ_バースト後に拒否(t *testing.T) {
	l := New(60, 3)
	for i := 0; i < 3; i++ {
		if !l.Allow("1.2.3.4") {
			t.Fatalf("request %d should be allowed within burst", i+1)
		}
	}
	if l.Allow("1.2.3.4") {
		t.Fatal("4th request should be denied after burst is exhausted")
	}
}

func Test_レートリミッタ_キーごとに独立(t *testing.T) {
	l := New(60, 1)
	if !l.Allow("a") || !l.Allow("b") {
		t.Fatal("different keys should have independent buckets")
	}
	if l.Allow("a") {
		t.Fatal("same key should be limited after its own burst")
	}
}

func Test_レートリミッタ_時間経過で回復(t *testing.T) {
	cur := time.Now()
	l := New(60, 1) // 1 token/sec
	l.now = func() time.Time { return cur }

	if !l.Allow("k") {
		t.Fatal("first request should pass")
	}
	if l.Allow("k") {
		t.Fatal("immediate second request should be denied")
	}
	cur = cur.Add(1100 * time.Millisecond)
	if !l.Allow("k") {
		t.Fatal("request after refill window should pass")
	}
}

func Test_レートリミッタ_Forgetでバケツごと消える(t *testing.T) {
	// 存在しない対象の鍵を残さないための口。消したあとは満タンから始まる。
	l := New(60, 1)
	if !l.Allow("gone") {
		t.Fatal("first request should pass")
	}
	if l.Allow("gone") {
		t.Fatal("second request should be denied before Forget")
	}
	l.Forget("gone")
	if !l.Allow("gone") {
		t.Fatal("bucket should start full again after Forget")
	}
	if got := len(l.buckets); got != 1 {
		t.Fatalf("Forget 後に残るバケツは Allow で作り直した 1 つだけのはず: %d", got)
	}
}

func Test_レートリミッタ_放置されたバケツは掃除される(t *testing.T) {
	cur := time.Now()
	l := New(60, 1)
	l.now = func() time.Time { return cur }
	l.lastCleanup = cur

	l.Allow("old")
	cur = cur.Add(11 * time.Minute)
	l.Allow("new")

	if _, ok := l.buckets["old"]; ok {
		t.Fatal("idleTTL を超えたバケツは掃除されるはず")
	}
}
