import { useNavigate } from 'react-router-dom';

interface WeakAxisAdviceCardProps {
  axis: string;
}

export default function WeakAxisAdviceCard({ axis }: WeakAxisAdviceCardProps) {
  const navigate = useNavigate();

  return (
    <div className="bg-surface-2 rounded-lg border border-[var(--color-border-hover)] p-4">
      <p className="text-xs font-semibold text-primary-300 mb-1">おすすめ練習</p>
      <p className="text-xs text-primary-400 mb-2">
        {`${axis}を伸ばすシナリオで練習しましょう`}
      </p>
      <button
        onClick={() => navigate('/practice')}
        className="text-xs font-medium text-primary-300 bg-surface-1 px-3 py-1.5 rounded-lg border border-[var(--color-border-hover)] hover:bg-surface-3 transition-colors"
      >
        練習一覧を見る
      </button>
    </div>
  );
}
