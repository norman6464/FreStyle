# 会社 / ユーザーのアカウント停止・削除（管理操作）

運営（super_admin）・会社管理者（company_admin）が、会社やユーザーのアカウントを**停止（無効化）**したり削除したりするための管理操作。停止は削除と違い**可逆**で、停止中はログイン/利用を**即時に**不可にする。

> **PR1: 会社アカウントの有効/無効（実装済）** / **PR2: ユーザーの有効/無効 + ソフト削除（実装済）**。

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

## 2. ユーザー（従業員）の有効/無効・論理削除（super_admin / company_admin）

- 画面: `/admin/members`（管理 → 従業員一覧）。各従業員に「有効/無効」状態、「無効化 / 有効化」ボタン、「削除」ボタン（確認ダイアログ付き）。
- API:
  - `PATCH /api/v2/admin/members/:userId/active` body `{ "active": false }`（停止/再開）
  - `DELETE /api/v2/admin/members/:userId`（論理削除）

### 認可（テナント境界 + 自己ロックアウト防止）

`super_admin` は全社の従業員を、`company_admin` は**自社（同一 company_id）の従業員のみ**操作できる。**自分自身は無効化/削除できない**（`ErrCannotManageSelf` → `400 cannot_manage_self`）。別会社は `403 forbidden`、存在しない従業員は `404 member_not_found`。

### enforcement（即時ブロック）

- **無効化**: `users.is_active = false` にすると、`CurrentUser` middleware が次のリクエストで `403 user_disabled` を返す（有効な JWT でも即時に弾く）。
- **論理削除**: `users.deleted_at = NOW()`。`FindByCognitoSub` / `FindByID` は `deleted_at IS NULL` で除外するため、削除済みユーザーは**認証時に自動で `401 user_not_found`** となり、一覧からも消える。

## 3. レイヤー構成（クリーンアーキテクチャ）

```text
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

ユーザー操作の主なファイル:

| 層 | ファイル |
|---|---|
| handler | `admin_member_handler.go`（`SetActive` / `Delete`） |
| middleware | `current_user.go`（ユーザー無効チェック） |
| usecase | `set_member_active_usecase.go`（認可ヘルパ `authorizeMemberManagement`）/ `soft_delete_member_usecase.go` |
| repository | `usecase/repository/user.go`（`UpdateActive` / `SoftDelete`）/ `adapter/persistence/user_repository.go` |
| domain | `backend/internal/domain/user.go`（`IsActive` / 既存 `DeletedAt`） |
| frontend | `AdminMembersPage.tsx` / `hooks/useAdminMembers.ts` / `repositories/AdminMemberRepository.ts` |

> `companies.is_active` / `users.is_active boolean NOT NULL DEFAULT true` を追加（AutoMigrate が本番に列を追加）。sqlc は `SELECT *` のため `make sqlc` 再生成で `IsActive` を取り込む。

## 4. テスト

- handler(company): `SetActive` は super_admin のみ 200・他ロール 403・body 不正 400・会社なし 404（`admin_company_handler_test.go`）
- usecase(member): super_admin は全社可 / company_admin は自社のみ / 別会社 403 / 自分自身 403(self) / 不在 404（`manage_member_usecase_test.go`）
- middleware: 無効会社・無効ユーザーを 403 で弾く / 有効は通す / 会社なし super_admin は通す（`current_user_test.go`）
- frontend: `CompanyRepository.test.ts` / `AdminMemberRepository.test.ts`（`updateActive` PATCH / `remove` DELETE）

## 5. 安全策

- **自己締め出し防止**: 会社無効化は super_admin のみ（会社に属さないため締め出されない）。ユーザー操作は**自分自身を無効化/削除できない**（`ErrCannotManageSelf`）。
- **テナント境界**: company_admin は自社の従業員のみ操作可能（別会社は 403）。
- 停止は**可逆**（有効化で即復帰）。論理削除も `deleted_at` のクリアで復旧可能（データは残る）。
