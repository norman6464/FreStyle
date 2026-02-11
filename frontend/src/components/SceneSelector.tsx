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
];

interface SceneSelectorProps {
  onSelect: (sceneId: string) => void;
  onCancel: () => void;
}

export default function SceneSelector({ onSelect, onCancel }: SceneSelectorProps) {
  return (
    <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
      <div className="bg-white rounded-2xl shadow-xl max-w-md w-full mx-4 p-6">
        <h3 className="text-lg font-bold text-gray-800 mb-2">
          フィードバックのシーンを選択
        </h3>
        <p className="text-sm text-gray-500 mb-4">
          シーンに応じた追加評価観点でフィードバックします
        </p>

        <div className="space-y-2">
          {SCENES.map((scene) => (
            <button
              key={scene.id}
              onClick={() => onSelect(scene.id)}
              className="w-full text-left p-3 rounded-lg border border-gray-200 hover:border-primary-400 hover:bg-primary-50 transition-colors"
            >
              <span className="font-medium text-gray-800">{scene.label}</span>
              <p className="text-xs text-gray-500 mt-0.5">{scene.description}</p>
            </button>
          ))}
        </div>

        <button
          onClick={onCancel}
          className="w-full mt-4 py-2 text-sm text-gray-500 hover:text-gray-700 transition-colors"
        >
          スキップ
        </button>
      </div>
    </div>
  );
}
