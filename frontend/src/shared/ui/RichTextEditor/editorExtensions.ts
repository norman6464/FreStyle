import type { Extensions } from '@tiptap/react';
import StarterKit from '@tiptap/starter-kit';
import { Placeholder } from '@tiptap/extensions';
import Image from '@tiptap/extension-image';

/** createEditorExtensions の組み立てオプション。拡張を増やすときはここに口を足す。 */
export interface CreateEditorExtensionsOptions {
  /** 空エディタに表示するプレースホルダ文言。 */
  placeholder?: string;
  /** 画像ノードを有効にするか（既定 true）。アップロードの配線は利用側が別途行う。 */
  image?: boolean;
}

/**
 * createEditorExtensions は RichTextEditor が使う tiptap 拡張一式を組み立てる。
 *
 * 拡張の追加・入れ替えはこの 1 関数に集約し、UI 本体（RichTextEditor）から構成の詳細を隠す。
 * バブルメニュー・スラッシュコマンド・ドラッグハンドルなどを足すときも、オプションを増やして
 * ここで合成する（＝拡張ポイントを 1 箇所に保つ）。
 */
export function createEditorExtensions(
  options: CreateEditorExtensionsOptions = {},
): Extensions {
  const { placeholder = '本文を入力…', image = true } = options;

  const extensions: Extensions = [
    StarterKit.configure({ heading: { levels: [1, 2, 3] } }),
    Placeholder.configure({ placeholder }),
  ];

  if (image) {
    extensions.push(Image.configure({ inline: false, allowBase64: false }));
  }

  return extensions;
}
