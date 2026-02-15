import type { AxisScore } from '../types';
import Card from './Card';

interface Session {
  scores: AxisScore[];
}

interface Props {
  sessions: Session[];
}

const STYLE_MAP: Record<string, { label: string; description: string; color: string }> = {
  '論理的構成力': {
    label: '論理型コミュニケーター',
    description: '情報を整理し、筋道立てて伝えることが得意です',
    color: 'text-blue-400 bg-blue-900/30 border-blue-800',
  },
  '配慮表現': {
    label: '共感型コミュニケーター',
    description: '相手の気持ちに寄り添い、丁寧に伝えることが得意です',
    color: 'text-pink-600 bg-pink-50 border-pink-200',
  },
  '要約力': {
    label: '簡潔型コミュニケーター',
    description: '要点を短くまとめて伝えることが得意です',
    color: 'text-emerald-400 bg-emerald-900/30 border-emerald-800',
  },
  '提案力': {
    label: '提案型コミュニケーター',
    description: '解決策を積極的に提示することが得意です',
    color: 'text-amber-400 bg-amber-900/30 border-amber-800',
  },
  '質問・傾聴力': {
    label: '傾聴型コミュニケーター',
    description: '相手の話を引き出し、深く理解することが得意です',
    color: 'text-violet-600 bg-violet-50 border-violet-200',
  },
};

function getAverageScores(sessions: Session[]): Map<string, number> {
  const totals = new Map<string, { sum: number; count: number }>();

  for (const session of sessions) {
    const scores = Array.isArray(session.scores) ? session.scores : [];
    for (const score of scores) {
      const current = totals.get(score.axis) || { sum: 0, count: 0 };
      current.sum += score.score;
      current.count += 1;
      totals.set(score.axis, current);
    }
  }

  const averages = new Map<string, number>();
  for (const [axis, { sum, count }] of totals) {
    averages.set(axis, sum / count);
  }
  return averages;
}

export default function CommunicationStyleCard({ sessions }: Props) {
  if (sessions.length === 0) {
    return (
      <Card>
        <p className="text-xs font-medium text-[var(--color-text-muted)] mb-1">あなたのスタイル</p>
        <p className="text-sm text-[var(--color-text-faint)]">まだスタイルが判定できません</p>
        <p className="text-xs text-[var(--color-text-faint)] mt-1">練習セッションを完了するとスタイルが判定されます</p>
      </Card>
    );
  }

  const averages = getAverageScores(sessions);
  let topAxis = '';
  let topScore = -1;

  for (const [axis, avg] of averages) {
    if (avg > topScore) {
      topScore = avg;
      topAxis = axis;
    }
  }

  const style = STYLE_MAP[topAxis];

  if (!style) {
    return null;
  }

  return (
    <Card className={style.color}>
      <p className="text-xs font-medium opacity-75 mb-1">あなたのスタイル</p>
      <p className="text-base font-bold">{style.label}</p>
      <p className="text-xs mt-1 opacity-75">{style.description}</p>
    </Card>
  );
}
