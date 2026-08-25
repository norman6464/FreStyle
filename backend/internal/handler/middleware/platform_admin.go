package middleware

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/norman6464/FreStyle/backend/internal/usecase"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
)

// SyncPlatformAdmin はトークンが示す運営権限の在否を、認可が走る前に DB へ反映する。
//
// 反映を /auth/me だけでやると穴が残る。管理 API は JWTAuth → CurrentUser → RequireAdmin と
// 進むだけで /auth/me を通らないので、admin グループから外れた退任者が /auth/me を呼ばずに
// 直接 /admin/* を叩けば、同期前の DB の値で認可されてしまう。だから認証済み経路の全リクエストで、
// 認可の前に一度だけ突き合わせる。
//
// CurrentUser の後ろに置く。前に置くと users を 2 回引くことになるが、後ろなら CurrentUser が
// 既に読んだ行と claim を突き合わせるだけで済み、食い違ったときにしか DB へ触らない。
// RequireAdmin と各 handler の役割検査はさらに後ろなので、認可より前という条件は満たす。
func SyncPlatformAdmin(sync *usecase.SyncPlatformAdminUseCase, users repository.UserRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		if sync == nil || users == nil {
			c.Next()
			return
		}
		user := CurrentUserFromContext(c)
		if user == nil {
			c.Next()
			return
		}
		claim := PlatformAdminClaimFromContext(c)
		grant, decided := claim.Decided()
		// claim が無ければ何も分からないので触らない（groups claim は federated ユーザーの
		// トークンに載らないことがあり、欠落を「グループに居ない」と読むと締め出す）。
		// 既に一致していれば書くことも読み直すこともない。
		if !decided || user.IsPlatformAdmin == grant {
			c.Next()
			return
		}

		sub, _ := c.Get(ContextKeyCognitoSub)
		cognitoSub, _ := sub.(string)
		if cognitoSub == "" {
			c.Next()
			return
		}

		ctx := c.Request.Context()
		if _, err := sync.Execute(ctx, usecase.SyncPlatformAdminInput{
			CognitoSub: cognitoSub,
			Claim:      claim,
		}); err != nil {
			slog.ErrorContext(ctx, "platform admin sync failed", "cognitoSub", cognitoSub, "err", err)
			if !grant {
				// 剥奪に失敗したまま通すと、退任者がそのリクエストで管理 API を使えてしまう。
				// 付与の失敗は権限が上がらないだけなので、そちらは素通しでよい。
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "platform_admin_sync_failed"})
				return
			}
			c.Next()
			return
		}

		// 反映後の事実で認可させる。実効役割を決める規則は persistence の 1 箇所だけが持つので、
		// ここで役割を組み立て直さず、同じ経路で読み直した行に差し替える。
		fresh, err := users.FindByCognitoSub(ctx, cognitoSub)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "user_lookup_failed"})
			return
		}
		if fresh == nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "user_not_found"})
			return
		}
		c.Set(ContextKeyCurrentUserID, fresh.ID)
		c.Set(ContextKeyCurrentUser, fresh)
		c.Next()
	}
}
