package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// 認証 Cookie 関連の定数。値はフロントエンド側 (Cognito callback / axios withCredentials)
// と整合させる必要があるため、変えるときは双方を同期更新すること。
const (
	// CookieRefreshToken は Cognito から取得した refresh_token を格納する Cookie 名。
	// CookieAccessToken は jwt.go に定義済み（middleware 全体で再利用）。
	CookieRefreshToken = "refresh_token"

	// AccessTokenDefaultMaxAgeSeconds は Cognito レスポンスに expires_in が
	// 含まれていなかったときの fallback (1 時間)。Cognito User Pool の標準寿命と一致。
	AccessTokenDefaultMaxAgeSeconds = 3600

	// RefreshTokenMaxAgeSeconds は refresh_token Cookie の寿命 (30 日)。
	// Cognito User Pool の refresh-token-validity 設定と整合させる。
	RefreshTokenMaxAgeSeconds = 30 * 24 * 3600
)

// SetAccessTokenCookie は HttpOnly + Secure + SameSite=None で access_token を設定する。
// maxAgeSeconds が 0 以下の場合は AccessTokenDefaultMaxAgeSeconds に丸める。
//
// Cognito refresh / callback 双方で同じヘッダー設定が必要だったため共通化した。
// 旧実装では SetSameSite と SetCookie の組み合わせが各 handler に散在し
// 設定漏れリスクがあったため、責務を 1 箇所に集約する。
func SetAccessTokenCookie(c *gin.Context, accessToken string, maxAgeSeconds int) {
	if maxAgeSeconds <= 0 {
		maxAgeSeconds = AccessTokenDefaultMaxAgeSeconds
	}
	c.SetSameSite(http.SameSiteNoneMode)
	c.SetCookie(CookieAccessToken, accessToken, maxAgeSeconds, "/", "", true, true)
}

// SetRefreshTokenCookie は refresh_token を 30 日寿命で書き込む。
// 空文字列のときは何もしない（Cognito が refresh_token を返さないケースに対応）。
func SetRefreshTokenCookie(c *gin.Context, refreshToken string) {
	if refreshToken == "" {
		return
	}
	c.SetSameSite(http.SameSiteNoneMode)
	c.SetCookie(CookieRefreshToken, refreshToken, RefreshTokenMaxAgeSeconds, "/", "", true, true)
}

// ClearAuthCookies は access / refresh の両 Cookie を即時失効させる。
// Logout / refresh 失敗時のフロント誘導でログイン状態をクリアするのに使う。
func ClearAuthCookies(c *gin.Context) {
	c.SetSameSite(http.SameSiteNoneMode)
	c.SetCookie(CookieAccessToken, "", -1, "/", "", true, true)
	c.SetCookie(CookieRefreshToken, "", -1, "/", "", true, true)
}
