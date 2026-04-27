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

	r := handler.NewRouter(db, cfg)
	addr := ":" + cfg.ServerPort
	// Cognito secret 不一致の診断用一時ログ。値そのものは出さず長さと末尾 4 文字のみ。
	// 解消後に削除する。
	csLen := len(cfg.Cognito.ClientSecret)
	csTail := ""
	if csLen >= 4 {
		csTail = cfg.Cognito.ClientSecret[csLen-4:]
	}
	log.Printf("FreStyle Go backend listening on %s (env=%s) cognito_client_id=%s cognito_secret_len=%d cognito_secret_tail=...%s cognito_token_uri=%s",
		addr, cfg.AppEnv, cfg.Cognito.ClientID, csLen, csTail, cfg.Cognito.TokenURI)
	if err := r.Run(addr); err != nil {
		log.Fatalf("server stopped: %v", err)
	}
}
