import {
  HomeIcon,
  ChatBubbleLeftRightIcon,
  SparklesIcon,
  AcademicCapIcon,
  MagnifyingGlassIcon,
  ChartBarIcon,
  StarIcon,
  DocumentTextIcon,
  UserCircleIcon,
  LightBulbIcon,
  SunIcon,
  DocumentPlusIcon,
} from '@heroicons/react/24/outline';
import type { ComponentType, SVGProps } from 'react';

export type CommandAction =
  | { type: 'navigate'; path: string }
  | { type: 'action'; id: string };

export interface CommandItem {
  id: string;
  label: string;
  description?: string;
  icon: ComponentType<SVGProps<SVGSVGElement>>;
  category: 'ページ移動' | 'アクション';
  action: CommandAction;
  keywords?: string[];
}

export const COMMAND_ITEMS: CommandItem[] = [
  // ページ移動
  {
    id: 'nav-home',
    label: 'ホーム',
    description: 'ホーム画面に移動',
    icon: HomeIcon,
    category: 'ページ移動',
    action: { type: 'navigate', path: '/' },
    keywords: ['home', 'メニュー', 'トップ'],
  },
  {
    id: 'nav-chat',
    label: 'チャット',
    description: 'チャット一覧に移動',
    icon: ChatBubbleLeftRightIcon,
    category: 'ページ移動',
    action: { type: 'navigate', path: '/chat' },
    keywords: ['chat', 'メッセージ', '会話'],
  },
  {
    id: 'nav-ai',
    label: 'AI アシスタント',
    description: 'AIアシスタントに移動',
    icon: SparklesIcon,
    category: 'ページ移動',
    action: { type: 'navigate', path: '/chat/ask-ai' },
    keywords: ['ai', '人工知能', 'アシスタント'],
  },
  {
    id: 'nav-practice',
    label: '練習モード',
    description: '練習モードに移動',
    icon: AcademicCapIcon,
    category: 'ページ移動',
    action: { type: 'navigate', path: '/practice' },
    keywords: ['practice', '練習', 'トレーニング'],
  },
  {
    id: 'nav-user-search',
    label: 'ユーザー検索',
    description: 'ユーザー検索に移動',
    icon: MagnifyingGlassIcon,
    category: 'ページ移動',
    action: { type: 'navigate', path: '/chat/users' },
    keywords: ['search', 'user', '検索', 'ユーザー'],
  },
  {
    id: 'nav-scores',
    label: 'スコア履歴',
    description: 'スコア履歴に移動',
    icon: ChartBarIcon,
    category: 'ページ移動',
    action: { type: 'navigate', path: '/scores' },
    keywords: ['score', 'history', '履歴', '成績'],
  },
  {
    id: 'nav-favorites',
    label: 'お気に入り',
    description: 'お気に入りフレーズに移動',
    icon: StarIcon,
    category: 'ページ移動',
    action: { type: 'navigate', path: '/favorites' },
    keywords: ['favorite', 'star', 'お気に入り', 'フレーズ'],
  },
  {
    id: 'nav-notes',
    label: 'ノート',
    description: 'ノートに移動',
    icon: DocumentTextIcon,
    category: 'ページ移動',
    action: { type: 'navigate', path: '/notes' },
    keywords: ['note', 'メモ', 'ノート'],
  },
  {
    id: 'nav-profile',
    label: 'プロフィール',
    description: 'プロフィールに移動',
    icon: UserCircleIcon,
    category: 'ページ移動',
    action: { type: 'navigate', path: '/profile/me' },
    keywords: ['profile', 'プロフィール', '設定'],
  },
  {
    id: 'nav-personality',
    label: 'パーソナリティ設定',
    description: 'パーソナリティ設定に移動',
    icon: LightBulbIcon,
    category: 'ページ移動',
    action: { type: 'navigate', path: '/profile/personality' },
    keywords: ['personality', 'パーソナリティ', '性格'],
  },
  // アクション
  {
    id: 'action-toggle-theme',
    label: 'テーマ切替',
    description: 'ダークモード / ライトモードを切り替え',
    icon: SunIcon,
    category: 'アクション',
    action: { type: 'action', id: 'toggle-theme' },
    keywords: ['theme', 'dark', 'light', 'テーマ', 'ダーク', 'ライト'],
  },
  {
    id: 'action-new-note',
    label: '新規ノート作成',
    description: '新しいノートを作成',
    icon: DocumentPlusIcon,
    category: 'アクション',
    action: { type: 'action', id: 'new-note' },
    keywords: ['new', 'note', 'create', '新規', '作成', 'ノート'],
  },
];
