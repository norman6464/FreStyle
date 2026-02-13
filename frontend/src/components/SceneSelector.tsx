interface Scene {
  id: string;
  label: string;
  description: string;
}

const SCENES: Scene[] = [
  { id: 'meeting', label: '会議', description: '発言のタイミング・議論の建設性・ファシリテーション' },
  { id: 'one_on_one', label: '1on1', description: '心理的安全性・フィードバックの具体性・傾聴の深さ' },
  { id: 'email', label: 'メール', description: '件名の明確さ・構成の読みやすさ・アクション明示' },
  { id: 'presentation', label: 'プレゼン', description: 'ストーリー構成・聞き手への配慮・質疑応答力' },
  { id: 'negotiation', label: '商談', description: 'ニーズヒアリング・価値提案・クロージング' },
  { id: 'code_review', label: 'コードレビュー', description: '指摘の具体性・代替案の提示・相手への配慮' },
  { id: 'incident', label: '障害対応', description: '状況報告の正確さ・エスカレーション判断・事後報告の構成' },
  { id: 'daily_report', label: '日報・週報', description: '成果の定量化・課題の明確化・ネクストアクション' },
];

interface SceneSelectorProps {
  onSelect: (sceneId: string) => void;
  onCancel: () => void;
}

export default function SceneSelector({ onSelect, onCancel }: SceneSelectorProps) {
  return (
    <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
      <div className="bg-white rounded-lg shadow-lg max-w-md w-full mx-4 p-5">
        <h3 className="text-sm font-semibold text-slate-800 mb-1">
          フィードバックのシーンを選択
        </h3>
        <p className="text-xs text-slate-500 mb-4">
          シーンに応じた追加評価観点でフィードバックします
        </p>

        <div className="space-y-1">
          {SCENES.map((scene) => (
            <button
              key={scene.id}
              onClick={() => onSelect(scene.id)}
              className="w-full text-left px-3 py-2.5 rounded-md hover:bg-primary-50 transition-colors"
            >
              <span className="text-sm font-medium text-slate-700">{scene.label}</span>
              <p className="text-xs text-slate-400 mt-0.5">{scene.description}</p>
            </button>
          ))}
        </div>

        <button
          onClick={onCancel}
          className="w-full mt-3 py-2 text-xs text-slate-500 hover:text-slate-700 transition-colors"
        >
          スキップ
        </button>
      </div>
    </div>
  );
}
