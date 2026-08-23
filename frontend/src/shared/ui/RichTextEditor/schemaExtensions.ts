import type { Extensions } from '@tiptap/core';
import Code from '@tiptap/extension-code';
import CodeBlockLowlight from '@tiptap/extension-code-block-lowlight';
import Heading from '@tiptap/extension-heading';
import Image from '@tiptap/extension-image';
import { TaskItem, TaskList } from '@tiptap/extension-list';
import { TableKit } from '@tiptap/extension-table';
import StarterKit from '@tiptap/starter-kit';
import { common, createLowlight } from 'lowlight';

/**
 * lowlight のインスタンス（highlight.js の common 言語 37 種を登録）。
 * codeBlock スキーマを configure する本モジュールが所有する。エディタ側では
 * トークンの配色に app 全体へ import 済みの Atom One Light テーマ（.hljs-*）が当たる。
 */
export const lowlight = createLowlight(common);

/**
 * CombinableCode は他のマークと共存できるインラインコード。
 *
 * 既定の Code は `excludes: '_'`（＝他の全マークを排他）で、コードを掛けると
 * 太字・斜体・下線・打ち消しがすべて外れてしまう。ノートでは「コード＋太字」等を
 * 重ねたい場面があるため、排他指定を解いて併用できるようにする。マーク名は 'code' のまま。
 */
const CombinableCode = Code.extend({ excludes: '' });

/** createSchemaExtensions の組み立てオプション。 */
export interface CreateSchemaExtensionsOptions {
  /** 画像ノードをスキーマに含めるか（既定 true）。 */
  image?: boolean;
}

/**
 * createSchemaExtensions は「ドキュメントのスキーマ（ノード/マーク名・attrs・content 式）を
 * 決める」拡張だけを組み立てる共有 factory。
 *
 * エディタ（editorExtensions.ts）と、Node で動く教材 Markdown 変換器（scripts/md2doc.mjs）が
 * 同一スキーマで doc(JSON) を扱うための単一ソース。変換器は本ファイルを jsdom なしの Node から
 * 直接 import するため、ここには React・CSS・DOM 依存を（推移的にも）置かないこと。
 * NodeView・input rule・プレースホルダ等の表示/入力の挙動は editorExtensions.ts 側で上掛けする。
 */
export function createSchemaExtensions(
  options: CreateSchemaExtensionsOptions = {},
): Extensions {
  const { image = true } = options;

  const extensions: Extensions = [
    // StarterKit の code は排他指定、heading は levels 無制限、codeBlock はハイライトなしのため
    // それぞれ無効化し、スキーマを決める拡張へ差し替える。
    StarterKit.configure({ heading: false, code: false, codeBlock: false }),
    CombinableCode,
    // 見出しは 1〜3 のみ（エディタ UI・教材の章構造とも 3 段で揃える）。
    Heading.configure({ levels: [1, 2, 3] }),
    // 構文ハイライト付きコードブロック。ノード名は 'codeBlock' のまま既存 doc と互換。
    CodeBlockLowlight.configure({ lowlight, defaultLanguage: 'plaintext' }),
    // 表（GFM テーブル相当）。教材の本文とノートの両方で使う。
    // resizable は列幅ドラッグ UI が必要になるため、まずは固定幅で表現力を優先する。
    TableKit.configure({ table: { resizable: false } }),
    // タスクリスト（チェックボックス）。教材のチェックリスト章とノートの TODO で使う。
    TaskList,
    TaskItem.configure({ nested: true }),
  ];

  if (image) {
    extensions.push(Image.configure({ inline: false, allowBase64: false }));
  }

  return extensions;
}
