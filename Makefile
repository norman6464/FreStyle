# FreStyle ルート Makefile — ステージング(Lightsail)のデプロイ / リリース前確認 / 本番反映フロー。
# backend 個別のターゲット（test / verify / openapi 等）は backend/Makefile を参照。
#
# 想定フロー:
#   1) make staging-release        # ステージングへ deploy → 検証（NG なら停止）
#   2) （検証 OK なら）make prod-deploy-backend / prod-deploy-frontend
#      → CD ワークフローが起動し、GitHub の production Environment 承認ゲートで停止 → 承認で本番反映

.PHONY: staging-deploy staging-verify staging-release prod-deploy-backend prod-deploy-frontend help

help:
	@echo "staging-deploy        STAGING_HOST=... [STAGING_SSH_KEY=...] [STAGING_DIR=...] [BRANCH=main]"
	@echo "staging-verify        STAGING_URL=... [STAGING_API_URL=...]"
	@echo "staging-release       上記 2 つを連続実行（検証 NG で停止）"
	@echo "prod-deploy-backend   本番 backend(ECS) デプロイを起動（production 承認ゲートで停止）"
	@echo "prod-deploy-frontend  本番 frontend(S3+CloudFront) デプロイを起動（同上）"

# 1) ステージングにデプロイ（SSH → git pull → build → 再起動）
staging-deploy:
	./scripts/staging-deploy.sh $(BRANCH)

# 2) リリース前確認（ステージングへスモーク）
staging-verify:
	./scripts/staging-verify.sh

# 3) deploy → verify を一括（検証 NG なら staging-verify が非ゼロ終了して止まる）
staging-release: staging-deploy staging-verify
	@echo ">>> ステージング検証まで完了。問題なければ make prod-deploy-* で本番反映を起動してください。"

# 4) 本番反映の起動（実際の反映は GitHub の production 承認ゲート通過後）
prod-deploy-backend:
	gh workflow run "CD - Backend Deploy to ECS" -R norman6464/FreStyle -f confirm=deploy

prod-deploy-frontend:
	gh workflow run "CD - Frontend Deploy to S3 + CloudFront" -R norman6464/FreStyle -f confirm=deploy
