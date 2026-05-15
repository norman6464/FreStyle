// Package repository は usecase 層 が 依存 する 永続化 境界 (port) を 定義 する。
//
// Clean Architecture の Use Case 層 が「自分 が 必要 と する 抽象」 を ここ で 宣言 し、
// 実装 は [github.com/norman6464/FreStyle/backend/internal/adapter/persistence]
// 配下 で 提供 さ れる (依存 方向: usecase ← persistence、 DIP)。
//
// 1 boundary = 1 fat interface の 方針 で 配置 する (Clean Architecture 通常 流儀)。
// 単一 メソッド の port (例: Presigner / Enqueuer / Publisher) は Effective Go 流 の
// -er 命名 で 別 ファイル に 配置 して 良い。
package repository

import (
	"context"

	"github.com/norman6464/FreStyle/backend/internal/domain"
)

// UserRepository は users テーブル へ の アクセス を 提供 する。
//
// 旧 internal/repository/user_repository.go を Clean Architecture 構造 に 沿って
// usecase 層 (port) と adapter 層 (実装) に 物理 分離 した うえ で、 interface 定義
// を こちら に 移植 した もの。 メソッド 個別 の WHY は impl 側 (persistence.userRepository)
// の メソッド コメント を 参照。
type UserRepository interface {
	FindByCognitoSub(ctx context.Context, sub string) (*domain.User, error)
	FindByID(ctx context.Context, id uint64) (*domain.User, error)
	Create(ctx context.Context, user *domain.User) error
	// UpdateDisplayName は ProfilePage の「氏名」変更、 および OIDC ログイン時に
	// 旧 displayName=email を id_token の name claim で自動補正するときに呼ばれる。
	UpdateDisplayName(ctx context.Context, userID uint64, displayName string) error
	// UpdateRole は Cognito group → DB role 同期、または招待受諾時に呼ばれる。
	UpdateRole(ctx context.Context, userID uint64, role string) error
	// UpdateCompanyID は既存ユーザーが招待を受けて company に紐付くときに呼ばれる。
	UpdateCompanyID(ctx context.Context, userID uint64, companyID uint64) error
	// MarkOnboarded は Welcome 画面の「はじめる」ボタン押下時に呼ばれ、
	// onboarded_at = NOW() に更新する。冪等（既に値が入っていても上書きしない）。
	MarkOnboarded(ctx context.Context, userID uint64) error
}
