import { Extension } from '@tiptap/core';

/** Enter押下時に全角＃（および半角#）のみの行を見出しに変換するパターン */
export const HEADING_ENTER_PATTERN = /^[#＃]{1,3}$/;

/**
 * 全角＃ + Enter で見出し変換する TipTap 拡張
 *
 * 日本語IMEで全角＃を入力した場合、TipTapの標準InputRule（スペーストリガー）が
 * 動作しないため、Enterキーでもハッシュ記号のみの行を見出しに変換する。
 */
export const FullWidthHeadingEnter = Extension.create({
  name: 'fullWidthHeadingEnter',

  addKeyboardShortcuts() {
    return {
      Enter: ({ editor }) => {
        const { state } = editor;
        const { $from, empty } = state.selection;

        if (!empty) return false;
        if ($from.parent.type.name !== 'paragraph') return false;

        const text = $from.parent.textContent;
        const match = text.match(HEADING_ENTER_PATTERN);
        if (!match) return false;

        const level = match[0].length as 1 | 2 | 3;

        editor
          .chain()
          .command(({ tr }) => {
            tr.delete($from.start(), $from.end());
            return true;
          })
          .setNode('heading', { level })
          .run();

        return true;
      },
    };
  },
});
