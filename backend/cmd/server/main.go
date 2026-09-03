// FreStyle backend (Go / Gin / sqlc).
//
// @title           FreStyle Backend API
// @version         2.0
// @description     新卒 IT エンジニア 向け 統合 研修 プラットフォーム の REST API。
// @description     Clean Architecture (port = usecase/repository / adapter = adapter/persistence) で 構築。
// @termsOfService  https://frestyle.jp/terms
//
// @contact.name    FreStyle Engineering
// @contact.url     https://frestyle.jp
//
// @BasePath        /api/v2
//
// host / schemes は spec に焼き込まない（ローカルの「Try it out」が本番を叩かないようにするため）。
//
// @securityDefinitions.apikey  CookieAuth
// @in                          header
// @name                        Cookie
// @description                 Cognito 由来 の JWT を HttpOnly Cookie で 送る。 ログイン 後 自動 付与。
package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/gin-gonic/gin"
	_ "github.com/norman6464/FreStyle/backend/docs" // swag init で生成される OpenAPI spec
	"github.com/norman6464/FreStyle/backend/internal/handler"
	"github.com/norman6464/FreStyle/backend/internal/infra/config"
	"github.com/norman6464/FreStyle/backend/internal/infra/database"
	"github.com/norman6464/FreStyle/backend/internal/infra/logging"
	"github.com/norman6464/FreStyle/backend/internal/infra/oidc"
)

// fatal は致命的エラーを構造化ログで出して終了する（log.Fatalf の slog 版）。
func fatal(msg string, err error) {
	slog.Error(msg, slog.Any("error", err))
	os.Exit(1)
}

func main() {
	cfg, err := config.Load()
	if err != nil {
		// logging.Setup 前なので env は分からない。既定(Info/JSON)で出す。
		logging.Setup("")
		fatal("config load failed", err)
	}

	// 構造化ログ(slog/JSON)を初期化する。以降は request middleware 含め JSON で出力する。
	logging.Setup(cfg.AppEnv)

	// 本番は gin を release モードにする。debug モードのルート登録ログ ([GIN-debug] ...) や
	// 起動時 warning を抑止して CloudWatch のログ量を減らす。ローカルは debug のまま。
	if cfg.AppEnv != "local" {
		gin.SetMode(gin.ReleaseMode)
	}

	sqlDB, err := database.NewPostgres(cfg)
	if err != nil {
		fatal("database connect failed", err)
	}

	// infra/database/schema/*.sql
	if err := database.Migrate(context.Background(), sqlDB); err != nil {
		fatal("migrate failed", err)
	}

	// トークンの検証器はここで組み立てる。設定が足りなければ起動を止める。
	// router の中で組み立ててエラーを飲み込むと、検証していないまま動く環境ができる。
	verifier, err := oidc.NewVerifier(oidc.Config{
		Issuer:    cfg.OIDC.Issuer,
		JWKSURI:   cfg.OIDC.JWKSURI,
		ClientID:  cfg.OIDC.ClientID,
		Audiences: cfg.OIDC.Audiences,
	})
	if err != nil {
		fatal("oidc verifier init failed", err)
	}

	r := handler.NewRouter(sqlDB, cfg, verifier)
	addr := ":" + cfg.ServerPort
	slog.Info("FreStyle Go backend listening", slog.String("addr", addr), slog.String("env", cfg.AppEnv))
	if err := r.Run(addr); err != nil {
		fatal("server stopped", err)
	}
}
