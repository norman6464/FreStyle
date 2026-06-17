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
