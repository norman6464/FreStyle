import type { Editor } from '@tiptap/react';

/**
 * EditorCommandGroup はコマンドの分類。UI（バブルメニュー・将来のスラッシュ/ツールバー）が
 * 「どのコマンドを出すか」を出し分けるために使う。
 * - mark: インラインのマーク（太字・斜体…）。選択テキストに掛ける
 * - turn: 現在ブロックの種類変換（見出し・リスト・引用・コードブロック）
 * - insert: カーソル位置への挿入（水平線…）
 * - history: 取り消し/やり直し
 */
export type EditorCommandGroup = 'mark' | 'turn' | 'insert' | 'history';

/**
 * EditorCommand はエディタの 1 操作を「データ」として表す記述子。
 * UI 側はこの配列を描画するだけにし、書式ロジックをレジストリへ一元化する
 * （＝バブルメニューにもスラッシュメニューにも、記述子を 1 つ足すだけで同じ操作が現れる）。
 */
export interface EditorCommand {
  /**
   * 一意 id（'bold' / 'heading1' / 'image' 等）。React key・テストのアンカーであり、
   * スラッシュコマンドの正規トリガ（英単語）も兼ねる（例: /bold・/heading1・/image）。
   */
  id: string;
  /** 日本語ラベル。アクセシブルネーム / tooltip / スラッシュメニューの表示名に共用する。 */
  label: string;
  /** 分類。 */
  group: EditorCommandGroup;
  /** ボタンに出す短い字面（B / I / H1 等）。アイコン化するまでの簡易表現。 */
  glyph: string;
  /** スラッシュ検索用キーワード（後続の '/' メニューで使う）。英単語のみ（日本語は入れない）。 */
  keywords?: string[];
  /** トグル状態（マーク・turn 系）。非トグル（insert/history）は未定義。 */
  isActive?: (editor: Editor) => boolean;
  /** 実行可否（undo/redo 等）。未定義なら常に実行可。 */
  isEnabled?: (editor: Editor) => boolean;
  /** 実行本体。フォーカス付き chain を内部で張って副作用を完結させる。 */
  run: (editor: Editor) => void;
}

// フォーカスを当ててから操作する（ボタン経由でも選択・キャレットを保って実行する）。
const focused = (editor: Editor) => editor.chain().focus();

/**
 * EDITOR_COMMANDS は RichTextEditor が提供する書式コマンドの正典（single source of truth）。
 * 新しい操作を足すときはこの配列に記述子を 1 つ加える（UI 側の分岐は増やさない）。
 */
