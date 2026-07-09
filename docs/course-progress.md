# コース学習進捗の表示（一覧カード / 詳細ページ）

コースの学習進捗（完了した章数 / 全章数）をどこでどう計算・表示しているかをまとめる。

- 一覧カードへの表示は FRESTYLE-98 で追加
- コース詳細ページの進捗バーは従来から存在（本ドキュメントで仕様を明文化）

## データの流れ

```
GET /api/v2/courses          → 各コースの materialCount（全章数）を含む一覧
GET /api/v2/lesson-progress  → current user の完了行（teachingMaterialId / courseId / completedAt）
```

- **全章数（分母）**: backend の `ListCoursesWithMaterialCountUseCase` が
  `TeachingMaterialRepository.CountByCourseForCompany`（`GROUP BY course_id` の 1 クエリ、N+1 回避）で集計し、
  一覧レスポンスの各要素に `materialCount` として付与する。
  - trainee: published の章のみカウント（詳細ページの分母と一致させる）
  - company_admin / super_admin: 下書き章も含む
- **完了章数（分子）**: frontend の `useCourseCompletionCounts` hook が既存の
  `GET /api/v2/lesson-progress`（current user 固定・1 リクエスト）を courseId 別に集計する。
  backend に新しい進捗集計 API は作らない。

## 表示ルール

| 画面 | コンポーネント | 表示条件 |
|---|---|---|
| コース一覧カード（`CoursesListPage`） | `components/CourseProgressBar` | 受講者（trainee）かつ `materialCount > 0` のコースのみ |
| コース詳細 左パネル（`CourseDetailPage`） | 同上（共有コンポーネント） | 受講者のみ（従来どおり） |

- `CourseProgressBar` は `{completed}/{total}（{pct}%）` のラベル + `role="progressbar"`（aria-valuenow/min/max）の緑バー。
  もともと `CourseDetailPage` 内のローカル実装だったものを `src/components/CourseProgressBar.tsx` へ抽出して共用している
- 管理ロールには進捗を表示しない（完了記録を持たないロールのため。`useCourseCompletionCounts(enabled=false)` は API を叩かない）

## 設計判断・落とし穴

- **完了数 > 全章数になり得る**: 完了記録（`user_lesson_progress`）は章の非公開化・削除時に消えないため、
  完了後に章が減ると分子が分母を上回る。一覧カードでは `Math.min(completed, total)` でクランプして 100% 止まりにしている
- **create/update 応答に materialCount は無い**: `POST/PUT /courses` は従来どおり `domain.Course` を返す。
  `useCourses` が作成時は `materialCount: 0`、更新時は取得済みの値を引き継いで補完する
- **0 件時の null 防御**: 一覧 usecase は 0 件でも空スライスを返す（FRESTYLE-70 と同じ理由）。
  hook 側も `rows ?? []` で防御している

## 関連

- Jira: FRESTYLE-98（一覧カード進捗）/ FRESTYLE-70（null 防御）/ FRESTYLE-68（カテゴリセクション + チップ）
- backend: `internal/usecase/list_courses_with_material_count_usecase.go` /
  `internal/adapter/persistence/teaching_material_repository.go`（`CountByCourseForCompany`）
- frontend: `src/hooks/useCourseCompletionCounts.ts` / `src/components/CourseProgressBar.tsx`
