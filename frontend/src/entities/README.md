# entities 層

ビジネス上の「もの」（コース・演習・ノート・ユーザーなど）を表す層。**その entity の
データ取得（`api`）・型（`model`）・単体表示（`ui`）**をひとまとめにする。

## Slice 一覧

| Slice | 担当 |
|---|---|
| `course` | コース・教材（章）・章の進捗 |
| `exercise` | コーディング演習・提出・採点結果 |
| `user` | ユーザー・プロフィール・認証状態・学習ダッシュボード |
| `note` | ノート・セッションノート |
| `ai-chat` | AI チャットのセッションとメッセージ |
| `notification` | 通知 |
| `company` | 会社・会社設定・利用申請 |
| `invitation` | 招待 |
| `member` | 会社メンバー |
| `audit` | 監査ログ |
| `learning-report` | 学習レポート |

## セグメント

```
entities/<slice>/
  api/      … リポジトリ（axios 呼び出し）
  model/    … 型・Redux slice
  ui/       … その entity 単体の表示コンポーネント
  lib/      … その entity 固有の純粋関数
  config/   … その entity 固有の定数
  index.ts  … Public API
```

## ルール

**Public API 以外から import しない。** 外からは `@/entities/course` だけを参照し、
`@/entities/course/api/courseRepository` のような内部パスは使わない。
`index.ts` は名前付きで re-export する（`export *` は公式が禁止）。

**Slice 内は相対パスで参照する。** `entities/note/api/noteRepository.ts` が同じ Slice の
型を使うときは `'../model/types'`。自分の barrel（`@/entities/note`）を参照すると
循環し、境界 lint も違反として検出する。

**entity 同士は直接 import できない。** 同一レイヤーだから。どうしても必要なら
次の `@x` を使う。

## `@x` 記法（entity 同士のクロス import）

FSD 公式が entities 層に限って認めている例外。**参照される側**が「誰に何を見せるか」を
明示する。

現在の唯一の例: `entities/course/@x/user.ts`

`UserDashboard`（`entities/user`）は「最近見た章」を持つため `UserChapterView`
（`entities/course`）を参照する。データ構造上の依存で、どちらかに寄せると
「章の型を user が定義する」か「ダッシュボードを course が知る」かの不自然さが出る。

```ts
// entities/course/@x/user.ts — user にだけ公開する
export type { UserChapterView } from '../model/types';

// entities/user/model/types.ts — @x 経由でのみ参照
import type { UserChapterView } from '@/entities/course/@x/user';
```

**増えたら Slice の切り方を疑う。** `@x` が増えるのは「その 2 つは実は 1 つの
entity では」というサイン。

境界 lint（`eslint.config.js`）は `@x` を正規表現で例外にしている。`group` の
gitignore 記法では「親ディレクトリが除外されていると子を再包含できない」ため
`!@/entities/*/@x/*` が効かず、否定先読みの `regex` で書く必要があった。

## 置き場所に迷ったときの実例

| 対象 | 置き場所 | 理由 |
|---|---|---|
| `TeachingMaterial`（教材＝章） | `entities/course` | コースの章であり単独では存在しない。別 Slice にすると `courseRepository` が型を参照した時点でクロス import になる |
| `UserDashboard` | `entities/user` | 独立した業務エンティティではなく「そのユーザーの学習集計」という read model。`entities/dashboard` を作ると実体のない Slice が増える |
| `MarkdownView` / `CodeBlock` | **shared/ui** | AI チャットと演習の両方が使う汎用レンダラ。ai-chat に置くと exercise からのクロス import になる |
| `LanguageBadge` / `LanguageIcon` | **shared/ui** | コースと演習の両方が使う。同上 |
| `MessageInput` | **features**（Phase 6 予定） | 「メッセージを送る」というユーザー操作であって entity の表示ではない |

## 移行状況

FSD 移行（FRESTYLE-154）の Phase 5b-1（FRESTYLE-164）と 5b-2（FRESTYLE-165）で構築した。
`hooks/` はまだ旧ディレクトリにあり、Phase 6 で `features/` へ振り分ける。
