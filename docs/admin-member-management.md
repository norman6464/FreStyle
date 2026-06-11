# 従業員管理（従業員一覧 + AI 利用可否の個別設定）

company_admin / super_admin が、自社の従業員一覧を見て、各従業員の AI チャット利用可否を**個別に**設定できる機能。サイドバー「管理」直下の「従業員一覧」（`/admin/members`）。

## AI 利用可否の優先順位

AI チャットを使ってよいかは [`usecase.AiChatEnabledForUserUseCase`](../backend/internal/usecase/ai_chat_access_usecase.go) が判定する。優先順位:

1. `company_admin` / `super_admin` → 常に許可（自社設定確認のため）
2. 会社未所属（`company_id = null`）→ 許可
3. **`users.ai_chat_enabled`（個別上書き）が設定済みなら、それを優先**（`true`=許可 / `false`=禁止）
4. それ以外 → 会社の `companies.ai_chat_enabled_for_trainees`（一括設定）に従う

つまり **「会社一括設定」は従来どおり残り、「個別設定」がそれを上書き**する。個別設定を「会社設定に従う」に戻すと `ai_chat_enabled` が `NULL` になり、再び会社一括設定に従う。

## データモデル

- `domain.User.AiChatEnabled *bool`（列 `users.ai_chat_enabled`、nullable）。`nil`=会社設定に従う / `true` / `false`=個別上書き
- 列追加は GORM AutoMigrate が起動時に行う（追加系なので無停止）

## API

| メソッド | パス | 役割 |
|---|---|---|
| GET | `/api/v2/admin/members` | 自社の従業員一覧（company_admin / super_admin のみ） |
| PATCH | `/api/v2/admin/members/{userId}/ai-access` | 従業員の AI 利用可否を個別更新（`{ "enabled": true\|false\|null }`） |

認可: 別会社の従業員は更新できない（`UpdateMemberAiAccessUseCase` がテナント境界をチェック、越権は 403）。

## 実装

- backend: `list_company_members_usecase.go` / `update_member_ai_access_usecase.go` / `admin_member_handler.go` / `routes_admin.go`。repository は `UserRepository.ListByCompanyID`（sqlc）+ `UpdateAiChatEnabled`（GORM、`nil` で `NULL` に戻す）
- frontend: `AdminMembersPage` / `useAdminMembers` / `AdminMemberRepository`。一覧表で各従業員の AI 利用を「会社設定に従う / 有効 / 無効」のセレクトで設定（trainee のみ対象）
- **一覧の検索**: 一覧上部の検索ボックスで **氏名・メールアドレス・役割**を部分一致（大文字小文字無視）でクライアント側フィルタ（`AdminMembersPage` の `useMemo`）。一覧は自社分のみで件数が少ないため、追加の API は持たずクライアントで絞り込む。一致が無いときは「〜に一致する従業員がいません」を表示。
