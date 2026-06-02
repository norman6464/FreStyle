package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// 認証 Cookie 関連の定数。フロント側の設定と同期更新すること。
const (
	// CookieRefreshToken は refresh_token を格納する Cookie 名（CookieAccessToken は jwt.go）。
	CookieRefreshToken = "refresh_token"

	// AccessTokenDefaultMaxAgeSeconds は expires_in 不在時の fallback（1 時間）。
	AccessTokenDefaultMaxAgeSeconds = 3600

	// RefreshTokenMaxAgeSeconds は refresh_token Cookie の寿命（30 日）。
	RefreshTokenMaxAgeSeconds = 30 * 24 * 3600
)

// SetAccessTokenCookie は HttpOnly + Secure + SameSite=None で access_token を設定する。
// maxAgeSeconds が 0 以下なら AccessTokenDefaultMaxAgeSeconds に丸める。
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
