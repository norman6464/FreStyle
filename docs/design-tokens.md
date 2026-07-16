# デザイントークン（配色・角丸）

frontend の見た目の基本トークンと、その決定経緯。値の正本は
[frontend/src/index.css](../frontend/src/index.css)（CSS 変数）と
[frontend/tailwind.config.js](../frontend/tailwind.config.js)（Tailwind テーマ）。

## 配色 — 灰青ボディ + 白カード（FRESTYLE-118 / 119）

- `--color-surface`（body・ページ背景）= **#EDF2F7**（rgb(237,242,247)・ユーザー指定）
- `--color-surface-1`（カード・パネル）= #FFFFFF、`--color-nav`（ヘッダー / サイドバー）= #FFFFFF
- 背景の灰青と白カードのコントラストで内容へ視線を集める構成。教材閲覧（FRESTYLE-118）で導入し、
  FRESTYLE-119 で body 全体へ統一した（旧「モノクロ最小主義 = 全面純白 + border 境界」から変更）

## 角丸 — 一段控えめのスケール（FRESTYLE-120）

「カードの角が丸すぎる」というユーザー要望で、トークン値をスケールごと一段引き締めた。
**使い分けルールは不変**（どの部品にどのトークンかは変えず、値だけで全画面一括反映）:

| トークン | 用途 | 値 |
|---|---|---|
| `rounded-md` | 小さいインライン要素（コードブロック枠等） | 4px |
| `rounded-lg` | ボタン / インプット / セレクト / タグ / バッジ | 6px |
| `rounded-xl` | カード / パネル / モーダル / フローティングメニュー | 8px（旧 16px） |
| `rounded-2xl` | 特大の囲み | 12px（旧 24px） |
| `rounded-full` | アバター / ピル | 変更なし |
