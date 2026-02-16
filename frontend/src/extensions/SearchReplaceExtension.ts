import { Extension } from '@tiptap/core';
import { Plugin, PluginKey } from '@tiptap/pm/state';
import { Decoration, DecorationSet } from '@tiptap/pm/view';

export interface SearchReplaceState {
  searchTerm: string;
  replaceTerm: string;
  caseSensitive: boolean;
  results: { from: number; to: number }[];
  currentIndex: number;
}

export const searchReplacePluginKey = new PluginKey('searchReplace');

function findMatches(
  doc: { textBetween: (from: number, to: number) => string; content: { size: number } },
  searchTerm: string,
  caseSensitive: boolean,
): { from: number; to: number }[] {
  if (!searchTerm) return [];

  const results: { from: number; to: number }[] = [];
  const text = doc.textBetween(0, doc.content.size, '\n');
  const search = caseSensitive ? searchTerm : searchTerm.toLowerCase();
  const content = caseSensitive ? text : text.toLowerCase();

  let pos = 0;
  while (pos < content.length) {
    const index = content.indexOf(search, pos);
    if (index === -1) break;
    // textBetween positions map 1:1 to doc positions offset by 0
    // but we need to account for node boundaries
    results.push({ from: index, to: index + search.length });
    pos = index + 1;
  }

  // textBetween returns flat text; we need to map back to doc positions
  return mapTextPositionsToDoc(doc, results);
}

function mapTextPositionsToDoc(
  doc: { nodesBetween?: (from: number, to: number, callback: (node: { isText: boolean; isBlock: boolean; nodeSize: number; text?: string }, pos: number) => boolean | void) => void; content: { size: number } },
  textPositions: { from: number; to: number }[],
): { from: number; to: number }[] {
  if (textPositions.length === 0 || !doc.nodesBetween) return [];

  // Build a mapping from text offset to doc position
  const mapping: number[] = [];
  let textOffset = 0;

  doc.nodesBetween(0, doc.content.size, (node, pos) => {
    if (node.isText && node.text) {
      for (let i = 0; i < node.text.length; i++) {
        mapping[textOffset] = pos + i;
        textOffset++;
      }
    } else if (node.isBlock && textOffset > 0) {
      // Block boundary adds a newline in textBetween
      mapping[textOffset] = pos;
      textOffset++;
    }
    return true;
  });

  return textPositions
    .map(({ from, to }) => {
      const docFrom = mapping[from];
      const docTo = mapping[to - 1];
      if (docFrom === undefined || docTo === undefined) return null;
      return { from: docFrom, to: docTo + 1 };
    })
    .filter((r): r is { from: number; to: number } => r !== null);
}

function createDecorations(
  results: { from: number; to: number }[],
  currentIndex: number,
): DecorationSet {
  if (results.length === 0) return DecorationSet.empty;

  // We can't create a DecorationSet without a doc, so return empty
  // Decorations are created inside the plugin's apply method
  return DecorationSet.empty;
}

