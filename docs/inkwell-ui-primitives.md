# inkwell UI プリミティブ（触感的コンポーネント群）

押下時の波紋・標高シャドウ・浮き上がるラベルを持つ触感的な UI プリミティブを、外部 UI ライブラリを使わず
**Tailwind CSS だけ**で実装したもの。青系プライマリ（#1976d2）・Roboto・角丸 4px・大文字ボタンといった
定番の Material 風ルックを、バンドルを増やさずに提供する。

> 「なぜこの見た目にするのか」という設計判断・採用理由は private リポ `frestyle-pdm` の
> `docs/architecture/inkwell-ui-primitives.md` に記載（本リポは使い方に集中する）。

## 置き場所と使い方

- 実装: `frontend/src/components/inkwell/`
- カタログ（見た目確認）: `/dev/inkwell`（認証不要ルート）
- アプリ本体のテーマ（`brand-*` / `taupe-*` のモノクロ最小構成）とは**分離**。既存 `components/ui/` は不変。

```tsx
import { useState } from 'react';
import {
  InkwellButton,
  InkwellTextField,
  InkwellCard,
  InkwellCardContent,
  InkwellCheckbox,
  InkwellSwitch,
} from '../components/inkwell';

function Example() {
  const [ok, setOk] = useState(false);
  const [on, setOn] = useState(true);
  const hasError = false;
  return (
    <>
      <InkwellButton color="primary">保存</InkwellButton>
      <InkwellButton variant="outlined">キャンセル</InkwellButton>
      <InkwellTextField label="メール" helperText="社内アドレス" error={hasError} />
      <InkwellCard elevation={2}>
        <InkwellCardContent>本文</InkwellCardContent>
      </InkwellCard>
      <InkwellCheckbox label="同意する" checked={ok} onChange={(e) => setOk(e.target.checked)} />
      <InkwellSwitch label="通知" checked={on} onChange={(e) => setOn(e.target.checked)} />
    </>
  );
}
```

## コンポーネント

| コンポーネント | 主な props | 見た目 |
|---|---|---|
| `InkwellButton` | `variant`(contained/outlined/text) / `color`(primary/secondary/error) / `size` / `startIcon` `endIcon` | 大文字ラベル・押下波紋・contained は標高が上がる |
| `InkwellTextField` | `label`(必須) / `helperText` / `error` / `fullWidth` | 枠線 + 浮き上がるラベル（枠を背景色で切り欠く）・エラー色 |
| `InkwellCard` + `InkwellCardContent` / `InkwellCardActions` | `elevation`(0/1/2/3/4/8) | 標高シャドウ。0 は影の代わりに枠線 |
| `InkwellCheckbox` | `label` / 標準 input 属性 | チェックで塗り＋レ点・押下波紋・ホバー円 |
| `InkwellSwitch` | `label` / 標準 input 属性 | トラック上をサムが滑る・ON で色付き |

## デザイントークン（`tailwind.config.js`、名前空間 `inkwell-*`）

- 色: `inkwell-primary`(#1976d2) / `-secondary`(#9c27b0) / `-error`(#d32f2f) と各 dark、`inkwell-text-*`、`inkwell-outline` / `-divider`
- フォント: `font-roboto`（`@fontsource/roboto` を自己ホスト。CSP `font-src 'self'` 準拠。付けた要素だけに適用）
- 影: `shadow-inkwell-1|2|3|4|8`（3 層合成の標高）
- アニメ: `animate-inkwell-ripple`（押下波紋）

いずれも名前空間付きで、既存の `brand-*` / `taupe-*` / `surface-*` には影響しない。

## 実装ノート（品質のための定番パターン）

- **波紋**: `useRipple` が pointerdown 座標に円を置き、`animate-inkwell-ripple` でスケール＋フェード。
  `onAnimationEnd` で DOM から除去してリークを防ぐ。
- **浮き上がるラベル**: ラベルを input の**後ろ**に置き `peer` で連動。既定を「浮いた状態」にして
  `peer-placeholder-shown:` で未入力時のみプレースホルダ位置へ戻す（入力後にラベルが本文へ落ちるバグの回避）。
  空白 placeholder（`' '`）で `:placeholder-shown` を機能させ、ラベル背景色で枠線を切り欠く。
- **アクセシビリティ**: Checkbox / Switch は実 `<input>`（`sr-only` の peer）+ `<label>` で構成し、
  role・checked・disabled がそのまま働く。Button は `focus-visible` リング、TextField は
  `aria-invalid` / `aria-describedby` を付与。

## テスト

`frontend/src/components/inkwell/__tests__/`（vitest + RTL、24 ケース）。role/aria・variant クラス・
クリック/トグル・波紋要素の生成を検証。

## 関連

- Jira: FRESTYLE-104
- 設計理由（why）: `frestyle-pdm` `docs/architecture/inkwell-ui-primitives.md`
