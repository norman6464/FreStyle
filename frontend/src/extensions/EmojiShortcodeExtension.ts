/**
 * `:smile:` のような絵文字 shortcode を Unicode 絵文字に変換する拡張。
 *
 * Zenn の絵文字補完（`:` トリガで候補を出す UI 補完）は SlashCommand 系で
 * 別途実装する余地を残し、本 PR では markdown 入力時の自動置換のみを実装する。
 *
 * 実装:
 *   - markdown-it-emoji を tiptap-markdown の parse setup で md.use()
 *   - serialize 側はテキストノードとして扱われるので特別な処理は不要
 *     (既存テキストに含まれる Unicode 絵文字はそのまま markdown 出力される)
 */
import { Extension } from '@tiptap/core';
// @ts-expect-error  no type declarations
import { full as mdEmoji } from 'markdown-it-emoji';

interface MarkdownItType {
  use: (plugin: unknown) => MarkdownItType;
}

export const EmojiShortcode = Extension.create({
  name: 'emojiShortcode',

  addStorage() {
    return {
      markdown: {
        // Extension の serialize は無し (テキストノードがそのまま markdown に出るので透過)
        parse: {
          setup(this: unknown, md: MarkdownItType) {
            md.use(mdEmoji);
          },
        },
      },
    };
  },
});
