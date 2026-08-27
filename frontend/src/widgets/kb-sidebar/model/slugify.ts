/**
 * slugify は入力した名前から、URL に使える短い名前の下書きを作る。
 *
 * **下書きでしかない。** 日本語の名前からは何も残らない（英数字だけを拾うため）ので、
 * 入力欄では書き換えられるようにしておくこと。自動生成だけに頼ると先へ進めなくなる。
 */
export function slugify(name: string): string {
  return name
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, '-')
    .replace(/^-+|-+$/g, '')
    .slice(0, 40);
}
