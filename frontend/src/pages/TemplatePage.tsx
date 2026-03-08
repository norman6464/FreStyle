import { useNavigate } from 'react-router-dom';
import { useTemplates } from '../hooks/useTemplates';
import Loading from '../components/Loading';
import { ConversationTemplate } from '../types';

const difficultyLabels: Record<string, { label: string; color: string }> = {
  beginner: { label: '初級', color: 'bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-400' },
  intermediate: { label: '中級', color: 'bg-yellow-100 text-yellow-700 dark:bg-yellow-900/30 dark:text-yellow-400' },
  advanced: { label: '上級', color: 'bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-400' },
};

function TemplateCard({ template, onStart }: { template: ConversationTemplate; onStart: () => void }) {
  const diff = difficultyLabels[template.difficulty] ?? difficultyLabels.intermediate;
  return (
    <div className="bg-white dark:bg-gray-800 rounded-lg border border-gray-200 dark:border-gray-700 p-4 space-y-2" data-testid={`template-${template.id}`}>
      <div className="flex items-center justify-between">
        <h3 className="font-semibold">{template.title}</h3>
        <span className={`text-xs px-2 py-0.5 rounded-full ${diff.color}`}>{diff.label}</span>
      </div>
      <p className="text-sm text-gray-500 dark:text-gray-400">{template.description}</p>
      <div className="bg-gray-50 dark:bg-gray-700/50 rounded p-2">
        <p className="text-xs text-gray-400 mb-1">会話の出だし:</p>
        <p className="text-sm italic">{template.openingMessage}</p>
      </div>
      <button
        onClick={onStart}
        className="w-full mt-2 px-4 py-2 bg-blue-500 text-white rounded-lg text-sm font-medium hover:bg-blue-600 transition-colors"
        data-testid={`start-template-${template.id}`}
      >
        このテンプレートで練習する
      </button>
    </div>
  );
}

export default function TemplatePage() {
  const navigate = useNavigate();
  const { templates, category, categories, changeCategory, loading, error } = useTemplates();

  const handleStart = (template: ConversationTemplate) => {
    navigate('/chat/ask-ai', {
      state: { templateTitle: template.title, templateMessage: template.openingMessage },
    });
  };

  return (
    <div className="max-w-lg mx-auto px-4 py-6 space-y-4">
      <h1 className="text-xl font-bold">会話テンプレート</h1>

      <div className="flex gap-2 overflow-x-auto pb-2">
        {categories.map((cat) => (
          <button
            key={cat.key}
            onClick={() => changeCategory(cat.key)}
            className={`px-3 py-1.5 rounded-full text-sm font-medium whitespace-nowrap transition-colors ${
              category === cat.key
                ? 'bg-blue-500 text-white'
                : 'bg-gray-100 dark:bg-gray-700 text-gray-600 dark:text-gray-300'
            }`}
            data-testid={`category-${cat.key || 'all'}`}
          >
            {cat.label}
          </button>
        ))}
      </div>

      {loading && <Loading message="テンプレートを読み込み中..." />}
      {error && <p className="text-red-500 text-center">{error}</p>}

      {!loading && !error && templates.length === 0 && (
        <p className="text-gray-400 text-center py-8">テンプレートがありません</p>
      )}

      {!loading && !error && (
        <div className="space-y-3">
          {templates.map((t) => (
            <TemplateCard key={t.id} template={t} onStart={() => handleStart(t)} />
          ))}
        </div>
      )}
    </div>
  );
}
