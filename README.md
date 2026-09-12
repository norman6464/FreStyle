# FreStyle

チームのナレッジ共有とチケット管理をまとめた Web アプリケーション。

- バックエンド: Go / Gin / sqlc（`backend/`）
- フロントエンド: React 19 / TypeScript / Vite / Tailwind（`frontend/`）
- DB: PostgreSQL 17

開発への参加方法（ブランチ運用・PR フロー・アーキテクチャ規約・テスト方針など）は [CONTRIBUTING.md](./CONTRIBUTING.md) を参照。

---

## セットアップ

### 前提

- Docker Desktop（または docker + docker compose plugin）
- Node.js（`corepack enable` で pnpm を使う）
- Go 1.x（backend のコードを編集する場合。アプリの実行自体は docker 経由でよく、ホストに Go を入れなくても起動できる）

### 起動

backend は **ローカルでは docker compose でだけ動かす**（`go run` は使わない）。DB と、ローカル用の OIDC 発行者 Dex を一緒に起動する前提のため。

```bash
# リポジトリ直下で: DB / Dex（ローカル用 OIDC 発行者）/ backend をまとめて起動
docker compose up -d

# 起動確認
curl http://localhost:8080/api/v2/health   # {"status":"UP","db":"UP"}

# 初回のみ: ダミーデータを投入
cd backend && make local-seed && cd ..
```

フロントエンド:

```bash
cd frontend
corepack enable        # pnpm を用意する（入らない環境では npm i -g pnpm）
cp .env.example .env   # 無ければ用意する
pnpm install
pnpm run dev
```

http://localhost:5173 を開く。

### ログイン（ローカルは Dex）

本番は GCIP（クライアント SDK が直接サインイン）だが、ローカルは OSS の OIDC 発行者 **Dex**（`docker-compose.yml` の `idp` サービス）を使い、本番と同じ認可コード + PKCE の経路をそのまま通す。

- 既定ユーザー: `admin@example.com` / `password`（`docker/idp/config.yaml` の `staticPasswords`。`backend/scripts/seed-local.sql` の管理者ユーザーと同じ資格情報）
- `frontend/.env` は gitignore 対象のため、`.env.example` が更新されても自分の `.env` には自動で反映されない。**ログインボタンを押しても何も起きない（エラーも出ない）**ときは、OIDC 系のキー（`VITE_OIDC_AUTHORIZE_URI` / `VITE_OIDC_TOKEN_URI` / `VITE_OIDC_CLIENT_ID` / `VITE_OIDC_REDIRECT_URI`）が `.env` に揃っているか `.env.example` と見比べる。1 つでも欠けると `readAuthConfig()`（`frontend/src/shared/lib/auth/authConfig.ts`）が「未設定」と判断し、認可 URL への遷移自体を始めない設計になっている
- `.env` を直したら `pnpm run dev` を再起動する（Vite の環境変数はプロセス起動時にしか読み込まれない）

### ローカル DB をやり直す

ローカル DB のスキーマは backend コンテナの**初回起動時にだけ**適用される（`internal/infra/database/schema.go` の `ApplySchema`。`users` テーブルが無いときだけ流す設計で、既存 DB への差分適用はしない）。スキーマ変更を取り込んだのに古いままで動いている・`make local-seed` が `column ... does not exist` のようなエラーで失敗するときは、DB を作り直す:

```bash
cd backend
make local-reset      # docker volume ごと破棄
cd ..
docker compose up -d   # 次回起動時に現行スキーマが再適用される
cd backend && make local-seed
```