export const SearchReplaceExtension = Extension.create({
  name: 'searchReplace',

  addStorage() {
    return {
      searchTerm: '',
      replaceTerm: '',
      caseSensitive: false,
      results: [] as { from: number; to: number }[],
      currentIndex: 0,
    } as SearchReplaceState;
  },

  addCommands() {
    return {
      setSearchTerm:
        (searchTerm: string) =>
        ({ editor }) => {
          const storage = editor.storage.searchReplace as SearchReplaceState;
          storage.searchTerm = searchTerm;
          storage.currentIndex = 0;
          // Trigger plugin state update
          editor.view.dispatch(editor.state.tr.setMeta(searchReplacePluginKey, { searchTerm }));
          return true;
        },

      setReplaceTerm:
        (replaceTerm: string) =>
        ({ editor }) => {
          const storage = editor.storage.searchReplace as SearchReplaceState;
          storage.replaceTerm = replaceTerm;
          return true;
        },

      setCaseSensitive:
        (caseSensitive: boolean) =>
        ({ editor }) => {
          const storage = editor.storage.searchReplace as SearchReplaceState;
          storage.caseSensitive = caseSensitive;
          editor.view.dispatch(editor.state.tr.setMeta(searchReplacePluginKey, { caseSensitive }));
          return true;
        },

      findNext:
        () =>
        ({ editor }) => {
          const storage = editor.storage.searchReplace as SearchReplaceState;
          if (storage.results.length === 0) return false;
          storage.currentIndex = (storage.currentIndex + 1) % storage.results.length;
          editor.view.dispatch(editor.state.tr.setMeta(searchReplacePluginKey, { navigate: true }));
          scrollToMatch(editor, storage.results[storage.currentIndex]);
          return true;
        },

      findPrev:
        () =>
        ({ editor }) => {
          const storage = editor.storage.searchReplace as SearchReplaceState;
          if (storage.results.length === 0) return false;
          storage.currentIndex = (storage.currentIndex - 1 + storage.results.length) % storage.results.length;
          editor.view.dispatch(editor.state.tr.setMeta(searchReplacePluginKey, { navigate: true }));
          scrollToMatch(editor, storage.results[storage.currentIndex]);
          return true;
        },

      replaceCurrent:
        () =>
        ({ editor }) => {
          const storage = editor.storage.searchReplace as SearchReplaceState;
          if (storage.results.length === 0) return false;
          const match = storage.results[storage.currentIndex];
          if (!match) return false;

          editor.chain()
            .command(({ tr }) => {
              tr.insertText(storage.replaceTerm, match.from, match.to);
              return true;
            })
            .run();

          // Re-trigger search
          editor.view.dispatch(editor.state.tr.setMeta(searchReplacePluginKey, { replaced: true }));
          return true;
        },

      replaceAll:
        () =>
        ({ editor }) => {
          const storage = editor.storage.searchReplace as SearchReplaceState;
          if (storage.results.length === 0) return false;

          // Replace from end to start to preserve positions
          const sorted = [...storage.results].sort((a, b) => b.from - a.from);
          editor.chain()
            .command(({ tr }) => {
              for (const match of sorted) {
                tr.insertText(storage.replaceTerm, match.from, match.to);
              }
              return true;
            })
            .run();

          editor.view.dispatch(editor.state.tr.setMeta(searchReplacePluginKey, { replacedAll: true }));
          return true;
        },

      clearSearch:
        () =>
        ({ editor }) => {
          const storage = editor.storage.searchReplace as SearchReplaceState;
          storage.searchTerm = '';
          storage.replaceTerm = '';
          storage.results = [];
          storage.currentIndex = 0;
          editor.view.dispatch(editor.state.tr.setMeta(searchReplacePluginKey, { clear: true }));
          return true;
        },
    };
  },

  addProseMirrorPlugins() {
    const extensionThis = this;

    return [
      new Plugin({
        key: searchReplacePluginKey,
        state: {
          init() {
            return DecorationSet.empty;
          },
          apply(tr, oldDecorations) {
            const meta = tr.getMeta(searchReplacePluginKey);
            if (meta || tr.docChanged) {
              const storage = extensionThis.storage as SearchReplaceState;
              const results = findMatches(tr.doc, storage.searchTerm, storage.caseSensitive);
              storage.results = results;

              if (storage.currentIndex >= results.length) {
                storage.currentIndex = Math.max(0, results.length - 1);
              }

              if (results.length === 0) return DecorationSet.empty;

              const decorations = results.map((result, index) => {
                const className = index === storage.currentIndex
                  ? 'search-match search-match-current'
                  : 'search-match';
                return Decoration.inline(result.from, result.to, { class: className });
              });

              return DecorationSet.create(tr.doc, decorations);
            }
            return oldDecorations.map(tr.mapping, tr.doc);
          },
        },
        props: {
          decorations(state) {
            return this.getState(state);
          },
        },
      }),
    ];
  },
});

function scrollToMatch(
  editor: { view: { domAtPos: (pos: number) => { node: Node; offset: number } } },
  match: { from: number; to: number } | undefined,
) {
  if (!match) return;
  try {
    const { node } = editor.view.domAtPos(match.from);
    const element = node instanceof HTMLElement ? node : node.parentElement;
    element?.scrollIntoView({ block: 'center', behavior: 'smooth' });
  } catch {
    // Position may be invalid
  }
}

declare module '@tiptap/core' {
  interface Commands<ReturnType> {
    searchReplace: {
      setSearchTerm: (searchTerm: string) => ReturnType;
      setReplaceTerm: (replaceTerm: string) => ReturnType;
      setCaseSensitive: (caseSensitive: boolean) => ReturnType;
      findNext: () => ReturnType;
      findPrev: () => ReturnType;
      replaceCurrent: () => ReturnType;
      replaceAll: () => ReturnType;
      clearSearch: () => ReturnType;
    };
  }
}
