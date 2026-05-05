# 招待マジックリンクフロー（Path B）

## 概要

CompanyAdmin 招待を **Cognito 事前作成 + Cognito 自動メール** から **token 付きマジックリンク + SES 独自メール** に切り替える。

理由:
- Google OIDC 利用時に「Cognito 事前作成済みアカウント」と「OIDC で作られるアカウント」が email で重複して識別できなくなる問題があった
- Cognito の招待メールは件名・本文・差出人名のカスタマイズに制約があり、ブランディング・日本語表現の自由度が低かった
- 招待リンクをユーザー側で「このメールが本物か」確認しづらかった

## 全体フロー

```
[SuperAdmin] 招待作成
  POST /api/v2/admin/invitations
    ↓
[Backend] invitations に UUID token を保存（Cognito 操作なし）
[Backend] SES で独自 HTML メール送信
  本文に https://normanblog.com/invitations/accept?token=<UUID>
    ↓
[ユーザー] メールのリンクをクリック
    ↓
[/invitations/accept?token=...] フロントエンド
  GET /api/v2/invitations/accept/:token で検証
    → 「ログインへ進む」ボタンを表示（招待先の company / role を表示）
    → token を sessionStorage に保持
    ↓
[/login] ログイン画面
  - Google OIDC（推奨）
  - or Cognito SignUp（メール+パスワードを自分で設定）
    ↓
[Backend] /auth/cognito/callback の upsertUserFromIDToken が
  sessionStorage の token（state パラメータ経由）を invitations と照合し、
  invitation の role / company / display_name を新規ユーザーに反映 + status=accepted
```

## データモデル

### `invitations` テーブル

| カラム | 型 | 用途 |
|---|---|---|
| `id` | uint64 PK | |
| `company_id` | uint64 INDEX | 招待先 company |
| `email` | string | 招待先メールアドレス |
| `role` | string | 付与する role（company_admin / trainee） |
| `display_name` | string | 表示名（任意） |
| `status` | string | `pending` / `accepted` / `canceled` |
| `token` | *string UNIQUE size:64 NULL | **PR-A で追加。** マジックリンクの不透明 UUID。NULL 許容（既存 pending を NULL のまま残し UNIQUE 競合を避ける） |
| `expires_at` | datetime | 招待有効期限（usecase で 7 日後をセット） |
| `created_at` | datetime | |

### Token 設計

- 形式: UUID v4（128 bit）。推測不能な値。
- 漏洩時の被害局所化のため:
  - 一招待につき 1 値（再発行は status=canceled で旧 token 失効 + 新規 invitation 作成）
  - DB 検証時に `status = 'pending' AND expires_at > NOW()` を必ず付与
  - JSON 応答には含めない（`json:"-"`）

## API

| メソッド | パス | 公開 | 用途 |
|---|---|---|---|
| POST | `/api/v2/admin/invitations` | 認証必須（admin） | 招待作成 + SES でメール送信 |
| GET | `/api/v2/admin/invitations` | 認証必須（admin） | 一覧 |
| DELETE | `/api/v2/admin/invitations/:id` | 認証必須（admin） | 取り消し |
| GET | `/api/v2/invitations/accept/:token` | **公開** | token 検証（招待先の company/role を返す） |

token 検証 API は認証不要だが、レート制限と「該当なし → 404」のみ返してメタ情報を漏らさない設計。

## PR 分割

| PR | スコープ | 状態 |
|---|---|---|
| PR-A | `invitations.token` カラム追加 + domain/repository 拡張 | ✅ マージ済 (#1629) |
| **PR-B** | SES クライアント新設 + Cognito 事前作成撤去 + 招待メール送信（コード側） | ✅ 本 PR |
| PR-C | token 検証 API + ログイン後 invitation 受諾フロー | 未着手 |
| PR-D | フロント `/invitations/accept` ページ + login 統合 | 未着手 |
| Infra | SES ドメイン検証（normanblog.com） + IAM `ses:SendEmail` 追加 | 未着手（[手順](./ses-setup.md)） |

## SES サンドボックスについて

サンドボックス解除（送信量上限 200 通/日 → 50,000 通/日）は本リリースでは申請しない。
当面は社内招待用途のみで使うため 200 通/日で十分。必要になったら AWS Console から申請する。

サンドボックス中は **送信元・宛先の両方が SES 上で検証済アドレス** である必要がある点に注意。
詳しい AWS Console 上の作業手順は [ses-setup.md](./ses-setup.md) を参照。

## PR-B で追加した環境変数

| 変数 | 用途 | 例 |
|---|---|---|
| `SES_REGION` | SES エンドポイントのリージョン | `ap-northeast-1` |
| `SES_FROM_ADDRESS` | 検証済の送信元（空なら送信スキップ） | `FreStyle <noreply@normanblog.com>` |
| `APP_BASE_URL` | マジックリンクのベース URL | `https://normanblog.com` |

`SES_FROM_ADDRESS` または `APP_BASE_URL` が未設定の場合は招待メール送信をスキップし、
バックエンドのログに `token=...` を出力するフォールバックモードで動作する（ローカル開発用）。

## PR-B で撤去したもの

- `infra/cognito/admin_client.go` （Cognito `AdminCreateUser` ラッパ）
- `repository.CognitoAdminClient` interface とその stub
- `CognitoConfig.UserPoolID` フィールドと `COGNITO_USER_POOL_ID` env
- `usecase.CreateAdminInvitationUseCase` の `cog repository.CognitoAdminClient` 引数
- `handler.AdminInvitationHandler.Create` の `ErrUserAlreadyConfirmed` 分岐
