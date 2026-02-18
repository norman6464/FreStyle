import { useRef, useState, useEffect, useCallback } from 'react';
import InputField from '../components/InputField';
import TextareaField from '../components/TextareaField';
import PrimaryButton from '../components/PrimaryButton';
import FormMessage from '../components/FormMessage';
import Avatar from '../components/Avatar';
import Loading from '../components/Loading';
import ProfileStatsSection from '../components/profile/ProfileStatsSection';
import PersonalityTraitSelector from '../components/PersonalityTraitSelector';
import { useProfileEdit } from '../hooks/useProfileEdit';
import { useProfileImageUpload } from '../hooks/useProfileImageUpload';
import { useUserProfilePage, COMMUNICATION_STYLES, PERSONALITY_OPTIONS, FEEDBACK_STYLES } from '../hooks/useUserProfilePage';
import ProfileStatsRepository, { type ProfileStats } from '../repositories/ProfileStatsRepository';
import { CameraIcon } from '@heroicons/react/24/outline';

export default function ProfilePage() {
  const { form, message, setMessage, loading, submitting, updateField, handleUpdate } = useProfileEdit();
  const { upload, uploading } = useProfileImageUpload();
  const fileInputRef = useRef<HTMLInputElement>(null);

  // UserProfile（コミュニケーション設定）
  const {
    form: upForm,
    setForm: setUpForm,
    message: upMessage,
    loading: upLoading,
    togglePersonalityTrait,
    handleSave: handleUserProfileSave,
  } = useUserProfilePage();

  // 学習統計
  const [stats, setStats] = useState<ProfileStats | null>(null);
  const [statsLoading, setStatsLoading] = useState(true);

  const loadStats = useCallback(async () => {
    try {
      const data = await ProfileStatsRepository.fetchStats();
      setStats(data);
    } catch {
      // 統計取得エラーは無視
    } finally {
      setStatsLoading(false);
    }
  }, []);

  useEffect(() => {
    loadStats();
  }, [loadStats]);

  const handleImageSelect = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (!file) return;

    const imageUrl = await upload(file);
    if (imageUrl) {
      updateField('iconUrl', imageUrl);
    } else {
      setMessage({ type: 'error', text: '画像のアップロードに失敗しました。' });
    }

    if (fileInputRef.current) {
      fileInputRef.current.value = '';
    }
  };

  if (loading) {
    return <Loading className="h-full" />;
  }

  return (
    <div className="p-6 max-w-2xl mx-auto space-y-6">
      <FormMessage message={message} />

      {/* セクション1: 基本情報 */}
      <div className="bg-surface-1 rounded-lg border border-surface-3 p-6">
        <div className="flex items-center gap-4 mb-6">
          <div className="relative">
            <Avatar name={form.name || 'U'} src={form.iconUrl || undefined} size="xl" />
            <button
              type="button"
              onClick={() => fileInputRef.current?.click()}
              disabled={uploading}
              className="absolute bottom-0 right-0 bg-primary-500 text-white rounded-full p-1 hover:bg-primary-600 transition-colors disabled:opacity-50"
              aria-label="プロフィール画像を変更"
            >
              <CameraIcon className="w-4 h-4" />
            </button>
            <input
              ref={fileInputRef}
              type="file"
              accept="image/png,image/jpeg,image/gif,image/webp"
              onChange={handleImageSelect}
              className="hidden"
              data-testid="profile-image-input"
            />
          </div>
          <div>
            <h2 className="text-lg font-bold text-[var(--color-text-primary)]">プロフィールを編集</h2>
            <p className="text-sm text-[var(--color-text-muted)]">
              {uploading ? '画像をアップロード中...' : 'あなたの情報を更新してください'}
            </p>
          </div>
        </div>

        <form
          onSubmit={(e) => {
            e.preventDefault();
            handleUpdate();
          }}
          className="space-y-4"
        >
          <InputField
            label="ニックネーム"
            name="name"
            value={form.name}
            onChange={(e: React.ChangeEvent<HTMLInputElement>) => updateField('name', e.target.value)}
          />
          <TextareaField
            label="自己紹介"
            name="bio"
            value={form.bio}
            onChange={(e: React.ChangeEvent<HTMLTextAreaElement>) => updateField('bio', e.target.value)}
            placeholder="あなたについて教えてください..."
            rows={4}
            maxLength={200}
          />
          <InputField
            label="ステータス"
            name="status"
            value={form.status}
            onChange={(e: React.ChangeEvent<HTMLInputElement>) => updateField('status', e.target.value)}
            placeholder="例: 学習中、チャット可能、取り込み中..."
            maxLength={100}
          />
          <PrimaryButton type="submit" disabled={submitting}>
            {submitting ? '更新中...' : '基本情報を保存'}
          </PrimaryButton>
        </form>
      </div>

      {/* セクション2: 学習統計 */}
      <ProfileStatsSection stats={stats} loading={statsLoading} />

      {/* セクション3: コミュニケーション設定 */}
      <div className="bg-surface-1 rounded-lg border border-surface-3 p-6">
        <h3 className="text-sm font-semibold text-[var(--color-text-secondary)] mb-4">コミュニケーション設定</h3>
        <FormMessage message={upMessage} />

        <form onSubmit={handleUserProfileSave} className="space-y-4">
          <div>
            <label className="block text-sm font-medium text-[var(--color-text-secondary)] mb-1">
              コミュニケーションスタイル
            </label>
            <select
              value={upForm.communicationStyle}
              onChange={(e) => setUpForm((prev) => ({ ...prev, communicationStyle: e.target.value }))}
              className="w-full px-3 py-2 rounded-md border border-surface-3 bg-surface-2 text-[var(--color-text-primary)] text-sm focus:outline-none focus:ring-2 focus:ring-primary-500"
            >
              {COMMUNICATION_STYLES.map((s) => (
                <option key={s.value} value={s.value}>{s.label}</option>
              ))}
            </select>
          </div>

          <PersonalityTraitSelector
            label="性格タグ"
            options={[...PERSONALITY_OPTIONS]}
            selected={upForm.personalityTraits}
            onToggle={togglePersonalityTrait}
          />

          <div className="border-t border-surface-3 pt-4 mt-4">
            <h4 className="text-sm font-semibold text-[var(--color-text-secondary)] mb-3">AIフィードバック設定</h4>

            <div className="space-y-4">
              <TextareaField
                label="学習目標"
                name="goals"
                value={upForm.goals}
                onChange={(e: React.ChangeEvent<HTMLTextAreaElement>) => setUpForm((prev) => ({ ...prev, goals: e.target.value }))}
                placeholder="コミュニケーションで達成したいことは？"
                rows={2}
              />
              <TextareaField
                label="悩み・課題"
                name="concerns"
                value={upForm.concerns}
                onChange={(e: React.ChangeEvent<HTMLTextAreaElement>) => setUpForm((prev) => ({ ...prev, concerns: e.target.value }))}
                placeholder="コミュニケーションで困っていることは？"
                rows={2}
              />
              <div>
                <label className="block text-sm font-medium text-[var(--color-text-secondary)] mb-1">
                  フィードバックスタイル
                </label>
                <select
                  value={upForm.preferredFeedbackStyle}
                  onChange={(e) => setUpForm((prev) => ({ ...prev, preferredFeedbackStyle: e.target.value }))}
                  className="w-full px-3 py-2 rounded-md border border-surface-3 bg-surface-2 text-[var(--color-text-primary)] text-sm focus:outline-none focus:ring-2 focus:ring-primary-500"
                >
                  {FEEDBACK_STYLES.map((s) => (
                    <option key={s.value} value={s.value}>{s.label}</option>
                  ))}
                </select>
              </div>
            </div>
          </div>

          {!upLoading && (
            <PrimaryButton type="submit">
              設定を保存
            </PrimaryButton>
          )}
        </form>
      </div>
    </div>
  );
}
