# 会社 / ユーザーのアカウント停止・削除（管理操作）

運営（super_admin）・会社管理者（company_admin）が、会社やユーザーのアカウントを**停止（無効化）**したり削除したりするための管理操作。停止は削除と違い**可逆**で、停止中はログイン/利用を**即時に**不可にする。

> 本ドキュメントは段階的に拡張する。**PR1: 会社アカウントの有効/無効（実装済）** / PR2: ユーザーの有効/無効 + ソフト削除（予定）。

## 1. 会社アカウントの有効/無効（super_admin）

- 画面: `/admin/companies`（管理 → 会社一覧）。各会社に「無効化 / 有効化」ボタン（super_admin にのみ表示）と「無効」バッジ。
- API: `PATCH /api/v2/admin/companies/:id/active` body `{ "active": false }`（super_admin 専用）

### 振る舞い

- `companies.is_active = false` にすると、**その会社に属する全ユーザー**が次のリクエストから利用不可になる。
- 有効化（`active=true`）で即座に復帰する（可逆）。

### 認可 / enforcement（多層）

1. **更新権限**: `AdminCompanyHandler.SetActive` が `super_admin` のみ許可（他ロールは `403`）。フロントもボタンを super_admin にのみ表示。
2. **利用ブロック（即時）**: `CurrentUser` middleware が全認証リクエストで、ユーザーの所属会社をひき、`is_active=false` なら `403 company_disabled` を返す。これにより、無効化された会社のユーザーは**有効な JWT を持っていても**その時点でアクセス不可になる。
   - 会社行が無い（データ不整合）場合は弾かない。会社ひきの DB エラーは `500`。
   - super_admin は `company_id` を持たないため、この会社チェックの対象外（自分が締め出されることはない）。

## 2. レイヤー構成（クリーンアーキテクチャ）

```
AdminCompaniesPage / CompanyRepository.updateActive (frontend)
        │  PATCH /api/v2/admin/companies/:id/active
        ▼
AdminCompanyHandler.SetActive (super_admin チェック)
        ▼
SetCompanyActiveUseCase
        ▼
CompanyRepository.UpdateActive(impl: 生 SQL の UPDATE companies SET is_active=...)

CurrentUser middleware ── 会社が無効なら 403 で全リクエストを弾く（enforcement）
```

| 層 | ファイル |
|---|---|
| handler | `backend/internal/handler/admin_company_handler.go`（`SetActive`） |
| middleware | `backend/internal/handler/middleware/current_user.go`（会社無効チェック） |
| usecase | `backend/internal/usecase/set_company_active_usecase.go` |
| repository | `usecase/repository/company.go`（`UpdateActive`）/ `adapter/persistence/company_repository.go` |
| domain | `backend/internal/domain/company.go`（`IsActive`） |
| frontend | `frontend/src/pages/AdminCompaniesPage.tsx` / `repositories/CompanyRepository.ts` |

> `companies.is_active boolean NOT NULL DEFAULT true` を追加（AutoMigrate が本番に列を追加）。sqlc は `SELECT *` のため `make sqlc` 再生成で `IsActive` を取り込む。

## 3. テスト

- handler: `SetActive` は super_admin のみ 200・他ロール 403・body 不正 400（`admin_company_handler_test.go`）
- middleware: 無効会社のユーザーを 403 で弾く / 有効会社は通す / 会社なし super_admin は通す / 会社行なしは通す（`current_user_test.go`）
- frontend: `CompanyRepository.test.ts`（`updateActive` の PATCH）

## 4. 安全策

- **自己締め出し防止**: 会社無効化は super_admin のみが行い、super_admin 自身は会社に属さないため締め出されない。
- 停止は**可逆**（有効化で即復帰）。破壊的なデータ削除は伴わない。
