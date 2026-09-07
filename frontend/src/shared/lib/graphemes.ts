/**
 * 見た目の 1 文字（書記素クラスタ）を数える。
 *
 * 絵文字は `string.length`（UTF-16 コード単位）で測れない。サロゲートペア（📘 は 2）は
 * もちろん、肌色修飾子や ZWJ で連結した家族の絵文字（👨‍👩‍👧 は 1 見た目だが
 * コードポイントは 5、UTF-16 単位ではさらに多い）まで含めて「1 文字」と数える必要がある。
 * `Intl.Segmenter` はロケール非依存の書記素境界規則（UAX #29）で区切るので、
 * これらを自前の正規表現で再実装せずに正しく 1 つと数えられる。
 */
const segmenter = new Intl.Segmenter(undefined, { granularity: 'grapheme' });

export function countGraphemes(text: string): number {
  if (text === '') return 0;
  let count = 0;
  for (const _ of segmenter.segment(text)) count += 1;
  return count;
}
