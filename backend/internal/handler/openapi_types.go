// これら の 型 は handler コード から 直接 参照 さ れず、 swaggo (= `make openapi`)
// が doc コメント 内 の `@Success` / `@Failure` の `{object}` 識別子 と して
// 名前 解決 する。 そのため staticcheck の U1000 (unused) 警告 を ファイル 全体 で
// 抑止 する。 該当 型 を 実際 に handler から JSON marshal に 使う ように なれば
// この 抑止 は 不要。

//lint:file-ignore U1000 referenced only by swaggo annotations

package handler

import "time"

// errorResponse は すべて の handler が 返す 共通 エラー JSON 形式。
// OpenAPI spec で @Failure の レスポンス 型 と して 使う。
type errorResponse struct {
	Error string `json:"error" example:"unauthorized"`
}

// successMessage は 単純 メッセージ のみ を 返す 成功 レスポンス (200 OK で 内容 が プロパティ 1 個 の とき)。
type successMessage struct {
	Success string `json:"success" example:"プロフィールを更新しました"`
}

// messageResponse は ログアウト / ログイン 完了 等 の 一般 success メッセージ。
type messageResponse struct {
	Message string `json:"message" example:"ログインしました。"`
}

// meResponse は /auth/me の 戻り 値 形 (gin.H で 返して いる の を OpenAPI 用 に 型 化)。
type meResponse struct {
	ID          uint64    `json:"id"          example:"42"`
	CognitoSub  string    `json:"cognitoSub"  example:"abc-123-uuid"`
	Email       string    `json:"email"       example:"user@example.com"`
	DisplayName string    `json:"displayName" example:"山田 太郎"`
	CompanyID   *uint64   `json:"companyId,omitempty" example:"1"`
	Role        string    `json:"role"        example:"trainee"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
	Groups      []string  `json:"groups"      example:"admin"`
	IsAdmin     bool      `json:"isAdmin"     example:"false"`
	Onboarded   bool      `json:"onboarded"   example:"true"`
}

// invitationValidateResponse は /invitations/accept/{token} (public) の 戻り 値 形。
// email を 含め ない (= メタ 情報 漏洩 防止)。
type invitationValidateResponse struct {
	Role        string `json:"role"        example:"trainee"`
	DisplayName string `json:"displayName" example:"山田 太郎"`
	CompanyID   uint64 `json:"companyId"   example:"1"`
	CompanyName string `json:"companyName" example:"Example Corp"`
}
