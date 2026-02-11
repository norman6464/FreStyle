import { useNavigate } from 'react-router-dom';
import { useSelector } from 'react-redux';
import { useState, useEffect } from 'react';
import HamburgerMenu from '../components/HamburgerMenu';
import {
  CalendarIcon,
  UserGroupIcon,
  SparklesIcon,
  ChatBubbleLeftRightIcon,
  LightBulbIcon,
} from '@heroicons/react/24/solid';
import type { RootState } from '../store';

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
  const message = useSelector((state: RootState) => state.flash?.message);
  const API_BASE_URL = import.meta.env.VITE_API_BASE_URL;
  const [stats, setStats] = useState<ChatStats | null>(null);
  const [currentTime, setCurrentTime] = useState(new Date());

  // 日替わりTIPS
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

  // 今日のTIPSを日付ベースで選択
  const todayTip = dailyTips[new Date().getDate() % dailyTips.length];

  // 時刻の自動更新
  useEffect(() => {
    const timer = setInterval(() => {
      setCurrentTime(new Date());
    }, 1000);
    return () => clearInterval(timer);
  }, []);

  // 統計情報の取得
  useEffect(() => {
    const fetchStats = async () => {
      try {
        const res = await fetch(`${API_BASE_URL}/api/chat/stats`, {
          headers: {
            'Content-Type': 'application/json',
          },
          credentials: 'include',
        });

        // トークン期限切れの場合
        if (res.status === 401) {
          const refreshRes = await fetch(
            `${API_BASE_URL}/api/auth/cognito/refresh-token`,
            {
              method: 'POST',
              credentials: 'include',
            }
          );

          if (!refreshRes.ok) {
            navigate('/login');
            return;
          }

          // アクセストークンはhttpOnly cookieで管理しているのでここでは取得しない
          await refreshRes.json();

          // リトライ
          const retryRes = await fetch(`${API_BASE_URL}/api/chat/stats`, {
            headers: {
              'Content-Type': 'application/json'
            },
            credentials: 'include',
          });

          if (!retryRes.ok) return;
          const retryData = await retryRes.json();
          setStats(retryData);
          return;
        }

        if (!res.ok) return;
        const data = await res.json();
        console.log('Fetched stats:', data);
        setStats(data);
      } catch (err) {
        console.error('Error fetching stats:', err);
      }
    };

    fetchStats();
  }, [navigate, API_BASE_URL]);

  const formatDate = (date: Date) => {
    return date.toLocaleDateString('ja-JP', {
      year: 'numeric',
      month: 'long',
      day: 'numeric',
      weekday: 'long',
    });
  };

  const formatTime = (date: Date) => {
    return date.toLocaleTimeString('ja-JP', {
      hour: '2-digit',
      minute: '2-digit',
      second: '2-digit',
    });
  };

  return (
    <>
      {message && (
        <div className="fixed top-20 left-0 right-0 flex justify-center z-50">
          <div className="bg-green-50 border-l-4 border-green-500 text-green-700 px-6 py-4 rounded-lg shadow-sm font-medium">
            ✓ {message}
          </div>
        </div>
      )}
      <HamburgerMenu title="ホーム" />
      <div className="min-h-screen bg-gray-50 px-4 pt-20 pb-8">
        <div className="w-full max-w-2xl mx-auto">
          {/* ユーザー情報セクション */}
          <div className="grid grid-cols-2 gap-4 mb-8">
            {/* 会話したユーザー数 */}
            <div className="bg-white rounded-xl shadow-sm p-5 border border-gray-200">
              <div className="flex items-center justify-between mb-2">
                <h3 className="text-xs font-medium text-gray-500 uppercase">
                  会話した人数
                </h3>
                <UserGroupIcon className="w-5 h-5 text-primary-400" />
              </div>
              <p className="text-2xl font-bold text-primary-600">
                {stats?.chatPartnerCount ?? '—'}
                <span className="text-sm font-normal text-gray-500 ml-1">人</span>
              </p>
            </div>

            {/* 日時 */}
            <div className="bg-white rounded-xl shadow-sm p-5 border border-gray-200">
              <div className="flex items-center justify-between mb-2">
                <h3 className="text-xs font-medium text-gray-500 uppercase">
                  {formatDate(currentTime).split('日')[0]}日
                </h3>
                <CalendarIcon className="w-5 h-5 text-primary-400" />
              </div>
              <p className="text-2xl font-bold text-primary-600 font-mono">
                {formatTime(currentTime).slice(0, 5)}
              </p>
            </div>
          </div>

          {/* クイックアクション見出し */}
          <div className="flex items-center gap-2 mb-4">
            <ChatBubbleLeftRightIcon className="w-5 h-5 text-gray-500" />
            <h2 className="text-lg font-bold text-gray-700">クイックアクション</h2>
          </div>

          {/* ヒーローセクション */}
          <div className="mb-8 bg-primary-500 rounded-2xl p-6 text-center">
            <h1 className="text-2xl font-bold text-white mb-2">
              FreStyle
            </h1>
            <p className="text-white/90 text-sm">
              チャット × 対面 ー「印象のズレ」をAIが解消します
            </p>
          </div>

          {/* 3ステップガイド */}
          <div className="mb-8 bg-white rounded-xl shadow-sm p-6 border border-gray-200">
            <div className="flex items-center gap-2 mb-4">
              <LightBulbIcon className="w-5 h-5 text-primary-500" />
              <h2 className="text-lg font-bold text-gray-800">FreStyleの使い方</h2>
            </div>
            <div className="flex flex-col md:flex-row gap-4">
              {/* Step 1 */}
              <div className="flex-1 bg-gray-50 rounded-lg p-4 text-center">
                <div className="w-10 h-10 bg-primary-500 text-white rounded-full flex items-center justify-center mx-auto mb-2 font-bold text-lg">
                  1
                </div>
                <p className="text-sm font-medium text-gray-800 mb-1">
                  チャットする
                </p>
                <p className="text-xs text-gray-500">
                  友達と普段通りに会話
                </p>
              </div>
              {/* 矢印 */}
              <div className="hidden md:flex items-center text-gray-300">
                <svg className="w-6 h-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 5l7 7-7 7" />
                </svg>
              </div>
              {/* Step 2 */}
              <div className="flex-1 bg-gray-50 rounded-lg p-4 text-center">
                <div className="w-10 h-10 bg-primary-500 text-white rounded-full flex items-center justify-center mx-auto mb-2 font-bold text-lg">
                  2
                </div>
                <p className="text-sm font-medium text-gray-800 mb-1">
                  AIに分析依頼
                </p>
                <p className="text-xs text-gray-500">
                  会話を選んでAIへ
                </p>
              </div>
              {/* 矢印 */}
              <div className="hidden md:flex items-center text-gray-300">
                <svg className="w-6 h-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 5l7 7-7 7" />
                </svg>
              </div>
              {/* Step 3 */}
              <div className="flex-1 bg-gray-50 rounded-lg p-4 text-center">
                <div className="w-10 h-10 bg-primary-500 text-white rounded-full flex items-center justify-center mx-auto mb-2 font-bold text-lg">
                  3
                </div>
                <p className="text-sm font-medium text-gray-800 mb-1">
                  改善のヒント
                </p>
                <p className="text-xs text-gray-500">
                  印象のギャップを知る
                </p>
              </div>
            </div>
          </div>

          {/* 今日のTIPS */}
          <div className="mb-8 bg-white rounded-xl shadow-sm p-6 border border-gray-200">
            <div className="flex items-start gap-3">
              <div className="text-3xl">{todayTip.emoji}</div>
              <div>
                <span className="text-xs font-medium text-primary-600 bg-primary-50 px-2 py-0.5 rounded-full">
                  TODAY'S TIP
                </span>
                <h3 className="text-lg font-bold text-gray-800 mt-2 mb-1">
                  {todayTip.title}
                </h3>
                <p className="text-sm text-gray-600">
                  {todayTip.content}
                </p>
              </div>
            </div>
          </div>

          {/* メニュー項目 */}
          <div className="space-y-4">
            {/* AIに聞いてみる（推奨・最初に配置） */}
            <div
              onClick={() => navigate('/chat/ask-ai')}
              className="bg-primary-500 rounded-xl p-6 cursor-pointer hover:bg-primary-600 transition-colors duration-150"
            >
              <div className="flex items-start">
                <div className="bg-white/20 rounded-lg p-3 mr-4">
                  <SparklesIcon className="w-6 h-6 text-white" />
                </div>
                <div className="flex-1">
                  <h2 className="text-xl font-bold text-white">
                    AIに分析してもらう
                  </h2>
                  <p className="text-white/80 text-sm mt-1">
                    チャットの印象を分析 → 対面との差を発見
                  </p>
                </div>
                <svg
                  className="w-5 h-5 text-white/70 ml-auto mt-2"
                  fill="none"
                  stroke="currentColor"
                  viewBox="0 0 24 24"
                >
                  <path
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    strokeWidth={2}
                    d="M9 5l7 7-7 7"
                  />
                </svg>
              </div>
            </div>

            {/* チャット一覧 */}
            <div
              onClick={() => navigate('/chat')}
              className="bg-white rounded-xl p-6 cursor-pointer border border-gray-200 hover:bg-gray-50 transition-colors duration-150"
            >
              <div className="flex items-start">
                <div className="bg-primary-100 rounded-lg p-3 mr-4">
                  <ChatBubbleLeftRightIcon className="w-6 h-6 text-primary-500" />
                </div>
                <div>
                  <h2 className="text-xl font-bold text-gray-800">
                    チャット一覧
                  </h2>
                  <p className="text-gray-500 text-sm mt-1">
                    友達とのチャットを見る・会話を続ける
                  </p>
                </div>
                <svg
                  className="w-5 h-5 text-gray-300 ml-auto"
                  fill="none"
                  stroke="currentColor"
                  viewBox="0 0 24 24"
                >
                  <path
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    strokeWidth={2}
                    d="M9 5l7 7-7 7"
                  />
                </svg>
              </div>
            </div>

            {/* ユーザー検索 */}
            <div
              onClick={() => navigate('/chat/users')}
              className="bg-white rounded-xl p-6 cursor-pointer border border-gray-200 hover:bg-gray-50 transition-colors duration-150"
            >
              <div className="flex items-start">
                <div className="bg-primary-100 rounded-lg p-3 mr-4">
                  <svg
                    className="w-6 h-6 text-primary-500"
                    fill="none"
                    stroke="currentColor"
                    viewBox="0 0 24 24"
                  >
                    <path
                      strokeLinecap="round"
                      strokeLinejoin="round"
                      strokeWidth={2}
                      d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z"
                    />
                  </svg>
                </div>
                <div>
                  <h2 className="text-xl font-bold text-gray-800">
                    ユーザー検索
                  </h2>
                  <p className="text-gray-500 text-sm mt-1">
                    メールアドレスで友達を追加
                  </p>
                </div>
                <svg
                  className="w-5 h-5 text-gray-300 ml-auto"
                  fill="none"
                  stroke="currentColor"
                  viewBox="0 0 24 24"
                >
                  <path
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    strokeWidth={2}
                    d="M9 5l7 7-7 7"
                  />
                </svg>
              </div>
            </div>

            {/* プロフィール編集 */}
            <div
              onClick={() => navigate('/profile/me')}
              className="bg-white rounded-xl p-6 cursor-pointer border border-gray-200 hover:bg-gray-50 transition-colors duration-150"
            >
              <div className="flex items-start">
                <div className="bg-primary-100 rounded-lg p-3 mr-4">
                  <svg
                    className="w-6 h-6 text-primary-500"
                    fill="none"
                    stroke="currentColor"
                    viewBox="0 0 24 24"
                  >
                    <path
                      strokeLinecap="round"
                      strokeLinejoin="round"
                      strokeWidth={2}
                      d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z"
                    />
                  </svg>
                </div>
                <div>
                  <h2 className="text-xl font-bold text-gray-800">
                    プロフィールを編集
                  </h2>
                  <p className="text-gray-500 text-sm mt-1">
                    ユーザー情報の編集
                  </p>
                </div>
                <svg
                  className="w-5 h-5 text-gray-300 ml-auto"
                  fill="none"
                  stroke="currentColor"
                  viewBox="0 0 24 24"
                >
                  <path
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    strokeWidth={2}
                    d="M9 5l7 7-7 7"
                  />
                </svg>
              </div>
            </div>
          </div>

          {/* フッター */}
          <div className="mt-8 text-center">
            <p className="text-xs text-gray-400">
              💡 チャットで感じた「なんか違う...」をAIと一緒に解決しよう
            </p>
          </div>
        </div>
      </div>
    </>
  );
}
