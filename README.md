# FreStyle

新卒エンジニア向けの研修プラットフォーム。

**いまゼロベースで設計をやり直している最中で、README は白紙に戻してある。**
形が決まってから書き直す。

## 開発

ローカルのみ。クラウド（AWS）では動かさない。将来はオンプレの Kubernetes + PostgreSQL に置く。

### 立ち上げ

```bash
cp .env.example .env
cp frontend/.env.example frontend/.env

# DB・メール・認証（Zitadel）をまとめて起動する
docker compose --profile auth up -d

# 認証の設定を作る（何度流しても同じ結果になる）
./backend/scripts/zitadel-setup.sh
```

最後のスクリプトが `.env` / `frontend/.env` に貼る値を出力するので、そのとおりに貼って
backend を起動し直す。**認証の設定が欠けていると backend は起動を止める**（検証していない
まま動く状態を作らないため）。

| 入口 | URL |
|---|---|
| フロントエンド | http://localhost:5173 |
| API | http://localhost:8080/api/v2 |
| ログイン画面 | http://zitadel.localhost:8081/ui/v2/login |
| 認証の管理画面 | http://zitadel.localhost:8081/ui/console |
| 受信メール | http://localhost:8025 |

初期管理者は `admin@frestyle.local`。パスワードは `.env` の
`LOCAL_ZITADEL_ADMIN_PASSWORD` で決める（未設定なら手元専用の既定値が入る）。

### 認証の構成

ログインは発行者（Zitadel）の画面で行い、アプリはメールとパスワードを受け取らない。
認可コードフロー + PKCE で、トークンは HttpOnly Cookie に入る。

コンテナは 3 つに分かれている。

- **zitadel** — 本体（トークンの発行と管理 API）
- **zitadel-login** — ログイン画面（v4 では本体と別のコンテナ）
- **zitadel-proxy** — 前段。上の 2 つを同じホスト名の下にまとめる

前段が要るのは、Zitadel が Host ヘッダで「どのインスタンス宛か」を決めるため。
ブラウザ・ログイン画面・backend の全部が `zitadel.localhost:8081` という同じ名前で
入るようにしてある。オンプレでも同じ形になるので、手元だけ構成を変えていない。

`.localhost` で終わる名前はブラウザが自動でループバックへ向ける（RFC 6761）ので、
`/etc/hosts` をいじる必要はない。
