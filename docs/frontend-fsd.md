# フロントエンドの Feature-Sliced Design（FSD）

`frontend/src/` は **Feature-Sliced Design** で構成する。AI コーディングとの相性（「この関心はどこにあるか」が一意に決まる）を理由に採用し、FRESTYLE-154 で移行を完了した。旧構造（`components/` `hooks/` `repositories/` `utils/` `constants/` `lib/` `store/` `types/`）は撤去済み。

このドキュメントは一次情報ではない（規約の正本は `.claude/CLAUDE.md` §2.5）。ここには**判断の根拠と移行で得た知見**を残す。

## レイヤー

上ほど上位。**import は下向きの一方通行**。

```
app > pages > widgets > features > entities > shared
```

| レイヤー | 責務 | 例 |
|---|---|---|
| `app` | エントリ・Provider・ルーティング・store 組み立て | `app/index.tsx` / `app/App.tsx` / `app/providers/*` / `app/store` |
| `pages` | 1 画面 = 1 Slice。その画面専用の hook / component を同居 | `pages/courses` / `pages/ask-ai` / `pages/settings` |
| `widgets` | 複数機能を組み合わせた自立 UI ブロック | `widgets/app-shell`（ヘッダ + サイドバー + コマンドパレット） |
| `features` | 再利用されるユーザー操作 | `features/auth`（ログイン / ログアウト / 認証状態取得） |
| `entities` | ビジネス上の「もの」 | `course` / `exercise` / `user` / `note` / `ai-chat` / `notification` / `company` / `invitation` / `member` / `audit` / `learning-report` |
| `shared` | ビジネスを知らない再利用資産 | `shared/ui`（UI キット）/ `shared/api`（axios）/ `shared/lib`（汎用 hook・関数・typed Redux hooks）/ `shared/config` |

`app` と `shared` は Slice を持たず、セグメント（`ui` `api` `model` `lib` `config`）が直下に来る。他のレイヤーは `<layer>/<slice>/<segment>` の 3 階層。

## ルール（境界 lint が CI で `error` 強制）

`frontend/eslint.config.js` の `no-restricted-imports` が以下を弾く（`eslint --max-warnings 0`）。

1. **下向きの一方通行**: 自分と同じか上の層は import できない。
2. **app ↔ shared だけ相互 import 可**（FSD 公式の例外）。typed Redux hooks（`shared/lib/store`）が `RootState`(app/store) を参照するため必要。
3. **Public API 経由**: 各 Slice は `index.ts` の名前付き re-export だけを外へ出す（`export *` 禁止）。Slice 内部は相対パスで参照し、**自分の barrel を読まない**。
4. **entity 同士は `@x` 記法のみ**: `entities/<相手>/@x/<自分>` で参照される側が「誰に何を見せるか」を宣言する。現在の唯一の例は `entities/course/@x/user`（`UserDashboard` が `UserChapterView` を参照）。

## 配置の判断基準

移行で最も時間を使ったのは「どこに置くか」。実際に迷った判断を残す。

- **単一画面からしか使われないなら `pages/<slice>/model`（hook）・`pages/<slice>/ui`（component）**。FreStyle の hook はほぼ「1 画面 = 1 hook」で、複数画面で共有される横断機能は少なかった（認証と AI チャットくらい）。`features` は 2 画面以上で共有される操作に限る。
- **「どのプロジェクトにコピーしても使えるか」で shared か上位かを決める**。使えるなら `shared`、FreStyle 固有の意味を含むなら上の層。
- **entity に密結合した hook は entity の model に置く**（例: `useNoteEditor` は `Note` 型を扱い notes / course-detail / NoteMarkdownEditor で共有 → `entities/note/model`）。
- **read model は独立 entity にしない**（例: `UserDashboard` は「そのユーザーの学習集計」なので `entities/dashboard` を作らず `entities/user`）。
- **複数 entity が使う純粋 UI は shared**（例: `MarkdownView` / `CodeBlock` / `LanguageBadge` は AI チャットと演習・コースの両方が使う → `shared/ui`）。片方の entity に置くともう片方から Slice 間 import になる。
- **状態を購読するものと見た目だけのものを分ける**（例: `Toast` は見た目だけ → `shared/ui`。`ToastContainer` は `useToast` で購読 → `app/providers`）。

