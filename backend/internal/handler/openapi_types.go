// このファイルの型は handler から直接参照されず、swaggo が @Success / @Failure の
// {object} 識別子として名前解決するためだけに存在する（U1000 を抑止する）。

//lint:file-ignore U1000 referenced only by swaggo annotations

package handler

import "time"

// errorResponse は全 handler 共通のエラー JSON 形式。
type errorResponse struct {
	Error string `json:"error" example:"unauthorized"`
}

// successMessage は success プロパティ 1 個の成功レスポンス。
type successMessage struct {
	Success string `json:"success" example:"プロフィールを更新しました"`
}

// messageResponse は一般的な success メッセージ。
type messageResponse struct {
	Message string `json:"message" example:"ログインしました。"`
}

// meResponse は /auth/me の戻り値形（OpenAPI 用の型化）。
type meResponse struct {
	ID          uint64    `json:"id"          example:"42"`
	Email       string    `json:"email"       example:"user@example.com"`
	Name        string    `json:"name"        example:"山田 太郎"`
	WorkspaceID *string   `json:"workspaceId,omitempty" example:"0198a000-0000-7000-8000-000000000001"`
	Role        string    `json:"role"        example:"trainee"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
	Groups      []string  `json:"groups"      example:"admin"`
	IsAdmin     bool      `json:"isAdmin"     example:"false"`
}

// invitationValidateResponse は /invitations/accept/{token} の戻り値形（email は含めない）。
type invitationValidateResponse struct {
	Role        string  `json:"role"        example:"trainee"`
	Name        string  `json:"name"        example:"山田 太郎"`
	CompanyName string  `json:"companyName" example:"Example Corp"`
	WorkspaceID *string `json:"workspaceId,omitempty" example:"0198a000-0000-7000-8000-000000000001"`
}
