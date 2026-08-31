import {
  HomeIcon,
  SparklesIcon,
  AcademicCapIcon,
  ChartBarIcon,
  StarIcon,
  DocumentTextIcon,
  UserCircleIcon,
} from '@heroicons/react/24/outline';
import type { ComponentType, SVGProps } from 'react';

export type CommandAction = { type: 'navigate'; path: string };

export interface CommandItem {
  id: string;
  label: string;
  description?: string;
  icon: ComponentType<SVGProps<SVGSVGElement>>;
  category: 'ページ移動';
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
    id: 'nav-practice',
    label: '練習モード',
    description: '練習モードに移動',
    icon: AcademicCapIcon,
    category: 'ページ移動',
    action: { type: 'navigate', path: '/practice' },
    keywords: ['practice', '練習', 'トレーニング'],
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
    // 旧「ナレッジ」の語でも引けるようにしておく（統合後も呼び名の記憶は残る）。
    keywords: ['note', 'メモ', 'ノート', 'kb', 'knowledge', 'ナレッジ', 'wiki', '共有'],
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
];
