import type { Extensions } from '@tiptap/react';
import type CodeBlockLowlight from '@tiptap/extension-code-block-lowlight';
import type Heading from '@tiptap/extension-heading';
import { Placeholder } from '@tiptap/extensions';
import { ReactNodeViewRenderer, textblockTypeInputRule } from '@tiptap/react';
import type { EditorCommand } from './editorCommands';
import CodeBlockView from './CodeBlockView';
import { ListNormalization } from './listNormalization';
import { MarkdownShortcuts } from './markdownShortcuts';
import { createSchemaExtensions } from './schemaExtensions';
import { SlashCommand } from './slashCommandExtension';

/**
 * withZenkakuHeadingInputRules は Heading（levels 1〜3）へ全角入力対応の input rule を上掛けする。
 *
 * 既定の input rule は半角の `#` ＋半角スペースのみ発火する。日本語 IME のまま打つと
 * `＃`（全角）や全角スペースになり変換されないため、全角 `＃`（半角と混在も可）＋
 * 半角/全角スペースでも同じ level に変換されるルールを追加する。スキーマ（ノード名・attrs）は
 * factory 側の Heading のまま変えない。
 */
function withZenkakuHeadingInputRules(heading: typeof Heading) {
  return heading.extend({
    addInputRules() {
      const defaultRules = this.parent?.() ?? [];
      const zenkakuRules = this.options.levels.map((level) =>
        textblockTypeInputRule({
          // \u3000 は全角スペース（リテラルで書くと no-irregular-whitespace に触れるためエスケープ表記）。
          find: new RegExp(`^([#\uFF03]{${level}})[ \u3000]$`),
          type: this.type,
          getAttributes: { level },
        }),
      );
      return [...defaultRules, ...zenkakuRules];
    },
  });
}

/**
 * withCodeBlockView は codeBlock へ右上の言語選択・コピー UI（CodeBlockView）の NodeView を
 * 上掛けする。ノード名は 'codeBlock' のままなので、スラッシュ（/codeblock）・バブルメニュー・
 * 既存 doc とそのまま互換。lowlight ハイライトの設定は factory 側が持つ。
 */
function withCodeBlockView(codeBlock: typeof CodeBlockLowlight) {
  return codeBlock.extend({
    addNodeView() {
      return ReactNodeViewRenderer(CodeBlockView);
    },
  });
}

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
 * スキーマを決める拡張は createSchemaExtensions()（教材 Markdown 変換器と共有の単一ソース）から
 * 受け取り、ここでは NodeView・input rule・プレースホルダ等の表示/入力の挙動だけを上掛けする。
 * バブルメニュー・スラッシュコマンド・ドラッグハンドルなどを足すときも、オプションを増やして
 * ここで合成する（＝拡張ポイントを 1 箇所に保つ）。
 */
export function createEditorExtensions(
  options: CreateEditorExtensionsOptions = {},
): Extensions {
  const { placeholder = '本文を入力…', image = true, slashItems } = options;

  const extensions: Extensions = createSchemaExtensions({ image }).map((extension) => {
    if (extension.name === 'heading') {
      return withZenkakuHeadingInputRules(extension as typeof Heading);
    }
    if (extension.name === 'codeBlock') {
      return withCodeBlockView(extension as typeof CodeBlockLowlight);
    }
    return extension;
  });

  extensions.push(
    // 隣接する同種リストを結合し、番号リストの番号リセットを防ぐ。
    ListNormalization,
    // IME（日本語入力）確定でも効く ＃ 見出し・``` コードブロック変換。
    MarkdownShortcuts,
    Placeholder.configure({ placeholder }),
  );

  if (slashItems && slashItems.length > 0) {
    extensions.push(SlashCommand.configure({ items: slashItems }));
  }

  return extensions;
}