## Redux store の置き方（FSD で悩みやすい点）

- **store 本体（`configureStore` / `RootState` / `AppDispatch`）は `app/store`**。各 slice の reducer を集約するのは app の責務。
- **型付き hooks（`useAppSelector` / `useAppDispatch`）は `shared/lib/store`**。`RootState` を `@/app/store` から import する（app ↔ shared 相互 import の例外を使う）。
- pages / widgets / features は `shared` の typed hooks 経由で状態を読む。**`RootState`(app) を直接 import しない**ので逆流しない。
- slice の実体は entity が持つ（`entities/user/model/authSlice`）。

## 移行の実績（Phase 0〜7）

すべて「移動 + import 追従で振る舞いを変えない」を不変条件に、1 フェーズ = 1 PR で直列に進めた。各フェーズで `tsc` / `eslint --max-warnings 0` / `vitest`（assert 不変）/ `build` / Playwright を緑にしてからマージ。

| Phase | 内容 | チケット |
|---|---|---|
| 0 | 足場（`@` エイリアス 3 か所・境界 lint（warn）・空レイヤー） | FRESTYLE-155 |
| 1 | `app` 層（エントリ / ルーティング / Provider / 全体スタイル） | FRESTYLE-156 |
| 2a / 2b | UI キット → `shared/ui` / API・ユーティリティ・設定 → `shared` | FRESTYLE-157 / 158 |
| 3 | `widgets` 層 | FRESTYLE-159 |
| 4 | 28 画面を `pages` の Slice へ | FRESTYLE-160 |
| （前処理） | 到達不能コードの削除（旧プロダクトの残骸） | FRESTYLE-162 |
| 5a | 汎用 UI → `shared/ui` | FRESTYLE-163 |
| 5b-1 / 5b-2 | `entities`（course/exercise → 残り全 entity・`types/index.ts` 撤去） | FRESTYLE-164 / 165 |
| 6a〜6h | hook 47 個を用途別に振り分け（shared / pages/model / widgets / features / entities）・store 再配置・`hooks/` `components/` 撤去 | FRESTYLE-166〜175 |
| 7 | 境界 lint を warn → **error** 昇格・規約と docs の更新 | FRESTYLE-176 |

## ハマりどころ（再発防止）

- **リネームの大文字小文字**: `XxxRepository.ts` → `xxxRepository.ts` のような camelCase 化で、テストの相対 import が旧名のままだとモジュールが二重ロードされる。**macOS は case-insensitive なので通るが Linux CI で落ちる**。カバレッジ表に同一ファイルが 2 行出るのが検知の手がかり。
- **barrel とカバレッジ分母**: vitest のカバレッジ分母は import されたファイルのみ。Public API（barrel）を作ると、それを import したテストが Slice 内の全ファイルを読み込み、未テストのファイルが分母に入って閾値割れする。**テストは barrel ではなく深いパスで mock する**（`vi.mock` の粒度も保てる）。
- **重いモジュールは barrel に載せない**: `CodeEditor`（monaco）を `shared/ui/index.ts` に re-export すると、`@/shared/ui` を import した全画面が monaco を巻き込みコード分割が壊れる。深いパスで直接 import する。
- **境界 lint の `@x` 例外は regex で書く**: `group`（gitignore 記法）は「親を除外すると子を再包含できない」制約があり `!@/entities/*/@x/*` が効かない。否定先読みの `regex` オプションを使う。
- **一括 import 置換のパターン漏れ**: `@/hooks/X` と `../../hooks/X` は拾えても **1 段の `../X`** を取りこぼしやすい。移動していないファイルのテストでも、移動した対象を `vi.mock` していれば追従が必要。
- **名前が汎用的な死んだコード**: 自動の到達不能検出は「誰からも参照されていないか」しか見ない。`BookmarkRepository` のように名前が汎用的だと生き残る。配置を決めるとき中身を読んで初めて気付く。

## 関連

- 規約の正本: `.claude/CLAUDE.md` §2.5
- レイヤー別 README: `frontend/src/entities/README.md` / `frontend/src/shared/README.md`
- 到達不能コード棚卸し: `docs/frontend-dead-code-audit.md`
- 親 Design Doc: FRESTYLE-154
