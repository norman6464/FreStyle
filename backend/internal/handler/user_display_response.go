package handler

import (
	"context"
	"log/slog"

	"github.com/norman6464/frestyle/backend/internal/domain"
	"github.com/norman6464/frestyle/backend/internal/usecase/user"
)

// userDisplayResponse は人を表示するのに要る最小限の応答形（id・表示名・アイコン・
// 状態メッセージ）。コメントの投稿者・ページの最終編集者・提案の作成者/解決者・
// チケットの発言者/作成者/変更履歴の実行者、どの screen もこの 1 つの形で返す
// （画面ごとに専用の {userId, name} 型を作らない — 段 5）。
type userDisplayResponse struct {
	UserID uint64 `json:"userId" example:"42"`
	// Name は引けなければ空文字（行は落とさない）。
	Name      string `json:"name"      example:"山田太郎"`
	AvatarURL string `json:"avatarUrl" example:"https://storage.googleapis.com/.../avatar.png"`
	Status    string `json:"status"    example:"会議中"`
}

func toUserDisplayResponse(userID uint64, d *domain.UserDisplay) userDisplayResponse {
	if d == nil {
		return userDisplayResponse{UserID: userID}
	}
	return userDisplayResponse{
		UserID: d.UserID, Name: d.Name, AvatarURL: d.AvatarURL, Status: d.StatusMessage,
	}
}

// userDisplayCache は 1 リクエストの応答を組み立てる間だけ使う、人の表示情報のその場限りの
// キャッシュ。同じ user_id が同じ応答内に何度も出る場合（同じ人が何度も返信する・同じ人が
// 何度も変更履歴に出る等）に LookupUserDisplayUseCase を何度も引かないためのもの。
// リクエストをまたいでは使わない。
type userDisplayCache map[uint64]userDisplayResponse

// resolveUserDisplay はユーザー ID を表示形へ解決する。解決に失敗しても応答は止めない
// （空文字で埋めてログだけ残す — 名前・アイコンが付随情報でしかない画面で、
// 本体の一覧・詳細ごと 500 にする理由が無いため）。
func resolveUserDisplay(ctx context.Context, lookup *user.LookupUserDisplayUseCase, userID uint64, cache userDisplayCache) userDisplayResponse {
	if v, ok := cache[userID]; ok {
		return v
	}
	d, err := lookup.Execute(ctx, userID)
	if err != nil {
		slog.WarnContext(ctx, "handler: user display resolve failed", "err", err, "userId", userID)
	}
	resp := toUserDisplayResponse(userID, d)
	cache[userID] = resp
	return resp
}
