import { Extension } from '@tiptap/react';
import { ReactRenderer } from '@tiptap/react';
import Suggestion, { type SuggestionProps, type SuggestionKeyDownProps } from '@tiptap/suggestion';
import type { EditorCommand } from './editorCommands';
import { filterSlashItems } from './slashItems';
import SlashMenuList, { type SlashMenuListHandle, type SlashMenuListProps } from './SlashMenuList';

export interface SlashCommandOptions {
  /** '/' メニューに出すコマンド。トリガは英単語（id / keywords）。 */
  items: EditorCommand[];
}

/**
 * SlashCommand は '/' でブロック挿入メニューを開く拡張。
 *
 * '/' に続けて英単語（/h1・/quote・/image …）を打つと絞り込まれ、Enter/クリックで
 * 実行される。確定時は入力中の "/query" を消してからコマンドを実行する。
 * ポップアップの生成は ReactRenderer、位置決めは Suggestion 内蔵の mount（floating-ui）が担う。
 */
export const SlashCommand = Extension.create<SlashCommandOptions>({
  name: 'slashCommand',

  addOptions() {
    return { items: [] };
  },

  addProseMirrorPlugins() {
    const { editor } = this;
    const allItems = () => this.options.items;

    return [
      Suggestion<EditorCommand, EditorCommand>({
        editor,
        char: '/',
        // 行頭に限定しない（文中でもブロック操作を呼べる方が実用的）。
        startOfLine: false,
        // 日本語ラベルでは照合しない（トリガは英単語のみ）ため、空白は区切りとして打ち切る。
        allowSpaces: false,
        pluginKey: undefined,
        items: ({ query }) => filterSlashItems(allItems(), query),
        command: ({ editor: currentEditor, range, props: item }) => {
          // 入力中の "/query" を取り除いてから実行する（本文にトリガ文字列を残さない）。
          currentEditor.chain().focus().deleteRange(range).run();
          item.run(currentEditor);
        },
        render: () => {
          let renderer: ReactRenderer<SlashMenuListHandle, SlashMenuListProps> | null = null;
          let unmount: (() => void) | null = null;

          return {
            onStart: (props: SuggestionProps<EditorCommand, EditorCommand>) => {
              renderer = new ReactRenderer(SlashMenuList, {
                editor: props.editor,
                props: {
                  items: props.items,
                  onSelect: (item: EditorCommand) => props.command(item),
                },
                className: 'rte-slash',
              });
              // Suggestion 内蔵の mount がキャレット位置への追従（floating-ui）まで面倒を見る。
              unmount = props.mount(renderer.element);
            },
            onUpdate: (props: SuggestionProps<EditorCommand, EditorCommand>) => {
              renderer?.updateProps({
                items: props.items,
                onSelect: (item: EditorCommand) => props.command(item),
              });
            },
            onKeyDown: (props: SuggestionKeyDownProps) => {
              if (props.event.key === 'Escape') {
                unmount?.();
                unmount = null;
                return true;
              }
              return renderer?.ref?.onKeyDown(props.event) ?? false;
            },
            onExit: () => {
              unmount?.();
              unmount = null;
              renderer?.destroy();
              renderer = null;
            },
          };
        },
      }),
    ];
  },
});
