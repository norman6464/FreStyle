export interface SlashCommand {
  label: string;
  description: string;
  icon: string;
  action: string;
}

export const SLASH_COMMANDS: SlashCommand[] = [
  { label: 'テキスト', description: '通常の段落', icon: 'Aa', action: 'paragraph' },
  { label: '見出し1', description: '大見出し', icon: 'H1', action: 'heading1' },
  { label: '見出し2', description: '中見出し', icon: 'H2', action: 'heading2' },
  { label: '見出し3', description: '小見出し', icon: 'H3', action: 'heading3' },
  { label: '箇条書き', description: '箇条書きリスト', icon: '•', action: 'bulletList' },
  { label: '番号付きリスト', description: '順序付きリスト', icon: '1.', action: 'orderedList' },
];
