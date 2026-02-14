import InputField from '../components/InputField';
import TextareaField from '../components/TextareaField';
import SelectField from '../components/SelectField';
import PrimaryButton from '../components/PrimaryButton';
import Loading from '../components/Loading';
import PersonalityTraitSelector from '../components/PersonalityTraitSelector';
import FormMessage from '../components/FormMessage';
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
      <FormMessage message={message} />

      {/* 説明カード */}
      <div className="bg-surface-2 rounded-lg p-3 mb-4 border border-[var(--color-border-hover)]">
        <p className="text-xs text-primary-300">
          FreStyleはあなたのコミュニケーションスタイルを理解し、チャットと対面の「印象のズレ」を分析します。より正確なフィードバックのために、あなたらしさを教えてください。
        </p>
      </div>

      {/* メインカード */}
      <div className="bg-surface-1 rounded-lg border border-surface-3 overflow-hidden">
        <form onSubmit={handleSave} className="divide-y divide-slate-100">
          {/* 基本情報セクション */}
          <div className="p-5">
            <div className="flex items-center gap-2 mb-3">
              <UserCircleIcon className="w-4 h-4 text-primary-500" />
              <h3 className="text-sm font-bold text-[var(--color-text-primary)]">基本情報</h3>
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
              <TextareaField
                label="自己紹介"
                name="selfIntroduction"
                value={form.selfIntroduction}
                onChange={(e: React.ChangeEvent<HTMLTextAreaElement>) =>
                  setForm((prev) => ({ ...prev, selfIntroduction: e.target.value }))
                }
                placeholder="あなた自身について自由に書いてください..."
                maxLength={300}
              />
            </div>
          </div>

          {/* コミュニケーションスタイル */}
          <div className="p-5">
            <div className="flex items-center gap-2 mb-3">
              <ChatBubbleLeftRightIcon className="w-4 h-4 text-primary-500" />
              <h3 className="text-sm font-bold text-[var(--color-text-primary)]">コミュニケーションスタイル</h3>
            </div>
            <div className="space-y-3">
              <SelectField
                label="あなたのコミュニケーションスタイル"
                name="communicationStyle"
                value={form.communicationStyle}
                onChange={(e: React.ChangeEvent<HTMLSelectElement>) =>
                  setForm((prev) => ({ ...prev, communicationStyle: e.target.value }))
                }
                options={COMMUNICATION_STYLES}
              />

              <PersonalityTraitSelector
                options={PERSONALITY_OPTIONS}
                selected={form.personalityTraits}
                onToggle={togglePersonalityTrait}
                label="性格特性（当てはまるものを選んでください）"
              />
            </div>
          </div>

          {/* AIフィードバック設定 */}
          <div className="p-5">
            <div className="flex items-center gap-2 mb-3">
              <LightBulbIcon className="w-4 h-4 text-primary-500" />
              <h3 className="text-sm font-bold text-[var(--color-text-primary)]">AIフィードバック設定</h3>
            </div>
            <div className="space-y-3">
              <TextareaField
                label="コミュニケーションで改善したい点・目標"
                name="goals"
                value={form.goals}
                onChange={(e: React.ChangeEvent<HTMLTextAreaElement>) =>
                  setForm((prev) => ({ ...prev, goals: e.target.value }))
                }
                placeholder="例：もっと簡潔に伝えられるようになりたい..."
                maxLength={300}
              />

              <TextareaField
                label="苦手なこと・気になっていること"
                name="concerns"
                value={form.concerns}
                onChange={(e: React.ChangeEvent<HTMLTextAreaElement>) =>
                  setForm((prev) => ({ ...prev, concerns: e.target.value }))
                }
                placeholder="例：話が長くなりがち、相手の反応が気になる..."
                maxLength={300}
              />

              <SelectField
                label="フィードバックの受け取り方"
                name="preferredFeedbackStyle"
                value={form.preferredFeedbackStyle}
                onChange={(e: React.ChangeEvent<HTMLSelectElement>) =>
                  setForm((prev) => ({ ...prev, preferredFeedbackStyle: e.target.value }))
                }
                options={FEEDBACK_STYLES}
              />
            </div>
          </div>

          {/* 保存ボタン */}
          <div className="p-5 bg-surface-2">
            <PrimaryButton type="submit">
              {isNewProfile ? 'パーソナリティを保存' : 'パーソナリティを更新'}
            </PrimaryButton>
          </div>
        </form>
      </div>
    </div>
  );
}
