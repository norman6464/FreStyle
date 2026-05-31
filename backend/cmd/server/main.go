// FreStyle backend (Go / Gin / GORM).
//
// @title           FreStyle Backend API
// @version         2.0
// @description     新卒 IT エンジニア 向け 統合 研修 プラットフォーム の REST API。
// @description     Clean Architecture (port = usecase/repository / adapter = adapter/persistence) で 構築。
// @termsOfService  https://normanblog.com/terms
//
// @contact.name    FreStyle Engineering
// @contact.url     https://normanblog.com
//
// @BasePath        /api/v2
//
// host / schemes は spec に 焼き込まず、 Swagger UI が ロード 元 と 同じ origin
// (本番 = api.normanblog.com、 ローカル = localhost:8080 等) を 自動 で 使う ように
// する。 そう しない と ローカル 開発 で 「Try it out」 が 本番 を 叩いて しまう。
//
// @securityDefinitions.apikey  CookieAuth
// @in                          header
// @name                        Cookie
// @description                 Cognito 由来 の JWT を HttpOnly Cookie で 送る。 ログイン 後 自動 付与。
package main

import (
	"log"

	"github.com/gin-gonic/gin"
	_ "github.com/norman6464/FreStyle/backend/docs" // swag init で 生成 さ れる OpenAPI spec
	"github.com/norman6464/FreStyle/backend/internal/handler"
	"github.com/norman6464/FreStyle/backend/internal/infra/config"
	"github.com/norman6464/FreStyle/backend/internal/infra/database"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config load failed: %v", err)
	}

	// 本番は gin を release モードにする。debug モードのルート登録ログ ([GIN-debug] ...) や
	// 起動時 warning を抑止して CloudWatch のログ量を減らす。ローカルは debug のまま。
	if cfg.AppEnv != "local" {
		gin.SetMode(gin.ReleaseMode)
	}

	db, err := database.NewPostgres(cfg)
	if err != nil {
		log.Fatalf("database connect failed: %v", err)
	}

	// Go domain を「正」とする AutoMigrate。
	// RESET_DB=true なら DROP SCHEMA → CREATE SCHEMA で完全リセット後 AutoMigrate。
	// 詳細は backend/internal/infra/database/migrate.go と
	// docs/16-go-schema-design.md (frestyle-infrastructure) を参照。
	if err := database.Migrate(db); err != nil {
		log.Fatalf("migrate failed: %v", err)
	}

	r := handler.NewRouter(db, cfg)
	addr := ":" + cfg.ServerPort
	log.Printf("FreStyle Go backend listening on %s (env=%s)", addr, cfg.AppEnv)
	if err := r.Run(addr); err != nil {
		log.Fatalf("server stopped: %v", err)
	}
}
