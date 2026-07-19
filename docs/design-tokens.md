# デザイントークン（配色・角丸）

frontend の見た目の基本トークンと、その決定経緯。値の正本は
[frontend/src/index.css](../frontend/src/index.css)（CSS 変数）と
[frontend/tailwind.config.js](../frontend/tailwind.config.js)（Tailwind テーマ）。

## 配色 — 白ボディ + 白カード。教材閲覧だけ灰青（FRESTYLE-118 / 119 → 147）

- `--color-surface`（body・ページ背景）= **#FFFFFF**。カードは border 1px + 影で白 body から浮かせる
- `--color-surface-1`（カード・パネル）= #FFFFFF、`--color-nav`（ヘッダー / サイドバー）= #FFFFFF
- `--color-reading-surface`（教材コース閲覧 `/courses/:id` の読み物背景）= **#F4F6F9**（灰青）。
  背景の灰青と白カードのコントラストで本文へ視線を集める（教材閲覧 FRESTYLE-118 の配色）
- 経緯: FRESTYLE-119 で body 全体を灰青(#F4F6F9)へ統一したが、FRESTYLE-147 でユーザー要望により
  **教材コース閲覧以外は白へ戻した**（灰青は `--color-reading-surface` として教材閲覧に限定）。
  白 body での「カードを border + 影で浮かせる」扱いは元の設計どおり

## 角丸 — 一段控えめのスケール（FRESTYLE-120 → 124 でさらに引き締め）

「カードの角が丸すぎる」というユーザー要望で、トークン値をスケールごと引き締めた
（FRESTYLE-120 で 16px→8px 等、FRESTYLE-124 で「まだ丸い」との再要望を受けもう一段）。
**使い分けルールは不変**（どの部品にどのトークンかは変えず、値だけで全画面一括反映）:

| トークン | 用途 | 値 |
|---|---|---|
| `rounded-md` | 小さいインライン要素（コードブロック枠等） | 3px（旧 4px） |
| `rounded-lg` | ボタン / インプット / セレクト / タグ / バッジ | 4px（旧 6px） |
| `rounded-xl` | カード / パネル / モーダル / フローティングメニュー | 6px（旧 8px ← 16px） |
| `rounded-2xl` | 特大の囲み | 8px（旧 12px ← 24px） |
| `rounded-full` | アバター / ピル | 変更なし |
