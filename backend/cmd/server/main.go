package main

import (
	"log"

	"github.com/norman6464/FreStyle/backend/internal/handler"
	"github.com/norman6464/FreStyle/backend/internal/infra/config"
	"github.com/norman6464/FreStyle/backend/internal/infra/database"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config load failed: %v", err)
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
