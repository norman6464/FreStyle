export type SlashCommandAction =
  | 'paragraph'
  | 'heading1'
  | 'heading2'
  | 'heading3'
  | 'bulletList'
  | 'orderedList'
  | 'toggleList'
  | 'image'
  | 'taskList'
  | 'codeBlock'
  | 'blockquote'
  | 'horizontalRule'
  | 'table'
  | 'callout'
  | 'youtube'
  | 'emoji';

export interface SlashCommand {
  label: string;
  description: string;
  icon: string;
  action: SlashCommandAction;
  shortcut?: string;
  category: string;
  attrs?: { calloutType: string; emoji: string };
}

export const SLASH_COMMANDS: SlashCommand[] = [
  // 基本
  { label: 'テキスト', description: '通常の段落', icon: 'T', action: 'paragraph', category: '基本' },
  { label: '見出し1', description: '大見出し', icon: 'H1', action: 'heading1', shortcut: '#', category: '基本' },
  { label: '見出し2', description: '中見出し', icon: 'H2', action: 'heading2', shortcut: '##', category: '基本' },
  { label: '見出し3', description: '小見出し', icon: 'H3', action: 'heading3', shortcut: '###', category: '基本' },
  { label: '箇条書きリスト', description: '箇条書き', icon: '≡', action: 'bulletList', shortcut: '・', category: '基本' },
  { label: '番号付きリスト', description: '順序付きリスト', icon: '≡', action: 'orderedList', shortcut: '1.', category: '基本' },
  { label: 'ToDoリスト', description: 'チェックリスト', icon: '☑', action: 'taskList', category: '基本' },
  { label: 'トグルリスト', description: '開閉可能なブロック', icon: '▶', action: 'toggleList', shortcut: '>', category: '基本' },
  // コンテンツ
  { label: 'コールアウト', description: '情報の強調', icon: '💡', action: 'callout', category: 'コンテンツ', attrs: { calloutType: 'info', emoji: '💡' } },
  { label: '引用', description: '引用ブロック', icon: '❝', action: 'blockquote', shortcut: '"', category: 'コンテンツ' },
  { label: 'テーブル', description: '表を挿入', icon: '▦', action: 'table', category: 'コンテンツ' },
  { label: '区切り線', description: '水平線で区切る', icon: '—', action: 'horizontalRule', shortcut: '---', category: 'コンテンツ' },
  // メディア
  { label: '画像', description: '画像をアップロード', icon: '🖼', action: 'image', category: 'メディア' },
  { label: 'YouTube', description: 'YouTube動画を埋め込み', icon: '▶', action: 'youtube', category: 'メディア' },
  { label: 'コード', description: 'コードブロック', icon: '</>', action: 'codeBlock', category: 'メディア' },
  { label: '絵文字', description: '絵文字を挿入', icon: '😀', action: 'emoji', category: 'メディア' },
];
