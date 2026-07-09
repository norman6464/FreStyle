# コース教材（章）の遅延ロードとローディング表示

`/courses/:id`（[CourseDetailPage](../frontend/src/pages/CourseDetailPage.tsx)）での章本文の取得と、
取得中の UX について。

## 遅延ロード方針

教材一覧（左パネル）は本文 `content` を含まない軽量メタデータで取得し、
**選択された章の本文だけ** を都度 `GET /teaching-materials/:id` で取得してキャッシュする
（全章を先読みしない）。状態管理は [useTeachingMaterials](../frontend/src/hooks/useTeachingMaterials.ts)。

- `materials` … 本文なしのメタデータ一覧（`stripContent` で `content` を空に）
- `detailCache` … 取得済みの本文込み教材（`{ [id]: TeachingMaterial }`）
- `selected` … `detailCache[selectedId] ?? null`（未取得なら `null` = 取得中）

## 取得中のローディング表示

章を選んでから本文取得が終わるまで、ローディングを表示してちらつきを防ぐ。

判定は `detailLoading`（fetch フラグ）ではなく **`selectedId != null && !selected && !error`** で行う。
理由: `detailLoading` を立てる effect は描画後に走るため、それを待つと選択直後の 1 フレームだけ
「教材を選択してください」の空状態がちらついてしまう。「章を選んだのに本文が無い = 取得中」と
みなして即座にローディングへ切り替える。

- 学習者（trainee）には記事レイアウトを模した **スケルトン**（`MaterialSkeleton`）を表示し、
  レイアウトの跳ねを抑えて体感速度を上げる
- 編集者（company_admin / super_admin）にはエディタへ遷移するためスピナー（`Loading`）を表示

## エラー時の挙動

`selectMaterial` は選択時に前章で出た取得エラーを先にクリアする。これをしないと別章へ
切り替えても古い `error` が残り、ローディングに戻せない。取得失敗時は `error` がセットされ、
本文領域はローディングを抜けて上部のエラーバナーで通知する。

関連テスト: [useTeachingMaterials.test.ts](../frontend/src/hooks/__tests__/useTeachingMaterials.test.ts)

## 続きから表示（レジューム、FRESTYLE-99）

受講者（trainee）がコース詳細を開くと、**最後に閲覧した章**（無ければ先頭の章）を自動選択して
「教材を選択してください」の空表示を経由せず学習を再開できる。

### 閲覧の記録

- 受講者が章を表示したとき [CourseDetailPage](../frontend/src/pages/CourseDetailPage.tsx) が
  `DashboardRepository.recordChapterView`（`POST /api/v2/teaching-materials/:id/view`）を呼ぶ。
  ベストエフォート（失敗は黙殺）で、`user_chapter_views` の `last_viewed_at` / `view_count` が upsert される
- この記録はダッシュボードの「続きから」基盤データも兼ねる。なお PR #2014 で API と repository は
  用意されていたが frontend からの呼び出し配線が無く、FRESTYLE-99 で初めて配線された

### 復元（自動選択）

- backend: `GET /api/v2/courses/:id/last-viewed` が「そのコース内で last_viewed_at 最大の 1 件」を返す
  （履歴なしは 204。usecase `GetLastViewedChapterUseCase` がコースの閲覧権限を検証）
- frontend: [useChapterResume](../frontend/src/hooks/useChapterResume.ts) が章一覧のロード完了後に
  一度だけ発火し、履歴の章（一覧に無ければ / 履歴なし / 取得失敗なら先頭の章）を `selectMaterial` する

### ガード（重要）

- **コースごとに一度だけ**発火（`resumedCourseRef`）。手動選択済みなら発火済み扱い
- 履歴の取得中にユーザーが手動で章を選んだ場合は**上書きしない**（`selectedIdRef` で解決時に再判定）
- コース切替直後は前コースの `materials` が残っている commit があるため、
  `materials[0].courseId === courseId` になってから発火する
- 管理ロール（company_admin / super_admin）は従来どおり自動選択なし（編集フローを変えない）

関連テスト: [useChapterResume.test.ts](../frontend/src/hooks/__tests__/useChapterResume.test.ts)
