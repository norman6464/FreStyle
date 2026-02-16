import { useNavigate } from 'react-router-dom';
import Card from './Card';
import { DIFFICULTY_STYLES } from '../constants/difficultyStyles';

interface Challenge {
  title: string;
  description: string;
  category: string;
  difficulty: '初級' | '中級' | '上級';
}

const CHALLENGES: Challenge[] = [
  { title: '報連相マスター', description: '今日は報告・連絡・相談の3つを意識して練習しましょう。結論から先に伝えることを心がけてください。', category: '報連相', difficulty: '初級' },
  { title: '敬語チャレンジ', description: '尊敬語・謙譲語・丁寧語を正しく使い分ける練習をしましょう。', category: '敬語', difficulty: '中級' },
  { title: '要約力トレーニング', description: '長い説明を3文以内にまとめる練習をしましょう。ポイントを絞って伝える力を鍛えます。', category: '要約', difficulty: '中級' },
  { title: '質問力アップ', description: '5W1Hを意識した質問を5つ以上考えてみましょう。相手から情報を引き出す力を磨きます。', category: '質問', difficulty: '初級' },
  { title: '提案力チャレンジ', description: '課題に対して3つ以上の解決策を提案する練習をしましょう。それぞれのメリット・デメリットも添えてください。', category: '提案', difficulty: '上級' },
  { title: 'クッション言葉', description: '「恐れ入りますが」「お手数ですが」など、クッション言葉を5種類以上使って会話しましょう。', category: '配慮', difficulty: '初級' },
  { title: '技術説明の平易化', description: '技術用語を非エンジニアにも分かるように説明する練習をしましょう。', category: '要約', difficulty: '上級' },
  { title: 'エスカレーション判断', description: '問題発生時に上長へのエスカレーションが必要かどうか判断する練習をしましょう。', category: '報連相', difficulty: '中級' },
  { title: 'メール文面作成', description: '社外向けのフォーマルなメールを作成する練習をしましょう。件名・挨拶・本文・結びの構成を意識してください。', category: '配慮', difficulty: '中級' },
  { title: '会議ファシリテーション', description: '会議で議題を整理し、参加者から意見を引き出す練習をしましょう。', category: '質問', difficulty: '上級' },
];

export default function DailyChallengeCard() {
  const navigate = useNavigate();
  const today = new Date();
  const dayIndex = Math.floor(today.getTime() / (1000 * 60 * 60 * 24)) % CHALLENGES.length;
  const challenge = CHALLENGES[dayIndex];

  return (
    <Card>
      <div className="flex items-center justify-between mb-2">
        <p className="text-xs font-medium text-[var(--color-text-secondary)]">本日のチャレンジ</p>
        <span
          data-testid="challenge-difficulty"
          className={`text-[10px] font-medium px-2 py-0.5 rounded ${DIFFICULTY_STYLES[challenge.difficulty]}`}
        >
          {challenge.difficulty}
        </span>
      </div>

      <p className="text-sm font-semibold text-[var(--color-text-primary)] mb-1">{challenge.title}</p>
      <p data-testid="challenge-description" className="text-xs text-[var(--color-text-muted)] mb-2">
        {challenge.description}
      </p>

      <div className="flex items-center justify-between">
        <span data-testid="challenge-category" className="text-[10px] text-[var(--color-text-faint)]">
          {challenge.category}
        </span>
        <button
          onClick={() => navigate('/practice')}
          className="text-xs font-medium text-primary-400 hover:text-primary-300 transition-colors"
        >
          チャレンジする
        </button>
      </div>
    </Card>
  );
}
