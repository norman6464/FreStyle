import { useEffect } from 'react';

interface Scene {
  id: string;
  label: string;
  description: string;
  example: string;
  recommended?: boolean;
  category: string;
}

const SCENES: Scene[] = [
  { id: 'meeting', label: '会議', description: '発言のタイミング・議論の建設性・ファシリテーション', example: '朝会やスプリントレビューで', recommended: true, category: '日常業務' },
  { id: 'code_review', label: 'コードレビュー', description: '指摘の具体性・代替案の提示・相手への配慮', example: 'PRへのコメントで', recommended: true, category: '日常業務' },
  { id: 'daily_report', label: '日報・週報', description: '成果の定量化・課題の明確化・ネクストアクション', example: '日次・週次の報告で', recommended: true, category: '日常業務' },
  { id: 'one_on_one', label: '1on1', description: '心理的安全性・フィードバックの具体性・傾聴の深さ', example: '上司との定期面談で', category: '対面コミュニケーション' },
  { id: 'presentation', label: 'プレゼン', description: 'ストーリー構成・聞き手への配慮・質疑応答力', example: '技術共有会や提案で', category: '対面コミュニケーション' },
  { id: 'negotiation', label: '商談', description: 'ニーズヒアリング・価値提案・クロージング', example: '顧客ミーティングで', category: '対面コミュニケーション' },
  { id: 'email', label: 'メール', description: '件名の明確さ・構成の読みやすさ・アクション明示', example: '社内外への連絡で', category: '文書・報告' },
  { id: 'incident', label: '障害対応', description: '状況報告の正確さ・エスカレーション判断・事後報告の構成', example: '本番障害発生時に', category: '文書・報告' },
];

interface SceneSelectorProps {
  onSelect: (sceneId: string) => void;
  onCancel: () => void;
}

export default function SceneSelector({ onSelect, onCancel }: SceneSelectorProps) {
  const categories = [...new Set(SCENES.map((s) => s.category))];

  useEffect(() => {
    const handleEsc = (e: KeyboardEvent) => {
      if (e.key === 'Escape') onCancel();
    };
    document.addEventListener('keydown', handleEsc);
    return () => document.removeEventListener('keydown', handleEsc);
  }, [onCancel]);

  return (
    <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
      <div
        role="dialog"
        aria-modal="true"
        aria-labelledby="scene-selector-heading"
        className="bg-surface-1 rounded-lg shadow-lg max-w-md w-full mx-4 p-5 max-h-[80vh] overflow-y-auto"
      >
        <h3 id="scene-selector-heading" className="text-sm font-semibold text-[var(--color-text-primary)] mb-1">
          フィードバックのシーンを選択
        </h3>
        <p className="text-xs text-[var(--color-text-muted)] mb-4">
          シーンに応じた追加評価観点でフィードバックします
        </p>

        <div className="space-y-4">
          {categories.map((category) => (
            <div key={category}>
              <p className="text-[10px] font-semibold text-[var(--color-text-faint)] uppercase tracking-wider mb-1 px-1">
                {category}
              </p>
              <div className="space-y-1">
                {SCENES.filter((s) => s.category === category).map((scene) => (
                  <button
                    key={scene.id}
                    onClick={() => onSelect(scene.id)}
                    className="w-full text-left px-3 py-2.5 rounded-md hover:bg-surface-2 transition-colors"
                  >
                    <div className="flex items-center gap-2">
                      <span className="text-sm font-medium text-[var(--color-text-secondary)]">{scene.label}</span>
                      {scene.recommended && (
                        <span className="text-[10px] font-medium bg-surface-2 text-primary-400 px-1.5 py-0.5 rounded">
                          おすすめ
                        </span>
                      )}
                    </div>
                    <p className="text-xs text-[var(--color-text-faint)] mt-0.5">{scene.description}</p>
                    <p className="text-[10px] text-[var(--color-text-subtle)] mt-0.5">{scene.example}</p>
                  </button>
                ))}
              </div>
            </div>
          ))}
        </div>

        <button
          onClick={onCancel}
          className="w-full mt-3 py-2 text-xs text-[var(--color-text-muted)] hover:text-[var(--color-text-secondary)] transition-colors"
        >
          スキップ
        </button>
      </div>
    </div>
  );
}
