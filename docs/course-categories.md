# コースのカテゴリ（学習領域）と色分け

コース一覧で「色＝学習領域」の連想を作るためのカテゴリ機構（FRESTYLE-67 / Design Doc 承認済）。
コースをグループ単位で色分けし、繰り返し利用で説明なしでも学習者が領域を識別できる UI を目指す。

## 仕組み

- `courses.category`（string / 既定 `''` = 未分類）に**定義済みカテゴリ key** を保存する
- カテゴリは自由入力ではなく**選択式**。自由入力にすると「色＝領域」の一貫性が崩れるため
- 値の正本は backend の [`domain.ValidCourseCategories`](../backend/internal/domain/course.go)。
  frontend の [`constants/courseCategories.ts`](../frontend/src/constants/courseCategories.ts) が
  同じ key に対して表示名（日本語）と色（Tailwind クラス）を持つ

## カテゴリ一覧（8 分類）

| key | 表示名 | 色 |
|---|---|---|
| `dev-basics` | 開発基礎 | 黄 (amber) |
| `backend` | バックエンド開発 | 青 (blue) |
| `architecture` | 設計・アーキテクチャ | 紫 (violet) |
| `database` | データベース | 緑 (emerald) |
| `infra` | インフラ・クラウド | 橙 (orange) |
| `security` | セキュリティ | 赤 (rose) |
| `product` | プロダクト・仕様 | 水色 (cyan) |
| `design` | デザインパターン | 桃 (pink)。コース「デザインパターン入門 (PHP)」向けに FRESTYLE-127 で追加 |

未分類（`''`）は無色 = 従来表示のまま。

## 表示（frontend）

- コース一覧カード: 左端の**色帯**（`border-l-4`）+ **カテゴリ名バッジ**
  - 色だけに依存しない（色覚特性・スクリーンリーダー対応のため必ず文字ラベルを併記）
- コース作成 / 編集フォーム: カテゴリの select（未分類 + 7 分類）

## バリデーション（backend）

- handler: `courseRequest.Category` の `binding:"omitempty,oneof=..."` で宣言的に 400
- usecase: `domain.IsValidCourseCategory` で防衛的に再検証（HTTP 以外の呼び出し元対策）

## カテゴリを増やす / 変えるとき

1. `backend/internal/domain/course.go` の定数 + `ValidCourseCategories` に追加
2. `backend/internal/handler/course_handler.go` の `oneof` リストを同期
3. `frontend/src/constants/courseCategories.ts` に key / label / 色を追加
4. `make openapi`（backend）→ `npm run openapi:generate`（frontend）で spec / 生成型を更新
5. 本ドキュメントの表を更新

## DB への反映

- 列追加は GORM AutoMigrate が ECS 起動時に自動適用（破壊なし）
- 既存 22 コースへのカテゴリ割当は
  [`backend/migrations/0007_course_categories_backfill.sql`](../backend/migrations/0007_course_categories_backfill.sql)
  を `make apply-migration-supabase` で適用済（2026-07-03）
- 教材リポ（frestyle-teaching-materials）の `course.yaml` にも `category` を追記して正本の整合を保つ

### 既存コースの割当

| カテゴリ | コース id |
|---|---|
| dev-basics | 1, 2, 3, 4 |
| backend | 5, 6, 15, 16, 18, 22 |
| architecture | 7, 8, 9, 14, 21 |
| database | 10 |
| product | 11, 12 |
| infra | 13, 17, 19 |
| security | 20 |
