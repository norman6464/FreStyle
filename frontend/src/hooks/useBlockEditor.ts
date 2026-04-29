import { useEffect, useMemo, useRef } from 'react';
import { useEditor } from '@tiptap/react';
import { createEditorExtensions } from '../utils/editorExtensions';
import { isLegacyMarkdown } from '../utils/isLegacyMarkdown';
import { markdownToTiptap } from '../utils/markdownToTiptap';
import { resolveNoteContent } from '../utils/noteContentFormat';

interface UseBlockEditorOptions {
  /** Note.content (markdown 文字列 or 旧 Tiptap JSON 文字列 or 旧 Markdown 文字列)。 */
  content: string;
  /** エディタ更新時に呼ばれる。値は Zenn 互換 Markdown 文字列。 */
  onChange: (markdown: string) => void;
}

/**
 * Tiptap エディタを Zenn 互換 Markdown 入出力で使うための共通フック。
 *
 * 入力 (content) の解釈:
 *   1. 空文字 → 空エディタ
 *   2. 旧 markdownToTiptap で読める legacy Markdown (heading / list のみ)
 *      → Tiptap JSON に変換して setContent
 *   3. JSON-shaped で type=doc → そのまま Tiptap JSON として setContent
 *   4. それ以外 → tiptap-markdown が markdown として parse する
 *
 * 出力 (onChange) は常に Markdown 文字列。
 * 既存ノートの保存値は JSON 文字列のままだが、ユーザーが一度編集して保存すれば
 * Markdown に置き換わるので段階的にマイグレートされる（オンライン進行的アップグレード）。
 */
export function useBlockEditor({ content, onChange }: UseBlockEditorOptions) {
  const onChangeRef = useRef(onChange);
  onChangeRef.current = onChange;

  // エディタ自身が発行した最後の Markdown 文字列を記録し、同期ループを防止
  const lastEmittedMarkdown = useRef<string>('');

  const initialContent = useMemo(() => {
    if (!content) return undefined;
    // 旧 Markdown (heading / list のみの簡易フォーマット) は markdownToTiptap で
    // 一度 Tiptap JSON 化してから setContent に渡す（tiptap-markdown より旧フォーマット解釈が安全）。
    if (isLegacyMarkdown(content)) {
      return markdownToTiptap(content);
    }
    return resolveNoteContent(content).value;
  }, []); // eslint-disable-line react-hooks/exhaustive-deps

  const editor = useEditor({
    extensions: createEditorExtensions(),
    // initialContent は string (markdown) か object (Tiptap doc JSON) か undefined。
    // tiptap-markdown が string を markdown としてパースするため、Tiptap の Content 型に
    // 合致するよう unknown 経由で渡す。
    content: initialContent as Parameters<typeof useEditor>[0] extends infer O
      ? O extends { content?: infer C } ? C : never
      : never,
    onUpdate: ({ editor }) => {
      // tiptap-markdown が editor.storage.markdown.getMarkdown() を提供する。
      // 旧実装は JSON.stringify(editor.getJSON()) を保存していたが、Zenn markdown の
      // 持ち出しやすさ・diff レビュー性を優先して Markdown 文字列保存に切り替える。
      const md = editor.storage.markdown.getMarkdown() as string;
      lastEmittedMarkdown.current = md;
      onChangeRef.current(md);
    },
  });

  useEffect(() => {
    if (!editor) return;

    // エディタの onUpdate から発行されたコンテンツが戻ってきた場合はスキップ
    if (content === lastEmittedMarkdown.current) return;

    if (!content) {
      editor.commands.clearContent();
      return;
    }

    if (isLegacyMarkdown(content)) {
      const json = markdownToTiptap(content);
      const currentJson = JSON.stringify(editor.getJSON());
      if (JSON.stringify(json) === currentJson) return;
      editor.commands.setContent(json);
      return;
    }

    const resolved = resolveNoteContent(content);
    if (resolved.form === 'empty') return;
    editor.commands.setContent(resolved.value as Parameters<typeof editor.commands.setContent>[0]);
  }, [content, editor]);

  return { editor };
}
