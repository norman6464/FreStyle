package domain

import "time"

// ShareLink はログイン不要でページを開ける公開 URL。
//
// 来訪者は kind='share_link' の Principal として扱う（PrincipalID）。主体の種類を
// 1 本に揃えておくと、権限解決の入口が主体ごとに分岐しない。
//
// 「そのリンクで何ができるか」は Capability だけで決まる。来訪者はワークスペースに
// 所属しないので、付与の 3 段はそもそも届かない。
//
// **共有リンクは広げる方向にしか働かない。** ログインしていない相手へ「見せる」を足すだけで、
// すでに見えている人から取り上げることはない。
type ShareLink struct {
	ID string `json:"id"`
	// WorkspaceID はテナント境界。
	WorkspaceID string `json:"workspaceId"`
	// PageID はリンクの対象ページ。このページとその子孫が対象になる。
	PageID string `json:"pageId"`
	// PrincipalID はこのリンクの来訪者を表す主体（kind='share_link'）。
	PrincipalID string `json:"principalId"`
	// Capability はリンク経由でできることの既定（view または edit）。
	Capability Capability `json:"capability"`
	// TokenHash は共有 URL に載るトークンの SHA-256（32 バイト）。平文は保存しない
	// （DB が漏れた時点で全リンクが開けるのを避けるため）。API へは絶対に出さない。
	TokenHash []byte `json:"-"`
	// PasswordHash はリンクを開くときのパスワードの bcrypt ハッシュ。nil ならパスワード無し。
	// トークンと違い人が選ぶ値で総当たりに弱いため、こちらは遅いハッシュで持つ。API へは出さない。
	PasswordHash *string `json:"-"`
	// ExpiresAt は有効期限。nil なら無期限。
	ExpiresAt *time.Time `json:"expiresAt,omitempty"`
	// RevokedAt は失効日時。nil なら有効。失効は行を消さず日付で残す（誰がいつ止めたかを追えるように）。
	RevokedAt *time.Time `json:"revokedAt,omitempty"`
	// CreatedByUserID は発行者（users.id）。
	CreatedByUserID uint64    `json:"createdByUserId"`
	CreatedAt       time.Time `json:"createdAt"`
	UpdatedAt       time.Time `json:"updatedAt"`
}

// Usable は now の時点でこのリンクが使えるか（失効しておらず期限内か）を返す。
// パスワードの照合は含まない（ハッシュ照合は infra 側の責務）。
func (l ShareLink) Usable(now time.Time) bool {
	if l.RevokedAt != nil {
		return false
	}
	if l.ExpiresAt != nil && !now.Before(*l.ExpiresAt) {
		return false
	}
	return true
}

// RequiresPassword はリンクを開くのにパスワードが要るかを返す。
func (l ShareLink) RequiresPassword() bool { return l.PasswordHash != nil }
