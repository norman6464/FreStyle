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

## 完了トグルの固定表示（FRESTYLE-100）

受講者向け本文（ReadOnlyDetail）のメタ行（最終更新日・目次トグル・完了トグル）は
`sticky top-0` + ページ背景色（`bg-surface`）+ 下ボーダーで、スクロールコンテナの
上部に常時残る。本文を読み進めている途中でも、先頭へ戻らずに「完了にする」を押せる。

- sticky の基準はウィンドウではなく ReadOnlyDetail の `overflow-y-auto` コンテナ
  （body は overflow:hidden。目次 aside の `sticky top-6` と同じ仕組み）
- z-index は 10（ツールチップ z-30 / モーダル z-50 の下）
- 本文末尾の大きい完了ボタン + 「次の章へ」は読了導線としてそのまま残している
- 連打の二重送信は useLessonProgress の in-flight ガードで防止済み

## 次のコースへの導線（FRESTYLE-102）

最終章（次の章が無い章）の本文末尾では、「次の章へ」の代わりに
**「次のコースへ: {次コース名}」** を表示し、一覧に戻らず次のコースへ直行できる。

- 「次のコース」= コース一覧 API の並び順（sort_order 昇順 → id 昇順）で現在のコースの次
  （[useNextCourse](../frontend/src/hooks/useNextCourse.ts)。受講者のみ・取得失敗や
  並び順で最後のコースでは導線を出さないだけ）
- 遷移先ではレジューム（FRESTYLE-99）が働き、履歴の無いコースなら 1 章目が自動表示される
  = 「読了 → 次のコースの 1 章目」がクリック 1 回でつながる
- 最終章以外の章は従来どおり「次の章へ」（変更なし）

## 本文タイポグラフィ（FRESTYLE-115）

教材の本文が詰まって見えるというユーザー要望（学習教材サイト一般の「読み物」余白のスクリーンショット提示あり）への対応。
`.prose.course-prose`（教材本文専用スコープ — ノート / AI チャットの prose には影響しない）で余白を上書きする:

- 本文 16px・行間 2.0、段落間 1.4em（prose-sm プリセットは 15px・行間 1.71 で読み物には窮屈だった）
- h2 = 話題の切り替わり: 上 3.2em + 下線区切り（ニュートラルな `--color-surface-3`。モノクロ最小主義のトーン維持）
- h3 上 2.4em / リスト項目間 0.6em / コードブロック・表・画像の上下 1.8〜2em
- ページ外側の余白も拡大（`px-6 sm:px-10 py-8 sm:py-10`）、章タイトルは text-3xl

コードブロックの文字は 0.875rem・行間 1.7 に据え置き（コードは詰まっていた方が読みやすい）。
