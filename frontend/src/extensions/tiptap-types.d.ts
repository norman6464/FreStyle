/**
 * Tiptap module augmentation。
 *
 * Tiptap の `Storage` 型は declaration merging で extension ごとに拡張する設計
 * （公式: https://tiptap.dev/api/extensions/extension#addstorage）。
 *
 * 各カスタム extension が登録する storage を Storage interface に追記することで、
 * `editor.storage.searchReplace` のような利用箇所で TS エラーが解消される。
 */
import '@tiptap/core';

interface SlashCommandStorage {
  onImageUpload: (() => void) | null;
  onEmojiPicker: (() => void) | null;
  onYoutubeUrl: (() => void) | null;
}

interface SearchReplaceExtensionStorage {
  searchTerm: string;
  replaceTerm: string;
  caseSensitive: boolean;
  results: { from: number; to: number }[];
  currentIndex: number;
}

/**
 * tiptap-markdown が提供する storage。
 * editor.storage.markdown.getMarkdown() で現在の editor 内容を Markdown 文字列で取得できる。
 * 公式 README: https://github.com/aguingand/tiptap-markdown
 */
interface TiptapMarkdownStorage {
  getMarkdown: () => string;
}

declare module '@tiptap/core' {
  interface Storage {
    /** SlashCommandExtension の storage（image upload / emoji / youtube URL ハンドラ） */
    slashCommand: SlashCommandStorage;
    /** SearchReplaceExtension の storage（検索語・置換語・カレント位置・マッチ結果） */
    searchReplace: SearchReplaceExtensionStorage;
    /** tiptap-markdown の storage（getMarkdown() で現在エディタ内容を Markdown 文字列で取得） */
    markdown: TiptapMarkdownStorage;
  }

  // Tiptap の RawCommands は Commands<T> のうち「value が object 型」のキーだけを
  // Pick して UnionToIntersection でフラット化するため、各 extension のコマンド群は
  // 名前空間 (object) でラップして登録する。
  interface Commands<ReturnType> {
    /** SearchReplaceExtension のコマンド */
    searchReplaceCommands: {
      setSearchTerm: (searchTerm: string) => ReturnType;
      setReplaceTerm: (replaceTerm: string) => ReturnType;
      setCaseSensitive: (caseSensitive: boolean) => ReturnType;
      findNext: () => ReturnType;
      findPrevious: () => ReturnType;
      replaceCurrent: () => ReturnType;
      replaceAll: () => ReturnType;
      clearSearch: () => ReturnType;
    };
    /** CalloutExtension のコマンド */
    calloutCommands: {
      setCallout: () => ReturnType;
      setCalloutWithType: (calloutType: string, emoji: string) => ReturnType;
    };
    /** ToggleListExtension のコマンド（namespace 名は built-in toggleList と衝突しないよう別名） */
    toggleListCustomCommands: {
      setToggleList: () => ReturnType;
    };
  }
}
