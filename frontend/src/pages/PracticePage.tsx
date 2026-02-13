import { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import ScenarioCard from '../components/ScenarioCard';
import type { PracticeScenario } from '../types';

const CATEGORIES = ['すべて', '顧客折衝', 'シニア・上司', 'チーム内'] as const;

export default function PracticePage() {
  const API_BASE_URL = import.meta.env.VITE_API_BASE_URL;
  const navigate = useNavigate();

  const [scenarios, setScenarios] = useState<PracticeScenario[]>([]);
  const [selectedCategory, setSelectedCategory] = useState<string>('すべて');
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const fetchScenarios = async () => {
      try {
        const res = await fetch(`${API_BASE_URL}/api/practice/scenarios`, {
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
          const retryRes = await fetch(`${API_BASE_URL}/api/practice/scenarios`, {
            headers: { 'Content-Type': 'application/json' },
            credentials: 'include',
          });
          if (!retryRes.ok) return;
          const retryData = await retryRes.json();
          setScenarios(retryData);
          setLoading(false);
          return;
        }

        if (!res.ok) {
          console.error('シナリオ取得エラー:', res.status);
          setLoading(false);
          return;
        }

        const data = await res.json();
        setScenarios(data);
        setLoading(false);
      } catch (err) {
        console.error('シナリオ取得失敗:', err);
        setLoading(false);
      }
    };

    fetchScenarios();
  }, [API_BASE_URL, navigate]);

  const handleSelectScenario = async (scenario: PracticeScenario) => {
    try {
      const res = await fetch(`${API_BASE_URL}/api/practice/sessions`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        credentials: 'include',
        body: JSON.stringify({ scenarioId: scenario.id }),
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
        return handleSelectScenario(scenario);
      }

      if (!res.ok) {
        console.error('練習セッション作成エラー:', res.status);
        return;
      }

      const session = await res.json();
      navigate(`/chat/ask-ai/${session.id}`, {
        state: {
          sessionType: 'practice',
          scenarioId: scenario.id,
          scenarioName: scenario.name,
          initialPrompt: '練習開始',
        },
      });
    } catch (err) {
      console.error('練習セッション作成失敗:', err);
    }
  };

  const filteredScenarios = selectedCategory === 'すべて'
    ? scenarios
    : scenarios.filter((s) => s.category === selectedCategory);

  return (
    <div className="p-6 max-w-2xl mx-auto">
      {/* ヘッダー */}
      <div className="mb-6">
        <h1 className="text-lg font-bold text-slate-800 mb-1">ビジネスシナリオ練習</h1>
        <p className="text-sm text-slate-500">
          AIが相手役を演じます。実践的なビジネスシーンでコミュニケーションスキルを磨きましょう。
        </p>
      </div>

      {/* カテゴリフィルター */}
      <div className="flex gap-2 mb-6 overflow-x-auto">
        {CATEGORIES.map((category) => (
          <button
            key={category}
            onClick={() => setSelectedCategory(category)}
            className={`px-3 py-1.5 rounded-full text-xs font-medium whitespace-nowrap transition-colors ${
              selectedCategory === category
                ? 'bg-primary-500 text-white'
                : 'bg-white text-slate-600 border border-slate-200 hover:bg-slate-50'
            }`}
          >
            {category}
          </button>
        ))}
      </div>

      {/* シナリオ一覧 */}
      {loading ? (
        <div className="flex items-center justify-center py-12">
          <div className="w-6 h-6 border-2 border-primary-200 border-t-primary-500 rounded-full animate-spin" />
        </div>
      ) : filteredScenarios.length === 0 ? (
        <div className="text-center py-12 text-sm text-slate-500">
          シナリオがありません
        </div>
      ) : (
        <div className="grid gap-3">
          {filteredScenarios.map((scenario) => (
            <ScenarioCard
              key={scenario.id}
              scenario={scenario}
              onSelect={handleSelectScenario}
            />
          ))}
        </div>
      )}
    </div>
  );
}
