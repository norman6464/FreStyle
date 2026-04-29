/**
 * Note エディタのコンテンツ形式判定 / 変換ユーティリティ。
 *
 * 設計方針:
 * - 新規ノートは Zenn 互換 Markdown で保存する（tiptap-markdown が双方向変換を担当）
 * - 過去のノートは Tiptap JSON 文字列で保存されているので、判別して Tiptap が読める形に渡す
 * - 過去の旧 Markdown (見出し/リストのみ) は isLegacyMarkdown で判別し markdownToTiptap で
 *   先に Tiptap JSON に変換した上で editor に渡す
 *
 * 形式判定の優先順位（先勝ち）:
 *   1. 空文字 / undefined → 空エディタ
 *   2. JSON.parse 可能で `type: "doc"` を含む → 旧 Tiptap JSON 形式
 *   3. それ以外 → Markdown 文字列としてそのまま editor に渡す
 *
 * editor 側はこの戻り値を `editor.commands.setContent(value)` に渡せば良い。
 * tiptap-markdown は string を受け取れば markdown としてパースし、object を受け取れば
 * Tiptap JSON としてセットする。
 */

export type NoteContentForm = 'empty' | 'json' | 'markdown';

export interface ResolvedNoteContent {
  /** editor.commands.setContent に渡す値（string=markdown / object=Tiptap JSON） */
  value: unknown;
  /** どの形式として解釈したか（debug / migration 統計用） */
  form: NoteContentForm;
}

/**
 * note.content 文字列を editor 投入可能な形に解決する。
 *
 * raw が空ならエディタを空にしたいので value=undefined を返す。
 * JSON-shaped (Tiptap doc node) なら parse 結果を返す。
 * それ以外は markdown として string をそのまま返す。
 */
export function resolveNoteContent(raw: string | null | undefined): ResolvedNoteContent {
  if (!raw) {
    return { value: undefined, form: 'empty' };
  }

  // JSON 形式かどうか判定。Tiptap JSON は必ず先頭が `{` で `"type":"doc"` を含む。
  // 全 raw を毎回 JSON.parse すると markdown 文字列で例外コストが嵩むので prefix 検査でガードする。
  const trimmed = raw.trimStart();
  if (trimmed.startsWith('{')) {
    try {
      const parsed = JSON.parse(raw) as { type?: unknown };
      if (parsed && typeof parsed === 'object' && parsed.type === 'doc') {
        return { value: parsed, form: 'json' };
      }
    } catch {
      // JSON.parse 失敗 → markdown とみなす
    }
  }

  return { value: raw, form: 'markdown' };
}
