import type { Extensions } from '@tiptap/react';
import StarterKit from '@tiptap/starter-kit';
import { Placeholder } from '@tiptap/extensions';
import Image from '@tiptap/extension-image';
import Code from '@tiptap/extension-code';
import type { EditorCommand } from './editorCommands';
import { SlashCommand } from './slashCommandExtension';

/**
 * CombinableCode は他のマークと共存できるインラインコード。
 *
 * 既定の Code は `excludes: '_'`（＝他の全マークを排他）で、コードを掛けると
 * 太字・斜体・下線・打ち消しがすべて外れてしまう。ノートでは「コード＋太字」等を
 * 重ねたい場面があるため、排他指定を解いて併用できるようにする。
 */
const CombinableCode = Code.extend({ excludes: '' });

/** createEditorExtensions の組み立てオプション。拡張を増やすときはここに口を足す。 */
export interface CreateEditorExtensionsOptions {
  /** 空エディタに表示するプレースホルダ文言。 */
  placeholder?: string;
  /** 画像ノードを有効にするか（既定 true）。アップロードの配線は利用側が別途行う。 */
  image?: boolean;
  /**
   * '/' メニューに出すコマンド（英単語トリガ）。未指定ならスラッシュメニュー自体を付けない。
   * 利用側だけが知る操作（画像アップロード等）もこの配列に差し込める。
   */
  slashItems?: EditorCommand[];
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
  const { placeholder = '本文を入力…', image = true, slashItems } = options;

  const extensions: Extensions = [
    // StarterKit の code は排他指定があるため無効化し、共存できる CombinableCode を使う。
    StarterKit.configure({ heading: { levels: [1, 2, 3] }, code: false }),
    CombinableCode,
    Placeholder.configure({ placeholder }),
  ];

  if (image) {
    extensions.push(Image.configure({ inline: false, allowBase64: false }));
  }

  if (slashItems && slashItems.length > 0) {
    extensions.push(SlashCommand.configure({ items: slashItems }));
  }

  return extensions;
}
