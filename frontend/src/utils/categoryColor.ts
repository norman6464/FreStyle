/**
 * カテゴリ名 → 固定パレットの色クラス。カテゴリは DB の自由文字列で種類が多く
 * （本番 36 種類）、手書きの対応表では保守できないため、名前の決定的ハッシュで
 * パレットから選ぶ。同じカテゴリは原則同じ色になり、新カテゴリの追加でも
 * コード変更は不要。
 *
 * 濃色塗り + 白文字にしているのは、淡色バッジだと白背景で薄く見えて区別しにくい
 * というユーザー要望のため(FRESTYLE-112)。
 */
const PALETTE = [
  'bg-rose-600 text-white',
  'bg-orange-600 text-white',
  'bg-amber-600 text-white',
  'bg-lime-600 text-white',
  'bg-emerald-600 text-white',
  'bg-teal-600 text-white',
  'bg-sky-600 text-white',
  'bg-blue-600 text-white',
  'bg-violet-600 text-white',
  'bg-fuchsia-600 text-white',
];

function paletteIndex(category: string): number {
  let hash = 0;
  for (let i = 0; i < category.length; i++) {
    hash = (hash * 31 + category.charCodeAt(i)) | 0;
  }
  return Math.abs(hash) % PALETTE.length;
}

/** カテゴリ名から決定的にパレットの色クラスを返す。 */
export function categoryColorClass(category: string): string {
  return PALETTE[paletteIndex(category)];
}

/**
 * 表示順のカテゴリ列に色を割り当てる。基本はカテゴリ名のハッシュ色（ページや
 * フィルタが変わっても同じカテゴリは同じ色）だが、ハッシュ衝突で**隣り合う
 * ブロックが同色になるときだけ**次の色へずらす。本番データでは名前ハッシュ
 * 単体だと隣接同色が複数箇所発生するため（例: bash の 入出力/ファイル操作）、
 * 「ブロックを色で区別したい」という本来の目的を隣接保証で担保する。
 */
export function assignCategoryColors(categories: string[]): Map<string, string> {
  const assigned = new Map<string, string>();
  let prev = -1;
  for (const category of categories) {
    let idx = paletteIndex(category);
    if (idx === prev) {
      idx = (idx + 1) % PALETTE.length;
    }
    assigned.set(category, PALETTE[idx]);
    prev = idx;
  }
  return assigned;
}
