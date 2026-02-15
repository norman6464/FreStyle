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
  | 'youtube';

export interface SlashCommand {
  label: string;
  description: string;
  icon: string;
  action: SlashCommandAction;
  attrs?: { calloutType: string; emoji: string };
}

export const SLASH_COMMANDS: SlashCommand[] = [
  { label: 'テキスト', description: '通常の段落', icon: 'Aa', action: 'paragraph' },
  { label: '見出し1', description: '大見出し', icon: 'H1', action: 'heading1' },
  { label: '見出し2', description: '中見出し', icon: 'H2', action: 'heading2' },
  { label: '見出し3', description: '小見出し', icon: 'H3', action: 'heading3' },
  { label: '箇条書き', description: '箇条書きリスト', icon: '•', action: 'bulletList' },
  { label: '番号付きリスト', description: '順序付きリスト', icon: '1.', action: 'orderedList' },
  { label: 'トグル', description: '開閉可能なブロック', icon: '▶', action: 'toggleList' },
  { label: '画像', description: '画像をアップロード', icon: '🖼', action: 'image' },
  { label: 'タスクリスト', description: 'チェックリスト', icon: '☑', action: 'taskList' },
  { label: 'コード', description: 'コードブロック', icon: '</>', action: 'codeBlock' },
  { label: '引用', description: '引用ブロック', icon: '❝', action: 'blockquote' },
  { label: '区切り線', description: '水平線で区切る', icon: '—', action: 'horizontalRule' },
  { label: 'テーブル', description: '表を挿入', icon: '▦', action: 'table' },
  { label: 'コールアウト（情報）', description: '情報の強調', icon: '💡', action: 'callout', attrs: { calloutType: 'info', emoji: '💡' } },
  { label: 'コールアウト（警告）', description: '注意喚起', icon: '⚠️', action: 'callout', attrs: { calloutType: 'warning', emoji: '⚠️' } },
  { label: 'コールアウト（成功）', description: '成功・完了', icon: '✅', action: 'callout', attrs: { calloutType: 'success', emoji: '✅' } },
  { label: 'コールアウト（エラー）', description: 'エラー・危険', icon: '❌', action: 'callout', attrs: { calloutType: 'error', emoji: '❌' } },
  { label: 'YouTube', description: 'YouTube動画を埋め込み', icon: '▶', action: 'youtube' },
];
