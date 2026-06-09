---
status: approved
area: frontend
date: 2026-06-09
---

# 0003 ブランドカラー（青）でアクションボタンを統一

> 関連: フロントエンド `frontend/tailwind.config.js` / `frontend/src/index.css`

## 背景

アクション系ボタンの背景が `primary-500`(#5C5850 の暗い taupe) で、ロゴ（青の二重矢印 `#2E7DF6`）と色が一致せず、ブランド印象がちぐはぐだった。ボタンが「濃い橙色」のように見えるという指摘もあった。

## 決定

FreStyle のブランドカラーをロゴの青 `#2E7DF6` と定義し、**アクション系ボタンの色だけ**これに統一する。

- tailwind に `brand` 青スケール（`#2E7DF6` を 500、hover=600 / active=700 を濃く）を追加。
- ボタン背景を `bg-primary-*` → `bg-brand-*` に置換。対象は「ボタン特有のパターン（`hover:bg-primary-600` を伴う要素）」に限定し、機械的な誤爆を防いだ。
- `primary`(taupe) スケールは **残す**。チャート / プログレスバー / バッジ / フォーカスリングなど **非ボタンのアクセント** には引き続き `primary` を使う。

## なぜ primary を青に塗り替えなかったか

`primary-500` はボタン以外（ProgressBar の塗り、スコアバー、Avatar、番号バッジ等）でも広く使われている。全部を青にするとボタン以外まで色が変わるため、要望（ボタンのみ）に合わせて `brand` を新設し、ボタン要素だけを差し替えた。

## 適用範囲

`.btn-primary`(共通クラス) / `PrimaryButton`(共通コンポーネント) と、各画面の primary アクションボタン（ヘッダーのログイン、ConfirmModal の確認、各ページの主要ボタン等）。`isDanger` の確認ボタンは従来通り赤系のまま。
