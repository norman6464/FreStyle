/**
 * ラベルの色から、読める塗り方を決める。
 *
 * 色は利用者が自由に決めるので、地に敷いて白文字か濃い文字かの二択にすると、
 * 中くらいの明るさでどちらを選んでも小さな文字の基準（4.5:1）に届かない色が残る。
 * 届かないときは<b>地に敷くのをやめて</b>、枠線と点にだけ色を使う。
 *
 * チケットのラベル（`pages/backlog`）とナレッジのページラベル（`entities/kb`）が同じ
 * `labels` テーブル（`color` 列）を共有するため、色そのものの扱いはどちらの業務も
 * 知らない汎用処理として shared に置く。
 */

/** 小さな文字に要る明暗の比。 */
const MIN_CONTRAST = 4.5;
/** 地に敷いたときの濃い方の文字色（--color-text-primary と同じ値）。 */
const DARK_TEXT = '#191919';

export type LabelPaint =
  /** 色を地に敷き、その上に読める文字色を載せる。 */
  | { kind: 'solid'; background: string; color: string }
  /** 地には敷かず、枠線と点にだけ色を使う。文字は既定の濃さ。 */
  | { kind: 'outline'; borderColor: string };

const HEX = /^#[0-9a-f]{6}$/i;

function toLinear(value: number): number {
  const c = value / 255;
  return c <= 0.03928 ? c / 12.92 : ((c + 0.055) / 1.055) ** 2.4;
}

/** 相対輝度。色の形が読めなければ null。 */
function luminance(hex: string): number | null {
  if (!HEX.test(hex)) return null;
  const r = toLinear(parseInt(hex.slice(1, 3), 16));
  const g = toLinear(parseInt(hex.slice(3, 5), 16));
  const b = toLinear(parseInt(hex.slice(5, 7), 16));
  return 0.2126 * r + 0.7152 * g + 0.0722 * b;
}

function contrast(a: number, b: number): number {
  const hi = Math.max(a, b);
  const lo = Math.min(a, b);
  return (hi + 0.05) / (lo + 0.05);
}

/**
 * 色の形が読めないときは枠線も引けないので、枠の色を既定の罫線に落とす。
 * 表示そのものは止めない（付いているラベルが画面から消える方が困る）。
 */
const UNREADABLE: LabelPaint = { kind: 'outline', borderColor: 'var(--color-surface-3)' };

export function labelPaint(color: string): LabelPaint {
  const lum = luminance(color);
  if (lum === null) return UNREADABLE;
  const white = luminance('#ffffff') as number;
  const dark = luminance(DARK_TEXT) as number;
  if (contrast(lum, white) >= MIN_CONTRAST) return { kind: 'solid', background: color, color: '#ffffff' };
  if (contrast(lum, dark) >= MIN_CONTRAST) return { kind: 'solid', background: color, color: DARK_TEXT };
  return { kind: 'outline', borderColor: color };
}
