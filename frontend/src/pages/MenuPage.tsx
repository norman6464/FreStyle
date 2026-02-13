import { useNavigate } from 'react-router-dom';
import { useState, useEffect } from 'react';
import {
  UserGroupIcon,
  ChatBubbleLeftRightIcon,
  SparklesIcon,
  AcademicCapIcon,
  ChartBarIcon,
} from '@heroicons/react/24/outline';

interface ChatStats {
  chatPartnerCount: number;
}

export default function MenuPage() {
  const navigate = useNavigate();
  const API_BASE_URL = import.meta.env.VITE_API_BASE_URL;
  const [stats, setStats] = useState<ChatStats | null>(null);

  useEffect(() => {
    const fetchStats = async () => {
      try {
        const res = await fetch(`${API_BASE_URL}/api/chat/stats`, {
          headers: { 'Content-Type': 'application/json' },
          credentials: 'include',
        });

        if (res.status === 401) {
          const refreshRes = await fetch(
            `${API_BASE_URL}/api/auth/cognito/refresh-token`,
            { method: 'POST', credentials: 'include' }
          );
          if (!refreshRes.ok) {
            navigate('/login');
            return;
          }
          await refreshRes.json();
          const retryRes = await fetch(`${API_BASE_URL}/api/chat/stats`, {
            headers: { 'Content-Type': 'application/json' },
            credentials: 'include',
          });
          if (!retryRes.ok) return;
          setStats(await retryRes.json());
          return;
        }

        if (!res.ok) return;
        setStats(await res.json());
      } catch (err) {
        console.error('Error fetching stats:', err);
      }
    };
    fetchStats();
  }, [navigate, API_BASE_URL]);

  const menuItems = [
    {
      icon: ChatBubbleLeftRightIcon,
      label: 'チャット',
      description: 'メンバーとメッセージをやり取り',
      to: '/chat',
    },
    {
      icon: SparklesIcon,
      label: 'AI アシスタント',
      description: 'AIにコミュニケーションを分析・フィードバック',
      to: '/chat/ask-ai',
    },
    {
      icon: AcademicCapIcon,
      label: '練習モード',
      description: 'ビジネスシナリオでロールプレイ練習',
      to: '/practice',
    },
    {
      icon: ChartBarIcon,
      label: 'スコア履歴',
      description: 'フィードバック結果の振り返り',
      to: '/scores',
    },
  ];

  return (
    <div className="p-6 max-w-2xl mx-auto">
      {/* サマリー */}
      <div className="bg-white rounded-lg border border-slate-200 p-4 mb-6">
        <div className="flex items-center gap-3">
          <UserGroupIcon className="w-5 h-5 text-slate-400" />
          <div>
            <p className="text-xs text-slate-500">会話した人数</p>
            <p className="text-lg font-semibold text-slate-800">
              {stats?.chatPartnerCount ?? '—'}
              <span className="text-sm font-normal text-slate-500 ml-1">人</span>
            </p>
          </div>
        </div>
      </div>

      {/* メニュー */}
      <div className="space-y-2">
        {menuItems.map((item) => (
          <button
            key={item.to}
            onClick={() => navigate(item.to)}
            className="w-full flex items-center gap-4 bg-white rounded-lg border border-slate-200 p-4 text-left hover:bg-slate-50 transition-colors"
          >
            <item.icon className="w-5 h-5 text-slate-500 flex-shrink-0" />
            <div className="min-w-0">
              <p className="text-sm font-medium text-slate-800">{item.label}</p>
              <p className="text-xs text-slate-500">{item.description}</p>
            </div>
          </button>
        ))}
      </div>
    </div>
  );
}
