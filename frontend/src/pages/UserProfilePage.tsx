import InputField from '../components/InputField';
import PrimaryButton from '../components/PrimaryButton';
import Loading from '../components/Loading';
import {
  ChatBubbleLeftRightIcon,
  LightBulbIcon,
  UserCircleIcon,
} from '@heroicons/react/24/solid';
import {
  useUserProfilePage,
  COMMUNICATION_STYLES,
  PERSONALITY_OPTIONS,
  FEEDBACK_STYLES,
} from '../hooks/useUserProfilePage';

export default function UserProfilePage() {
  const {
    form,
    setForm,
    message,
    isNewProfile,
    loading,
    togglePersonalityTrait,
    handleSave,
  } = useUserProfilePage();

  if (loading) {
    return <Loading fullscreen message="プロファイルを読み込み中..." />;
  }

  return (
    <div className="p-6 max-w-2xl mx-auto">
      {/* メッセージ */}
      {message && (
        <div
          className={`mb-4 p-3 rounded-lg text-sm font-medium ${
            message.type === 'error'
              ? 'bg-rose-50 text-rose-700 border border-rose-200'
              : 'bg-emerald-50 text-emerald-700 border border-emerald-200'
          }`}
        >
          {message.text}
        </div>
      )}

      {/* 説明カード */}
      <div className="bg-primary-50 rounded-lg p-3 mb-4 border border-primary-200">
        <p className="text-xs text-primary-700">
          FreStyleはあなたのコミュニケーションスタイルを理解し、チャットと対面の「印象のズレ」を分析します。より正確なフィードバックのために、あなたらしさを教えてください。
        </p>
      </div>

      {/* メインカード */}
      <div className="bg-white rounded-lg border border-slate-200 overflow-hidden">
        <form onSubmit={handleSave} className="divide-y divide-slate-100">
          {/* 基本情報セクション */}
          <div className="p-5">
            <div className="flex items-center gap-2 mb-3">
              <UserCircleIcon className="w-4 h-4 text-primary-500" />
              <h3 className="text-sm font-bold text-slate-800">基本情報</h3>
            </div>
            <div className="space-y-3">
              <InputField
                label="呼ばれたい名前"
                name="displayName"
                value={form.displayName}
                onChange={(e: React.ChangeEvent<HTMLInputElement>) =>
                  setForm((prev) => ({ ...prev, displayName: e.target.value }))
                }
                placeholder="例：タロウ、たろちゃん"
              />
              <div>
                <label className="block text-sm font-medium text-slate-700 mb-1">
                  自己紹介
                </label>
                <textarea
                  name="selfIntroduction"
                  value={form.selfIntroduction}
                  onChange={(e: React.ChangeEvent<HTMLTextAreaElement>) =>
                    setForm((prev) => ({ ...prev, selfIntroduction: e.target.value }))
                  }
                  placeholder="あなた自身について自由に書いてください..."
                  rows={3}
                  className="w-full border border-slate-300 rounded-lg px-3 py-2 text-sm focus:border-primary-500 focus:ring-1 focus:ring-primary-500 transition-colors resize-none"
                />
              </div>
            </div>
          </div>

          {/* コミュニケーションスタイル */}
          <div className="p-5">
            <div className="flex items-center gap-2 mb-3">
              <ChatBubbleLeftRightIcon className="w-4 h-4 text-primary-500" />
              <h3 className="text-sm font-bold text-slate-800">コミュニケーションスタイル</h3>
            </div>
            <div className="space-y-3">
              <div>
                <label className="block text-sm font-medium text-slate-700 mb-1">
                  あなたのコミュニケーションスタイル
                </label>
                <select
                  value={form.communicationStyle}
                  onChange={(e: React.ChangeEvent<HTMLSelectElement>) =>
                    setForm((prev) => ({ ...prev, communicationStyle: e.target.value }))
                  }
                  className="w-full border border-slate-300 rounded-lg px-3 py-2 text-sm focus:border-primary-500 focus:ring-1 focus:ring-primary-500 transition-colors"
                >
                  {COMMUNICATION_STYLES.map((style) => (
                    <option key={style.value} value={style.value}>{style.label}</option>
                  ))}
                </select>
              </div>

              <div>
                <label className="block text-sm font-medium text-slate-700 mb-2">
                  性格特性（当てはまるものを選んでください）
                </label>
                <div className="flex flex-wrap gap-1.5">
                  {PERSONALITY_OPTIONS.map((trait) => (
                    <button
                      key={trait}
                      type="button"
                      onClick={() => togglePersonalityTrait(trait)}
                      className={`px-3 py-1.5 rounded-full text-xs font-medium transition-colors ${
                        form.personalityTraits.includes(trait)
                          ? 'bg-primary-500 text-white'
                          : 'bg-slate-100 text-slate-700 hover:bg-slate-200'
                      }`}
                    >
                      {trait}
                    </button>
                  ))}
                </div>
              </div>
            </div>
          </div>

          {/* AIフィードバック設定 */}
          <div className="p-5">
            <div className="flex items-center gap-2 mb-3">
              <LightBulbIcon className="w-4 h-4 text-primary-500" />
              <h3 className="text-sm font-bold text-slate-800">AIフィードバック設定</h3>
            </div>
            <div className="space-y-3">
              <div>
                <label className="block text-sm font-medium text-slate-700 mb-1">
                  コミュニケーションで改善したい点・目標
                </label>
                <textarea
                  name="goals"
                  value={form.goals}
                  onChange={(e: React.ChangeEvent<HTMLTextAreaElement>) =>
                    setForm((prev) => ({ ...prev, goals: e.target.value }))
                  }
                  placeholder="例：もっと簡潔に伝えられるようになりたい..."
                  rows={3}
                  className="w-full border border-slate-300 rounded-lg px-3 py-2 text-sm focus:border-primary-500 focus:ring-1 focus:ring-primary-500 transition-colors resize-none"
                />
              </div>

              <div>
                <label className="block text-sm font-medium text-slate-700 mb-1">
                  苦手なこと・気になっていること
                </label>
                <textarea
                  name="concerns"
                  value={form.concerns}
                  onChange={(e: React.ChangeEvent<HTMLTextAreaElement>) =>
                    setForm((prev) => ({ ...prev, concerns: e.target.value }))
                  }
                  placeholder="例：話が長くなりがち、相手の反応が気になる..."
                  rows={3}
                  className="w-full border border-slate-300 rounded-lg px-3 py-2 text-sm focus:border-primary-500 focus:ring-1 focus:ring-primary-500 transition-colors resize-none"
                />
              </div>

              <div>
                <label className="block text-sm font-medium text-slate-700 mb-1">
                  フィードバックの受け取り方
                </label>
                <select
                  value={form.preferredFeedbackStyle}
                  onChange={(e: React.ChangeEvent<HTMLSelectElement>) =>
                    setForm((prev) => ({ ...prev, preferredFeedbackStyle: e.target.value }))
                  }
                  className="w-full border border-slate-300 rounded-lg px-3 py-2 text-sm focus:border-primary-500 focus:ring-1 focus:ring-primary-500 transition-colors"
                >
                  {FEEDBACK_STYLES.map((style) => (
                    <option key={style.value} value={style.value}>{style.label}</option>
                  ))}
                </select>
              </div>
            </div>
          </div>

          {/* 保存ボタン */}
          <div className="p-5 bg-primary-50">
            <PrimaryButton type="submit">
              {isNewProfile ? 'パーソナリティを保存' : 'パーソナリティを更新'}
            </PrimaryButton>
          </div>
        </form>
      </div>
    </div>
  );
}
