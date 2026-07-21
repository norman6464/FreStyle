# フロントエンドの到達不能コード棚卸し（FSD 移行 Phase 5 の前処理）

FSD 移行（FRESTYLE-154）で `entities` / `features` レイヤーを切る前に、**どの画面からも到達できないモジュール**を洗い出して削除した記録。移行は「置き場所を決める」作業なので、死んだコードを先に落としておかないと、実体のない Slice を作ってしまい、以後ずっと嘘の構造を維持することになる。

## なぜ移動の前にやるのか

到達不能なモジュールをそのまま `entities/` に移すと、次の 2 つが同時に起きる。

1. **存在しないビジネス概念の Slice ができる**。たとえば「練習シナリオ」は現行プロダクトに存在しないが、型とコンポーネントだけが残っていた。これを移すと `entities/practice-scenario` という嘘の Slice が生まれる。
2. **消せなくなる**。移動後は「移行で作ったものだから意味があるはず」と読めてしまい、削除の判断コストが上がる。

## 見つけ方（再現手順）

`import` 文の解析で「誰からも import されていないモジュール」を求め、**不動点まで反復**する。A だけが B を使っていて A が死んでいるなら B も死ぬ、という連鎖を拾うため 1 回では足りない。

```bash
# frontend/src で実行
for f in $(find components constants utils lib hooks repositories store \
            \( -name '*.ts' -o -name '*.tsx' \) | grep -v __tests__ | sort); do
  b=${f##*/}; base=${b%.*}
  out=$(grep -rn "from '[^']*/$base'\|from '\./$base'\|import('[^']*/$base')\|vi\.mock('[^']*/$base'" \
        --include='*.ts' --include='*.tsx' . 2>/dev/null)
  # 自分自身のテストからの参照は「生きている」根拠にしない
  live=$(echo "$out" | grep -v "__tests__/$base\.test\." | grep -c .)
  [ "$live" -eq 0 ] && echo "DEAD $f"
done
```

これを「死んだと判定済みのファイルからの参照も根拠にしない」条件付きで、新たに見つからなくなるまで繰り返す。

### 判定の 2 つのルール

**自分のテストからの参照は生存の根拠にしない。** これを入れないと、テストだけが残っている死んだコンポーネントが永久に生き残る。実際に `ScoreCard` / `ScenarioCard` などはテストが通っていたが、どの画面からも描画されていなかった。

**すでに死んだモジュールからの参照も根拠にしない。** これが不動点反復の本体。今回は `ScoreCard` → `AxisScoreBar` → `Card` / `scoreColor` と 3 段の連鎖があった。

### ハマったところ

**BSD `sed` は `\?` を解釈しない。** macOS で `basename "$f" | sed 's/\.tsx\?$//'` と書くと拡張子が落ちず、`base` が `Avatar.tsx` のままになる。結果 `from '.../Avatar.tsx'` を探すことになり **1 件もヒットせず、全ファイルが「死んでいる」と誤判定される**。全部が DEAD になったら、まずこれを疑う。上のスクリプトでは shell の `${b%.*}` を使って回避している。

**ディレクトリの `index.ts` は別扱い。** `store/index.ts` は `from '@/store'` で import されるため、`from '.../index'` のパターンには掛からない。誤検出するので個別に確認する。

**型は tsc に判定させる。** `types/index.ts` は 1 ファイルに全型が同居しており、**ファイル内で他の型から参照されている**ものがある（例: `ExerciseSubmissionStats` は生きている `MasterExerciseWithStatus` の一部）。grep で「外部参照 0 件」だけを見ると消してはいけない型まで巻き込む。候補をまとめて消してから `tsc --noEmit` を回し、エラーが出たものだけ戻すのが確実。

## 消したもの

### 旧プロダクト（英会話練習）の残骸

FreStyle は現在 IT エンジニア研修プラットフォームだが、前身のプロダクトのコードが**互いを参照し合う閉じた部分グラフ**として丸ごと残っていた。**Go バックエンド側に対応する domain は 1 つも存在しない**（`backend/internal/` を `grep` して 0 件を確認済み）。

- コンポーネント: `ScenarioCard` / `ScoreCard` / `ScoreImprovementAdvice` / `AxisScoreBar` / `Card` / `CardHeading`
- 定数: `scenarioLabels` / `axisAdvice` / `axisScenarioMap` / `conversationTemplates`
- ユーティリティ: `scoreColor` / `skillRadarHelpers` / `calendarHelpers`
- 型: `PracticeScenario` / `ScoreCard` / `AxisScore` / `WeeklyChallenge` / `Ranking` / `FavoritePhrase` / `DailyGoal` / `ReminderSetting` / `SharedSession` ほか

これらの JSDoc には「Go backend `domain.PracticeScenario` と 1:1」といった記述が残っていたが、その domain 自体がすでに存在せず、**コメントが嘘になっていた**。

### 使われなくなった現行機能の部品

- `DifficultyFilter` / `difficultyStyles`（PR #1641「コア機能のみに整理」で練習系を削除した際に呼び出し元が消えたが、ファイルだけ残っていた。現在の演習一覧は難易度フィルタを持たず、`ExerciseListPage` 内で定義した `DifficultyBadge` を表示するだけ）
- `SortSelector`（`NoteSortMenu` に置き換え済み。`NoteSortMenu` 自身のコメントに「旧: タブ風の SortSelector」と書かれていた）
- `EmojiPicker` / `emojiData`
- `useBookmark` / `useDebounce` / `useRecentNotes` / `useSessionNote` / `useTableOfContents`
- `EmbedRepository` / `NoteImageRepository`
- `noteContentFormat`

### store に登録されていない slice

`store/flashSlice.ts` は **`configureStore` の `reducer` に登録されていなかった**。つまりマウントされておらず、仮に dispatch しても何も起きない状態だった。フラッシュメッセージは現在 `app/providers/ToastProvider` が担っている。

## 結果

| 項目 | 変更前 | 変更後 |
|---|---|---|
| 削除モジュール | — | 27 |
| 削除テストファイル | — | 18 |
| テストファイル数 | 164 | 146 |
| テスト件数 | 1346 | 1175 |
| Statements | 86.78% | 86.01% |
| Branches | 81.23% | 80.74% |

テスト件数の減少はすべて削除したモジュールに対するテストで、**残したテストの assert は 1 つも変更していない**。カバレッジは下がっているが、これは「よくテストされた死んだコード」が分母から抜けたため。閾値（Statements 85 / Branches 78 / Functions 80 / Lines 85）はすべて維持している。

## 関連

- FRESTYLE-154（FSD 移行 Design Doc）
- FRESTYLE-162（この棚卸し）
- `frontend/src/shared/README.md`（`shared` レイヤーの使い方）

---

## 追記: Phase 5b-2 での取りこぼし（2026-07-20）

上記の不動点検出のあとにも 1 件見つかった。**`BookmarkRepository`（シナリオのブックマーク）** で、参照は自分自身のテストのみ、バックエンドに対応するエンドポイントも存在しなかった（`SCENARIO_BOOKMARKS` ごと削除）。

見つかったきっかけは、entity へ振り分けるときに「これは*何の*ブックマークか」を確認したこと。ファイル名が `BookmarkRepository` だったので当初は `entities/course` に置こうとしたが、中身が `SCENARIO_BOOKMARKS` と `freestyle_scenario_bookmarks`（削除済みの練習シナリオ機能）だった。

**教訓**: 自動検出は「誰からも参照されていないか」しか見ない。**名前が汎用的なせいで生き残るコード**は、置き場所を決めるときに中身を読んで初めて気付く。移行は棚卸しの機会でもある。
