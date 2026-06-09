package handler

import (
	"context"
	"errors"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/norman6464/FreStyle/backend/internal/adapter/persistence"
	"github.com/norman6464/FreStyle/backend/internal/handler/middleware"
	"github.com/norman6464/FreStyle/backend/internal/infra/bedrock"
	"github.com/norman6464/FreStyle/backend/internal/infra/cognito"
	"github.com/norman6464/FreStyle/backend/internal/infra/config"
	"github.com/norman6464/FreStyle/backend/internal/usecase"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
	swaggerfiles "github.com/swaggo/files"
	ginswagger "github.com/swaggo/gin-swagger"
	"gorm.io/gorm"
)

// routeDeps はドメインごとの register*Routes 関数に渡す共通依存。
type routeDeps struct {
	db            *gorm.DB
	cfg           *config.Config
	userRepo      repository.UserRepository
	bedrockClient *bedrock.Client
	msgRepo       repository.AiChatMessageRepository
}

// NewRouter は API ルーティングを組み立てる。
func NewRouter(db *gorm.DB, cfg *config.Config) *gin.Engine {
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

	ctx := context.Background()

	bc, err := bedrock.NewClient(ctx, cfg.Bedrock.Region, cfg.Bedrock.ModelID)
	if err != nil {
		log.Printf("WARN: Bedrock client init failed (AI chat WS will be unavailable): %v", err)
	}

	msgRepo, err := persistence.NewAiChatMessageRepository(ctx, cfg.DynamoDB.Region, cfg.DynamoDB.AiChatTable)
	if err != nil {
		log.Printf("WARN: DynamoDB client init failed (AI chat WS will be unavailable): %v", err)
	}

	deps := &routeDeps{
		db:            db,
		cfg:           cfg,
		userRepo:      persistence.NewUserRepository(db),
		bedrockClient: bc,
		msgRepo:       msgRepo,
	}

	v2 := r.Group("/api/v2")

	// Swagger UI。ALB が /api/v2/* のみ backend に流すため v2 group 内で配信する（認証不要）。
	v2.GET("/swagger/*any", ginswagger.WrapHandler(swaggerfiles.Handler))

	registerHealthRoutes(v2, deps)
	registerInvitationPublicRoutes(v2, deps)
	companyAppHandler := newCompanyApplicationHandler(deps)
	registerCompanyApplicationPublicRoutes(v2, companyAppHandler)
	authHandler := registerAuthPublicRoutes(v2, deps)

	authed := v2.Group("")
	authed.Use(middleware.JWTAuth(buildJWTVerify(cfg)))
	authed.Use(middleware.CurrentUser(deps.userRepo))

	registerAuthAuthedRoutes(authed, authHandler)
	registerChatRoutes(authed, deps)
	registerProfileRoutes(authed, deps)
	registerNoteRoutes(authed, deps)
	registerSocialRoutes(authed, deps)
	registerAdminRoutes(authed, deps)
	registerEmbedRoutes(authed)
	registerExerciseRoutes(authed, deps)
	registerCourseRoutes(authed, deps)
	registerTeachingMaterialRoutes(authed, deps)
	registerLearningReportRoutes(authed, deps)
	registerCompanySettingsRoutes(authed, deps)
	registerCompanyApplicationAdminRoutes(authed, companyAppHandler)
	// WebSocket (/ws/ai-chat) は SSE (/ai-chat/stream) への置換で廃止 (PR-D)。
	return r
}

// buildJWTVerify は JWTAuth に渡す access_token 検証関数を組み立てる。
//   - JWKS URI 設定あり: cognito.Verifier で JWKS 署名検証する（本番の正規経路）
//   - JWKS 未設定 & local: 署名未検証の decode にフォールバック（Cognito 無しのローカル開発用）
//   - JWKS 未設定 & 非 local: fail closed（全リクエストを拒否し、未検証で本番が動くのを防ぐ）
func buildJWTVerify(cfg *config.Config) middleware.VerifyFunc {
	if cfg.Cognito.JwkSetURI != "" {
		v := cognito.NewVerifier(cfg.Cognito.JwkSetURI)
		return v.Verify
	}
	if cfg.AppEnv == "local" {
		log.Printf("WARN: COGNITO_JWK_SET_URI 未設定 — ローカル開発のため JWT 署名検証をスキップします（本番では設定必須）")
		return func(_ context.Context, token string) (map[string]any, error) {
			return middleware.DecodeClaims(token)
		}
	}
	log.Printf("ERROR: COGNITO_JWK_SET_URI 未設定 — 署名検証できないため認証付きリクエストを全拒否します")
	return func(_ context.Context, _ string) (map[string]any, error) {
		return nil, errors.New("jwt verifier not configured")
	}
}

// registerHealthRoutes は認証不要のヘルスチェック (/api/v2/health) を登録する。
func registerHealthRoutes(g *gin.RouterGroup, deps *routeDeps) {
	h := NewHealthHandler(usecase.NewCheckHealthUseCase(persistence.NewHealthRepository(deps.db)))
	g.GET("/health", h.Get)
}
