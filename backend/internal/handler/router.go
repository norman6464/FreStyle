package handler

import (
	"database/sql"

	"github.com/gin-gonic/gin"
	"github.com/norman6464/FreStyle/backend/internal/adapter/persistence"
	"github.com/norman6464/FreStyle/backend/internal/handler/middleware"
	"github.com/norman6464/FreStyle/backend/internal/infra/config"
	"github.com/norman6464/FreStyle/backend/internal/infra/oidc"
	"github.com/norman6464/FreStyle/backend/internal/usecase"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
)

// routeDeps はドメインごとの register*Routes 関数に渡す共通依存。
type routeDeps struct {
	db       *sql.DB
	cfg      *config.Config
	userRepo repository.UserRepository
	// verifier は access_token / id_token の署名とクレームを検証する。
	// handler も使う（id_token を署名未検証で読まないため）。
	verifier *oidc.Verifier
}

// NewRouter は API ルーティングを組み立てる。
//
// verifier は呼び出し側（cmd/server）が組み立てて渡す。ここで組み立てて
// エラーを飲み込むと、設定が足りない状態のまま起動してしまう。
func NewRouter(db *sql.DB, cfg *config.Config, verifier *oidc.Verifier) *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())
	// 構造化アクセスログ(slog/JSON)。request_id 採番 + status 別レベルで出力する。
	// ヘルスチェック (ALB が 30 秒間隔で叩く /api/v2/health) と root の access log は出さない。
	// 大量の health ログが CloudWatch の取り込み課金を押し上げるのを防ぐ。
	r.Use(middleware.RequestLogger("/api/v2/health", "/"))
	r.Use(middleware.CORS())

	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "FreStyle Go backend"})
	})

	deps := &routeDeps{
		db:       db,
		cfg:      cfg,
		userRepo: persistence.NewUserRepository(db),
		verifier: verifier,
	}

	v2 := r.Group("/api/v2")

	registerHealthRoutes(v2, deps)
	registerInvitationPublicRoutes(v2, deps)
	// 共有リンクの検証だけは未認証。リンクを受け取った相手はログインしていない
	// （認可はトークンとパスワードそのものが担う）。
	registerKnowledgeBasePublicRoutes(v2, deps)
	authHandler := registerAuthPublicRoutes(v2, deps)

	authed := v2.Group("")
	authed.Use(middleware.JWTAuth(buildJWTVerify(verifier), cfg.OIDC.AdminRoleClaim))
	authed.Use(middleware.CurrentUser(deps.userRepo, persistence.NewKnowledgeBaseRepository(deps.db)))

	registerAuthAuthedRoutes(authed, authHandler)
	registerProfileRoutes(authed, deps)
	registerNoteRoutes(authed, deps)
	registerDocumentRoutes(authed, deps)
	registerSocialRoutes(authed, deps)
	registerAdminRoutes(authed, deps)
	registerEmbedRoutes(authed)
	registerExerciseRoutes(authed, deps)
	registerCourseRoutes(authed, deps)
	registerTeachingMaterialRoutes(authed, deps)
	registerLessonProgressRoutes(authed, deps)
	registerDashboardRoutes(authed, deps)
	registerDailyGoalsRoutes(authed, deps)
	registerKnowledgeBaseRoutes(authed, deps)
	return r
}

// buildJWTVerify は JWTAuth に渡す access_token 検証関数を組み立てる。
//
// **分岐は無い。** 検証器は 1 つで、設定が足りなければ config.Load が起動を止める。
//
// 以前はここに「JWKS が無く APP_ENV が local なら署名検証をしない」経路と、
// ローカル専用のパスワードログインが発行するトークンを受ける経路があった。
// どちらも Cognito を通さずに手元を動かすためのもので、Cognito をやめた今は
// 用が無い。逃げ道を残すと、設定を書き忘れた環境が「認証が効いているように
// 見えて実は素通し」という一番気づけない壊れ方をする。
func buildJWTVerify(v *oidc.Verifier) middleware.VerifyFunc {
	return v.Verify
}

// registerHealthRoutes は認証不要のヘルスチェック (/api/v2/health) を登録する。
func registerHealthRoutes(g *gin.RouterGroup, deps *routeDeps) {
	h := NewHealthHandler(usecase.NewCheckHealthUseCase(persistence.NewHealthRepository(deps.db)))
	g.GET("/health", h.Get)
}
