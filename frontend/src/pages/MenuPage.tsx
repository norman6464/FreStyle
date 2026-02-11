import { useNavigate } from 'react-router-dom';
import { useState, useEffect } from 'react';
import {
  CalendarIcon,
  UserGroupIcon,
  SparklesIcon,
  ChatBubbleLeftRightIcon,
  AcademicCapIcon,
} from '@heroicons/react/24/outline';

interface ChatStats {
  chatPartnerCount: number;
}

interface DailyTip {
  emoji: string;
  title: string;
  content: string;
}

export default function MenuPage() {
  const navigate = useNavigate();
  const API_BASE_URL = import.meta.env.VITE_API_BASE_URL;
  const [stats, setStats] = useState<ChatStats | null>(null);
  const [currentTime, setCurrentTime] = useState(new Date());

  const dailyTips: DailyTip[] = [
    {
      emoji: '💬',
      title: '文字だけでは伝わりにくい感情',
      content: 'テキストでは声のトーンや表情が伝わりません。絵文字を活用したり、AIに印象をチェックしてもらいましょう。',
    },
    {
      emoji: '🤔',
      title: '相手の立場で読み返してみる',
      content: '送信前に一度、相手の気持ちになって読み返すと、誤解を防げることがあります。',
    },
    {
      emoji: '✨',
      title: 'ポジティブな言葉を意識する',
      content: '「でも」より「そして」、「〜できない」より「〜してみよう」を使うと印象が変わります。',
    },
    {
      emoji: '🎯',
      title: '具体的に伝える',
      content: '「ちゃんとやって」より「〇〇を△△までにお願い」の方が誤解なく伝わります。',
    },
    {
      emoji: '👂',
      title: '質問で会話を広げる',
      content: '「そうなんだ」で終わらせず、「それでどうなった？」と聞くと会話が深まります。',
    },
    {
      emoji: '🌈',
      title: '感謝を言葉にする',
      content: '「ありがとう」は対面でもチャットでも、最強のコミュニケーションツールです。',
    },
    {
      emoji: '⏰',
      title: '返信のタイミング',
      content: '即レスが良いとは限りません。落ち着いて考えてから返信することも大切です。',
    },
  ];

  const todayTip = dailyTips[new Date().getDate() % dailyTips.length];

  useEffect(() => {
    const timer = setInterval(() => {
      setCurrentTime(new Date());
    }, 1000);
    return () => clearInterval(timer);
  }, []);

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

  const formatDate = (date: Date) =>
    date.toLocaleDateString('ja-JP', {
      year: 'numeric',
      month: 'long',
      day: 'numeric',
      weekday: 'short',
    });

  const formatTime = (date: Date) =>
    date.toLocaleTimeString('ja-JP', {
      hour: '2-digit',
      minute: '2-digit',
    });

  return (
    <div className="p-6 max-w-3xl mx-auto">
      {/* 統計カード */}
      <div className="grid grid-cols-2 gap-4 mb-6">
        <div className="bg-white rounded-lg border border-slate-200 p-4">
          <div className="flex items-center justify-between mb-1">
            <span className="text-xs font-medium text-slate-500">会話した人数</span>
            <UserGroupIcon className="w-4 h-4 text-slate-400" />
          </div>
          <p className="text-2xl font-bold text-slate-800">
            {stats?.chatPartnerCount ?? '—'}
            <span className="text-sm font-normal text-slate-500 ml-1">人</span>
          </p>
        </div>

        <div className="bg-white rounded-lg border border-slate-200 p-4">
          <div className="flex items-center justify-between mb-1">
            <span className="text-xs font-medium text-slate-500">
              {formatDate(currentTime)}
            </span>
            <CalendarIcon className="w-4 h-4 text-slate-400" />
          </div>
          <p className="text-2xl font-bold text-slate-800 font-mono">
            {formatTime(currentTime)}
          </p>
        </div>
      </div>

      {/* 今日のTIPS */}
      <div className="bg-white rounded-lg border border-slate-200 p-4 mb-6">
        <div className="flex items-start gap-3">
          <div className="text-xl">{todayTip.emoji}</div>
          <div className="min-w-0">
            <span className="text-[10px] font-semibold text-primary-600 bg-primary-50 px-1.5 py-0.5 rounded">
              TODAY'S TIP
            </span>
            <h3 className="text-sm font-semibold text-slate-800 mt-1">
              {todayTip.title}
            </h3>
            <p className="text-xs text-slate-500 mt-0.5">
              {todayTip.content}
            </p>
          </div>
        </div>
      </div>

      {/* クイックスタート */}
      <h2 className="text-xs font-semibold text-slate-500 uppercase tracking-wide mb-3">
        クイックスタート
      </h2>
      <div className="grid grid-cols-3 gap-3">
        <button
          onClick={() => navigate('/chat')}
          className="bg-white rounded-lg border border-slate-200 p-4 text-center hover:bg-slate-50 transition-colors"
        >
          <ChatBubbleLeftRightIcon className="w-6 h-6 text-primary-500 mx-auto mb-2" />
          <span className="text-xs font-medium text-slate-700">チャット</span>
        </button>
        <button
          onClick={() => navigate('/chat/ask-ai')}
          className="bg-white rounded-lg border border-slate-200 p-4 text-center hover:bg-slate-50 transition-colors"
        >
          <SparklesIcon className="w-6 h-6 text-primary-500 mx-auto mb-2" />
          <span className="text-xs font-medium text-slate-700">AI分析</span>
        </button>
        <button
          onClick={() => navigate('/practice')}
          className="bg-white rounded-lg border border-slate-200 p-4 text-center hover:bg-slate-50 transition-colors"
        >
          <AcademicCapIcon className="w-6 h-6 text-primary-500 mx-auto mb-2" />
          <span className="text-xs font-medium text-slate-700">練習</span>
        </button>
      </div>
    </div>
  );
}
