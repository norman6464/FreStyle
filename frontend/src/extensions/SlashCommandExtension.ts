import { Extension } from '@tiptap/core';
import Suggestion from '@tiptap/suggestion';
import type { SuggestionOptions } from '@tiptap/suggestion';
import type { Editor } from '@tiptap/core';
import { SLASH_COMMANDS, type SlashCommand } from '../constants/slashCommands';

function executeCommand(editor: Editor, command: SlashCommand) {
  const chain = editor.chain().focus();

  switch (command.action) {
    case 'paragraph':
      chain.setParagraph().run();
      break;
    case 'heading1':
      chain.setHeading({ level: 1 }).run();
      break;
    case 'heading2':
      chain.setHeading({ level: 2 }).run();
      break;
    case 'heading3':
      chain.setHeading({ level: 3 }).run();
      break;
    case 'bulletList':
      chain.toggleBulletList().run();
      break;
    case 'orderedList':
      chain.toggleOrderedList().run();
      break;
    case 'toggleList': {
      const setToggleList = (editor.commands as { setToggleList?: () => boolean }).setToggleList;
      if (typeof setToggleList === 'function') {
        setToggleList();
      }
      break;
    }
    case 'image': {
      const onImageUpload = editor.storage.slashCommand?.onImageUpload;
      if (typeof onImageUpload === 'function') {
        onImageUpload();
      }
      break;
    }
    case 'taskList':
      chain.toggleTaskList().run();
      break;
    case 'codeBlock':
      chain.toggleCodeBlock().run();
      break;
    case 'blockquote':
      chain.toggleBlockquote().run();
      break;
    case 'horizontalRule':
      chain.setHorizontalRule().run();
      break;
    case 'table':
      chain.insertTable({ rows: 3, cols: 3, withHeaderRow: true }).run();
      break;
    case 'callout': {
      if (command.attrs) {
        const setCalloutWithType = (editor.commands as { setCalloutWithType?: (type: string, emoji: string) => boolean }).setCalloutWithType;
        if (typeof setCalloutWithType === 'function') {
          setCalloutWithType(command.attrs.calloutType, command.attrs.emoji);
        }
      } else {
        const setCallout = (editor.commands as { setCallout?: () => boolean }).setCallout;
        if (typeof setCallout === 'function') {
          setCallout();
        }
      }
      break;
    }
    case 'youtube': {
      const onYoutubeUrl = editor.storage.slashCommand?.onYoutubeUrl;
      if (typeof onYoutubeUrl === 'function') {
        onYoutubeUrl();
      }
      break;
    }
    case 'emoji': {
      const onEmojiPicker = editor.storage.slashCommand?.onEmojiPicker;
      if (typeof onEmojiPicker === 'function') {
        onEmojiPicker();
      }
      break;
    }
    default: {
      const _exhaustive: never = command.action;
      console.error('Unknown slash command action:', _exhaustive);
      break;
    }
  }
}

export type SlashCommandSuggestionOptions = Omit<SuggestionOptions<SlashCommand>, 'editor'>;

export const SlashCommandExtension = Extension.create({
  name: 'slashCommand',

  addStorage() {
    return {
      onImageUpload: null as (() => void) | null,
      onEmojiPicker: null as (() => void) | null,
      onYoutubeUrl: null as (() => void) | null,
    };
  },

  addOptions() {
    return {
      suggestion: {
        char: '/',
        command: ({ editor, range, props }: { editor: Editor; range: { from: number; to: number }; props: SlashCommand }) => {
          editor.chain().focus().deleteRange(range).run();
          executeCommand(editor, props);
        },
        items: ({ query }: { query: string }) => {
          return SLASH_COMMANDS.filter(item =>
            item.label.toLowerCase().includes(query.toLowerCase())
          );
        },
      } as SlashCommandSuggestionOptions,
    };
  },

  addProseMirrorPlugins() {
    return [
      Suggestion({
        editor: this.editor,
        ...this.options.suggestion,
      }),
    ];
  },
});

export { executeCommand };
