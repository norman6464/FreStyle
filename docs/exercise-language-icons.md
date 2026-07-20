# コード学習の言語アイコン（Devicon）

コード学習の言語選択カード（FRESTYLE-152）で使う言語ロゴ。実体は `frontend/public/lang/<key>.svg`。

## 出典・ライセンス

[Devicon](https://devicon.dev/) の `*-original` アイコンをそのまま配置している。

- リポジトリ: https://github.com/devicons/devicon
- ライセンス: MIT License（Copyright (c) 2015 konpa）
- 取得元: `https://raw.githubusercontent.com/devicons/devicon/master/icons/<name>/<name>-original.svg`

各言語ロゴの商標は各権利者に帰属する。ここでは「その言語の教材である」ことを示す
識別目的でのみ使用している。

## なぜ npm パッケージではなく vendoring か

`devicon` パッケージは全言語ぶん（数百ファイル）を含み node_modules が重くなる。
本アプリで必要なのは下記 7 個だけなので、SVG を直接置いてバンドルと依存を軽く保つ。
CDN 参照にしないのは、外部ドメイン障害でアイコンが消えるのを避けるため。

## 収録アイコン

| ファイル | 言語 |
|---|---|
| `php.svg` | PHP |
| `go.svg` | Go |
| `javascript.svg` | JavaScript |
| `typescript.svg` | TypeScript |
| `git.svg` | Git |
| `bash.svg` | Bash / Linux |
| `docker.svg` | Docker |

## 追加するとき

1. 上記 URL から `<name>-original.svg` を取得して `<key>.svg` として置く（`<key>` は
   backend の `master_exercises.language` の値）。
2. [`frontend/src/constants/exerciseLanguages.ts`](../frontend/src/constants/exerciseLanguages.ts) に
   エントリを追加する（label / badgeClass）。アイコンは key から自動解決される。

## 画面での使われ方

- [`LanguageIcon`](../frontend/src/components/LanguageIcon.tsx) が `/lang/<language>.svg` を `<img>` で描画する。
  未知の言語・読み込み失敗時は Heroicons の汎用コードアイコンにフォールバックするので、
  アイコンが無くてもカードは壊れない。
- 汎用 UI アイコン（ダッシュボードのグラフ・カレンダー等）は引き続き **Heroicons** を使う。
  ブランドロゴ＝アイコンセット（Devicon / Simple Icons）、汎用 UI＝Heroicons、という住み分け。
  オリジナルの独自アイコンを起こす場合のみ Figma 等でベクター作成して SVG で持ち込む。
