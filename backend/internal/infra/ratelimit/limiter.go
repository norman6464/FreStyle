// Package ratelimit は「鍵ごとに、単位時間あたり何回まで」を数えるトークンバケットを提供する。
//
// 鍵に何を使うかはこのパッケージの関心事ではない（呼び出し側が決める）。HTTP の入口で
// IP を鍵にするのも、共有リンクの検証でリンクを鍵にするのも、同じこの実装を使う。
// gin を import しないのはそのためで、middleware 側が *gin.Context から鍵を作って渡す。
//
// 単一 ECS タスク（desiredCount=1）前提の in-memory 実装。スケールアウトすると
// インスタンスごとに別カウントになるため、その際は共有ストア（Redis 等）が要る。
package ratelimit

import (
	"sync"
	"time"
)

// bucket は 1 つの鍵ぶんのトークンバケット状態。
type bucket struct {
	tokens float64
	last   time.Time
}

// Limiter は鍵ごとのトークンバケットで流量を制限する。ゼロ値は使えない（New を使う）。
type Limiter struct {
	mu          sync.Mutex
	buckets     map[string]*bucket
	rate        float64 // 毎秒の補充トークン数
	burst       float64 // バケット上限
	idleTTL     time.Duration
	lastCleanup time.Time
	now         func() time.Time // テスト差し替え用
}

// New は「鍵あたり perMinute 回（短期は burst まで許容）」の Limiter を作る。
func New(perMinute float64, burst int) *Limiter {
	return &Limiter{
		buckets:     map[string]*bucket{},
		rate:        perMinute / 60.0,
		burst:       float64(burst),
		idleTTL:     10 * time.Minute,
		lastCleanup: time.Now(),
		now:         time.Now,
	}
}

// Allow は key のトークンを 1 つ消費できれば true を返す。
// 消費できなければ false（＝ 上限に達している）。
func (l *Limiter) Allow(key string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	now := l.now()
	l.cleanupLocked(now)

	b, ok := l.buckets[key]
	if !ok {
		b = &bucket{tokens: l.burst, last: now}
		l.buckets[key] = b
	}
	b.tokens += now.Sub(b.last).Seconds() * l.rate
	if b.tokens > l.burst {
		b.tokens = l.burst
	}
	b.last = now

	if b.tokens >= 1 {
		b.tokens--
		return true
	}
	return false
}

// Forget は key のバケツを捨てる（次の Allow は満タンから始まる）。
//
// 用途は 1 つだけ: **守る対象が存在しなかったと分かったとき**に、その鍵の状態を残さないこと。
// 共有リンクの検証は「トークンを見つけてからでないと、それが実在するリンクか分からない」ので、
// まず Allow で 1 つ消費し、存在しなかった鍵だけをここで消す。そうしないと、でたらめな
// トークンを投げ続けるだけで、攻撃者が好きなだけバケツを作らせられる（鍵は要求ごとに変えられる）。
//
// 実在するリンクの鍵をここへ渡してはいけない（それは上限を自分でリセットさせる操作になる）。
func (l *Limiter) Forget(key string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	delete(l.buckets, key)
}

// cleanupLocked は idleTTL を超えて使われていないバケツを掃除する（メモリ肥大化防止）。
func (l *Limiter) cleanupLocked(now time.Time) {
	if now.Sub(l.lastCleanup) < l.idleTTL {
		return
	}
	for k, b := range l.buckets {
		if now.Sub(b.last) > l.idleTTL {
			delete(l.buckets, k)
		}
	}
	l.lastCleanup = now
}
