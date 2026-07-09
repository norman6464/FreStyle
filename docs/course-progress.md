# コース学習進捗の表示（一覧カード / 詳細ページ）

コースの学習進捗（完了した章数 / 全章数）をどこでどう計算・表示しているかをまとめる。

- 一覧カードへの表示は FRESTYLE-98 で追加
- コース詳細ページの進捗バーは従来から存在（本ドキュメントで仕様を明文化）

## データの流れ

進捗の分母・分子とも **backend がコース一覧 API のレスポンスに同梱**する。
frontend は一覧表示のために追加のリクエストを送らない。

```
GET /api/v2/courses → CourseWithProgress[]（Course + materialCount + completedCount）
```

`ListCoursesWithProgressUseCase` が 1 回の Execute で次を集計する:

- **materialCount（分母 = 章数）**: `TeachingMaterialRepository.CountByCourseForCompany`
  （`GROUP BY course_id` の 1 クエリ、N+1 回避）。trainee は published のみ / admin 系は下書き込み
- **completedCount（分子 = 完了章数）**: `LessonProgressRepository.CountCompletedByUserGroupedByCourse`。
  `user_lesson_progress` を `teaching_materials` に **JOIN して「現存する published 章」のみ**数える。
  管理ロールは完了記録を持たないため集計自体をスキップ（completedCount = 0）

## 分子を JOIN で数える理由（重要な設計判断）

完了記録（`user_lesson_progress`）は章の**非公開化・削除では消えない**。
生の完了行数を分子にすると「幽霊完了行」で進捗が過大表示され、
コース詳細ページ（表示中の章一覧と完了記録の積集合で分子を計算）と数値が食い違う。
JOIN で「現存する published 章の完了行」だけを数えることで、

- 分子 ≤ 分母が常に成立する
- 一覧カードと詳細ページの進捗が同じ意味論になる

## 表示ルール

| 画面 | コンポーネント | 表示条件 |
|---|---|---|
| コース一覧カード（`CoursesListPage`） | `components/CourseProgressBar` | 受講者（trainee）かつ `materialCount > 0` のコースのみ |
| コース詳細 左パネル（`CourseDetailPage`） | 同上（共有コンポーネント） | 受講者のみ（従来どおり） |

- `CourseProgressBar` は `{completed}/{total}（{pct}%）` のラベル + `role="progressbar"`（aria-valuenow/min/max）の緑バー。
  もともと `CourseDetailPage` 内のローカル実装だったものを `src/components/CourseProgressBar.tsx` へ抽出して共用している
- コンポーネント内部で `completed` を `[0, total]` にクランプする（呼び出し元のデータ差に対する防御）

## そのほかの実装メモ

- `POST/PUT /courses` の応答は従来どおり `domain.Course`（進捗フィールド無し）。
  `useCourses` が作成時は 0、更新時は取得済みの値を引き継いで補完する
- 一覧 usecase は 0 件でも空スライスを返す（FRESTYLE-70 の null 防御と同じ理由）
- 旧 `CourseUseCase.List` は一覧が進捗集計を伴うようになったため削除し、
  一覧の可視性ロジック（company / role フィルタ）は `ListCoursesWithProgressUseCase` に一本化した

## 関連

- Jira: FRESTYLE-98（一覧カード進捗）/ FRESTYLE-70（null 防御）/ FRESTYLE-68（カテゴリセクション + チップ）
- backend: `internal/usecase/list_courses_with_progress_usecase.go` /
  `internal/adapter/persistence/teaching_material_repository.go`（`CountByCourseForCompany`）/
  `internal/adapter/persistence/lesson_progress_repository.go`（`CountCompletedByUserGroupedByCourse`）
- frontend: `src/components/CourseProgressBar.tsx` / `src/types/index.ts`（`CourseWithProgress`）
