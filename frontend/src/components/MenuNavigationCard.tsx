import { useNavigate } from 'react-router-dom';
import {
  ChatBubbleLeftRightIcon,
  SparklesIcon,
  AcademicCapIcon,
  ChartBarIcon,
  TrophyIcon,
  DocumentTextIcon,
  BellAlertIcon,
} from '@heroicons/react/24/outline';
import Card from './Card';

interface MenuNavigationCardProps {
  totalUnread: number;
  latestScore: number | null;
}

const MENU_ITEMS = [
  {
    icon: ChatBubbleLeftRightIcon,
    label: 'チャット',
    description: 'メンバーとメッセージをやり取り',
    to: '/chat',
    badgeKey: 'unread' as const,
  },
  {
    icon: SparklesIcon,
    label: 'AI アシスタント',
    description: 'AIにコミュニケーションを分析・フィードバック',
    to: '/chat/ask-ai',
    badgeKey: null,
  },
  {
    icon: AcademicCapIcon,
    label: '練習モード',
    description: 'ビジネスシナリオでロールプレイ練習',
    to: '/practice',
    badgeKey: null,
  },
  {
    icon: ChartBarIcon,
    label: 'スコア履歴',
    description: 'フィードバック結果の振り返り',
    to: '/scores',
    badgeKey: 'score' as const,
  },
  {
    icon: TrophyIcon,
    label: 'ランキング',
    description: 'ユーザー間のスコアランキング',
    to: '/ranking',
    badgeKey: null,
  },
  {
    icon: DocumentTextIcon,
    label: '会話テンプレート',
    description: 'シーン別の会話テンプレートで練習',
    to: '/templates',
    badgeKey: null,
  },
  {
    icon: BellAlertIcon,
    label: '練習リマインダー',
    description: '練習の通知時間・曜日を設定',
    to: '/reminder',
    badgeKey: null,
  },
];

export default function MenuNavigationCard({ totalUnread, latestScore }: MenuNavigationCardProps) {
  const navigate = useNavigate();

  const getBadge = (badgeKey: 'unread' | 'score' | null): string | null => {
    if (badgeKey === 'unread' && totalUnread > 0) return `${totalUnread}件の未読`;
    if (badgeKey === 'score' && latestScore !== null) return `最新: ${latestScore}`;
    return null;
  };

  return (
    <div className="space-y-2">
      {MENU_ITEMS.map((item) => {
        const badge = getBadge(item.badgeKey);
        return (
          <Card
            key={item.to}
            role="button"
            tabIndex={0}
            onClick={() => navigate(item.to)}
            onKeyDown={(e: React.KeyboardEvent) => {
              if (e.key === 'Enter' || e.key === ' ') {
                e.preventDefault();
                navigate(item.to);
              }
            }}
            aria-label={item.label}
            className="w-full flex items-center gap-4 text-left hover:bg-surface-2 transition-colors cursor-pointer"
          >
            <item.icon className="w-5 h-5 text-[var(--color-text-muted)] flex-shrink-0" />
            <div className="flex-1 min-w-0">
              <div className="flex items-center gap-2">
                <p className="text-sm font-medium text-[var(--color-text-primary)]">{item.label}</p>
                {badge && (
                  <span className="text-[10px] font-medium bg-primary-500 text-white px-1.5 py-0.5 rounded-full">
                    {badge}
                  </span>
                )}
              </div>
              <p className="text-xs text-[var(--color-text-muted)]">{item.description}</p>
            </div>
          </Card>
        );
      })}
    </div>
  );
}
