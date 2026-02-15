import { ChatBubbleLeftRightIcon, EnvelopeIcon, AcademicCapIcon, PresentationChartBarIcon } from '@heroicons/react/24/outline';

interface Template {
  title: string;
  prompt: string;
  category: string;
}

interface ConversationTemplatesProps {
  onSelect: (prompt: string) => void;
}

const CATEGORIES = [
  { name: 'メール添削', icon: EnvelopeIcon },
  { name: '敬語・表現', icon: AcademicCapIcon },
  { name: '報連相', icon: ChatBubbleLeftRightIcon },
  { name: '会議・発表', icon: PresentationChartBarIcon },
];

const TEMPLATES: Template[] = [
  { category: 'メール添削', title: 'メールの添削をお願い', prompt: '以下のメールをビジネスメールとして添削してください。改善点とその理由も教えてください。' },
  { category: 'メール添削', title: '件名の改善提案', prompt: 'メールの件名を効果的にするコツを教えてください。具体的な良い例・悪い例も挙げてください。' },
  { category: '敬語・表現', title: '敬語の使い分け', prompt: '尊敬語・謙譲語・丁寧語の使い分けについて、よくある間違いと正しい使い方を教えてください。' },
  { category: '敬語・表現', title: 'クッション言葉の練習', prompt: 'ビジネスシーンで使えるクッション言葉を状況別に教えてください。依頼・断り・質問それぞれの場面で使える表現をお願いします。' },
  { category: '報連相', title: '報告の構成方法', prompt: '上司への報告を分かりやすく伝えるための構成方法を教えてください。「結論→理由→詳細」の型で練習したいです。' },
  { category: '報連相', title: '悪い報告の伝え方', prompt: '問題やトラブルが発生した場合の報告の仕方を教えてください。相手を不安にさせずに正確に伝えるコツを知りたいです。' },
  { category: '会議・発表', title: '会議での発言の仕方', prompt: '会議で自分の意見を効果的に伝える方法を教えてください。発言のタイミングや話し方のコツもお願いします。' },
  { category: '会議・発表', title: '質疑応答の対処法', prompt: 'プレゼンテーション後の質疑応答で、うまく答えられない質問が来た場合の対処法を教えてください。' },
];

export default function ConversationTemplates({ onSelect }: ConversationTemplatesProps) {
  return (
    <div className="max-w-2xl mx-auto w-full space-y-6">
      <p className="text-sm font-medium text-[var(--color-text-secondary)] text-center">
        テンプレートから始める
      </p>

      <div className="flex flex-wrap justify-center gap-2 mb-2">
        {CATEGORIES.map(({ name, icon: Icon }) => (
          <span key={name} className="inline-flex items-center gap-1 text-xs text-[var(--color-text-tertiary)] bg-surface-2 px-2.5 py-1 rounded-full">
            <Icon className="w-3.5 h-3.5" />
            {name}
          </span>
        ))}
      </div>

      <div className="grid grid-cols-1 sm:grid-cols-2 gap-2">
        {TEMPLATES.map((template) => (
          <button
            key={template.title}
            onClick={() => onSelect(template.prompt)}
            className="text-left p-3 rounded-lg border border-surface-3 hover:border-[var(--color-text-muted)] hover:bg-surface-2 transition-colors group"
          >
            <p className="text-xs font-medium text-[var(--color-text-primary)] group-hover:text-primary-400 transition-colors">
              {template.title}
            </p>
            <p className="text-[10px] text-[var(--color-text-tertiary)] mt-0.5 line-clamp-2">
              {template.prompt}
            </p>
          </button>
        ))}
      </div>
    </div>
  );
}