export const EDITOR_COMMANDS: EditorCommand[] = [
  // --- インラインのマーク ---
  {
    id: 'bold',
    label: '太字',
    group: 'mark',
    glyph: 'B',
    keywords: ['bold', 'strong'],
    isActive: (editor) => editor.isActive('bold'),
    run: (editor) => focused(editor).toggleBold().run(),
  },
  {
    id: 'italic',
    label: '斜体',
    group: 'mark',
    glyph: 'I',
    keywords: ['italic', 'em'],
    isActive: (editor) => editor.isActive('italic'),
    run: (editor) => focused(editor).toggleItalic().run(),
  },
  {
    id: 'underline',
    label: '下線',
    group: 'mark',
    glyph: 'U',
    keywords: ['underline'],
    isActive: (editor) => editor.isActive('underline'),
    run: (editor) => focused(editor).toggleUnderline().run(),
  },
  {
    id: 'strike',
    label: '打ち消し線',
    group: 'mark',
    glyph: 'S',
    keywords: ['strike', 'strikethrough'],
    isActive: (editor) => editor.isActive('strike'),
    run: (editor) => focused(editor).toggleStrike().run(),
  },
  {
    id: 'code',
    label: 'インラインコード',
    group: 'mark',
    glyph: '</>',
    keywords: ['code', 'inline', 'inlinecode'],
    isActive: (editor) => editor.isActive('code'),
    run: (editor) => focused(editor).toggleCode().run(),
  },
  // --- ブロックの種類変換（turn into） ---
  {
    id: 'heading1',
    label: '見出し1',
    group: 'turn',
    glyph: 'H1',
    keywords: ['h1', 'heading1', 'heading', 'title'],
    isActive: (editor) => editor.isActive('heading', { level: 1 }),
    run: (editor) => focused(editor).toggleHeading({ level: 1 }).run(),
  },
  {
    id: 'heading2',
    label: '見出し2',
    group: 'turn',
    glyph: 'H2',
    keywords: ['h2', 'heading2', 'heading', 'subtitle'],
    isActive: (editor) => editor.isActive('heading', { level: 2 }),
    run: (editor) => focused(editor).toggleHeading({ level: 2 }).run(),
  },
  {
    id: 'heading3',
    label: '見出し3',
    group: 'turn',
    glyph: 'H3',
    keywords: ['h3', 'heading3', 'heading'],
    isActive: (editor) => editor.isActive('heading', { level: 3 }),
    run: (editor) => focused(editor).toggleHeading({ level: 3 }).run(),
  },
  {
    id: 'bulletList',
    label: '箇条書き',
    group: 'turn',
    glyph: '•',
    keywords: ['bullet', 'bulletlist', 'list', 'ul', 'unordered'],
    isActive: (editor) => editor.isActive('bulletList'),
    run: (editor) => focused(editor).toggleBulletList().run(),
  },
  {
    id: 'orderedList',
    label: '番号付きリスト',
    group: 'turn',
    glyph: '1.',
    keywords: ['ordered', 'orderedlist', 'number', 'numbered', 'list', 'ol'],
    isActive: (editor) => editor.isActive('orderedList'),
    run: (editor) => focused(editor).toggleOrderedList().run(),
  },
  {
    id: 'blockquote',
    label: '引用',
    group: 'turn',
    glyph: '“',
    keywords: ['quote', 'blockquote', 'citation'],
    isActive: (editor) => editor.isActive('blockquote'),
    run: (editor) => focused(editor).toggleBlockquote().run(),
  },
  {
    id: 'codeBlock',
    label: 'コードブロック',
    group: 'turn',
    glyph: '{ }',
    keywords: ['codeblock', 'pre', 'fence'],
    isActive: (editor) => editor.isActive('codeBlock'),
    run: (editor) => focused(editor).toggleCodeBlock().run(),
  },
  // --- カーソル位置への挿入 ---
  {
    id: 'taskList',
    label: 'タスクリスト',
    group: 'turn',
    glyph: '☑',
    keywords: ['task', 'tasklist', 'todo', 'check', 'checkbox', 'checklist'],
    isActive: (editor) => editor.isActive('taskList'),
    run: (editor) => focused(editor).toggleTaskList().run(),
  },
  {
    id: 'table',
    label: '表',
    group: 'insert',
    glyph: '⊞',
    keywords: ['table', 'grid'],
    run: (editor) => focused(editor).insertTable({ rows: 3, cols: 3, withHeaderRow: true }).run(),
  },
  {
    id: 'horizontalRule',
    label: '水平線',
    group: 'insert',
    glyph: '—',
    keywords: ['hr', 'rule', 'divider', 'separator'],
    run: (editor) => focused(editor).setHorizontalRule().run(),
  },
  // --- 履歴 ---
  {
    id: 'undo',
    label: '元に戻す',
    group: 'history',
    glyph: '↺',
    keywords: ['undo'],
    isEnabled: (editor) => editor.can().undo(),
    run: (editor) => focused(editor).undo().run(),
  },
  {
    id: 'redo',
    label: 'やり直す',
    group: 'history',
    glyph: '↻',
    keywords: ['redo'],
    isEnabled: (editor) => editor.can().redo(),
    run: (editor) => focused(editor).redo().run(),
  },
];

/**
 * getEditorCommands は指定グループのコマンドだけを、EDITOR_COMMANDS の並び順のまま返す。
 * 引数なしなら全件。UI（バブル=mark+turn 等）の出し分けに使う。
 */
export function getEditorCommands(...groups: EditorCommandGroup[]): EditorCommand[] {
  if (groups.length === 0) return EDITOR_COMMANDS;
  const wanted = new Set(groups);
  return EDITOR_COMMANDS.filter((command) => wanted.has(command.group));
}
